package config

import (
	"go.opentelemetry.io/otel"
)

var Tracer = otel.Tracer("github.com/tedyst/licenta")
var Meter = otel.Meter("github.com/tedyst/licenta")
var JaegerEndpoint string
var SendgridAPIKey string

const ResetPasswordTokenValidity = 60 * 60 // 1 hour
const EmailSenderName = "Licenta"
const EmailSender = "no-reply@tedyst.ro"
const BaseURL = "http://localhost:8080"

type DatabaseConfiguration struct {
	Host     string `default:"localhost" validate:"required"`
	Port     int    `default:"5432"`
	User     string `default:"postgres"`
	Password string `default:"postgres"`
	Database string `default:"licenta"`
}

type SendgridConfiguration struct {
	UseSendgrid    bool   `default:"false"`
	SendgridAPIKey string `default:""`
}

type EmailConfiguration struct {
	EmailSenderName string                `default:"Licenta"`
	EmailSender     string                `default:"no-reply@tedyst.ro"`
	SendGrid        SendgridConfiguration `required:"false"`
}

type JaegerConfiguration struct {
	UseJaeger bool   `default:"false"`
	Endpoint  string `default:"http://localhost:14268/api/traces"`
}

type OpenTelemetryConfiguration struct {
	UseMetrics bool `default:"true"`
	UseTracing bool `default:"true"`
	Jaeger     JaegerConfiguration
}

type Configuration struct {
	Database      DatabaseConfiguration
	Email         EmailConfiguration
	Debug         bool   `default:"false" flag:"debug"`
	BaseURL       string `default:"http://localhost:8080"`
	OpenTelemetry OpenTelemetryConfiguration
}

var Config Configuration
