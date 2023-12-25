package auth

import (
	"context"
	"log/slog"

	"github.com/volatiletech/authboss/v3"
)

type contextAuthbossLogger struct {
	ctx context.Context
}

func (a *contextAuthbossLogger) Info(msg string) {
	slog.InfoContext(a.ctx, msg)
}

func (a *contextAuthbossLogger) Error(msg string) {
	slog.ErrorContext(a.ctx, msg)
}

type authbossLogger struct {
}

func (*authbossLogger) FromContext(ctx context.Context) authboss.Logger {
	return &contextAuthbossLogger{
		ctx: ctx,
	}
}

func (*authbossLogger) Error(msg string) {
	slog.Error(msg)
}

func (*authbossLogger) Info(msg string) {
	slog.Debug(msg)
}

var _ authboss.Logger = (*contextAuthbossLogger)(nil)
var _ authboss.ContextLogger = (*authbossLogger)(nil)
