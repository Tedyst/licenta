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
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/saver"
)

func (server *serverHandler) GetProjectsId(ctx context.Context, request generated.GetProjectsIdRequestObject) (generated.GetProjectsIdResponseObject, error) {
	project, err := server.DatabaseProvider.GetProjectWithStats(ctx, request.Id)
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
			Scans:          int(project.Scans),
		},
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

	gitRepositories, err := server.DatabaseProvider.GetGitRepositoriesForProject(ctx, queries.GetGitRepositoriesForProjectParams{
		ProjectID: request.Id,
		SaltKey:   server.saltKey,
	})
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting git repositories: %w", err)
	}

	for _, gitRepository := range gitRepositories {
		scan, err := server.DatabaseProvider.CreateScan(ctx, queries.CreateScanParams{
			Status:      models.SCAN_NOT_STARTED,
			ScanGroupID: scanGroup.ID,
			ScanType:    models.SCAN_GIT,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating scan: %w", err)
		}

		_, err = server.DatabaseProvider.CreateGitScan(ctx, queries.CreateGitScanParams{
			ScanID:       scan.ID,
			RepositoryID: gitRepository.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating git scan: %w", err)
		}

		scans = append(scans, scan)
	}

	dockerImages, err := server.DatabaseProvider.GetDockerImagesForProject(ctx, queries.GetDockerImagesForProjectParams{
		ProjectID: project.ID,
		SaltKey:   server.saltKey,
	})
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting docker images: %w", err)
	}

	for _, dockerImage := range dockerImages {
		scan, err := server.DatabaseProvider.CreateScan(ctx, queries.CreateScanParams{
			Status:      models.SCAN_NOT_STARTED,
			ScanGroupID: scanGroup.ID,
			ScanType:    models.SCAN_DOCKER,
		})

		if err != nil {
			return nil, fmt.Errorf("error creating scan: %w", err)
		}

		_, err = server.DatabaseProvider.CreateDockerScan(ctx, queries.CreateDockerScanParams{
			ScanID:  scan.ID,
			ImageID: dockerImage.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating docker scan: %w", err)
		}
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
			ScanType:    int(scan.ScanType),
		}
	}

	go func() {
		ctx := context.WithoutCancel(ctx)
		if err := server.TaskRunner.ScheduleFullRun(ctx, project, scanGroup, "all", "all"); err != nil {
			slog.ErrorContext(ctx, "error scheduling full run", "error", err, "project", project, "scangroup", scanGroup)
		}
	}()

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
			Scans:          0,
		},
	}, nil
}

func (server *serverHandler) DeleteProjectsId(ctx context.Context, request generated.DeleteProjectsIdRequestObject) (generated.DeleteProjectsIdResponseObject, error) {
	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	project, err := server.DatabaseProvider.GetProject(ctx, request.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting project: %w", err)
	}

	ok, err := server.authorization.UserHasPermissionForProject(ctx, project, user, authorization.Admin)
	if err != nil {
		return nil, fmt.Errorf("error checking permissions: %w", err)
	}

	if !ok {
		return generated.DeleteProjectsId401JSONResponse{
			Message: "Not allowed to delete project",
			Success: false,
		}, nil
	}

	_, err = server.DatabaseProvider.DeleteProject(ctx, request.Id)
	if err != nil {
		return nil, fmt.Errorf("error deleting project: %w", err)
	}

	return generated.DeleteProjectsId204JSONResponse{
		Success: true,
	}, nil
}
