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
	conn    *nats.Conn
	saltKey string
}

func NewDockerScannerTaskSender(conn *nats.Conn) *dockerScannerTaskSender {
	return &dockerScannerTaskSender{
		conn: conn,
	}
}

const scanDockerRepositoryQueue = "scan-Docker-repository"

func (gs *dockerScannerTaskSender) ScanDockerRepository(ctx context.Context, img *queries.DockerImage, scan *queries.Scan) error {
	return publishMessage(ctx, gs.conn, scanDockerRepositoryQueue, &ScanDockerRepoMessage{
		ImageId: img.ID,
		ScanId:  scan.ID,
	}, 0)
}

type dockerScannerTaskRunner struct {
	conn        *nats.Conn
	localRunner tasks.DockerTasksRunner
	semaphore   *semaphore.Weighted
	querier     DockerQuerier
	saltKey     string
}

type DockerQuerier interface {
	local.DockerQuerier
	GetDockerImage(context.Context, queries.GetDockerImageParams) (*queries.GetDockerImageRow, error)
	GetScan(ctx context.Context, id int64) (*queries.GetScanRow, error)
}

func NewDockerScannerTaskRunner(conn *nats.Conn, localRunner tasks.DockerTasksRunner, querier DockerQuerier, concurrency int, saltKey string) *dockerScannerTaskRunner {
	return &dockerScannerTaskRunner{
		conn:        conn,
		localRunner: localRunner,
		semaphore:   semaphore.NewWeighted(int64(concurrency)),
		querier:     querier,
		saltKey:     saltKey,
	}
}

func (gs *dockerScannerTaskRunner) Run(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := receiveMessage(ctx, gs.conn, gs.semaphore, scanDockerRepositoryQueue, func(ctx context.Context, message *ScanDockerRepoMessage) error {
			repo, err := gs.querier.GetDockerImage(ctx, queries.GetDockerImageParams{
				ID:      message.ImageId,
				SaltKey: gs.saltKey,
			})
			if err != nil {
				return nil
			}

			scan, err := gs.querier.GetScan(ctx, message.ScanId)
			if err != nil {
				return nil
			}

			err = gs.localRunner.ScanDockerRepository(ctx, &queries.DockerImage{
				ID:          repo.ID,
				ProjectID:   repo.ProjectID,
				DockerImage: repo.DockerImage,
				Username:    repo.Username,
				Password:    repo.Password,
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
