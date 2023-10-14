package email

import (
	email_lib "github.com/jordan-wright/email"
)

type consoleEmailSender struct {
	senderName string
	sender     string
}

func NewConsoleEmailSender(senderName, sender string) EmailSender {
	return &consoleEmailSender{
		senderName: senderName,
		sender:     sender,
	}
}

func (s *consoleEmailSender) SendMultipartEmail(address string, subject string, html string, text string) error {
	e := email_lib.NewEmail()
	e.From = s.senderName + " <" + s.sender + ">"
	e.To = []string{address}
	e.Subject = subject
	e.Text = []byte(text)
	e.HTML = []byte(html)

	b, err := e.Bytes()
	if err != nil {
		return err
	}
	println(string(b))
	return nil
}
