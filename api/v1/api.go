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

	TaskRunner      tasks.TaskRunner
	MessageExchange messages.Exchange

	WorkerAuth workerAuth
	UserAuth   userAuth

	DatabaseProvider db.TransactionQuerier
}

type workerAuth interface {
	GetWorker(ctx context.Context) (*models.Worker, error)
}

type userAuth interface {
	GetUser(ctx context.Context) (*models.User, error)
	VerifyPassword(ctx context.Context, user *models.User, password string) (bool, error)
	UpdatePassword(ctx context.Context, user *models.User, newPassword string) error
}

func RegisterHandler(app chi.Router, config ApiV1Config) http.Handler {
	serverHandler := handlers.NewServerHandler(handlers.HandlerConfig{
		DatabaseProvider: config.DatabaseProvider,
		TaskRunner:       config.TaskRunner,
		MessageExchange:  config.MessageExchange,
		WorkerAuth:       config.WorkerAuth,
		UserAuth:         config.UserAuth,
	})
	api := generated.NewStrictHandlerWithOptions(serverHandler, nil, generated.StrictHTTPServerOptions{
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
