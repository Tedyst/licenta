package handlers

import (
	"context"
	"log/slog"

	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
)

func (server *serverHandler) GetProjectProjectid(ctx context.Context, request generated.GetProjectProjectidRequestObject) (generated.GetProjectProjectidResponseObject, error) {
	project, err := server.Queries.GetProjectByID(ctx, request.Projectid)
	if err != nil {
		return nil, err
	}

	postgres_databases_q, err := server.Queries.GetPostgresDatabasesForProject(ctx, request.Projectid)
	if err != nil {
		return nil, err
	}
	postgres_databases := make([]generated.PostgresDatabase, len(postgres_databases_q))
	for i, db := range postgres_databases_q {
		postgres_databases[i] = generated.PostgresDatabase{
			CreatedAt:    db.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			Id:           int(db.ID),
			DatabaseName: db.DatabaseName,
			Host:         db.Host,
			Password:     db.Password,
			Port:         int(db.Port),
			ProjectId:    int(db.ProjectID),
			Remote:       db.Remote,
			Username:     db.Username,
			Version:      db.Version.String,
		}
	}

	return generated.GetProjectProjectid200JSONResponse{
		Success: true,
		Project: generated.Project{
			CreatedAt:      project.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			Id:             project.ID,
			Name:           project.Name,
			OrganizationId: project.OrganizationID,
		},
		PostgresDatabases: postgres_databases,
	}, nil
}

func (server *serverHandler) PostProjectProjectidRun(ctx context.Context, request generated.PostProjectProjectidRunRequestObject) (generated.PostProjectProjectidRunResponseObject, error) {
	postgres_databases, err := server.Queries.GetPostgresDatabasesForProject(ctx, request.Projectid)
	if err != nil {
		return nil, err
	}

	var postgresScans []generated.PostgresScan
	for _, db := range postgres_databases {
		scan, err := server.Queries.CreatePostgresScan(ctx, queries.CreatePostgresScanParams{
			PostgresDatabaseID: int64(db.ID),
			Status:             models.SCAN_NOT_STARTED,
		})
		if err != nil {
			return nil, err
		}

		go func() {
			ctx := context.WithoutCancel(ctx)
			err := server.TaskRunner.SchedulePostgresScan(ctx, scan)
			if err != nil {
				slog.Error("Error scheduling postgres scan", "error", err)
			}
		}()

		postgresScans = append(postgresScans, generated.PostgresScan{
			CreatedAt: db.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			EndedAt:   db.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			Error:     "",
			Id:        int(scan.ID),
			Status:    int(scan.Status),
		})
	}

	return generated.PostProjectProjectidRun200JSONResponse{
		Success:       true,
		PostgresScans: postgresScans,
	}, nil
}
