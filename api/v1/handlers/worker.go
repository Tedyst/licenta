package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
)

func (server *serverHandler) GetWorkerGetTask(ctx context.Context, request generated.GetWorkerGetTaskRequestObject) (generated.GetWorkerGetTaskResponseObject, error) {
	w, err := server.workerauth.GetWorker(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get worker: %w", err)
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
		return nil, fmt.Errorf("cannot receive message: %w", err)
	}

	if !ok {
		return generated.GetWorkerGetTask202JSONResponse{
			Success: false,
			Message: "No task available",
		}, nil
	}

	boundScan, err := server.DatabaseProvider.BindScanToWorker(ctx, queries.BindScanToWorkerParams{
		ID:       int64(message.ScanID),
		WorkerID: sql.NullInt64{Int64: int64(w.ID), Valid: true},
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("cannot bind scan to worker: %w", err)
	}

	if boundScan == nil {
		return generated.GetWorkerGetTask202JSONResponse{
			Success: false,
			Message: "Task already taken",
		}, nil
	}

	scan, err := server.DatabaseProvider.GetScan(ctx, int64(message.ScanID))
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("cannot get scan: %w", err)
	}

	scanGroup, err := server.DatabaseProvider.GetScanGroup(ctx, scan.Scan.ScanGroupID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("cannot get scan group: %w", err)
	}

	return generated.GetWorkerGetTask200JSONResponse{
		Success: true,
		Scan: generated.Scan{
			Id:              int(scan.Scan.ID),
			CreatedAt:       scan.Scan.CreatedAt.Time.Format(time.RFC3339Nano),
			EndedAt:         scan.Scan.EndedAt.Time.Format(time.RFC3339Nano),
			Error:           scan.Scan.Error.String,
			Status:          int(scan.Scan.Status),
			MaximumSeverity: int(scan.MaximumSeverity),
			ScanGroupId:     int(scan.Scan.ScanGroupID),
		},
		ScanGroup: generated.ScanGroup{
			Id:        int(scanGroup.ID),
			ProjectId: int(scanGroup.ProjectID),
		},
	}, nil
}
