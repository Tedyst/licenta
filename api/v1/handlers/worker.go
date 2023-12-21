package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/models"
)

func (server *serverHandler) GetWorkerGetTask(ctx context.Context, request generated.GetWorkerGetTaskRequestObject) (generated.GetWorkerGetTaskResponseObject, error) {
	worker := models.Worker{
		ID:    1,
		Token: "asdasd",
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	message, ok, err := server.MessageExchange.ReceiveSendScanToWorkerMessage(ctx, worker)
	if err != nil && err != context.DeadlineExceeded {
		return nil, err
	}

	if !ok {
		return generated.GetWorkerGetTask204JSONResponse{
			Success: false,
			Message: "No task available",
		}, nil
	}

	return generated.GetWorkerGetTask200JSONResponse{
		Success: true,
		Task: struct {
			Id       int64  "json:\"id\""
			TaskData string "json:\"task_data\""
			TaskType string "json:\"task_type\""
		}{
			Id:       1,
			TaskData: strconv.Itoa(int(message.PostgresScanID)),
			TaskType: "postgres_scan",
		},
	}, nil
}

func (server *serverHandler) PostWorkerGetTask(ctx context.Context, request generated.PostWorkerGetTaskRequestObject) (generated.PostWorkerGetTaskResponseObject, error) {
	worker := models.Worker{
		ID:    1,
		Token: "asdasd",
	}
	err := server.MessageExchange.PublishSendScanToWorkerMessage(ctx, worker, 2, 2)
	if err != nil {
		return nil, err
	}

	return generated.PostWorkerGetTask200JSONResponse{
		Success: true,
	}, nil
}
