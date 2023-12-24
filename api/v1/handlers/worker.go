package handlers

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/messages"
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

	if message.ScanType != messages.PostgresScan {
		return nil, errors.New("invalid scan type")
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
					CreatedAt: scan.CreatedAt.Time.Format(time.RFC3339),
					EndedAt:   scan.EndedAt.Time.Format(time.RFC3339),
					Error:     scan.Error.String,
					Id:        int(scan.ID),
					Status:    int(scan.Status),
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
