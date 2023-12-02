package local

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/scanner"
	"github.com/tedyst/licenta/scanner/postgres"
)

func getPostgresConnectString(db *models.PostgresDatabases) string {
	return fmt.Sprintf("host=%s port=%d database=%s user=%s password=%s", db.Host, db.Port, db.DatabaseName, db.Username, db.Password)
}

func (runner *localRunner) ScanPostgresDB(ctx context.Context, scan *models.PostgresScan) error {
	if err := runner.queries.UpdatePostgresScanStatus(ctx, queries.UpdatePostgresScanStatusParams{
		ID:     scan.ID,
		Status: models.SCAN_RUNNING,
	}); err != nil {
		return err
	}

	db, err := runner.queries.GetPostgresDatabase(ctx, scan.PostgresDatabaseID)
	if err != nil {
		return err
	}

	notifyError := func(err error) error {
		if err2 := runner.queries.UpdatePostgresScanStatus(ctx, queries.UpdatePostgresScanStatusParams{
			ID:     scan.ID,
			Status: models.SCAN_FINISHED,
			Error:  sql.NullString{String: err.Error(), Valid: true},
		}); err2 != nil {
			return errors.Join(err, err2)
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
				return err
			}
		}
		return nil
	}

	conn, err := pgx.Connect(ctx, getPostgresConnectString(db))
	if err != nil {
		return notifyError(err)
	}
	defer conn.Close(ctx)

	sc, err := postgres.NewScanner(ctx, conn)
	if err != nil {
		return notifyError(err)
	}

	if err := sc.Ping(ctx); err != nil {
		return notifyError(err)
	}

	if err := sc.CheckPermissions(ctx); err != nil {
		return notifyError(err)
	}

	results, err := sc.ScanConfig(ctx)
	if err != nil {
		return notifyError(err)
	}
	insertResults(results)

	_, err = sc.GetUsers(ctx)
	if err != nil {
		return notifyError(err)
	}

	bruteforceResults := map[scanner.User]int64{}
	notifyBruteforceStatus := func(status map[scanner.User]bruteforce.BruteforceUserStatus) error {
		for user, entry := range status {
			if _, ok := bruteforceResults[user]; !ok {
				username, err := user.GetUsername()
				if err != nil {
					return err
				}
				bfuser, err := runner.queries.CreatePostgresScanBruteforceResult(ctx, queries.CreatePostgresScanBruteforceResultParams{
					PostgresScanID: scan.ID,
					Username:       username,
					Password:       sql.NullString{String: entry.FoundPassword, Valid: entry.FoundPassword != ""},
					Tried:          int32(entry.Tried),
				})
				if err != nil {
					return err
				}
				bruteforceResults[user] = bfuser.ID
			} else {
				if err := runner.queries.UpdatePostgresScanBruteforceResult(ctx, queries.UpdatePostgresScanBruteforceResultParams{
					ID:       bruteforceResults[user],
					Password: sql.NullString{String: entry.FoundPassword, Valid: true},
					Tried:    int32(entry.Tried),
					Total:    int32(entry.Total),
				}); err != nil {
					return err
				}
			}
		}
		return nil
	}

	bruteforceResult, err := bruteforce.BruteforcePasswordAllUsers(ctx, sc, runner.queries, notifyBruteforceStatus)
	if err != nil {
		return notifyError(err)
	}

	if err := insertResults(bruteforceResult); err != nil {
		return notifyError(err)
	}

	if err := runner.queries.UpdatePostgresScanStatus(ctx, queries.UpdatePostgresScanStatusParams{
		ID:     scan.ID,
		Status: models.SCAN_FINISHED,
	}); err != nil {
		return err
	}

	return nil
}
