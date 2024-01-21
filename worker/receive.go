package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	localexchange "github.com/tedyst/licenta/messages/local"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/tasks/local"
)

func ReceiveTasks(ctx context.Context, client generated.ClientWithResponsesInterface) error {
	slog.Info("Starting to receive tasks", "client", client)

	for {
		newCtx, cancel := context.WithTimeout(ctx, 15*time.Second)

		task, err := client.GetWorkerGetTaskWithResponse(newCtx)
		if err != nil && err != context.DeadlineExceeded {
			cancel()
			return errors.Wrap(err, "error getting task")
		}

		cancel()

		switch task.StatusCode() {
		case http.StatusOK:
			slog.Info("Received task", "task", string(task.Body))

			scan := models.Scan{
				ID:        int64(task.JSON200.Scan.Id),
				ProjectID: int64(task.JSON200.Scan.ProjectId),
				Status:    int32(task.JSON200.Scan.Status),
				Error:     sql.NullString{String: task.JSON200.Scan.Error, Valid: task.JSON200.Scan.Error != ""},
			}
			var postgresScan *models.PostgresScan
			if task.JSON200.Scan.PostgresScan != nil {
				postgresScan = &models.PostgresScan{
					ID:         int64(task.JSON200.Scan.PostgresScan.Id),
					DatabaseID: int64(task.JSON200.Scan.PostgresScan.DatabaseId),
					ScanID:     int64(task.JSON200.Scan.Id),
				}
			}

			slog.DebugContext(ctx, "Got task from remote server", "scan", scan, "postgres_scan", postgresScan)

			localExchange := localexchange.NewLocalExchange()
			runner := local.NewAllScannerRunner(&remoteQuerier{
				client:       client,
				scan:         &scan,
				postgresScan: postgresScan,
			}, localExchange, &remoteBruteforceProvider{
				client: client,
				scan:   &scan,
			})

			err := runner.RunAllScanners(ctx, &scan, true)
			if err != nil {
				return errors.Wrap(err, "error running task")
			}
		case http.StatusAccepted:
			slog.Debug("No task available yet, retrying in 5 seconds...")
		default:
			slog.ErrorContext(ctx, "got invalid response from server", "response", task)
			return errors.New("error receiving task")
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Second):
		}
	}
}
