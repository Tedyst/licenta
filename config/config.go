package config

import (
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/tedyst/licenta/db/generated"
	"go.opentelemetry.io/otel"
)

var DatabasePool *pgxpool.Pool
var DatabaseQueries *db.Queries
var Debug bool
var Tracer = otel.Tracer("github.com/tedyst/licenta")
var Meter = otel.Meter("github.com/tedyst/licenta")
var JaegerEndpoint string
var SendgridAPIKey string

const ResetPasswordTokenValidity = 60 * 60 // 1 hour
const EmailSenderName = "Licenta"
const EmailSender = "no-reply@tedyst.ro"
const BaseURL = "http://localhost:8080"
