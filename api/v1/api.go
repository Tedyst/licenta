package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/api/v1/handlers"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/tasks"
)

type ApiV1Config struct {
	Debug   bool
	BaseURL string
}

type workerAuth interface {
	Handler(next http.Handler) http.Handler
	GetWorker(ctx context.Context) *models.Worker
}

type sessionStore interface {
	GetUser(ctx context.Context) *models.User
	SetUser(ctx context.Context, user *models.User)
	ClearSession(ctx context.Context)
	Handler(next http.Handler) http.Handler
}

func RegisterHandler(app *chi.Mux, database db.TransactionQuerier, sessionStore sessionStore, config ApiV1Config, messageExchange messages.Exchange, taskRunner tasks.TaskRunner, workerAuth workerAuth) http.Handler {
	api := generated.NewStrictHandlerWithOptions(handlers.NewServerHandler(database, sessionStore, messageExchange, taskRunner), nil, generated.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			var message = "Invalid request"
			if config.Debug {
				message = err.Error()
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			data, err := json.Marshal(generated.Error{
				Success: false,
				Message: message,
			})
			if err != nil {
				panic(err)
			}
			_, err = w.Write(data)
			if err != nil {
				panic(err)
			}
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			if err == nil {
				return
			}
			var message = "Internal server error"
			if config.Debug {
				message = err.Error()
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)

			data, err := json.Marshal(generated.Error{
				Success: false,
				Message: message,
			})
			if err != nil {
				panic(err)
			}
			_, err = w.Write(data)
			if err != nil {
				panic(err)
			}
		},
	})
	return generated.HandlerFromMuxWithBaseURL(api, app, config.BaseURL)
}
