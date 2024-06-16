package nats

import (
	"context"
	"fmt"
	"log/slog"
	sync "sync"

	"github.com/nats-io/nats.go"
	"github.com/tedyst/licenta/nvd"
	"github.com/tedyst/licenta/tasks"
	"github.com/tedyst/licenta/tasks/local"
	"golang.org/x/sync/semaphore"
)

type nvdScannerTaskSender struct {
	conn *nats.Conn
}

func NewNVDScannerTaskSender(conn *nats.Conn) *nvdScannerTaskSender {
	return &nvdScannerTaskSender{
		conn: conn,
	}
}

const updateNVDVulnerabilitiesForProductQueue = "update-nvd-vulnerabilities-for-product"

func (ns *nvdScannerTaskSender) UpdateNVDVulnerabilitiesForProduct(ctx context.Context, product nvd.Product) error {
	return publishMessage(ctx, ns.conn, updateNVDVulnerabilitiesForProductQueue, &UpdateNVDVulnerabilitiesForProductMessage{
		ProductId: int64(product),
	}, 0)
}

type nvdScannerTaskRunner struct {
	conn        *nats.Conn
	localRunner tasks.VulnerabilityTasksRunner
	semaphore   *semaphore.Weighted
	querier     local.NvdQuerier
}

func NewNVDScannerTaskRunner(conn *nats.Conn, localRunner tasks.VulnerabilityTasksRunner, querier local.NvdQuerier, concurrency int) *nvdScannerTaskRunner {
	return &nvdScannerTaskRunner{
		conn:        conn,
		localRunner: localRunner,
		semaphore:   semaphore.NewWeighted(int64(concurrency)),
		querier:     querier,
	}
}

func (nr *nvdScannerTaskRunner) Run(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := receiveMessage(ctx, nr.conn, nr.semaphore, updateNVDVulnerabilitiesForProductQueue, func(ctx context.Context, message *UpdateNVDVulnerabilitiesForProductMessage) error {
			err := nr.localRunner.UpdateNVDVulnerabilitiesForProduct(ctx, nvd.Product(message.ProductId))
			if err != nil {
				return fmt.Errorf("failed to update NVD vulnerabilities for product %d: %w", message.ProductId, err)
			}

			return nil
		})

		if err != nil {
			slog.Error("failed to receive message", "error", err)
		}
	}()

	return nil
}
