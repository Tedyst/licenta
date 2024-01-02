package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/volatiletech/authboss/v3"
)

type authbossErrorHandler struct {
	LogWriter authboss.Logger
}

// Wrap an http handler with an error
func (e *authbossErrorHandler) Wrap(handler func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return &innerAuthbossErrorHandler{
		Handler:   handler,
		LogWriter: e.LogWriter,
	}
}

type innerAuthbossErrorHandler struct {
	Handler   func(w http.ResponseWriter, r *http.Request) error
	LogWriter authboss.Logger
}

func (e *innerAuthbossErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := e.Handler(w, r)
	if err == nil {
		return
	}

	if errors.Is(err, authboss.ErrUserNotFound) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
		return
	} else if errors.Is(err, &json.UnmarshalTypeError{}) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
		return
	}

	e.LogWriter.Error(fmt.Sprintf("request error from (%s) %s: %+v", r.RemoteAddr, r.URL.String(), err))
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Server Error"))
}
