package local

import (
	"context"
	"fmt"

	"github.com/tedyst/licenta/db/queries"
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

func (r *emailRunner) SendResetEmail(ctx context.Context, address string, subject string, html string, text string) error {
	mailsSent.Add(ctx, 1)

	err := r.emailSender.SendMultipartEmail(ctx, address, subject, html, text)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (r *emailRunner) SendCVEVulnerabilityEmail(ctx context.Context, project *queries.Project) error {
	return nil
}

func (r *emailRunner) SendCVEMailsToAllProjectMembers(ctx context.Context, projectID int64) error {
	return nil
}

func (r *emailRunner) SendCVEMailsToAllProjects(ctx context.Context) error {
	return nil
}
