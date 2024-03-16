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
	if err == pgx.ErrNoRows {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("could not get postgres scan: %w", err))
	}
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
		connection:   conn,
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

	return saver.connection.Close(ctx)
}

type postgresSaver struct {
	queries PostgresQuerier

	connection *pgx.Conn

	postgresScan *queries.PostgresScan
	database     *queries.PostgresDatabase

	baseSaver
}

func init() {
	savers["postgres"] = NewPostgresSaver
	creaters["postgres"] = CreatePostgresScan
}

type CreatePostgresScanQuerier interface {
	BaseCreater
	GetPostgresDatabasesForProject(ctx context.Context, projectID int64) ([]*queries.PostgresDatabase, error)
	CreatePostgresScan(ctx context.Context, params queries.CreatePostgresScanParams) (*queries.PostgresScan, error)
}

var CreatePostgresScan = CreateBaseScan(
	func(q BaseCreater) (func(context.Context, int64) ([]*queries.PostgresDatabase, error), error) {
		mq, ok := q.(CreatePostgresScanQuerier)
		if !ok {
			return nil, errors.New("querier is not a CreatePostgresScanQuerier")
		}
		return mq.GetPostgresDatabasesForProject, nil
	},
	func(ctx context.Context, q BaseCreater, scanID int64, db *queries.PostgresDatabase) (any, error) {
		mq, ok := q.(CreatePostgresScanQuerier)
		if !ok {
			return nil, errors.New("querier is not a CreatePostgresScanQuerier")
		}
		return mq.CreatePostgresScan(ctx, queries.CreatePostgresScanParams{
			ScanID:     scanID,
			DatabaseID: db.ID,
		})
	},
)
