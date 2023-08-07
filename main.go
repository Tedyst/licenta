package main

import (
	"context"
	"log"
	"os"

	"github.com/exaring/otelpgx"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/num30/config"
	v1 "github.com/tedyst/licenta/api/v1"
	conf "github.com/tedyst/licenta/config"
	database "github.com/tedyst/licenta/db"
	db "github.com/tedyst/licenta/db/generated"
	"go.opentelemetry.io/otel/trace"
)

func run() error {
	err := config.NewConfReader("").Read(&conf.Config)
	if err != nil {
		panic(err)
	}

	print(conf.Config.OpenTelemetry.UseMetrics)
	print(conf.Config.OpenTelemetry.UseTracing)

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

	database.DatabasePool = pool
	database.DatabaseQueries = queries

	// app.Use(recover.New())
	// if config.Debug {
	// 	app.Use(logger.New())
	// 	app.Use(pprof.New())
	// }
	// if config.Debug {
	// 	runtime.SetMutexProfileFraction(5)
	// 	runtime.SetBlockProfileRate(5)
	// }
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

	return app.Listen(":5000")
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
