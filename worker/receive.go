package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"errors"

	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	localexchange "github.com/tedyst/licenta/messages/local"
	"github.com/tedyst/licenta/tasks/local"
)

func ReceiveTasks(ctx context.Context, client generated.ClientWithResponsesInterface) error {
	slog.Info("Starting to receive tasks", "client", client)

	for {
		newCtx, cancel := context.WithTimeout(ctx, 15*time.Second)

		task, err := client.GetWorkerGetTaskWithResponse(newCtx)
		if err != nil && err != context.DeadlineExceeded {
			cancel()
			return fmt.Errorf("error getting task: %w", err)
		}

		cancel()

		switch task.StatusCode() {
		case http.StatusOK:
			slog.Info("Received task", "task", string(task.Body))

			scan := queries.Scan{
				ID:     int64(task.JSON200.Scan.Id),
				Status: int32(task.JSON200.Scan.Status),
				Error:  sql.NullString{String: task.JSON200.Scan.Error, Valid: task.JSON200.Scan.Error != ""},
			}
			scanGroup := queries.ScanGroup{
				ID:        int64(task.JSON200.ScanGroup.Id),
				ProjectID: int64(task.JSON200.ScanGroup.ProjectId),
				CreatedBy: sql.NullInt64{Int64: int64(task.JSON200.ScanGroup.CreatedBy.Id), Valid: task.JSON200.ScanGroup.CreatedBy != nil},
			}

			var postgresScan *queries.PostgresScan
			if task.JSON200.Scan.PostgresScan != nil {
				postgresScan = &queries.PostgresScan{
					ID:         int64(task.JSON200.Scan.PostgresScan.Id),
					DatabaseID: int64(task.JSON200.Scan.PostgresScan.DatabaseId),
					ScanID:     int64(task.JSON200.Scan.Id),
				}
			}

			slog.DebugContext(ctx, "Got task from remote server", "scan", scan, "postgres_scan", postgresScan)

			localExchange := localexchange.NewLocalExchange()

			database := &remoteQuerier{
				client:       client,
				scan:         &scan,
				postgresScan: postgresScan,
				scanGroup:    &scanGroup,
			}
			passProvider := bruteforce.NewDatabaseBruteforceProvider(database)

			runner := local.NewAllScannerRunner(database, localExchange, passProvider)

			err := runner.RunAllScanners(ctx, &scan, true)
			if err != nil {
				return fmt.Errorf("error running task: %w", err)
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
