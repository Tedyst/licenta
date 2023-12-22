package handlers

import (
	"context"
	"time"

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
