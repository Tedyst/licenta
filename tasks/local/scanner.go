package local

import (
	"context"
	"database/sql"
	errs "errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/scanner"
	"github.com/tedyst/licenta/scanner/postgres"
)

func getPostgresConnectString(db *models.PostgresDatabases) string {
	return fmt.Sprintf("host=%s port=%d database=%s user=%s password=%s", db.Host, db.Port, db.DatabaseName, db.Username, db.Password)
}

type scannerRunner struct {
	queries            postgresQuerier
	bruteforceProvider bruteforce.BruteforceProvider
	messageExchange    messages.Exchange
}

type postgresQuerier interface {
	GetPostgresDatabase(ctx context.Context, id int64) (*models.PostgresDatabases, error)
	UpdatePostgresScanStatus(ctx context.Context, params queries.UpdatePostgresScanStatusParams) error
	CreatePostgresScanResult(ctx context.Context, params queries.CreatePostgresScanResultParams) (*models.PostgresScanResult, error)
	CreatePostgresScanBruteforceResult(ctx context.Context, arg queries.CreatePostgresScanBruteforceResultParams) (*models.PostgresScanBruteforceResult, error)
	UpdatePostgresScanBruteforceResult(ctx context.Context, params queries.UpdatePostgresScanBruteforceResultParams) error
	GetWorkersForProject(ctx context.Context, projectID int64) ([]*queries.GetWorkersForProjectRow, error)
}

func NewScannerRunner(queries postgresQuerier, bruteforceProvider bruteforce.BruteforceProvider, exchange messages.Exchange) *scannerRunner {
	return &scannerRunner{
		queries:            queries,
		bruteforceProvider: bruteforceProvider,
		messageExchange:    exchange,
	}
}

func (runner *scannerRunner) ScanPostgresDB(ctx context.Context, scan *models.PostgresScan) error {
	ctx, span := tracer.Start(ctx, "ScanPostgresDB")
	defer span.End()

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

	logger := slog.With(
		"scan", scan.ID,
		"database_id", scan.PostgresDatabaseID,
		"project_id", db.ProjectID,
	)

	logger.InfoContext(ctx, "Starting Postgres DB scan")
	defer logger.InfoContext(ctx, "Finished Postgres DB scan")

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

	logger.DebugContext(ctx, "Connected to database")

	sc, err := postgres.NewScanner(ctx, conn)
	if err != nil {
		return notifyError(errors.Wrap(err, "could not create scanner"))
	}

	logger.DebugContext(ctx, "Created scanner")

	if err := sc.Ping(ctx); err != nil {
		return notifyError(errors.Wrap(err, "could not ping database"))
	}

	logger.DebugContext(ctx, "Pinged database")

	if err := sc.CheckPermissions(ctx); err != nil {
		return notifyError(errors.Wrap(err, "could not check permissions"))
	}

	logger.DebugContext(ctx, "Checked permissions")

	results, err := sc.ScanConfig(ctx)
	if err != nil {
		return notifyError(errors.Wrap(err, "could not scan config"))
	}
	insertResults(results)

	logger.DebugContext(ctx, "Scanned config")

	_, err = sc.GetUsers(ctx)
	if err != nil {
		return notifyError(errors.Wrap(err, "could not get users"))
	}

	logger.DebugContext(ctx, "Got users")

	return errors.Wrap(runner.bruteforcePostgres(ctx, scan, sc, notifyError, insertResults, logger), "could not bruteforce passwords")
}

func (runner *scannerRunner) bruteforcePostgres(
	ctx context.Context,
	scan *models.PostgresScan,
	sc scanner.Scanner,
	notifyError func(error) error,
	insertResults func([]scanner.ScanResult) error,
	logger *slog.Logger,
) error {
	logger.DebugContext(ctx, "Bruteforcing passwords for all users")

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

	bruteforcer, err := runner.bruteforceProvider.NewBruteforcer(ctx, sc, notifyBruteforceStatus, int(database.ProjectID))
	if err != nil {
		return notifyError(errors.Wrap(err, "could not create bruteforcer"))
	}

	bruteforceResult, err := bruteforcer.BruteforcePasswordAllUsers(ctx)
	if err != nil {
		return notifyError(errors.Wrap(err, "could not bruteforce passwords"))
	}
	if err := insertResults(bruteforceResult); err != nil {
		return notifyError(errors.Wrap(err, "could not insert bruteforce results"))
	}

	logger.DebugContext(ctx, "Bruteforced passwords for all users")

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

func (runner *scannerRunner) ScanPostgresDBForPublicAccess(ctx context.Context, scan *models.PostgresScan) error {
	ctx, span := tracer.Start(ctx, "ScanPostgresDBForPublicAccess")
	defer span.End()

	db, err := runner.queries.GetPostgresDatabase(ctx, scan.PostgresDatabaseID)
	if err != nil {
		return errors.Wrap(err, "could not get database")
	}

	if !db.Remote {
		return nil
	}

	logger := slog.With(
		"scan", scan.ID,
		"database_id", scan.PostgresDatabaseID,
		"project_id", db.ProjectID,
	)

	logger.InfoContext(ctx, "Starting Postgres DB scan for public access")
	defer logger.InfoContext(ctx, "Finished Postgres DB scan for public access")

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

	conn, err := pgx.Connect(ctx, getPostgresConnectString(db))
	if err != nil {
		return notifyError(errors.Wrap(err, "could not connect to database"))
	}
	defer conn.Close(ctx)

	logger.DebugContext(ctx, "Connected to database")

	sc, err := postgres.NewScanner(ctx, conn)
	if err != nil {
		return notifyError(errors.Wrap(err, "could not create scanner"))
	}

	logger.DebugContext(ctx, "Created scanner")

	err = sc.Ping(ctx)
	if err == nil {
		logger.DebugContext(ctx, "Database is accessible from the internet")

		if _, err := runner.queries.CreatePostgresScanResult(ctx, queries.CreatePostgresScanResultParams{
			PostgresScanID: scan.ID,
			Severity:       int32(scanner.SEVERITY_HIGH),
			Message:        "Database is accessible from the internet",
		}); err != nil {
			return errors.Wrap(err, "could not insert scan result")
		}

		return nil
	}

	return nil
}

func (runner *scannerRunner) SchedulePostgresScan(ctx context.Context, scan *models.PostgresScan) error {
	database, err := runner.queries.GetPostgresDatabase(ctx, scan.PostgresDatabaseID)
	if err != nil {
		return errors.Wrap(err, "could not get database")
	}

	if database.Remote {
		err = runner.ScanPostgresDBForPublicAccess(ctx, scan)
		if err != nil {
			return errors.Wrap(err, "could not scan database for public access")
		}

		workers, err := runner.queries.GetWorkersForProject(ctx, database.ProjectID)
		if err != nil {
			return errors.Wrap(err, "could not get workers for project")
		}

		if len(workers) == 0 {
			return errors.New("no workers available")
		}

		for _, worker := range workers {
			runner.messageExchange.PublishSendScanToWorkerMessage(ctx, worker.Worker, int(scan.ID), int(database.ProjectID))
		}

		return nil
	}

	return runner.ScanPostgresDB(ctx, scan)
}
