package saver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/scanner/mongodb"
	m "go.mongodb.org/mongo-driver/mongo"
)

type MongoQuerier interface {
	BaseQuerier

	GetMongoScanByScanID(ctx context.Context, scanID int64) (*queries.MongoScan, error)
	GetMongoDatabase(context.Context, queries.GetMongoDatabaseParams) (*queries.GetMongoDatabaseRow, error)
	UpdateMongoVersion(ctx context.Context, params queries.UpdateMongoVersionParams) error
}

func getMongoConnectString(db *queries.MongoDatabase) string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", db.Username, db.Password, db.Host, db.Port, db.DatabaseName)
}

func NewMongoSaver(ctx context.Context, baseQuerier BaseQuerier, bruteforceProvider bruteforce.BruteforceProvider, scan *queries.Scan, projectIsRemote bool, saltKey string) (Saver, error) {
	q, ok := baseQuerier.(MongoQuerier)
	if !ok {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("queries is not a MongoQuerier"))
	}

	mongoScan, err := q.GetMongoScanByScanID(ctx, scan.ID)
	if err == pgx.ErrNoRows {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("could not get mongo scan: %w", err))
	}
	if err != nil {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("could not get mongo scan: %w", err))
	}

	db, err := q.GetMongoDatabase(ctx, queries.GetMongoDatabaseParams{
		ID:      mongoScan.DatabaseID,
		SaltKey: saltKey,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get database: %w", err)
	}

	conn, err := m.Connect(ctx, options.Client().
		ApplyURI(getMongoConnectString(&queries.MongoDatabase{
			ID:           db.ID,
			ProjectID:    db.ProjectID,
			Host:         db.Host,
			Port:         db.Port,
			DatabaseName: db.DatabaseName,
			Username:     db.Username,
			Password:     db.Password,
			Version:      db.Version,
			CreatedAt:    db.CreatedAt,
		})),
		options.Client().SetConnectTimeout(time.Second*5),
		options.Client().SetTimeout(time.Second*5),
		options.Client().SetSocketTimeout(time.Second*5),
		options.Client().SetServerSelectionTimeout(time.Second*5),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create database connection: %w", err)
	}

	sc, err := mongodb.NewScanner(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("could not create scanner: %w", err)
	}

	logger := slog.With(
		"scan", scan.ID,
		"mongo_scan", mongoScan.ID,
		"mongo_database_id", mongoScan.DatabaseID,
	)

	saver := &mongoSaver{
		queries:   q,
		baseSaver: *createBaseSaver(q, bruteforceProvider, logger, scan, sc, projectIsRemote),
		mongoScan: mongoScan,
		database: &queries.MongoDatabase{
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

func (saver *mongoSaver) hookAfterScan(ctx context.Context) error {
	version, err := saver.scanner.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("could not get version: %w", err)
	}
	if err := saver.queries.UpdateMongoVersion(ctx, queries.UpdateMongoVersionParams{
		ID:      saver.mongoScan.DatabaseID,
		Version: sql.NullString{String: version, Valid: true},
	}); err != nil {
		return fmt.Errorf("could not update version: %w", err)
	}

	return saver.connection.Disconnect(ctx)
}

type mongoSaver struct {
	queries MongoQuerier

	connection *m.Client

	mongoScan *queries.MongoScan
	database  *queries.MongoDatabase

	baseSaver
}

func init() {
	savers["mongo"] = NewMongoSaver
	creaters["mongo"] = CreateMongoScan
}

type CreateMongoScanQuerier interface {
	BaseCreater
	GetMongoDatabasesForProject(context.Context, queries.GetMongoDatabasesForProjectParams) ([]*queries.GetMongoDatabasesForProjectRow, error)
	CreateMongoScan(ctx context.Context, params queries.CreateMongoScanParams) (*queries.MongoScan, error)
}

var CreateMongoScan = CreateBaseScan(
	func(q BaseCreater) (func(context.Context, int64) ([]*queries.MongoDatabase, error), error) {
		mq, ok := q.(CreateMongoScanQuerier)
		if !ok {
			return nil, errors.New("querier is not a CreateMongoScanQuerier")
		}
		return func(ctx context.Context, projectID int64) ([]*queries.MongoDatabase, error) {
			rows, err := mq.GetMongoDatabasesForProject(ctx, queries.GetMongoDatabasesForProjectParams{
				ProjectID: projectID,
				SaltKey:   viper.GetString("db-encryption-salt"),
			})
			if err != nil {
				return nil, fmt.Errorf("could not get databases: %w", err)
			}

			dbs := make([]*queries.MongoDatabase, 0, len(rows))
			for _, row := range rows {
				dbs = append(dbs, &queries.MongoDatabase{
					ID:           row.ID,
					ProjectID:    row.ProjectID,
					Host:         row.Host,
					Port:         row.Port,
					DatabaseName: row.DatabaseName,
					Username:     row.Username,
					Password:     row.Password,
					Version:      row.Version,
					CreatedAt:    row.CreatedAt,
				})
			}
			return dbs, nil
		}, nil
	},
	func(ctx context.Context, q BaseCreater, scanID int64, db *queries.MongoDatabase) (any, error) {
		mq, ok := q.(CreateMongoScanQuerier)
		if !ok {
			return nil, errors.New("querier is not a CreateMongoScanQuerier")
		}
		return mq.CreateMongoScan(ctx, queries.CreateMongoScanParams{
			ScanID:     scanID,
			DatabaseID: db.ID,
		})
	},
	mongodb.GetScannerID(),
)

var _ MongoQuerier = (db.TransactionQuerier)(nil)
var _ CreateMongoScanQuerier = (db.TransactionQuerier)(nil)
