package email

import "github.com/spf13/viper"

func SendMultipartEmail(address string, subject string, html string, text string) error {
	if viper.IsSet("sendgrid") {
		return SendMultipartEmailSendgrid(address, subject, html, text)
	}
	return SendMultipartEmailDebug(address, subject, html, text)
}
