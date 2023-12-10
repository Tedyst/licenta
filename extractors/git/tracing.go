package git

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var tracer = otel.Tracer("github.com/tedyst/licenta/extractors/git")
var meter = otel.Meter("github.com/tedyst/licenta/extractors/git")

var commitsInspected metric.Int64Counter
var foundResults metric.Int64Counter

func init() {
	var err error
	commitsInspected, err = meter.Int64Counter("commits_inspected", metric.WithDescription("Number of commits inspected"))
	if err != nil {
		panic(err)
	}
	foundResults, err = meter.Int64Counter("found_results", metric.WithDescription("Number of results found"))
	if err != nil {
		panic(err)
	}
}
