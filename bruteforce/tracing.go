package bruteforce

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var tracer = otel.Tracer("github.com/tedyst/licenta/bruteforce")
var meter = otel.Meter("github.com/tedyst/licenta/bruteforce")

var passwordsTried metric.Int64Counter
var passwordsAlreadyFound metric.Int64Counter

func init() {
	var err error
	passwordsTried, err = meter.Int64Counter("passwords_tried", metric.WithDescription("Number password tried bruteforcing"))
	if err != nil {
		panic(err)
	}
	passwordsAlreadyFound, err = meter.Int64Counter("passwords_already_found", metric.WithDescription("Number password tried bruteforcing"))
	if err != nil {
		panic(err)
	}
}
