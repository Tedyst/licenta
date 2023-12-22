package worker

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/models"
)

func ReceiveTasks(ctx context.Context, remoteURL string, authToken string) error {
	slog.Info("Starting to receive tasks", "remoteURL", remoteURL, "authToken", authToken)

	for {
		req, err := http.NewRequest("GET", remoteURL+"/api/v1/worker/get-task", nil)
		if err != nil {
			return err
		}

		req.Header.Set("X-Worker-Token", authToken)

		newCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		req = req.WithContext(newCtx)

		resp, err := http.DefaultClient.Do(req)
		if err != nil && err != context.DeadlineExceeded {
			return errors.Wrap(err, "error sending request")
		}

		if resp.StatusCode == http.StatusNoContent {
			slog.DebugContext(ctx, "No task available yet, retrying in 5 seconds...")
		}

		if resp.StatusCode == http.StatusOK {
			httpBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return errors.Wrap(err, "error reading body")
			}

			slog.InfoContext(ctx, "Received task", "body", string(httpBody))

			var response generated.GetWorkerGetTask200JSONResponse
			err = json.NewDecoder(bytes.NewReader(httpBody)).Decode(&response)
			if err != nil {
				return errors.Wrap(err, "error decoding task")
			}

			slog.InfoContext(ctx, "Received task", "task", response)

			if !response.Success {
				return errors.New("error receiving task")
			}

			task := Task{
				Type: TaskType(response.Task.Type),
				PostgresScan: PostgresScan{
					Scan: models.PostgresScan{
						ID:                 int64(response.Task.PostgresScan.Scan.Id),
						PostgresDatabaseID: int64(response.Task.PostgresScan.Scan.PostgresDatabaseId),
						Status:             int32(response.Task.PostgresScan.Scan.Status),
						CreatedAt:          pgtype.Timestamptz{Time: time.Now()},
						EndedAt:            pgtype.Timestamptz{Time: time.Now()},
						Error:              sql.NullString{String: response.Task.PostgresScan.Scan.Error},
					},
					Database: models.PostgresDatabases{
						ID:           int64(response.Task.PostgresScan.PostgresDatabase.Id),
						Host:         response.Task.PostgresScan.PostgresDatabase.Host,
						Port:         int32(response.Task.PostgresScan.PostgresDatabase.Port),
						DatabaseName: response.Task.PostgresScan.PostgresDatabase.DatabaseName,
						Password:     response.Task.PostgresScan.PostgresDatabase.Password,
						Username:     response.Task.PostgresScan.PostgresDatabase.Username,
						Remote:       response.Task.PostgresScan.PostgresDatabase.Remote,
						ProjectID:    int64(response.Task.PostgresScan.PostgresDatabase.ProjectId),
					},
				},
			}

			err = runTask(ctx, remoteURL, authToken, task)
			if err != nil {
				return errors.Wrap(err, "error running task")
			}
		}

		cancel()

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Second):
		}
	}
}

func runTask(ctx context.Context, remoteURL string, authData string, task Task) error {
	switch task.Type {
	case TaskTypePostgresScan:
		return ScanPostgresDB(ctx, remoteURL, authData, task)
	default:
		return errors.New("unknown task type")
	}
}
