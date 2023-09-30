package handlers

import (
	"context"
	"net/http"

	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/models"
)

type SessionStoreType interface {
	GetUser(ctx context.Context) *models.User
	SetUser(ctx context.Context, user *models.User)
	GetWaiting2FA(ctx context.Context) *models.User
	SetWaiting2FA(ctx context.Context, waitingUser *models.User)
	GetTOTPKey(ctx context.Context) (string, error)
	SetTOTPKey(ctx context.Context, key string)
	ClearSession(ctx context.Context)

	Handler(next http.Handler) http.Handler
}

type serverHandler struct {
	Queries      db.TransactionQuerier
	SessionStore SessionStoreType
}

func NewServerHandler(queries db.TransactionQuerier, sessionStore SessionStoreType) *serverHandler {
	return &serverHandler{
		Queries:      queries,
		SessionStore: sessionStore,
	}
}
