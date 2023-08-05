package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tedyst/licenta/api/v1/generated"
)

func (*ServerHandler) GetUsers(*fiber.Ctx, generated.GetUsersParams) error {
	return nil
}

func (*ServerHandler) PostUsers(c *fiber.Ctx) error {
	return nil
}

func (*ServerHandler) GetUsersMe(c *fiber.Ctx) error {
	return nil
}
