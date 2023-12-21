package handlers

import (
	"github.com/tedyst/licenta/api/v1/middleware/session"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/messages"
)

type serverHandler struct {
	Queries         db.TransactionQuerier
	SessionStore    session.SessionStore
	MessageExchange messages.Exchange
}

func NewServerHandler(queries db.TransactionQuerier, sessionStore session.SessionStore, messageExchange messages.Exchange) *serverHandler {
	return &serverHandler{
		Queries:         queries,
		SessionStore:    sessionStore,
		MessageExchange: messageExchange,
	}
}
