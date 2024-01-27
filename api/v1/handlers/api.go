package handlers

import (
	"context"

	"github.com/tedyst/licenta/api/authorization"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/tasks"
)

type serverHandler struct {
	DatabaseProvider db.TransactionQuerier
	TaskRunner       tasks.TaskRunner
	MessageExchange  messages.Exchange

	workerauth    workerAuth
	userAuth      userAuth
	authorization AuthorizationManager
}

type workerAuth interface {
	GetWorker(ctx context.Context) (*models.Worker, error)
}

type userAuth interface {
	GetUser(ctx context.Context) (*models.User, error)
	VerifyPassword(ctx context.Context, user *models.User, password string) (bool, error)
	UpdatePassword(ctx context.Context, user *models.User, newPassword string) error
}

type AuthorizationManager = authorization.AuthorizationManager

type HandlerConfig struct {
	DatabaseProvider db.TransactionQuerier
	TaskRunner       tasks.TaskRunner
	MessageExchange  messages.Exchange

	WorkerAuth           workerAuth
	UserAuth             userAuth
	AuthorizationManager AuthorizationManager
}

func NewServerHandler(config HandlerConfig) *serverHandler {
	return &serverHandler{
		DatabaseProvider: config.DatabaseProvider,
		MessageExchange:  config.MessageExchange,
		TaskRunner:       config.TaskRunner,
		workerauth:       config.WorkerAuth,
		userAuth:         config.UserAuth,
		authorization:    config.AuthorizationManager,
	}
}

var _ generated.StrictServerInterface = (*serverHandler)(nil)
