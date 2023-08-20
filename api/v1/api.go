package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/api/v1/handlers"
	db "github.com/tedyst/licenta/db/generated"
)

type ApiV1Config struct {
	Debug   bool
	BaseURL string
}

func RegisterHandler(app *chi.Mux, database *db.Queries, sessionStore handlers.SessionStoreType, config ApiV1Config) http.Handler {
	api := generated.NewStrictHandlerWithOptions(handlers.NewServerHandler(database, sessionStore), nil, generated.StrictHTTPServerOptions{
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
