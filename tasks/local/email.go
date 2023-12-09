package local

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/email"
	"github.com/tedyst/licenta/models"
)

type emailRunner struct {
	queries     db.TransactionQuerier
	emailSender email.EmailSender
}

func NewEmailRunner(queries db.TransactionQuerier, emailSender email.EmailSender) *emailRunner {
	return &emailRunner{
		queries:     queries,
		emailSender: emailSender,
	}
}

func (r *emailRunner) SendResetEmail(ctx context.Context, address string, subject string, html string, text string) error {
	mailsSent.Add(ctx, 1)

	err := r.emailSender.SendMultipartEmail(ctx, address, subject, html, text)
	if err != nil {
		return errors.Wrap(err, "failed to send email")
	}

	return nil
}

func (r *emailRunner) SendCVEVulnerabilityEmail(ctx context.Context, project *models.Project) error {
	return nil
}

func (r *emailRunner) SendCVEMailsToAllProjectMembers(ctx context.Context, projectID int64) error {
	return nil
}

func (r *emailRunner) SendCVEMailsToAllProjects(ctx context.Context) error {
	return nil
}
