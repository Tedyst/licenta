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

func (server *serverHandler) GetProjectsId(ctx context.Context, request generated.GetProjectsIdRequestObject) (generated.GetProjectsIdResponseObject, error) {
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

	return generated.GetProjectsId200JSONResponse{
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

func (server *serverHandler) PostProjectsIdRun(ctx context.Context, request generated.PostProjectsIdRunRequestObject) (generated.PostProjectsIdRunResponseObject, error) {
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
		return generated.PostProjectsIdRun404JSONResponse{
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
			return generated.PostProjectsIdRun401JSONResponse{
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
			return generated.PostProjectsIdRun401JSONResponse{
				Message: "Not allowed to run scans on this project",
				Success: false,
			}, nil
		}
	} else {
		return generated.PostProjectsIdRun401JSONResponse{
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

	return generated.PostProjectsIdRun200JSONResponse{
		Success: true,
		ScanGroup: &generated.ScanGroup{
			Id:    int(scanGroup.ID),
			Scans: resultScans,
		},
	}, nil
}

func (server *serverHandler) PostProjects(ctx context.Context, request generated.PostProjectsRequestObject) (generated.PostProjectsResponseObject, error) {
	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	organization, err := server.DatabaseProvider.GetOrganization(ctx, int64(request.Body.OrganizationId))
	if err != nil {
		return nil, fmt.Errorf("error getting organization: %w", err)
	}

	ok, err := server.authorization.UserHasPermissionForOrganization(ctx, organization, user, authorization.Admin)
	if err != nil {
		return nil, fmt.Errorf("error checking permissions: %w", err)
	}

	if !ok {
		return generated.PostProjects401JSONResponse{
			Message: "Not allowed to create projects in this organization",
			Success: false,
		}, nil
	}

	project, err := server.DatabaseProvider.CreateProject(ctx, queries.CreateProjectParams{
		Name:           request.Body.Name,
		OrganizationID: int64(request.Body.OrganizationId),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating project: %w", err)
	}

	return generated.PostProjects201JSONResponse{
		Success: true,
		Project: generated.Project{
			CreatedAt:      project.CreatedAt.Time.Format(time.RFC3339Nano),
			Id:             project.ID,
			Name:           project.Name,
			OrganizationId: project.OrganizationID,
		},
	}, nil
}
