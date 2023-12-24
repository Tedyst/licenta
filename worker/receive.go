package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/models"
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
			slog.Info("Received task", "task", task.JSON200.Task)
			err = runTask(ctx, client, Task{
				Type: TaskType(task.JSON200.Task.Type),
				PostgresScan: PostgresScan{
					Scan: models.PostgresScan{
						ID:        int64(task.JSON200.Task.PostgresScan.Scan.Id),
						Status:    int32(task.JSON200.Task.PostgresScan.Scan.Status),
						CreatedAt: pgtype.Timestamptz{Time: time.Now()},
						EndedAt:   pgtype.Timestamptz{Time: time.Now()},
						Error:     sql.NullString{String: task.JSON200.Task.PostgresScan.Scan.Error},
					},
					Database: models.PostgresDatabases{
						ID:           int64(task.JSON200.Task.PostgresScan.PostgresDatabase.Id),
						Host:         task.JSON200.Task.PostgresScan.PostgresDatabase.Host,
						Port:         int32(task.JSON200.Task.PostgresScan.PostgresDatabase.Port),
						DatabaseName: task.JSON200.Task.PostgresScan.PostgresDatabase.DatabaseName,
						Password:     task.JSON200.Task.PostgresScan.PostgresDatabase.Password,
						Username:     task.JSON200.Task.PostgresScan.PostgresDatabase.Username,
						Remote:       task.JSON200.Task.PostgresScan.PostgresDatabase.Remote,
						ProjectID:    int64(task.JSON200.Task.PostgresScan.PostgresDatabase.ProjectId),
					},
				},
			})
			if err != nil {
				return errors.Wrap(err, "error running task")
			}
		case http.StatusAccepted:
			slog.Debug("No task available yet, retrying in 5 seconds...")
		default:
			return errors.New("error receiving task")
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Second):
		}
	}
}

func runTask(ctx context.Context, client generated.ClientWithResponsesInterface, task Task) error {
	switch task.Type {
	case TaskTypePostgresScan:
		return ScanPostgresDB(ctx, client, task)
	default:
		return errors.New("unknown task type")
	}
}
