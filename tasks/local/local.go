package local

import (
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/email"
	"github.com/tedyst/licenta/tasks"
)

type localRunner struct {
	scannerRunner
	nvdRunner
	gitRunner
	emailRunner
	dockerRunner
}

func NewLocalRunner(debug bool, emailSender email.EmailSender, queries db.TransactionQuerier) *localRunner {
	return &localRunner{
		scannerRunner: scannerRunner{
			queries: queries,
		},
		nvdRunner: nvdRunner{
			queries: queries,
		},
		gitRunner: gitRunner{
			queries: queries,
		},
		emailRunner: emailRunner{
			queries:     queries,
			emailSender: emailSender,
		},
		dockerRunner: dockerRunner{
			queries: queries,
		},
	}
}

var _ tasks.TaskRunner = (*localRunner)(nil)
