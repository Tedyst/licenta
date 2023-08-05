package errorhandler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/config"
	"go.opentelemetry.io/otel/trace"
)

const InternalServerError = "Internal server error"

func New() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span := trace.SpanFromContext(c.UserContext())
		err := c.Next()
		if err != nil {
			span.RecordError(err)
			var fiberError *fiber.Error
			if errors.As(err, &fiberError) {
				return c.Status(fiberError.Code).JSON(generated.Error{
					Success: false,
					Message: fiberError.Message,
				})
			}
			message := InternalServerError
			if config.Debug {
				message = err.Error()
			}
			return c.Status(fiber.StatusInternalServerError).JSON(generated.Error{
				Success: false,
				Message: message,
			})
		}
		return err
	}
}
