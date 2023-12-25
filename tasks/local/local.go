package local

import (
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/email"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/tasks"
)

type localRunner struct {
	scannerRunner
	nvdRunner
	gitRunner
	emailRunner
	dockerRunner

	queries db.TransactionQuerier
}

func NewLocalRunner(debug bool, emailSender email.EmailSender, queries db.TransactionQuerier, exchange messages.Exchange, bruteforceProvider bruteforce.BruteforceProvider) *localRunner {
	return &localRunner{
		scannerRunner: *NewScannerRunner(queries, bruteforceProvider, exchange),
		nvdRunner:     *NewNVDRunner(queries),
		gitRunner:     *NewGitRunner(queries),
		emailRunner:   *NewEmailRunner(queries, emailSender),
		dockerRunner:  *NewDockerRunner(queries),
		queries:       queries,
	}
}

var _ tasks.TaskRunner = (*localRunner)(nil)
