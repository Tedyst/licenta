package handlers

import "go.opentelemetry.io/otel"

const DefaultPaginationLimit = 10
const Prefix = "/api/v1"

var Tracer = otel.Tracer("github.com/tedyst/licenta/api/v1")
var Meter = otel.Meter("github.com/tedyst/licenta/api/v1")
