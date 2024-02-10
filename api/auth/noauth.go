package auth

import (
	"context"
	"net/http"

	"github.com/tedyst/licenta/db/queries"
)

type noAuth struct {
	user *queries.User
}

func NewNoAuth(user *queries.User) *noAuth {
	return &noAuth{
		user: user,
	}
}

func (auth *noAuth) Middleware(next http.Handler) http.Handler {
	return next
}

func (auth *noAuth) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})
}

func (auth *noAuth) GetUser(ctx context.Context) (*queries.User, error) {
	return auth.user, nil
}
