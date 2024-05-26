package saver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/scanner/postgres"
)

func getPostgresConnectString(db *queries.PostgresDatabase) string {
	return fmt.Sprintf("host=%s port=%d database=%s user=%s password=%s", db.Host, db.Port, db.DatabaseName, db.Username, db.Password)
}

type PostgresQuerier interface {
	BaseQuerier

	GetPostgresScanByScanID(ctx context.Context, scanID int64) (*queries.PostgresScan, error)
	GetPostgresDatabase(context.Context, queries.GetPostgresDatabaseParams) (*queries.GetPostgresDatabaseRow, error)
	UpdatePostgresVersion(ctx context.Context, params queries.UpdatePostgresVersionParams) error
}

func NewPostgresSaver(ctx context.Context, baseQuerier BaseQuerier, bruteforceProvider bruteforce.BruteforceProvider, scan *queries.Scan, projectIsRemote bool, saltKey string) (Saver, error) {
	q, ok := baseQuerier.(PostgresQuerier)
	if !ok {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("queries is not a PostgresQuerier"))
	}

	postgresScan, err := q.GetPostgresScanByScanID(ctx, scan.ID)
	if err == pgx.ErrNoRows {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("could not get postgres scan: %w", err))
	}
	if err != nil {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("could not get postgres scan: %w", err))
	}

	db, err := q.GetPostgresDatabase(ctx, queries.GetPostgresDatabaseParams{
		ID:      postgresScan.DatabaseID,
		SaltKey: saltKey,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get database: %w", err)
	}

	conn, err := pgx.Connect(ctx, getPostgresConnectString(&queries.PostgresDatabase{
		ID:           db.ID,
		ProjectID:    db.ProjectID,
		Host:         db.Host,
		Port:         db.Port,
		DatabaseName: db.DatabaseName,
		Username:     db.Username,
		Password:     db.Password,
		Version:      db.Version,
		CreatedAt:    db.CreatedAt,
	}))
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
		queries:      q,
		baseSaver:    *createBaseSaver(q, bruteforceProvider, logger, scan, sc, projectIsRemote),
		postgresScan: postgresScan,
		database: &queries.PostgresDatabase{
			ID:           db.ID,
			ProjectID:    db.ProjectID,
			Host:         db.Host,
			Port:         db.Port,
			DatabaseName: db.DatabaseName,
			Username:     db.Username,
			Password:     db.Password,
			Version:      db.Version,
			CreatedAt:    db.CreatedAt,
		},
		connection: conn,
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
	GetPostgresDatabasesForProject(context.Context, queries.GetPostgresDatabasesForProjectParams) ([]*queries.GetPostgresDatabasesForProjectRow, error)
	CreatePostgresScan(ctx context.Context, params queries.CreatePostgresScanParams) (*queries.PostgresScan, error)
}

var CreatePostgresScan = CreateBaseScan(
	func(q BaseCreater) (func(context.Context, int64) ([]*queries.PostgresDatabase, error), error) {
		mq, ok := q.(CreatePostgresScanQuerier)
		if !ok {
			return nil, errors.New("querier is not a CreatePostgresScanQuerier")
		}
		return func(ctx context.Context, projectID int64) ([]*queries.PostgresDatabase, error) {
			rows, err := mq.GetPostgresDatabasesForProject(ctx, queries.GetPostgresDatabasesForProjectParams{
				ProjectID: projectID,
				SaltKey:   viper.GetString("db-encryption-salt"),
			})
			if err != nil {
				return nil, fmt.Errorf("could not get postgres databases: %w", err)
			}

			databases := make([]*queries.PostgresDatabase, len(rows))
			for i, row := range rows {
				databases[i] = &queries.PostgresDatabase{
					ID:           row.ID,
					ProjectID:    row.ProjectID,
					Host:         row.Host,
					Port:         row.Port,
					DatabaseName: row.DatabaseName,
					Username:     row.Username,
					Password:     row.Password,
					Version:      row.Version,
					CreatedAt:    row.CreatedAt,
				}
			}
			return databases, nil
		}, nil
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
	postgres.GetScannerID(),
)

var _ PostgresQuerier = (db.TransactionQuerier)(nil)
var _ CreatePostgresScanQuerier = (db.TransactionQuerier)(nil)
