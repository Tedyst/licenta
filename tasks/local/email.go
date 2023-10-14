package local

import "context"

func (r *localRunner) SendResetEmail(ctx context.Context, address string, subject string, html string, text string) {
	go r.emailSender.SendMultipartEmail(address, subject, html, text)
}
