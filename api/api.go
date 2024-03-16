package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-http-utils/etag"
	"github.com/justinas/nosurf"
	slogchi "github.com/samber/slog-chi"
	"github.com/tedyst/licenta/api/authorization"
	v1 "github.com/tedyst/licenta/api/v1"
	"github.com/tedyst/licenta/api/v1/middleware/cache"
	"github.com/tedyst/licenta/api/v1/middleware/options"
	requestid "github.com/tedyst/licenta/api/v1/middleware/requestID"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/messages"
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

	AuthorizationManager authorization.AuthorizationManager

	Database db.TransactionQuerier
}

type workerAuth interface {
	Handler(next http.Handler) http.Handler
	GetWorker(ctx context.Context) (*queries.Worker, error)
}

type userAuth interface {
	Middleware(next http.Handler) http.Handler
	APIMiddleware(next http.Handler) http.Handler
	Handler() http.Handler
	GetUser(ctx context.Context) (*queries.User, error)
	VerifyPassword(ctx context.Context, user *queries.User, password string) (bool, error)
	UpdatePassword(ctx context.Context, user *queries.User, newPassword string) error
}

const limitRatePerMinute = 100
const timeout = 30 * time.Second

func Initialize(config ApiConfig) (http.Handler, error) {
	app := chi.NewRouter()
	app.Use(middleware.RealIP)
	app.Use(slogchi.New(slog.Default()))
	app.Use(middleware.Recoverer)
	app.Use(cache.CacheControlHeaderMiddleware)
	app.Use(func(h http.Handler) http.Handler {
		return etag.Handler(h, false)
	})
	app.Use(nosurf.NewPure)
	app.Use(middleware.CleanPath)
	app.Use(middleware.GetHead)
	app.Use(options.HandleOptions(config.Origin))
	app.Use(requestid.RequestIDMiddleware)
	app.Use(httprate.LimitByIP(limitRatePerMinute, 1*time.Minute))

	if !config.Debug {
		app.Use(middleware.Timeout(timeout))
	}
	app.Use(config.WorkerAuth.Handler)

	app.Use(config.UserAuth.Middleware)
	app.Mount("/api/auth", http.StripPrefix("/api/auth", config.UserAuth.Handler()))

	app.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := w.Write([]byte("Method not allowed"))
		if err != nil {
			slog.Error("Error writing response", "error", err)
		}
	})

	apiRouter := app.Route("/api/v1/", func(r chi.Router) {
		// r.Use(app.Middlewares()...)
		r.Use(config.UserAuth.APIMiddleware)
	})

	v1.RegisterHandler(apiRouter, v1.ApiV1Config{
		Debug:                config.Debug,
		BaseURL:              "",
		TaskRunner:           config.TaskRunner,
		MessageExchange:      config.MessageExchange,
		WorkerAuth:           config.WorkerAuth,
		UserAuth:             config.UserAuth,
		DatabaseProvider:     config.Database,
		AuthorizationManager: config.AuthorizationManager,
	})
	return otelhttp.NewHandler(app, "api"), nil
}
