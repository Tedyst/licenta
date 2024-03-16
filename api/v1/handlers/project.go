package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/tedyst/licenta/api/authorization"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/saver"
)

func (server *serverHandler) GetProjectId(ctx context.Context, request generated.GetProjectIdRequestObject) (generated.GetProjectIdResponseObject, error) {
	project, err := server.DatabaseProvider.GetProject(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	postgresDatabasesQ, err := server.DatabaseProvider.GetPostgresDatabasesForProject(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	postgresDatabases := make([]generated.PostgresDatabase, len(postgresDatabasesQ))
	for i, db := range postgresDatabasesQ {
		postgresDatabases[i] = generated.PostgresDatabase{
			CreatedAt:    db.CreatedAt.Time.Format(time.RFC3339Nano),
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
			CreatedAt:      project.CreatedAt.Time.Format(time.RFC3339Nano),
			Id:             project.ID,
			Name:           project.Name,
			OrganizationId: project.OrganizationID,
		},
		PostgresDatabases: postgresDatabases,
	}, nil
}

func (server *serverHandler) PostProjectIdRun(ctx context.Context, request generated.PostProjectIdRunRequestObject) (generated.PostProjectIdRunResponseObject, error) {
	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	worker, err := server.workerauth.GetWorker(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting worker: %w", err)
	}

	project, err := server.DatabaseProvider.GetProject(ctx, request.Id)
	if err != nil {
		return generated.PostProjectIdRun404JSONResponse{
			Message: "Project not found",
			Success: false,
		}, nil
	}

	if user != nil {
		authorized, err := server.authorization.UserHasPermissionForProject(ctx, project, user, authorization.Admin)
		if err != nil {
			return nil, fmt.Errorf("error checking permissions: %w", err)
		}
		if !authorized {
			return generated.PostProjectIdRun401JSONResponse{
				Message: "Not allowed to run scans on this project",
				Success: false,
			}, nil
		}
	} else if worker != nil {
		authorized, err := server.authorization.WorkerHasPermissionForProject(ctx, project, worker, authorization.Admin)
		if err != nil {
			return nil, fmt.Errorf("error checking permissions: %w", err)
		}
		if !authorized {
			return generated.PostProjectIdRun401JSONResponse{
				Message: "Not allowed to run scans on this project",
				Success: false,
			}, nil
		}
	} else {
		return generated.PostProjectIdRun401JSONResponse{
			Message: "Not allowed to run scans on this project",
			Success: false,
		}, nil
	}

	createdBy := sql.NullInt64{}
	if user != nil {
		createdBy = sql.NullInt64{Int64: user.ID, Valid: true}
	}

	scanGroup, err := server.DatabaseProvider.CreateScanGroup(ctx, queries.CreateScanGroupParams{
		ProjectID: request.Id,
		CreatedBy: createdBy,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating scan group: %w", err)
	}

	scans, err := saver.CreateScans(ctx, server.DatabaseProvider, request.Id, scanGroup.ID, "all")
	if err != nil {
		return nil, fmt.Errorf("error creating scans: %w", err)
	}

	resultScans := make([]generated.Scan, len(scans))
	for i, scan := range scans {
		resultScans[i] = generated.Scan{
			CreatedAt:   scan.CreatedAt.Time.Format(time.RFC3339Nano),
			EndedAt:     scan.EndedAt.Time.Format(time.RFC3339Nano),
			Error:       scan.Error.String,
			Id:          int(scan.ID),
			ScanGroupId: int(scan.ScanGroupID),
			Status:      int(scan.Status),
		}
	}

	for _, scan := range scans {
		scan := scan
		go func() {
			ctx := context.WithoutCancel(ctx)
			err := server.TaskRunner.ScheduleSaverRun(ctx, scan, "all")
			if err != nil {
				slog.Error("Error scheduling postgres scan", "error", err)
			}
		}()
	}

	return generated.PostProjectIdRun200JSONResponse{
		Success: true,
		ScanGroup: &generated.ScanGroup{
			Id:    int(scanGroup.ID),
			Scans: resultScans,
		},
	}, nil
}
