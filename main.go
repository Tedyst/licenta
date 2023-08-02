package main

import (
	"context"
	"log"
	"os"
	"runtime"

	"github.com/exaring/otelpgx"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	v1 "github.com/tedyst/licenta/api/v1"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"
	"go.opentelemetry.io/otel/trace"

	_ "github.com/tedyst/licenta/docs"
)

//go:generate go run github.com/swaggo/swag/cmd/swag init
//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
//go:generate go run github.com/valyala/quicktemplate/qtc -dir=templates

// @title			Proiect Licenta
// @version		1.0
// @description	This is a sample swagger for Fiber
// @termsOfService	http://swagger.io/terms/
// @contact.name	API Support
// @contact.email	fiber@swagger.io
// @license.name	Apache 2.0
// @license.url	http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath		/
func run() error {
	err := parseConfig()
	if err != nil {
		return err
	}
	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	mp := initMetric()
	defer func() {
		if err := mp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down metric provider: %v", err)
		}
	}()

	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return err
	}

	queries := db.New(pool)

	app := fiber.New()

	registerPrometheus(app)

	app.Use(otelfiber.Middleware())

	config.DatabasePool = pool
	config.DatabaseQueries = queries

	app.Use(recover.New())
	if config.Debug {
		app.Use(logger.New())
		app.Use(pprof.New())
	}
	if config.Debug {
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
	err = v1.InitV1Router(api_v1)
	if err != nil {
		return err
	}

	return app.Listen(":5000")
}

func parseConfig() error {
	if os.Getenv("DATABASE_URL") == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	config.Debug = os.Getenv("DEBUG") == "true"
	config.JaegerEndpoint = os.Getenv("JAEGER_ENDPOINT")
	db.PasswordPepper = []byte(os.Getenv("PASSWORD_PEPPER"))
	config.SendgridAPIKey = os.Getenv("SENDGRID_API_KEY")
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
