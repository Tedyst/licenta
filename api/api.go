package api

import (
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
	"github.com/tedyst/licenta/api/v1/middleware/session"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/tasks"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type ApiConfig struct {
	Debug  bool
	Origin string

	TaskRunner tasks.TaskRunner
}

func Initialize(database db.TransactionQuerier, sessionStore session.SessionStore, config ApiConfig, messageExchange messages.Exchange) http.Handler {
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
		app.Use(middleware.Timeout(10 * time.Second))
	}
	app.Use(sessionStore.Handler)

	app.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := w.Write([]byte("Method not allowed"))
		if err != nil {
			slog.Error("Error writing response", "error", err)
		}
	})

	v1.RegisterHandler(app, database, sessionStore, v1.ApiV1Config{
		Debug:   config.Debug,
		BaseURL: "/api/v1",
	}, messageExchange, config.TaskRunner)
	return otelhttp.NewHandler(app, "api")
}
