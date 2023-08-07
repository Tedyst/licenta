package api

import (
	"runtime"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/spf13/viper"
	v1 "github.com/tedyst/licenta/api/v1"
	"github.com/tedyst/licenta/telemetry"
	"go.opentelemetry.io/otel/trace"
)

func InitializeFiber() *fiber.App {
	app := fiber.New()

	telemetry.RegisterPrometheus(app)

	app.Use(otelfiber.Middleware())

	app.Use(recover.New())
	app.Use(logger.New())
	if viper.GetBool("debug") {
		app.Use(pprof.New())
		runtime.SetMutexProfileFraction(5)
		runtime.SetBlockProfileRate(5)
	}

	app.Use(func(c *fiber.Ctx) error {
		span := trace.SpanFromContext(c.UserContext())
		c.Response().Header.Set("X-Trace-Id", span.SpanContext().TraceID().String())
		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	api_v1 := app.Group("/api/v1")
	v1.RegisterHandlers(api_v1)

	return app
}
