package email

type EmailSender interface {
	SendMultipartEmail(address string, subject string, html string, text string) error
}
