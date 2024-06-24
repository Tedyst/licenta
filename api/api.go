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
	"github.com/riandyrn/otelchi"
	slogchi "github.com/samber/slog-chi"
	v1 "github.com/tedyst/licenta/api/v1"
	buildid "github.com/tedyst/licenta/api/v1/middleware/buildID"
	"github.com/tedyst/licenta/api/v1/middleware/cache"
	"github.com/tedyst/licenta/api/v1/middleware/options"
	requestid "github.com/tedyst/licenta/api/v1/middleware/requestID"
	"github.com/tedyst/licenta/db/queries"
)

type ApiConfig struct {
	v1.ApiV1Config

	Origin string
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
	app.Use(otelchi.Middleware("api", otelchi.WithRequestMethodInSpanName(true)))
	app.Use(buildid.BuildIDMiddleware)
	app.Use(middleware.RealIP)
	app.Use(slogchi.NewWithConfig(slog.Default(), slogchi.Config{
		WithUserAgent: true,
		WithRequestID: false,
		WithSpanID:    true,
		WithTraceID:   true,
	}))
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
	if !config.Debug {
		app.Use(httprate.LimitByIP(limitRatePerMinute, 1*time.Minute))
	}

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
		r.Use(config.UserAuth.APIMiddleware)
	})

	v1.RegisterHandler(apiRouter, config.ApiV1Config)
	return app, nil
}
