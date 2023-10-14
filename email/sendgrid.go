package email

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type sendGridEmailSender struct {
	client *sendgrid.Client

	senderName string
	sender     string
}

func NewSendGridEmailSender(apiKey, senderName, sender string) EmailSender {
	return &sendGridEmailSender{
		client:     sendgrid.NewSendClient(apiKey),
		senderName: senderName,
		sender:     sender,
	}
}

func (s *sendGridEmailSender) SendMultipartEmail(address string, subject string, html string, text string) error {
	from := mail.NewEmail(s.senderName, s.sender)

	to := mail.NewEmail("", address)
	message := mail.NewSingleEmail(from, subject, to, text, html)
	_, err := s.client.Send(message)
	return err
}
