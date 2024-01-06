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
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/scanner"
	"github.com/tedyst/licenta/scanner/postgres"
)

func getPostgresConnectString(db *models.PostgresDatabases) string {
	return fmt.Sprintf("host=%s port=%d database=%s user=%s password=%s", db.Host, db.Port, db.DatabaseName, db.Username, db.Password)
}

type postgresQuerier interface {
	scanQuerier

	GetPostgresDatabase(ctx context.Context, id int64) (*queries.GetPostgresDatabaseRow, error)
	UpdatePostgresVersion(ctx context.Context, params queries.UpdatePostgresVersionParams) error
}

type postgresScanRunner struct {
	queries            postgresQuerier
	bruteforceProvider bruteforce.BruteforceProvider

	PostgresScannerProvider func(ctx context.Context, db *pgx.Conn) (scanner.Scanner, error)
}

func NewPostgresScanRunner(queries postgresQuerier, bruteforceProvider bruteforce.BruteforceProvider) *postgresScanRunner {
	return &postgresScanRunner{
		queries:                 queries,
		bruteforceProvider:      bruteforceProvider,
		PostgresScannerProvider: postgres.NewScanner,
	}
}

func (runner *postgresScanRunner) ScanPostgresDB(ctx context.Context, postgresScan *models.PostgresScan) error {
	ctx, span := tracer.Start(ctx, "ScanPostgresDB")
	defer span.End()

	scan, err := runner.queries.GetScan(ctx, postgresScan.ScanID)
	if err != nil {
		return errors.Wrap(err, "could not get scan")
	}

	if err := runner.scanPostgresDB(ctx, postgresScan, &scan.Scan, false); err != nil {
		if err2 := runner.queries.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
			ID:     scan.Scan.ID,
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

	if err := runner.queries.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
		ID:     postgresScan.ScanID,
		Status: models.SCAN_FINISHED,
	}); err != nil {
		return errors.Wrap(err, "could not update scan status")
	}

	return nil
}

func (runner *postgresScanRunner) createBaseScanner(ctx context.Context, postgresScan *models.PostgresScan, scan *models.Scan) (scanner.Scanner, *baseScanRunner, error) {
	logger := slog.With(
		"scan", scan.ID,
		"postgres_scan", postgresScan.ID,
		"postgres_database_id", postgresScan.DatabaseID,
		"project_id", scan.ProjectID,
	)

	db, err := runner.queries.GetPostgresDatabase(ctx, postgresScan.DatabaseID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not get database")
	}

	logger.InfoContext(ctx, "Starting Postgres DB scan")
	defer logger.InfoContext(ctx, "Finished Postgres DB scan")

	conn, err := pgx.Connect(ctx, getPostgresConnectString(&db.PostgresDatabase))
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not connect to database")
	}

	logger.DebugContext(ctx, "Connected to database")

	sc, err := runner.PostgresScannerProvider(ctx, conn)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not create scanner")
	}

	logger.DebugContext(ctx, "Created scanner")

	return sc, createScanner(ctx, runner.queries, runner.bruteforceProvider, logger, scan, sc), nil
}

func (runner *postgresScanRunner) scanPostgresDB(ctx context.Context, postgresScan *models.PostgresScan, scan *models.Scan, shouldOnlyScanRemote bool) error {
	sc, scanRunner, err := runner.createBaseScanner(ctx, postgresScan, scan)
	if err != nil {
		return errors.Wrap(err, "could not create scanner")
	}

	if shouldOnlyScanRemote {
		return errors.Wrap(sc.ScanForPublicAccess(ctx), "could not scan for public access")
	}

	if err := scanRunner.run(ctx); err != nil {
		return errors.Wrap(err, "could not scan")
	}

	version, err := sc.GetVersion(ctx)
	if err != nil {
		return errors.Wrap(err, "could not get version")
	}
	if err := runner.queries.UpdatePostgresVersion(ctx, queries.UpdatePostgresVersionParams{
		ID:      postgresScan.DatabaseID,
		Version: sql.NullString{String: version, Valid: true},
	}); err != nil {
		return errors.Wrap(err, "could not update version")
	}

	return nil
}
