package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	v1 "github.com/tedyst/licenta/api/v1"
	"github.com/tedyst/licenta/db"
)

func run() error {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}

	queries := db.New(pool)

	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", pool)
		c.Locals("queries", queries)
		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	api_v1 := app.Group("/api/v1")
	v1.InitV1Router(api_v1)

	app.Listen(":5000")
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
