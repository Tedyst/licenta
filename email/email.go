package email

import "github.com/tedyst/licenta/config"

func SendMultipartEmail(address string, subject string, html string, text string) error {
	if config.Debug {
		return SendMultipartEmailDebug(address, subject, html, text)
	}
	return SendMultipartEmailSendgrid(address, subject, html, text)
}
