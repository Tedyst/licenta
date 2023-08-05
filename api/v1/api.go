package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/api/v1/handlers"
	"github.com/tedyst/licenta/api/v1/middleware/errorhandler"
)

func GetServerHandler() generated.ServerInterface {
	return &handlers.ServerHandler{}
}

func RegisterHandlers(router fiber.Router) {
	router.Use(errorhandler.New())
	generated.RegisterHandlers(router, GetServerHandler())
}
