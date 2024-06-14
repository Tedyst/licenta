package nats

import (
	"context"
	"log/slog"
	sync "sync"

	"github.com/nats-io/nats.go"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/tasks"
	"github.com/tedyst/licenta/tasks/local"
	"golang.org/x/sync/semaphore"
)

type gitScannerTaskSender struct {
	conn *nats.Conn
}

func NewGitScannerTaskSender(conn *nats.Conn) *gitScannerTaskSender {
	return &gitScannerTaskSender{
		conn: conn,
	}
}

const scanGitRepositoryQueue = "scan-git-repository"

func (gs *gitScannerTaskSender) ScanGitRepository(ctx context.Context, repo *queries.GitRepository, scan *queries.Scan) error {
	return publishMessage(ctx, gs.conn, scanGitRepositoryQueue, &ScanGitRepositoryMessage{
		RepoId: repo.ID,
		ScanId: scan.ID,
	}, 0)
}

type gitScannerTaskRunner struct {
	conn        *nats.Conn
	localRunner tasks.GitTasksRunner
	semaphore   *semaphore.Weighted
	querier     GitQuerier
	saltKey     string
}

type GitQuerier interface {
	local.GitQuerier
	GetGitRepository(context.Context, queries.GetGitRepositoryParams) (*queries.GetGitRepositoryRow, error)
	GetScan(ctx context.Context, id int64) (*queries.GetScanRow, error)
}

func NewGitScannerTaskRunner(conn *nats.Conn, localRunner tasks.GitTasksRunner, querier GitQuerier, concurrency int, saltKey string) *gitScannerTaskRunner {
	return &gitScannerTaskRunner{
		conn:        conn,
		localRunner: localRunner,
		semaphore:   semaphore.NewWeighted(int64(concurrency)),
		querier:     querier,
		saltKey:     saltKey,
	}
}

func (gs *gitScannerTaskRunner) Run(ctx context.Context, wg *sync.WaitGroup) error {
	go func() {
		defer wg.Done()

		err := receiveMessage(ctx, gs.conn, gs.semaphore, scanGitRepositoryQueue, func(ctx context.Context, message *ScanGitRepositoryMessage) error {
			repo, err := gs.querier.GetGitRepository(ctx, queries.GetGitRepositoryParams{
				ID:      message.RepoId,
				SaltKey: gs.saltKey,
			})
			if err != nil {
				return nil
			}

			scan, err := gs.querier.GetScan(ctx, message.ScanId)
			if err != nil {
				return nil
			}

			err = gs.localRunner.ScanGitRepository(ctx, &queries.GitRepository{
				ID:            repo.ID,
				ProjectID:     repo.ProjectID,
				GitRepository: repo.GitRepository,
				Username:      repo.Username,
				Password:      repo.Password,
				PrivateKey:    repo.PrivateKey,
			}, &scan.Scan)
			if err != nil {
				return nil
			}

			return nil
		})

		if err != nil {
			slog.Error("failed to receive message", "error", err)
		}
	}()

	return nil
}
