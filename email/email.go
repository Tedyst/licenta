package email

import "context"

type EmailSender interface {
	SendMultipartEmail(ctx context.Context, address string, subject string, html string, text string) error
}
