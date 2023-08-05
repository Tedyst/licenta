package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tedyst/licenta/api/v1/generated"
)

func sendError(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(generated.Error{
		Code:    int32(code),
		Message: message,
	})
}
