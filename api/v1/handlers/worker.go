package handlers

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
)

func (server *serverHandler) GetWorkerGetTask(ctx context.Context, request generated.GetWorkerGetTaskRequestObject) (generated.GetWorkerGetTaskResponseObject, error) {
	w, err := server.workerauth.GetWorker(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get worker")
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
		return nil, errors.Wrap(err, "cannot receive message")
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
		return nil, errors.Wrap(err, "cannot bind scan to worker")
	}

	if boundScan == nil {
		return generated.GetWorkerGetTask202JSONResponse{
			Success: false,
			Message: "Task already taken",
		}, nil
	}

	scan, err := server.DatabaseProvider.GetScan(ctx, int64(message.ScanID))
	if err != nil && err != pgx.ErrNoRows {
		return nil, errors.Wrap(err, "cannot get scan")
	}

	postgresScan, err := server.DatabaseProvider.GetPostgresScanByScanID(ctx, message.ScanID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, errors.Wrap(err, "cannot get postgres scan")
	}

	var postgresScanResponse *generated.PostgresScan = nil
	if postgresScan != nil {
		postgresScanResponse = &generated.PostgresScan{
			DatabaseId: int(postgresScan.DatabaseID),
			Id:         int(postgresScan.ID),
		}
	}

	return generated.GetWorkerGetTask200JSONResponse{
		Success: true,
		Scan: generated.Scan{
			Id:              int(scan.Scan.ID),
			CreatedAt:       scan.Scan.CreatedAt.Time.Format(time.RFC3339),
			EndedAt:         scan.Scan.EndedAt.Time.Format(time.RFC3339),
			Error:           scan.Scan.Error.String,
			PostgresScan:    postgresScanResponse,
			Status:          int(scan.Scan.Status),
			MaximumSeverity: int(scan.MaximumSeverity),
		},
	}, nil
}
