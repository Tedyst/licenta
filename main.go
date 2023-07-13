package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	postgres_session "github.com/gofiber/storage/postgres/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	v1 "github.com/tedyst/licenta/api/v1"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"

	_ "github.com/tedyst/licenta/docs"
)

//go:generate swag init
//go:generate sqlc generate

// @title			Proiect Licenta
// @version		1.0
// @description	This is a sample swagger for Fiber
// @termsOfService	http://swagger.io/terms/
// @contact.name	API Support
// @contact.email	fiber@swagger.io
// @license.name	Apache 2.0
// @license.url	http://www.apache.org/licenses/LICENSE-2.0.html
// @host			localhost:8080
// @BasePath		/
func run() error {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	queries := db.New(pool)
	sessionStorage := postgres_session.New(postgres_session.Config{
		DB: pool,
	})
	store := session.New(session.Config{
		Storage:   sessionStorage,
		KeyLookup: "cookie:session",
	})
	config.Debug = os.Getenv("DEBUG") == "true"

	app := fiber.New(fiber.Config{
		Prefork: !config.Debug,
	})

	config.DatabasePool = pool
	config.DatabaseQueries = queries
	config.SessionStore = store
	db.PasswordPepper = []byte(os.Getenv("PASSWORD_PEPPER"))

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

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
