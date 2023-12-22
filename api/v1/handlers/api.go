package handlers

import (
	"github.com/tedyst/licenta/api/v1/middleware/session"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/tasks"
)

type serverHandler struct {
	Queries         db.TransactionQuerier
	SessionStore    session.SessionStore
	TaskRunner      tasks.TaskRunner
	MessageExchange messages.Exchange
}

func NewServerHandler(queries db.TransactionQuerier, sessionStore session.SessionStore, messageExchange messages.Exchange, taskRunner tasks.TaskRunner) *serverHandler {
	return &serverHandler{
		Queries:         queries,
		SessionStore:    sessionStore,
		MessageExchange: messageExchange,
		TaskRunner:      taskRunner,
	}
}
