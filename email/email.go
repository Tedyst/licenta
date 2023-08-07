package email

import "github.com/spf13/viper"

func SendMultipartEmail(address string, subject string, html string, text string) error {
	if viper.GetBool("debug") {
		return SendMultipartEmailDebug(address, subject, html, text)
	}
	return SendMultipartEmailSendgrid(address, subject, html, text)
}
