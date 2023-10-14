package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	v1 "github.com/tedyst/licenta/api/v1"
	"github.com/tedyst/licenta/api/v1/middleware/options"
	requestid "github.com/tedyst/licenta/api/v1/middleware/requestID"
	"github.com/tedyst/licenta/api/v1/middleware/session"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/tasks"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type ApiConfig struct {
	Debug  bool
	Origin string

	TaskRunner tasks.TaskRunner
}

func Initialize(database db.TransactionQuerier, sessionStore session.SessionStore, config ApiConfig) http.Handler {
	app := chi.NewRouter()
	app.Use(middleware.RealIP)
	app.Use(middleware.Logger)
	app.Use(middleware.Recoverer)
	app.Use(middleware.CleanPath)
	app.Use(middleware.GetHead)
	app.Use(options.HandleOptions(config.Origin))
	app.Use(requestid.RequestIDMiddleware)
	if !config.Debug {
		app.Use(middleware.Timeout(10 * time.Second))
	}
	app.Use(sessionStore.Handler)

	app.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
	})

	v1.RegisterHandler(app, database, sessionStore, v1.ApiV1Config{
		Debug:   config.Debug,
		BaseURL: "/api/v1",
	})
	return otelhttp.NewHandler(app, "api")
}
