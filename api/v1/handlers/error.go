package handlers

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

func traceError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}
