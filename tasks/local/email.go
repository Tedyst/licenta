package local

import (
	"context"
	"fmt"

	"github.com/tedyst/licenta/email"
)

type emailRunner struct {
	emailSender email.EmailSender
}

func NewEmailRunner(emailSender email.EmailSender) *emailRunner {
	return &emailRunner{
		emailSender: emailSender,
	}
}

func (r *emailRunner) SendEmail(ctx context.Context, address string, subject string, html string, text string) error {
	mailsSent.Add(ctx, 1)

	err := r.emailSender.SendMultipartEmail(ctx, address, subject, html, text)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
