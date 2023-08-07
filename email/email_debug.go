package email

import (
	email_lib "github.com/jordan-wright/email"
	"github.com/spf13/viper"
)

func SendMultipartEmailDebug(address string, subject string, html string, text string) error {
	mail := email_lib.NewEmail()
	mail.Subject = subject
	mail.From = viper.GetString("email.senderName") + " <" + viper.GetString("email.sender") + ">"
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
