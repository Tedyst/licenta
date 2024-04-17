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

type dockerScannerTaskSender struct {
	conn *nats.Conn
}

func NewDockerScannerTaskSender(conn *nats.Conn) *dockerScannerTaskSender {
	return &dockerScannerTaskSender{
		conn: conn,
	}
}

const scanDockerRepositoryQueue = "scan-Docker-repository"

func (gs *dockerScannerTaskSender) ScanDockerRepository(ctx context.Context, img *queries.ProjectDockerImage) error {
	return publishMessage(ctx, gs.conn, scanDockerRepositoryQueue, &ScanDockerRepoMessage{
		ImageId: img.ID,
	}, 0)
}

type dockerScannerTaskRunner struct {
	conn        *nats.Conn
	localRunner tasks.DockerTasksRunner
	semaphore   *semaphore.Weighted
	querier     DockerQuerier
}

type DockerQuerier interface {
	local.DockerQuerier
	GetDockerImage(ctx context.Context, id int64) (*queries.ProjectDockerImage, error)
}

func NewDockerScannerTaskRunner(conn *nats.Conn, localRunner tasks.DockerTasksRunner, querier DockerQuerier, concurrency int) *dockerScannerTaskRunner {
	return &dockerScannerTaskRunner{
		conn:        conn,
		localRunner: localRunner,
		semaphore:   semaphore.NewWeighted(int64(concurrency)),
		querier:     querier,
	}
}

func (gs *dockerScannerTaskRunner) Run(ctx context.Context, wg *sync.WaitGroup) error {
	go func() {
		defer wg.Done()

		err := receiveMessage(ctx, gs.conn, gs.semaphore, scanDockerRepositoryQueue, func(ctx context.Context, message *ScanDockerRepoMessage) error {
			repo, err := gs.querier.GetDockerImage(ctx, message.ImageId)
			if err != nil {
				return nil
			}

			err = gs.localRunner.ScanDockerRepository(ctx, repo)
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
