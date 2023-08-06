package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

func traceError(c *fiber.Ctx, err error) {
	span := trace.SpanFromContext(c.UserContext())
	span.RecordError(err)
}
