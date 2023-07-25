package email

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/tedyst/licenta/config"
)

var client *sendgrid.Client

func initSendGridClient() {
	if client == nil {
		println("Using postmark")
		if config.SendgridAPIKey == "" {
			panic("SendgridAPIKey is empty")
		}
		client = sendgrid.NewSendClient(config.SendgridAPIKey)
	}
}

func SendMultipartEmailSendgrid(address string, subject string, html string, text string) error {
	initSendGridClient()

	from := mail.NewEmail(config.EmailSenderName, config.EmailSender)
	to := mail.NewEmail("", address)
	message := mail.NewSingleEmail(from, subject, to, text, html)
	response, err := client.Send(message)
	println(response.StatusCode)
	return err
}
