package email

import (
	email_lib "github.com/jordan-wright/email"
	"github.com/tedyst/licenta/config"
)

func SendMultipartEmail(address string, html string, text string) error {
	mail := email_lib.NewEmail()
	mail.From = config.EmailSender
	mail.To = []string{address}
	mail.HTML = []byte(html)
	mail.Text = []byte(text)
	bytes, err := mail.Bytes()
	if err != nil {
		return err
	}
	println(string(bytes))
	return err
}
