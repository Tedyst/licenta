package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/api/v1/handlers"
	"github.com/tedyst/licenta/api/v1/middleware/session"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/messages"
)

type ApiV1Config struct {
	Debug   bool
	BaseURL string
}

func RegisterHandler(app *chi.Mux, database db.TransactionQuerier, sessionStore session.SessionStore, config ApiV1Config, messageExchange messages.Exchange) http.Handler {
	api := generated.NewStrictHandlerWithOptions(handlers.NewServerHandler(database, sessionStore, messageExchange), nil, generated.StrictHTTPServerOptions{
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
			w.Write(data)
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
			w.Write(data)
		},
	})
	return generated.HandlerFromMuxWithBaseURL(api, app, config.BaseURL)
}
