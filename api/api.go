package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-http-utils/etag"
	slogchi "github.com/samber/slog-chi"
	v1 "github.com/tedyst/licenta/api/v1"
	"github.com/tedyst/licenta/api/v1/middleware/cache"
	"github.com/tedyst/licenta/api/v1/middleware/options"
	requestid "github.com/tedyst/licenta/api/v1/middleware/requestID"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/tasks"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type ApiConfig struct {
	Debug  bool
	Origin string

	TaskRunner      tasks.TaskRunner
	MessageExchange messages.Exchange

	WorkerAuth workerAuth
	UserAuth   userAuth

	Database db.TransactionQuerier
}

type workerAuth interface {
	Handler(next http.Handler) http.Handler
	GetWorker(ctx context.Context) (*models.Worker, error)
}

type userAuth interface {
	Middleware(next http.Handler) http.Handler
	Handler() http.Handler
	GetUser(ctx context.Context) (*models.User, error)
}

func Initialize(config ApiConfig) (http.Handler, error) {
	app := chi.NewRouter()
	app.Use(middleware.RealIP)
	app.Use(slogchi.New(slog.Default()))
	app.Use(middleware.Recoverer)
	app.Use(cache.CacheControlHeaderMiddleware)
	app.Use(func(h http.Handler) http.Handler {
		return etag.Handler(h, false)
	})
	app.Use(middleware.CleanPath)
	app.Use(middleware.GetHead)
	app.Use(options.HandleOptions(config.Origin))
	app.Use(requestid.RequestIDMiddleware)
	if !config.Debug {
		app.Use(middleware.Timeout(30 * time.Second))
	}
	app.Use(config.WorkerAuth.Handler)

	app.Use(config.UserAuth.Middleware)
	app.Mount("/auth", http.StripPrefix("/auth", config.UserAuth.Handler()))

	app.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := w.Write([]byte("Method not allowed"))
		if err != nil {
			slog.Error("Error writing response", "error", err)
		}
	})

	v1.RegisterHandler(app, v1.ApiV1Config{
		Debug:            config.Debug,
		BaseURL:          "/api/v1",
		TaskRunner:       config.TaskRunner,
		MessageExchange:  config.MessageExchange,
		WorkerAuth:       config.WorkerAuth,
		UserAuth:         config.UserAuth,
		DatabaseProvider: config.Database,
	})
	return otelhttp.NewHandler(app, "api"), nil
}
