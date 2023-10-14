package local

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var tracer = otel.Tracer("github.com/tedyst/licenta/tasks/local")
var meter = otel.Meter("github.com/tedyst/licenta/tasks/local")
var mailsSent metric.Int64Counter

func init() {
	var err error
	mailsSent, err = meter.Int64Counter("mails_sent", metric.WithDescription("Number of mails sent"))
	if err != nil {
		panic(err)
	}
}
