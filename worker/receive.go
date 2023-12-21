package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

func ReceiveTasks(ctx context.Context, remoteURL string, authToken string) error {
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
			return err
		}

		if resp.StatusCode == http.StatusNoContent {
			slog.Info("No task available")
		}

		if resp.StatusCode == http.StatusOK {
			var task Task
			err := json.NewDecoder(resp.Body).Decode(&task)
			if err != nil {
				return err
			}

			slog.Info("Received task", "task", task)

			err = runTask(ctx, remoteURL, authToken, task)
			if err != nil {
				return err
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
