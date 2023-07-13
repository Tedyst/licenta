package config

import (
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tedyst/licenta/db"
)

var DatabasePool *pgxpool.Pool
var DatabaseQueries *db.Queries
var SessionStore *session.Store
var Debug bool
