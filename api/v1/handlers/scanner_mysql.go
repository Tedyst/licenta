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
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting Mysql scan: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.GetMysqlScans404JSONResponse{
			Success: false,
			Message: "Scan not found",
		}, nil
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

func (server *serverHandler) GetMysql(ctx context.Context, request generated.GetMysqlRequestObject) (generated.GetMysqlResponseObject, error) {
	_, project, response, err := checkUserHasProjectPermission[generated.GetMysql401JSONResponse](server, ctx, int64(request.Params.Project), authorization.Viewer)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	databases, err := server.DatabaseProvider.GetMysqlDatabasesForProject(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	mysqlDatabases := make([]generated.MysqlDatabase, len(databases))
	for i, db := range databases {
		mysqlDatabases[i] = generated.MysqlDatabase{
			CreatedAt:    db.CreatedAt.Time.Format(time.RFC3339Nano),
			Host:         db.Host,
			DatabaseName: db.DatabaseName,
			Id:           int(db.ID),
			Port:         int(db.Port),
			Username:     db.Username,
			Version:      db.Version.String,
			ProjectId:    int(db.ProjectID),
		}
	}

	return generated.GetMysql200JSONResponse{
		Success:        true,
		MysqlDatabases: mysqlDatabases,
	}, nil
}

func (server *serverHandler) PostMysql(ctx context.Context, request generated.PostMysqlRequestObject) (generated.PostMysqlResponseObject, error) {
	_, project, response, err := checkUserHasProjectPermission[generated.PostMysql401JSONResponse](server, ctx, int64(request.Body.ProjectId), authorization.Admin)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	db, err := server.DatabaseProvider.CreateMysqlDatabase(ctx, queries.CreateMysqlDatabaseParams{
		Host:         request.Body.Host,
		Username:     request.Body.Username,
		Password:     request.Body.Password,
		DatabaseName: request.Body.DatabaseName,
		Port:         int32(request.Body.Port),
		Version:      sql.NullString{Valid: false},
		ProjectID:    project.ID,
	})
	if err != nil {
		return nil, err
	}

	return generated.PostMysql201JSONResponse{
		Success: true,
		MysqlDatabase: generated.MysqlDatabase{
			CreatedAt:    time.Now().Format(time.RFC3339Nano),
			Host:         db.Host,
			DatabaseName: db.DatabaseName,
			Id:           int(db.ID),
			Port:         int(db.Port),
			Username:     db.Username,
			Version:      db.Version.String,
			ProjectId:    int(db.ProjectID),
		},
	}, nil
}

func (server *serverHandler) DeleteMysqlId(ctx context.Context, request generated.DeleteMysqlIdRequestObject) (generated.DeleteMysqlIdResponseObject, error) {
	database, err := server.DatabaseProvider.GetMysqlDatabase(ctx, request.Id)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting Mysql database: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.DeleteMysqlId404JSONResponse{
			Success: false,
			Message: "Database not found",
		}, nil
	}

	_, project, response, err := checkUserHasProjectPermission[generated.DeleteMysqlId401JSONResponse](server, ctx, int64(database.MysqlDatabase.ProjectID), authorization.Admin)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}
	if project.ID != database.MysqlDatabase.ProjectID {
		return generated.DeleteMysqlId401JSONResponse{
			Success: false,
			Message: "Database not found in project",
		}, nil
	}

	err = server.DatabaseProvider.DeleteMysqlDatabase(ctx, database.MysqlDatabase.ID)
	if err != nil {
		return nil, err
	}

	return generated.DeleteMysqlId204JSONResponse{
		Success: true,
	}, nil
}
