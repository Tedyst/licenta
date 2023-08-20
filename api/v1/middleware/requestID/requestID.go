package requestid

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		w.Header().Set("X-Trace-Id", span.SpanContext().TraceID().String())
		next.ServeHTTP(w, r)
	})
}
