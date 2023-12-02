package db

import (
	"context"
	"errors"
	"log"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tedyst/licenta/db/queries"
)

type TransactionQuerier interface {
	queries.Querier

	StartTransaction(ctx context.Context) (TransactionQuerier, error)
	EndTransaction(ctx context.Context, err error) error
	GetRawPool() *pgxpool.Pool
}

type querierImpl struct {
	*queries.Queries

	pool *pgxpool.Pool
	tx   pgx.Tx
}

func (q querierImpl) StartTransaction(ctx context.Context) (TransactionQuerier, error) {
	if q.tx != nil {
		newTx, err := q.tx.Begin(ctx)
		if err != nil {
			return nil, err
		}
		return querierImpl{
			pool:    q.pool,
			tx:      newTx,
			Queries: queries.New(newTx),
		}, nil
	}

	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return querierImpl{
		pool:    q.pool,
		tx:      tx,
		Queries: queries.New(tx),
	}, nil
}

func (q querierImpl) EndTransaction(ctx context.Context, err error) error {
	if err != nil {
		return errors.Join(err, q.tx.Rollback(ctx))
	}
	return errors.Join(err, q.tx.Commit(ctx))
}

func NewQuerier(pool *pgxpool.Pool) *querierImpl {
	return &querierImpl{
		pool:    pool,
		Queries: queries.New(pool),
	}
}

func (q querierImpl) GetRawTx() pgx.Tx {
	return q.tx
}

func (q querierImpl) GetRawPool() *pgxpool.Pool {
	return q.pool
}

func InitDatabase(uri string) *querierImpl {
	cfg, err := pgxpool.ParseConfig(uri)
	if err != nil {
		log.Fatal(err)
	}
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	return NewQuerier(pool)
}
