package auth

import (
	"context"

	"github.com/volatiletech/authboss/v3"
)

type emailTaskRunner interface {
	SendResetEmail(ctx context.Context, address string, subject string, html string, text string) error
}

type authbossMailer struct {
	runner emailTaskRunner
}

func (a *authbossMailer) Send(ctx context.Context, email authboss.Email) error {
	return nil
}
