package handlers

import (
	"github.com/tedyst/licenta/api/v1/middleware/session"
	"github.com/tedyst/licenta/db"
)

type serverHandler struct {
	Queries      db.TransactionQuerier
	SessionStore session.SessionStore
}

func NewServerHandler(queries db.TransactionQuerier, sessionStore session.SessionStore) *serverHandler {
	return &serverHandler{
		Queries:      queries,
		SessionStore: sessionStore,
	}
}
