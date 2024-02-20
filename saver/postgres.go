package saver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/scanner/postgres"
)

func getPostgresConnectString(db *queries.PostgresDatabase) string {
	return fmt.Sprintf("host=%s port=%d database=%s user=%s password=%s", db.Host, db.Port, db.DatabaseName, db.Username, db.Password)
}

type PostgresQuerier interface {
	BaseQuerier

	GetPostgresScanByScanID(ctx context.Context, scanID int64) (*queries.PostgresScan, error)
	GetPostgresDatabase(ctx context.Context, id int64) (*queries.GetPostgresDatabaseRow, error)
	UpdatePostgresVersion(ctx context.Context, params queries.UpdatePostgresVersionParams) error
}

func NewPostgresSaver(ctx context.Context, baseQuerier BaseQuerier, bruteforceProvider bruteforce.BruteforceProvider, scan *queries.Scan) (Saver, error) {
	queries, ok := baseQuerier.(PostgresQuerier)
	if !ok {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("queries is not a PostgresQuerier"))
	}

	postgresScan, err := queries.GetPostgresScanByScanID(ctx, scan.ID)
	if err != nil {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("could not get postgres scan: %w", err))
	}

	db, err := queries.GetPostgresDatabase(ctx, postgresScan.DatabaseID)
	if err != nil {
		return nil, fmt.Errorf("could not get database: %w", err)
	}

	conn, err := pgx.Connect(ctx, getPostgresConnectString(&db.PostgresDatabase))
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	sc, err := postgres.NewScanner(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("could not create scanner: %w", err)
	}

	logger := slog.With(
		"scan", scan.ID,
		"postgres_scan", postgresScan.ID,
		"postgres_database_id", postgresScan.DatabaseID,
	)

	saver := &postgresSaver{
		queries:      queries,
		baseSaver:    *createBaseSaver(queries, bruteforceProvider, logger, scan, sc),
		postgresScan: postgresScan,
		database:     &db.PostgresDatabase,
	}
	saver.runAfterScan = saver.hookAfterScan
	return saver, nil
}

func (saver *postgresSaver) hookAfterScan(ctx context.Context) error {
	version, err := saver.scanner.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("could not get version: %w", err)
	}
	if err := saver.queries.UpdatePostgresVersion(ctx, queries.UpdatePostgresVersionParams{
		ID:      saver.postgresScan.DatabaseID,
		Version: sql.NullString{String: version, Valid: true},
	}); err != nil {
		return fmt.Errorf("could not update version: %w", err)
	}

	return nil
}

type postgresSaver struct {
	queries PostgresQuerier

	postgresScan *queries.PostgresScan
	database     *queries.PostgresDatabase

	baseSaver
}
