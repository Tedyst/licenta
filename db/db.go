package database

import (
	"context"
	"log"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	db "github.com/tedyst/licenta/db/generated"
)

var DatabasePool *pgxpool.Pool
var DatabaseQueries *db.Queries

func InitDatabase() *db.Queries {
	cfg, err := pgxpool.ParseConfig(viper.GetString("database"))
	if err != nil {
		log.Fatal(err)
	}
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()
	DatabasePool, err = pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}
	DatabaseQueries = db.New(DatabasePool)
	err = DatabasePool.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return DatabaseQueries
}
