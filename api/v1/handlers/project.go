package handlers

import (
	"context"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
)

func (server *serverHandler) GetProjectId(ctx context.Context, request generated.GetProjectIdRequestObject) (generated.GetProjectIdResponseObject, error) {
	project, err := server.DatabaseProvider.GetProject(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	postgres_databases_q, err := server.DatabaseProvider.GetPostgresDatabasesForProject(ctx, request.Id)
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
			Username:     db.Username,
			Version:      db.Version.String,
		}
	}

	return generated.GetProjectId200JSONResponse{
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

func (server *serverHandler) PostProjectIdRun(ctx context.Context, request generated.PostProjectIdRunRequestObject) (generated.PostProjectIdRunResponseObject, error) {
	postgres_databases, err := server.DatabaseProvider.GetPostgresDatabasesForProject(ctx, request.Id)
	if err != nil {
		return nil, errors.Wrap(err, "error getting postgres databases for project")
	}

	var scans []generated.Scan
	for _, db := range postgres_databases {
		scan, err := server.DatabaseProvider.CreateScan(ctx, queries.CreateScanParams{
			Status:    models.SCAN_NOT_STARTED,
			ProjectID: db.ProjectID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "error creating postgres scan")
		}

		postgresScan, err := server.DatabaseProvider.CreatePostgresScan(ctx, queries.CreatePostgresScanParams{
			ScanID:     scan.ID,
			DatabaseID: db.ID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "error creating postgres scan")
		}

		go func() {
			ctx := context.WithoutCancel(ctx)
			err := server.TaskRunner.RunAllScanners(ctx, scan, false)
			if err != nil {
				slog.Error("Error scheduling postgres scan", "error", err)
			}
		}()

		scans = append(scans, generated.Scan{
			CreatedAt:       db.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			EndedAt:         db.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			Error:           "",
			Id:              int(scan.ID),
			Status:          int(scan.Status),
			MaximumSeverity: 0,
			PostgresScan: &generated.PostgresScan{
				Id:         int(postgresScan.ID),
				DatabaseId: int(db.ID),
			},
		})
	}

	return generated.PostProjectIdRun200JSONResponse{
		Success: true,
		Scans:   &scans,
	}, nil
}
