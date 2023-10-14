package local

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"golang.org/x/exp/slog"
)

func (r *localRunner) SendResetEmail(ctx context.Context, address string, subject string, html string, text string) {
	_, span := tracer.Start(ctx, "SendResetEmail")
	defer span.End()

	mailsSent.Add(ctx, 1)

	err := r.emailSender.SendMultipartEmail(ctx, address, subject, html, text)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Error sending email: %v", err)
		return
	}
}
