package local

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tedyst/licenta/models"
)

func (r *localRunner) SendResetEmail(ctx context.Context, address string, subject string, html string, text string) error {
	mailsSent.Add(ctx, 1)

	err := r.emailSender.SendMultipartEmail(ctx, address, subject, html, text)
	if err != nil {
		return errors.Wrap(err, "failed to send email")
	}

	return nil
}

func (r *localRunner) SendCVEVulnerabilityEmail(ctx context.Context, project *models.Project) error {
	return nil
}
