package local

import (
	"context"
	"database/sql"
	errs "errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/scanner"
	"github.com/tedyst/licenta/scanner/postgres"
)

func getPostgresConnectString(db *models.PostgresDatabases) string {
	return fmt.Sprintf("host=%s port=%d database=%s user=%s password=%s", db.Host, db.Port, db.DatabaseName, db.Username, db.Password)
}

type scannerRunner struct {
	queries db.TransactionQuerier
}

func NewScannerRunner(queries db.TransactionQuerier) *scannerRunner {
	return &scannerRunner{
		queries: queries,
	}
}

func (runner *scannerRunner) ScanPostgresDB(ctx context.Context, scan *models.PostgresScan) error {
	if err := runner.queries.UpdatePostgresScanStatus(ctx, queries.UpdatePostgresScanStatusParams{
		ID:     scan.ID,
		Status: models.SCAN_RUNNING,
	}); err != nil {
		return errors.Wrap(err, "could not update scan status")
	}

	db, err := runner.queries.GetPostgresDatabase(ctx, scan.PostgresDatabaseID)
	if err != nil {
		return errors.Wrap(err, "could not get database")
	}

	notifyError := func(err error) error {
		if err2 := runner.queries.UpdatePostgresScanStatus(ctx, queries.UpdatePostgresScanStatusParams{
			ID:     scan.ID,
			Status: models.SCAN_FINISHED,
			Error:  sql.NullString{String: err.Error(), Valid: true},
			EndedAt: pgtype.Timestamptz{
				Time:  time.Now(),
				Valid: true,
			},
		}); err2 != nil {
			return errs.Join(err, errors.Wrap(err2, "could not update scan status"))
		}
		return err
	}

	insertResults := func(results []scanner.ScanResult) error {
		for _, result := range results {
			if _, err := runner.queries.CreatePostgresScanResult(ctx, queries.CreatePostgresScanResultParams{
				PostgresScanID: scan.ID,
				Severity:       int32(result.Severity()),
				Message:        result.Detail(),
			}); err != nil {
				return errors.Wrap(err, "could not insert scan result")
			}
		}
		return nil
	}

	conn, err := pgx.Connect(ctx, getPostgresConnectString(db))
	if err != nil {
		return notifyError(errors.Wrap(err, "could not connect to database"))
	}
	defer conn.Close(ctx)

	sc, err := postgres.NewScanner(ctx, conn)
	if err != nil {
		return notifyError(errors.Wrap(err, "could not create scanner"))
	}

	if err := sc.Ping(ctx); err != nil {
		return notifyError(errors.Wrap(err, "could not ping database"))
	}

	if err := sc.CheckPermissions(ctx); err != nil {
		return notifyError(errors.Wrap(err, "could not check permissions"))
	}

	results, err := sc.ScanConfig(ctx)
	if err != nil {
		return notifyError(errors.Wrap(err, "could not scan config"))
	}
	insertResults(results)

	_, err = sc.GetUsers(ctx)
	if err != nil {
		return notifyError(errors.Wrap(err, "could not get users"))
	}

	return errors.Wrap(runner.bruteforcePostgres(ctx, scan, sc, notifyError, insertResults), "could not bruteforce passwords")
}

func (runner *scannerRunner) bruteforcePostgres(
	ctx context.Context,
	scan *models.PostgresScan,
	sc scanner.Scanner,
	notifyError func(error) error,
	insertResults func([]scanner.ScanResult) error,
) error {
	bruteforceResults := map[scanner.User]int64{}
	notifyBruteforceStatus := func(status map[scanner.User]bruteforce.BruteforceUserStatus) error {
		for user, entry := range status {
			if _, ok := bruteforceResults[user]; !ok {
				username, err := user.GetUsername()
				if err != nil {
					return errors.Wrap(err, "could not get username")
				}
				bfuser, err := runner.queries.CreatePostgresScanBruteforceResult(ctx, queries.CreatePostgresScanBruteforceResultParams{
					PostgresScanID: scan.ID,
					Username:       username,
					Password:       sql.NullString{String: entry.FoundPassword, Valid: entry.FoundPassword != ""},
					Tried:          int32(entry.Tried),
					Total:          int32(entry.Total),
				})
				if err != nil {
					return errors.Wrap(err, "could not insert bruteforce result")
				}
				bruteforceResults[user] = bfuser.ID
			} else {
				if err := runner.queries.UpdatePostgresScanBruteforceResult(ctx, queries.UpdatePostgresScanBruteforceResultParams{
					ID:       bruteforceResults[user],
					Password: sql.NullString{String: entry.FoundPassword, Valid: entry.FoundPassword != ""},
					Tried:    int32(entry.Tried),
					Total:    int32(entry.Total),
				}); err != nil {
					return errors.Wrap(err, "could not update bruteforce result")
				}
			}
		}
		return nil
	}

	database, err := runner.queries.GetPostgresDatabase(ctx, scan.PostgresDatabaseID)
	if err != nil {
		return notifyError(errors.Wrap(err, "could not get database"))
	}

	passProvider, err := bruteforce.NewDatabasePasswordProvider(ctx, runner.queries, database.ProjectID)
	if err != nil {
		return notifyError(errors.Wrap(err, "could not create password provider"))
	}
	defer passProvider.Close()

	bruteforcer := bruteforce.NewBruteforcer(passProvider, sc, notifyBruteforceStatus)

	bruteforceResult, err := bruteforcer.BruteforcePasswordAllUsers(ctx)
	if err != nil {
		return notifyError(errors.Wrap(err, "could not bruteforce passwords"))
	}
	if err := insertResults(bruteforceResult); err != nil {
		return notifyError(errors.Wrap(err, "could not insert bruteforce results"))
	}

	if err := runner.queries.UpdatePostgresScanStatus(ctx, queries.UpdatePostgresScanStatusParams{
		ID:     scan.ID,
		Status: models.SCAN_FINISHED,
		EndedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}); err != nil {
		return errors.Wrap(err, "could not update scan status")
	}

	return nil
}
