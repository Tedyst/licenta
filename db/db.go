package database

import (
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/tedyst/licenta/db/generated"
)

var DatabasePool *pgxpool.Pool
var DatabaseQueries *db.Queries
