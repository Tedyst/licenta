package tasks

import (
	"context"
)

type TaskRunner interface {
	EmailTasksRunner
}

type EmailTasksRunner interface {
	SendResetEmail(ctx context.Context, address string, subject string, html string, text string)
}
