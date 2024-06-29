package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/api/authorization"
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

	scan, err := server.DatabaseProvider.GetScan(ctx, int64(message.ScanID))
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("cannot get scan: %w", err)
	}

	if scan.Scan.WorkerID.Valid && scan.Scan.WorkerID.Int64 != int64(w.ID) {
		return generated.GetWorkerGetTask202JSONResponse{
			Success: false,
			Message: "Task already taken",
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

func (server *serverHandler) DeleteWorkerId(ctx context.Context, request generated.DeleteWorkerIdRequestObject) (generated.DeleteWorkerIdResponseObject, error) {
	w, err := server.DatabaseProvider.GetWorker(ctx, request.Id)
	if err != nil {
		return nil, fmt.Errorf("cannot get worker: %w", err)
	}
	if w == nil {
		return generated.DeleteWorkerId401JSONResponse{
			Success: false,
			Message: "Unauthorized",
		}, nil
	}

	_, _, hasPerm, _, err := server.checkForOrganizationPermission(ctx, w.Organization, authorization.Admin)
	if err != nil {
		return nil, fmt.Errorf("cannot check for organization permission: %w", err)
	}

	if !hasPerm {
		return generated.DeleteWorkerId401JSONResponse{
			Success: false,
			Message: "Unauthorized",
		}, nil
	}

	_, err = server.DatabaseProvider.DeleteWorker(ctx, int64(w.ID))
	if err != nil {
		return nil, fmt.Errorf("cannot delete worker: %w", err)
	}

	return generated.DeleteWorkerId204JSONResponse{
		Success: true,
	}, nil
}

func (server *serverHandler) GetWorker(ctx context.Context, request generated.GetWorkerRequestObject) (generated.GetWorkerResponseObject, error) {
	_, _, hasPerm, _, err := server.checkForOrganizationPermission(ctx, int64(request.Params.Organization), authorization.Viewer)
	if err != nil {
		return nil, fmt.Errorf("cannot check for organization permission: %w", err)
	}
	if !hasPerm {
		return generated.GetWorker401JSONResponse{
			Success: false,
			Message: "Unauthorized",
		}, nil
	}

	workers, err := server.DatabaseProvider.GetWorkersForOrganization(ctx, int64(request.Params.Organization))
	if err != nil {
		return nil, fmt.Errorf("cannot get workers: %w", err)
	}

	workersResponse := make([]generated.Worker, len(workers))
	for i, worker := range workers {
		workersResponse[i] = generated.Worker{
			Id:           int64(worker.ID),
			Organization: int(worker.Organization),
			Token:        worker.Token,
			CreatedAt:    worker.CreatedAt.Time.Format(time.RFC3339Nano),
			Name:         worker.Name,
		}
	}

	return generated.GetWorker200JSONResponse{
		Success: true,
		Workers: workersResponse,
	}, nil
}

func (server *serverHandler) PostWorker(ctx context.Context, request generated.PostWorkerRequestObject) (generated.PostWorkerResponseObject, error) {
	_, organization, hasPerm, _, err := server.checkForOrganizationPermission(ctx, int64(request.Body.Organization), authorization.Admin)
	if err != nil {
		return nil, fmt.Errorf("cannot check for organization permission: %w", err)
	}
	if !hasPerm {
		return generated.PostWorker401JSONResponse{
			Success: false,
			Message: "Unauthorized",
		}, nil
	}

	w, err := server.DatabaseProvider.CreateWorker(ctx, queries.CreateWorkerParams{
		Organization: organization.ID,
		Token:        uuid.New().String(),
		Name:         request.Body.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create worker: %w", err)
	}

	return generated.PostWorker201JSONResponse{
		Success: true,
		Worker: generated.Worker{
			Id:           int64(w.ID),
			Organization: int(w.Organization),
			Token:        w.Token,
			CreatedAt:    w.CreatedAt.Time.Format(time.RFC3339Nano),
			Name:         w.Name,
		},
	}, nil
}
