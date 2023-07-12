package config

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tedyst/licenta/db"
)

var DatabasePool *pgxpool.Pool
var DatabaseQueries *db.Queries
