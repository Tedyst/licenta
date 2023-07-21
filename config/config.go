package config

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tedyst/licenta/db"
	"go.opentelemetry.io/otel"
)

var DatabasePool *pgxpool.Pool
var DatabaseQueries *db.Queries
var Debug bool
var Tracer = otel.Tracer("github.com/tedyst/licenta")
var Meter = otel.Meter("github.com/tedyst/licenta")
var JaegerEndpoint string
