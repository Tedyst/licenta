package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tedyst/licenta/api/v1/generated"
	"go.opentelemetry.io/otel/trace"
)

func sendError(c *fiber.Ctx, code int, message string) error {
	span := trace.SpanFromContext(c.UserContext())
	span.AddEvent(message)
	return c.Status(code).JSON(generated.Error{
		Success: false,
		Message: message,
	})
}

func traceError(c *fiber.Ctx, err error) {
	span := trace.SpanFromContext(c.UserContext())
	span.RecordError(err)
}
