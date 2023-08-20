package handlers

import (
	"context"
	"net/http"

	db "github.com/tedyst/licenta/db/generated"
)

type SessionStoreType interface {
	GetUser(ctx context.Context) *db.User
	SetUser(ctx context.Context, user *db.User)
	GetWaiting2FA(ctx context.Context) *db.User
	SetWaiting2FA(ctx context.Context, waitingUser *db.User)
	GetTOTPKey(ctx context.Context) (string, error)
	SetTOTPKey(ctx context.Context, key string)
	ClearSession(ctx context.Context)

	Handler(next http.Handler) http.Handler
}

type serverHandler struct {
	Queries      *db.Queries
	SessionStore SessionStoreType
}

func NewServerHandler(queries *db.Queries, sessionStore SessionStoreType) *serverHandler {
	return &serverHandler{
		Queries:      queries,
		SessionStore: sessionStore,
	}
}
