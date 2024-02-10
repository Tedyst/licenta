package local

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/scanner"
	"github.com/tedyst/licenta/scanner/postgres"
)

func getPostgresConnectString(db *queries.PostgresDatabase) string {
	return fmt.Sprintf("host=%s port=%d database=%s user=%s password=%s", db.Host, db.Port, db.DatabaseName, db.Username, db.Password)
}

type PostgresQuerier interface {
	ScanQuerier

	GetPostgresDatabase(ctx context.Context, id int64) (*queries.GetPostgresDatabaseRow, error)
	UpdatePostgresVersion(ctx context.Context, params queries.UpdatePostgresVersionParams) error
}

type postgresScanRunner struct {
	queries            PostgresQuerier
	bruteforceProvider bruteforce.BruteforceProvider

	PostgresScannerProvider func(ctx context.Context, db *pgx.Conn) (scanner.Scanner, error)
}

func NewPostgresScanRunner(queries PostgresQuerier, bruteforceProvider bruteforce.BruteforceProvider) *postgresScanRunner {
	return &postgresScanRunner{
		queries:                 queries,
		bruteforceProvider:      bruteforceProvider,
		PostgresScannerProvider: postgres.NewScanner,
	}
}

func (runner *postgresScanRunner) ScanPostgresDB(ctx context.Context, postgresScan *queries.PostgresScan) error {
	ctx, span := tracer.Start(ctx, "ScanPostgresDB")
	defer span.End()

	scan, err := runner.queries.GetScan(ctx, postgresScan.ScanID)
	if err != nil {
		return fmt.Errorf("could not get scan: %w", err)
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
			return errors.Join(err, fmt.Errorf("could not update scan status: %w", err2))
		}
		return err
	}

	if err := runner.queries.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
		ID:     scan.Scan.ID,
		Status: models.SCAN_FINISHED,
	}); err != nil {
		return fmt.Errorf("could not update scan status: %w", err)
	}

	return nil
}

func (runner *postgresScanRunner) createBaseScanner(ctx context.Context, postgresScan *queries.PostgresScan, scan *queries.Scan) (scanner.Scanner, *baseScanRunner, error) {
	logger := slog.With(
		"scan", scan.ID,
		"postgres_scan", postgresScan.ID,
		"postgres_database_id", postgresScan.DatabaseID,
	)

	db, err := runner.queries.GetPostgresDatabase(ctx, postgresScan.DatabaseID)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get database: %w", err)
	}

	logger.InfoContext(ctx, "Starting Postgres DB scan")
	defer logger.InfoContext(ctx, "Finished Postgres DB scan")

	conn, err := pgx.Connect(ctx, getPostgresConnectString(&db.PostgresDatabase))
	if err != nil {
		return nil, nil, fmt.Errorf("could not connect to database: %w", err)
	}

	logger.DebugContext(ctx, "Connected to database")

	sc, err := runner.PostgresScannerProvider(ctx, conn)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create scanner: %w", err)
	}

	logger.DebugContext(ctx, "Created scanner")

	return sc, createScanner(ctx, runner.queries, runner.bruteforceProvider, logger, scan, sc), nil
}

func (runner *postgresScanRunner) scanPostgresDB(ctx context.Context, postgresScan *queries.PostgresScan, scan *queries.Scan, shouldOnlyScanRemote bool) error {
	sc, scanRunner, err := runner.createBaseScanner(ctx, postgresScan, scan)
	if err != nil {
		return fmt.Errorf("could not create scanner: %w", err)
	}

	if shouldOnlyScanRemote {
		err := scanRunner.scanForPublicAccess(ctx)
		if err != nil {
			return fmt.Errorf("could not scan for public access: %w", err)
		}
	}

	if err := scanRunner.run(ctx); err != nil {
		return fmt.Errorf("could not scan: %w", err)
	}

	version, err := sc.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("could not get version: %w", err)
	}
	if err := runner.queries.UpdatePostgresVersion(ctx, queries.UpdatePostgresVersionParams{
		ID:      postgresScan.DatabaseID,
		Version: sql.NullString{String: version, Valid: true},
	}); err != nil {
		return fmt.Errorf("could not update version: %w", err)
	}

	return nil
}
