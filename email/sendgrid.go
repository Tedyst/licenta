package email

import (
	"context"
	"log/slog"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type sendGridEmailSender struct {
	client *sendgrid.Client

	senderName string
	sender     string
}

func NewSendGridEmailSender(apiKey, senderName, sender string) EmailSender {
	slog.Debug("Using SendGrid email sender")

	return &sendGridEmailSender{
		client:     sendgrid.NewSendClient(apiKey),
		senderName: senderName,
		sender:     sender,
	}
}

func (s *sendGridEmailSender) SendMultipartEmail(ctx context.Context, address string, subject string, html string, text string) error {
	slog.DebugContext(ctx, "Sending email", "address", address, "subject", subject, "html", html, "text", text)

	from := mail.NewEmail(s.senderName, s.sender)

	to := mail.NewEmail("", address)
	message := mail.NewSingleEmail(from, subject, to, text, html)
	resp, err := s.client.Send(message)

	slog.DebugContext(ctx, "Email sent", "response", resp)

	return err
}
