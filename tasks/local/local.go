package local

import "github.com/tedyst/licenta/email"

type localRunner struct {
	emailSender email.EmailSender
}

func NewLocalRunner(debug bool, emailSender email.EmailSender) *localRunner {
	return &localRunner{
		emailSender: emailSender,
	}
}
