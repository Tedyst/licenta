package nats

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("github.com/tedyst/licenta/tasks/nats")
