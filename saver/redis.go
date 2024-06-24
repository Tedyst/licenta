package saver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	r "github.com/redis/go-redis/v9"
	"github.com/spf13/viper"

	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/scanner/redis"
)

type RedisQuerier interface {
	BaseQuerier

	GetRedisScanByScanID(ctx context.Context, scanID int64) (*queries.RedisScan, error)
	GetRedisDatabase(context.Context, queries.GetRedisDatabaseParams) (*queries.GetRedisDatabaseRow, error)
	UpdateRedisVersion(ctx context.Context, params queries.UpdateRedisVersionParams) error
}

func NewRedisSaver(ctx context.Context, baseQuerier BaseQuerier, bruteforceProvider bruteforce.BruteforceProvider, scan *queries.Scan, projectIsRemote bool, saltKey string) (Saver, error) {
	q, ok := baseQuerier.(RedisQuerier)
	if !ok {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("queries is not a RedisQuerier"))
	}

	redisScan, err := q.GetRedisScanByScanID(ctx, scan.ID)
	if err == pgx.ErrNoRows {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("could not get redis scan: %w", err))
	}
	if err != nil {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("could not get redis scan: %w", err))
	}

	db, err := q.GetRedisDatabase(ctx, queries.GetRedisDatabaseParams{
		ID:      redisScan.DatabaseID,
		SaltKey: saltKey,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get database: %w", err)
	}

	conn := r.NewClient(&r.Options{
		Addr:     db.Host + ":" + fmt.Sprint(db.Port),
		Username: db.Username,
		Password: db.Password,
		DB:       0,
	})

	sc, err := redis.NewScanner(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("could not create scanner: %w", err)
	}

	logger := slog.With(
		"scan", scan.ID,
		"redis_scan", redisScan.ID,
		"redis_database_id", redisScan.DatabaseID,
	)

	saver := &redisSaver{
		queries:   q,
		baseSaver: *createBaseSaver(q, bruteforceProvider, logger, scan, sc, projectIsRemote),
		redisScan: redisScan,
		database: &queries.RedisDatabase{
			ID:        db.ID,
			ProjectID: db.ProjectID,
			Host:      db.Host,
			Port:      db.Port,
			Username:  db.Username,
			Password:  db.Password,
			Version:   db.Version,
			CreatedAt: db.CreatedAt,
		},
		connection: conn,
	}
	saver.runAfterScan = saver.hookAfterScan
	return saver, nil
}

func (saver *redisSaver) hookAfterScan(ctx context.Context) error {
	version, err := saver.scanner.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("could not get version: %w", err)
	}
	if err := saver.queries.UpdateRedisVersion(ctx, queries.UpdateRedisVersionParams{
		ID:      saver.redisScan.DatabaseID,
		Version: sql.NullString{String: version, Valid: true},
	}); err != nil {
		return fmt.Errorf("could not update version: %w", err)
	}

	return saver.connection.Close()
}

type redisSaver struct {
	queries RedisQuerier

	connection *r.Client

	redisScan *queries.RedisScan
	database  *queries.RedisDatabase

	baseSaver
}

func init() {
	savers["redis"] = NewRedisSaver
	creaters["redis"] = CreateRedisScan
}

type CreateRedisScanQuerier interface {
	BaseCreater
	GetRedisDatabasesForProject(context.Context, queries.GetRedisDatabasesForProjectParams) ([]*queries.GetRedisDatabasesForProjectRow, error)
	CreateRedisScan(ctx context.Context, params queries.CreateRedisScanParams) (*queries.RedisScan, error)
}

var CreateRedisScan = CreateBaseScan(
	func(q BaseCreater) (func(context.Context, int64) ([]*queries.RedisDatabase, error), error) {
		mq, ok := q.(CreateRedisScanQuerier)
		if !ok {
			return nil, errors.New("querier is not a CreateRedisScanQuerier")
		}
		return func(ctx context.Context, projectID int64) ([]*queries.RedisDatabase, error) {
			rows, err := mq.GetRedisDatabasesForProject(ctx, queries.GetRedisDatabasesForProjectParams{
				ProjectID: projectID,
				SaltKey:   viper.GetString("db-encryption-salt"),
			})
			if err != nil {
				return nil, fmt.Errorf("could not get databases: %w", err)
			}

			dbs := make([]*queries.RedisDatabase, 0, len(rows))
			for _, row := range rows {
				dbs = append(dbs, &queries.RedisDatabase{
					ID:        row.ID,
					ProjectID: row.ProjectID,
					Host:      row.Host,
					Port:      row.Port,
					Username:  row.Username,
					Password:  row.Password,
					Version:   row.Version,
					CreatedAt: row.CreatedAt,
				})
			}
			return dbs, nil
		}, nil
	},
	func(ctx context.Context, q BaseCreater, scanID int64, db *queries.RedisDatabase) (any, error) {
		mq, ok := q.(CreateRedisScanQuerier)
		if !ok {
			return nil, errors.New("querier is not a CreateRedisScanQuerier")
		}
		return mq.CreateRedisScan(ctx, queries.CreateRedisScanParams{
			ScanID:     scanID,
			DatabaseID: db.ID,
		})
	},
	redis.GetScannerID(),
)

var _ RedisQuerier = (db.TransactionQuerier)(nil)
var _ CreateRedisScanQuerier = (db.TransactionQuerier)(nil)
