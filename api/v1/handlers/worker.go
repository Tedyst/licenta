package handlers

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/worker"
)

func (server *serverHandler) GetWorkerGetTask(ctx context.Context, request generated.GetWorkerGetTaskRequestObject) (generated.GetWorkerGetTaskResponseObject, error) {
	workerA := models.Worker{
		ID:    1,
		Token: "asdasd",
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	message, ok, err := server.MessageExchange.ReceiveSendScanToWorkerMessage(ctx, workerA)
	if err != nil && err != context.DeadlineExceeded {
		return nil, err
	}

	if !ok {
		return generated.GetWorkerGetTask204JSONResponse{
			Success: false,
			Message: "No task available",
		}, nil
	}

	scan, err := server.Queries.GetPostgresScan(ctx, int64(message.PostgresScanID))
	if err != nil {
		return nil, err
	}

	if scan.WorkerID.Valid {
		return generated.GetWorkerGetTask204JSONResponse{
			Success: false,
			Message: "Task already taken",
		}, nil
	}

	err = server.Queries.BindPostgresScanToWorker(ctx, queries.BindPostgresScanToWorkerParams{
		ID:       int64(message.PostgresScanID),
		WorkerID: sql.NullInt64{Int64: int64(workerA.ID), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	database, err := server.Queries.GetPostgresDatabase(ctx, scan.PostgresDatabaseID)
	if err != nil {
		return nil, err
	}

	return generated.GetWorkerGetTask200JSONResponse{
		Success: true,
		Task: generated.WorkerTask{
			Type: generated.WorkerTaskType(worker.TaskTypePostgresScan),
			PostgresScan: &struct {
				PostgresDatabase *generated.PostgresDatabase "json:\"postgres_database,omitempty\""
				Scan             *generated.PostgresScan     "json:\"scan,omitempty\""
			}{
				Scan: &generated.PostgresScan{
					CreatedAt:          scan.CreatedAt.Time.Format(time.RFC3339),
					EndedAt:            scan.EndedAt.Time.Format(time.RFC3339),
					Error:              scan.Error.String,
					Id:                 int(scan.ID),
					PostgresDatabaseId: int(scan.PostgresDatabaseID),
					Status:             int(scan.Status),
				},
				PostgresDatabase: &generated.PostgresDatabase{
					CreatedAt:    database.CreatedAt.Time.Format(time.RFC3339),
					DatabaseName: database.DatabaseName,
					Host:         database.Host,
					Id:           int(database.ID),
					Password:     database.Password,
					Port:         int(database.Port),
					ProjectId:    int(database.ProjectID),
					Remote:       database.Remote,
					Username:     database.Username,
				},
			},
		},
	}, nil
}

func (server *serverHandler) PostWorkerGetTask(ctx context.Context, request generated.PostWorkerGetTaskRequestObject) (generated.PostWorkerGetTaskResponseObject, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	scan, err := server.Queries.CreatePostgresScan(ctx, queries.CreatePostgresScanParams{
		PostgresDatabaseID: 1,
		Status:             int32(models.SCAN_QUEUED),
	})
	if err != nil {
		return nil, err
	}

	err = server.TaskRunner.SchedulePostgresScan(ctx, scan)
	if err != nil {
		return nil, err
	}

	return generated.PostWorkerGetTask200JSONResponse{
		Success: true,
	}, nil
}

func (server *serverHandler) GetProjectProjectidScannerPostgresScanid(ctx context.Context, request generated.GetProjectProjectidScannerPostgresScanidRequestObject) (generated.GetProjectProjectidScannerPostgresScanidResponseObject, error) {
	return nil, nil
}

func (server *serverHandler) PatchProjectProjectidScannerPostgresScanid(ctx context.Context, request generated.PatchProjectProjectidScannerPostgresScanidRequestObject) (generated.PatchProjectProjectidScannerPostgresScanidResponseObject, error) {
	if request.Body == nil {
		return generated.PatchProjectProjectidScannerPostgresScanid401JSONResponse{
			Success: false,
			Message: "Invalid request",
		}, nil
	}

	scan, err := server.Queries.GetPostgresScan(ctx, request.Scanid)
	if err != nil {
		return nil, err
	}

	database, err := server.Queries.GetPostgresDatabase(ctx, scan.PostgresDatabaseID)
	if err != nil {
		return nil, err
	}

	if database.ProjectID != int64(request.Projectid) {
		return generated.PatchProjectProjectidScannerPostgresScanid401JSONResponse{
			Success: false,
			Message: "Invalid request",
		}, nil
	}

	t, err := time.Parse(time.RFC3339, request.Body.EndedAt)
	if err != nil {
		return nil, err
	}

	err = server.Queries.UpdatePostgresScanStatus(ctx, queries.UpdatePostgresScanStatusParams{
		ID:     int64(request.Scanid),
		Status: int32(request.Body.Status),
		Error:  sql.NullString{String: request.Body.Error, Valid: request.Body.Error != ""},
		EndedAt: pgtype.Timestamptz{
			Time:  t,
			Valid: true,
		},
	})

	if err != nil {
		return nil, err
	}

	return generated.PatchProjectProjectidScannerPostgresScanid200JSONResponse{
		Success: true,
		Scan: &generated.PostgresScan{
			CreatedAt:          scan.CreatedAt.Time.Format(time.RFC3339),
			EndedAt:            scan.EndedAt.Time.Format(time.RFC3339),
			Error:              scan.Error.String,
			Id:                 int(scan.ID),
			PostgresDatabaseId: int(scan.PostgresDatabaseID),
			Status:             int(scan.Status),
		},
	}, nil
}

func (server *serverHandler) PostProjectProjectidScannerPostgresScanidResult(ctx context.Context, request generated.PostProjectProjectidScannerPostgresScanidResultRequestObject) (generated.PostProjectProjectidScannerPostgresScanidResultResponseObject, error) {
	if request.Body == nil {
		return generated.PostProjectProjectidScannerPostgresScanidResult400JSONResponse{
			Success: false,
			Message: "Invalid request",
		}, nil
	}

	scan, err := server.Queries.GetPostgresScan(ctx, request.Scanid)
	if err != nil {
		return nil, err
	}

	database, err := server.Queries.GetPostgresDatabase(ctx, scan.PostgresDatabaseID)
	if err != nil {
		return nil, err
	}

	if database.ProjectID != int64(request.Projectid) {
		return generated.PostProjectProjectidScannerPostgresScanidResult400JSONResponse{
			Success: false,
			Message: "Invalid request",
		}, nil
	}

	scanresult, err := server.Queries.CreatePostgresScanResult(ctx, queries.CreatePostgresScanResultParams{
		PostgresScanID: int64(request.Scanid),
		Severity:       int32(request.Body.Severity),
		Message:        request.Body.Message,
	})
	if err != nil {
		return nil, err
	}

	return generated.PostProjectProjectidScannerPostgresScanidResult200JSONResponse{
		Success: true,
		Scan: &generated.PostgresScanResult{
			CreatedAt:      scanresult.CreatedAt.Time.Format(time.RFC3339),
			Id:             int(scanresult.ID),
			Message:        scanresult.Message,
			PostgresScanId: int(scanresult.PostgresScanID),
			Severity:       int(scanresult.Severity),
		},
	}, nil

}
