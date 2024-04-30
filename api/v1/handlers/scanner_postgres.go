package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/api/authorization"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
)

const bruteforcePasswordsPerPage = 10000

func (server *serverHandler) GetPostgresId(ctx context.Context, request generated.GetPostgresIdRequestObject) (generated.GetPostgresIdResponseObject, error) {
	database, err := server.DatabaseProvider.GetPostgresDatabase(ctx, request.Id)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting postgres database: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.GetPostgresId404JSONResponse{
			Success: false,
			Message: "Database not found",
		}, nil
	}

	return generated.GetPostgresId200JSONResponse{
		Success: true,
		PostgresDatabase: generated.PostgresDatabase{
			CreatedAt:    database.PostgresDatabase.CreatedAt.Time.Format(time.RFC3339Nano),
			Host:         database.PostgresDatabase.Host,
			DatabaseName: database.PostgresDatabase.DatabaseName,
			Password:     database.PostgresDatabase.Password,
			Id:           int(database.PostgresDatabase.ID),
			Port:         int(database.PostgresDatabase.Port),
			Username:     database.PostgresDatabase.Username,
			Version:      database.PostgresDatabase.Version.String,
			ProjectId:    int(database.PostgresDatabase.ProjectID),
		},
	}, nil
}

func (server *serverHandler) PatchPostgresId(ctx context.Context, request generated.PatchPostgresIdRequestObject) (generated.PatchPostgresIdResponseObject, error) {
	database, err := server.DatabaseProvider.GetPostgresDatabase(ctx, request.Id)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting postgres database: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.PatchPostgresId404JSONResponse{
			Success: false,
			Message: "Database not found",
		}, nil
	}

	host := database.PostgresDatabase.Host
	if request.Body.Host != nil {
		host = *request.Body.Host
	}
	username := database.PostgresDatabase.Username
	if request.Body.Username != nil {
		username = *request.Body.Username
	}
	password := database.PostgresDatabase.Password
	if request.Body.Password != nil {
		password = *request.Body.Password
	}
	databaseName := database.PostgresDatabase.DatabaseName
	if request.Body.DatabaseName != nil {
		databaseName = *request.Body.DatabaseName
	}
	port := database.PostgresDatabase.Port
	if request.Body.Port != nil {
		port = int32(*request.Body.Port)
	}
	version := database.PostgresDatabase.Version
	if request.Body.Version != nil {
		version = sql.NullString{String: *request.Body.Version, Valid: true}
	}

	err = server.DatabaseProvider.UpdatePostgresDatabase(ctx, queries.UpdatePostgresDatabaseParams{
		ID:           int64(request.Id),
		Host:         host,
		Username:     username,
		Password:     password,
		DatabaseName: databaseName,
		Port:         port,
		Version:      version,
	})
	if err != nil {
		return nil, err
	}

	return generated.PatchPostgresId200JSONResponse{
		Success: true,
		PostgresDatabase: generated.PostgresDatabase{
			CreatedAt:    database.PostgresDatabase.CreatedAt.Time.Format(time.RFC3339Nano),
			Host:         host,
			DatabaseName: databaseName,
			Password:     password,
			Id:           int(database.PostgresDatabase.ID),
			Port:         int(port),
			Username:     username,
			ProjectId:    int(database.PostgresDatabase.ProjectID),
			Version:      version.String,
		},
	}, nil
}

func (server *serverHandler) GetPostgresScans(ctx context.Context, request generated.GetPostgresScansRequestObject) (generated.GetPostgresScansResponseObject, error) {
	worker, err := server.workerauth.GetWorker(ctx)
	if err != nil {
		return nil, err
	}

	info, err := server.DatabaseProvider.GetProjectInfoForPostgresScanByScanID(ctx, request.Params.Scan)
	if err == pgx.ErrNoRows {
		return generated.GetPostgresScans404JSONResponse{
			Success: false,
			Message: "Scan not found",
		}, nil
	}
	if err != nil {
		return nil, err
	}

	hasPerm, err := server.authorization.WorkerHasPermissionForProject(ctx, &info.Project, worker, authorization.Worker)
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return generated.GetPostgresScans401JSONResponse{
			Success: false,
			Message: "Worker does not have permission for project",
		}, nil
	}

	return generated.GetPostgresScans200JSONResponse{
		Success: true,
		Scans: []generated.PostgresScan{{
			DatabaseId: int(info.PostgresScan.DatabaseID),
			Id:         int(info.PostgresScan.ID),
		}},
	}, nil
}

func (server *serverHandler) GetPostgres(ctx context.Context, request generated.GetPostgresRequestObject) (generated.GetPostgresResponseObject, error) {
	_, project, response, err := checkUserHasProjectPermission[generated.GetPostgres401JSONResponse](server, ctx, int64(request.Params.Project), authorization.Viewer)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	databases, err := server.DatabaseProvider.GetPostgresDatabasesForProject(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	postgresDatabases := make([]generated.PostgresDatabase, 0, len(databases))
	for i, db := range databases {
		postgresDatabases[i] = generated.PostgresDatabase{
			CreatedAt:    db.CreatedAt.Time.Format(time.RFC3339Nano),
			DatabaseName: db.DatabaseName,
			Host:         db.Host,
			Password:     db.Password,
			Port:         int(db.Port),
			ProjectId:    int(db.ProjectID),
			Username:     db.Username,
			Version:      db.Version.String,
			Id:           int(db.ID),
		}
	}

	return generated.GetPostgres200JSONResponse{
		Success:           true,
		PostgresDatabases: postgresDatabases,
	}, nil
}

func (server *serverHandler) PostPostgres(ctx context.Context, request generated.PostPostgresRequestObject) (generated.PostPostgresResponseObject, error) {
	_, project, response, err := checkUserHasProjectPermission[generated.PostPostgres401JSONResponse](server, ctx, int64(request.Body.ProjectId), authorization.Admin)
	if err != nil {
		return nil, err
	}

	if response.Success == false {
		return response, nil
	}

	db, err := server.DatabaseProvider.CreatePostgresDatabase(ctx, queries.CreatePostgresDatabaseParams{
		Host:         request.Body.Host,
		Username:     request.Body.Username,
		ProjectID:    project.ID,
		DatabaseName: request.Body.DatabaseName,
		Port:         int32(request.Body.Port),
		Password:     request.Body.Password,
		Version:      sql.NullString{Valid: false},
	})
	if err != nil {
		return nil, err
	}

	return generated.PostPostgres201JSONResponse{
		Success: true,
		PostgresDatabase: generated.PostgresDatabase{
			CreatedAt:    time.Now().Format(time.RFC3339Nano),
			Host:         db.Host,
			DatabaseName: db.DatabaseName,
			Password:     db.Password,
			Id:           int(db.ID),
			Port:         int(db.Port),
			Username:     db.Username,
			Version:      db.Version.String,
			ProjectId:    int(db.ProjectID),
		},
	}, nil
}
