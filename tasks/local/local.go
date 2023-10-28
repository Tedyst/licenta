package local

import (
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/email"
	"github.com/tedyst/licenta/tasks"
)

type localRunner struct {
	emailSender email.EmailSender
	queries     db.TransactionQuerier
}

func NewLocalRunner(debug bool, emailSender email.EmailSender, queries db.TransactionQuerier) *localRunner {
	return &localRunner{
		emailSender: emailSender,
		queries:     queries,
	}
}

var _ tasks.TaskRunner = (*localRunner)(nil)
