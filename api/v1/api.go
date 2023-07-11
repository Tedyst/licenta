package v1

import (
	"github.com/gofiber/fiber/v2"
)

func InitV1Router(router fiber.Router) error {
	router.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!\n" + c.Path())
	})
	router.Get("/users", HandleGetUsers)
	router.Post("/users", HandleCreateUser)
	return nil
}
