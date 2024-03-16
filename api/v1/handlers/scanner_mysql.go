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

func (server *serverHandler) GetMysqlId(ctx context.Context, request generated.GetMysqlIdRequestObject) (generated.GetMysqlIdResponseObject, error) {
	database, err := server.DatabaseProvider.GetMysqlDatabase(ctx, request.Id)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting mysql database: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.GetMysqlId404JSONResponse{
			Success: false,
			Message: "Database not found",
		}, nil
	}

	return generated.GetMysqlId200JSONResponse{
		Success: true,
		MysqlDatabase: generated.MysqlDatabase{
			CreatedAt:    database.MysqlDatabase.CreatedAt.Time.Format(time.RFC3339Nano),
			Host:         database.MysqlDatabase.Host,
			DatabaseName: database.MysqlDatabase.DatabaseName,
			Password:     database.MysqlDatabase.Password,
			Id:           int(database.MysqlDatabase.ID),
			Port:         int(database.MysqlDatabase.Port),
			Username:     database.MysqlDatabase.Username,
			Version:      database.MysqlDatabase.Version.String,
			ProjectId:    int(database.MysqlDatabase.ProjectID),
		},
	}, nil
}

func (server *serverHandler) PatchMysqlId(ctx context.Context, request generated.PatchMysqlIdRequestObject) (generated.PatchMysqlIdResponseObject, error) {
	database, err := server.DatabaseProvider.GetMysqlDatabase(ctx, request.Id)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting Mysql database: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.PatchMysqlId404JSONResponse{
			Success: false,
			Message: "Database not found",
		}, nil
	}

	host := database.MysqlDatabase.Host
	if request.Body.Host != nil {
		host = *request.Body.Host
	}
	username := database.MysqlDatabase.Username
	if request.Body.Username != nil {
		username = *request.Body.Username
	}
	password := database.MysqlDatabase.Password
	if request.Body.Password != nil {
		password = *request.Body.Password
	}
	databaseName := database.MysqlDatabase.DatabaseName
	if request.Body.DatabaseName != nil {
		databaseName = *request.Body.DatabaseName
	}
	port := database.MysqlDatabase.Port
	if request.Body.Port != nil {
		port = int32(*request.Body.Port)
	}
	version := database.MysqlDatabase.Version
	if request.Body.Version != nil {
		version = sql.NullString{String: *request.Body.Version, Valid: true}
	}

	err = server.DatabaseProvider.UpdateMysqlDatabase(ctx, queries.UpdateMysqlDatabaseParams{
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

	return generated.PatchMysqlId200JSONResponse{
		Success: true,
		MysqlDatabase: generated.MysqlDatabase{
			CreatedAt:    database.MysqlDatabase.CreatedAt.Time.Format(time.RFC3339Nano),
			Host:         host,
			DatabaseName: databaseName,
			Password:     password,
			Id:           int(database.MysqlDatabase.ID),
			Port:         int(port),
			Username:     username,
			ProjectId:    int(database.MysqlDatabase.ProjectID),
			Version:      version.String,
		},
	}, nil
}

func (server *serverHandler) GetMysqlScans(ctx context.Context, request generated.GetMysqlScansRequestObject) (generated.GetMysqlScansResponseObject, error) {
	worker, err := server.workerauth.GetWorker(ctx)
	if err != nil {
		return nil, err
	}

	MysqlScan, err := server.DatabaseProvider.GetProjectInfoForMysqlScanByScanID(ctx, request.Params.Scan)
	if err != nil {
		return nil, err
	}

	hasPerm, err := server.authorization.WorkerHasPermissionForProject(ctx, &MysqlScan.Project, worker, authorization.Worker)
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return generated.GetMysqlScans401JSONResponse{
			Success: false,
			Message: "Worker does not have permission for project",
		}, nil
	}

	return generated.GetMysqlScans200JSONResponse{
		Success: true,
		Scans: []generated.MysqlScan{{
			DatabaseId: int(MysqlScan.MysqlScan.DatabaseID),
			Id:         int(MysqlScan.MysqlScan.ID),
		}},
	}, nil
}
