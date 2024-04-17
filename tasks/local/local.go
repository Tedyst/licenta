package local

import (
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/email"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/tasks"
)

type localRunner struct {
	SaverRunner
	NvdRunner
	GitRunner
	emailRunner
	DockerRunner

	queries db.TransactionQuerier
}

func NewLocalRunner(debug bool, emailSender email.EmailSender, queries db.TransactionQuerier, exchange messages.Exchange, bruteforceProvider bruteforce.BruteforceProvider) *localRunner {
	return &localRunner{
		NvdRunner:    *NewNVDRunner(queries),
		GitRunner:    *NewGitRunner(queries),
		emailRunner:  *NewEmailRunner(emailSender),
		DockerRunner: *NewDockerRunner(queries),
		queries:      queries,
		SaverRunner:  *NewSaverRunner(queries, exchange, bruteforceProvider),
	}
}

var _ tasks.TaskRunner = (*localRunner)(nil)
