package handlers

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

	message, ok, err := server.MessageExchange.ReceiveSendScanToWorkerMessage(ctx, *w)
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

	scan, err := server.DatabaseProvider.GetScan(ctx, int64(message.PostgresScanID))
	if err != nil {
		return nil, err
	}

	if scan.Scan.WorkerID.Valid {
		return generated.GetWorkerGetTask202JSONResponse{
			Success: false,
			Message: "Task already taken",
		}, nil
	}

	err = server.DatabaseProvider.BindScanToWorker(ctx, queries.BindScanToWorkerParams{
		ID:       int64(message.PostgresScanID),
		WorkerID: sql.NullInt64{Int64: int64(w.ID), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	var postgresDatabase *generated.PostgresDatabase
	if scan.PostgresScan != 0 {
		database, err := server.DatabaseProvider.GetPostgresDatabase(ctx, scan.PostgresScan)
		if err != nil {
			return nil, err
		}

		postgresDatabase = &generated.PostgresDatabase{
			CreatedAt:    database.PostgresDatabase.CreatedAt.Time.Format(time.RFC3339),
			DatabaseName: database.PostgresDatabase.DatabaseName,
			Host:         database.PostgresDatabase.Host,
			Id:           int(database.PostgresDatabase.ID),
			Password:     database.PostgresDatabase.Password,
			Port:         int(database.PostgresDatabase.Port),
			ProjectId:    int(database.PostgresDatabase.ProjectID),
			Remote:       database.PostgresDatabase.Remote,
			Username:     database.PostgresDatabase.Username,
		}
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
					CreatedAt: scan.Scan.CreatedAt.Time.Format(time.RFC3339),
					EndedAt:   scan.Scan.EndedAt.Time.Format(time.RFC3339),
					Error:     scan.Scan.Error.String,
					Id:        int(scan.Scan.ID),
					Status:    int(scan.Scan.Status),
				},
				PostgresDatabase: postgresDatabase,
			},
		},
	}, nil
}
