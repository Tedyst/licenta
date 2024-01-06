package handlers

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/worker"
)

func (server *serverHandler) GetWorkerGetTask(ctx context.Context, request generated.GetWorkerGetTaskRequestObject) (generated.GetWorkerGetTaskResponseObject, error) {
	w, err := server.workerauth.GetWorker(ctx)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return generated.GetWorkerGetTask401JSONResponse{
			Success: false,
			Message: "Unauthorized",
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	message, ok, err := server.MessageExchange.ReceiveSendScanToWorkerMessage(ctx, w)
	if err != nil && err != context.DeadlineExceeded {
		return nil, err
	}

	if !ok {
		return generated.GetWorkerGetTask202JSONResponse{
			Success: false,
			Message: "No task available",
		}, nil
	}

	if message.ScanType != messages.PostgresScan {
		return nil, errors.New("invalid scan type")
	}

	boundScan, err := server.DatabaseProvider.BindScanToWorker(ctx, queries.BindScanToWorkerParams{
		ID:       int64(message.ScanID),
		WorkerID: sql.NullInt64{Int64: int64(w.ID), Valid: true},
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	if boundScan == nil {
		return generated.GetWorkerGetTask202JSONResponse{
			Success: false,
			Message: "Task already taken",
		}, nil
	}

	scan, err := server.DatabaseProvider.GetScan(ctx, int64(message.ScanID))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	postgresScan, err := server.DatabaseProvider.GetPostgresScan(ctx, int64(message.ScanID))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	var postgresScanResponse *generated.PostgresScan
	if postgresScan != nil {
		postgresScanResponse = &generated.PostgresScan{
			DatabaseId: int(postgresScan.DatabaseID),
			Id:         int(postgresScan.ID),
		}
	}

	return generated.GetWorkerGetTask200JSONResponse{
		Success: true,
		Task: generated.WorkerTask{
			Type: generated.WorkerTaskType(worker.TaskTypePostgresScan),
			Scan: generated.Scan{
				Id:              int(scan.Scan.ID),
				CreatedAt:       scan.Scan.CreatedAt.Time.Format(time.RFC3339),
				EndedAt:         scan.Scan.EndedAt.Time.Format(time.RFC3339),
				Error:           scan.Scan.Error.String,
				PostgresScan:    postgresScanResponse,
				Status:          int(scan.Scan.Status),
				MaximumSeverity: int(scan.MaximumSeverity),
			},
		},
	}, nil
}
