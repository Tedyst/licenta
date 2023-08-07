package email

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/viper"
)

var client *sendgrid.Client

func initSendGridClient() {
	if client == nil {
		client = sendgrid.NewSendClient(viper.GetString("sendgrid"))
	}
}

func SendMultipartEmailSendgrid(address string, subject string, html string, text string) error {
	initSendGridClient()

	from := mail.NewEmail(viper.GetString("email.senderName"), viper.GetString("email.sender"))
	to := mail.NewEmail("", address)
	message := mail.NewSingleEmail(from, subject, to, text, html)
	response, err := client.Send(message)
	return err
}
