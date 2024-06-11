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

func (server *serverHandler) GetMongoId(ctx context.Context, request generated.GetMongoIdRequestObject) (generated.GetMongoIdResponseObject, error) {
	database, err := server.DatabaseProvider.GetMongoDatabase(ctx, queries.GetMongoDatabaseParams{
		ID:      request.Id,
		SaltKey: server.saltKey,
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting Mongo database: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.GetMongoId404JSONResponse{
			Success: false,
			Message: "Database not found",
		}, nil
	}

	return generated.GetMongoId200JSONResponse{
		Success: true,
		MongoDatabase: generated.MongoDatabase{
			CreatedAt:    database.CreatedAt.Time.Format(time.RFC3339Nano),
			Host:         database.Host,
			DatabaseName: database.DatabaseName,
			Password:     database.Password,
			Id:           int(database.ID),
			Port:         int(database.Port),
			Username:     database.Username,
			Version:      database.Version.String,
			ProjectId:    int(database.ProjectID),
		},
	}, nil
}

func (server *serverHandler) PatchMongoId(ctx context.Context, request generated.PatchMongoIdRequestObject) (generated.PatchMongoIdResponseObject, error) {
	database, err := server.DatabaseProvider.GetMongoDatabase(ctx, queries.GetMongoDatabaseParams{
		ID:      request.Id,
		SaltKey: server.saltKey,
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting Mongo database: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.PatchMongoId404JSONResponse{
			Success: false,
			Message: "Database not found",
		}, nil
	}

	host := database.Host
	if request.Body.Host != nil {
		host = *request.Body.Host
	}
	username := database.Username
	if request.Body.Username != nil {
		username = *request.Body.Username
	}
	password := database.Password
	if request.Body.Password != nil {
		password = *request.Body.Password
	}
	databaseName := database.DatabaseName
	if request.Body.DatabaseName != nil {
		databaseName = *request.Body.DatabaseName
	}
	port := database.Port
	if request.Body.Port != nil {
		port = int32(*request.Body.Port)
	}
	version := database.Version
	if request.Body.Version != nil {
		version = sql.NullString{String: *request.Body.Version, Valid: true}
	}

	err = server.DatabaseProvider.UpdateMongoDatabase(ctx, queries.UpdateMongoDatabaseParams{
		ID:           int64(request.Id),
		Host:         host,
		Username:     username,
		Password:     password,
		DatabaseName: databaseName,
		Port:         port,
		Version:      version,
		ProjectID:    database.ProjectID,
		SaltKey:      server.saltKey,
	})
	if err != nil {
		return nil, err
	}

	return generated.PatchMongoId200JSONResponse{
		Success: true,
		MongoDatabase: generated.MongoDatabase{
			CreatedAt:    database.CreatedAt.Time.Format(time.RFC3339Nano),
			Host:         host,
			DatabaseName: databaseName,
			Password:     password,
			Id:           int(database.ID),
			Port:         int(port),
			Username:     username,
			ProjectId:    int(database.ProjectID),
			Version:      version.String,
		},
	}, nil
}

func (server *serverHandler) GetMongoScans(ctx context.Context, request generated.GetMongoScansRequestObject) (generated.GetMongoScansResponseObject, error) {
	worker, err := server.workerauth.GetWorker(ctx)
	if err != nil {
		return nil, err
	}

	MongoScan, err := server.DatabaseProvider.GetProjectInfoForMongoScanByScanID(ctx, queries.GetProjectInfoForMongoScanByScanIDParams{
		ScanID:  request.Params.Scan,
		SaltKey: server.saltKey,
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting Mongo scan: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.GetMongoScans404JSONResponse{
			Success: false,
			Message: "Scan not found",
		}, nil
	}

	hasPerm, err := server.authorization.WorkerHasPermissionForProject(ctx, &MongoScan.Project, worker, authorization.Worker)
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return generated.GetMongoScans401JSONResponse{
			Success: false,
			Message: "Worker does not have permission for project",
		}, nil
	}

	return generated.GetMongoScans200JSONResponse{
		Success: true,
		Scans: []generated.MongoScan{{
			DatabaseId: int(MongoScan.MongoScan.DatabaseID),
			Id:         int(MongoScan.MongoScan.ID),
		}},
	}, nil
}

func (server *serverHandler) GetMongo(ctx context.Context, request generated.GetMongoRequestObject) (generated.GetMongoResponseObject, error) {
	_, project, response, err := checkUserHasProjectPermission[generated.GetMongo401JSONResponse](server, ctx, int64(request.Params.Project), authorization.Viewer)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	databases, err := server.DatabaseProvider.GetMongoDatabasesForProject(ctx, queries.GetMongoDatabasesForProjectParams{
		ProjectID: project.ID,
		SaltKey:   server.saltKey,
	})
	if err != nil {
		return nil, err
	}

	MongoDatabases := make([]generated.MongoDatabase, len(databases))
	for i, db := range databases {
		MongoDatabases[i] = generated.MongoDatabase{
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

	return generated.GetMongo200JSONResponse{
		Success:        true,
		MongoDatabases: MongoDatabases,
	}, nil
}

func (server *serverHandler) PostMongo(ctx context.Context, request generated.PostMongoRequestObject) (generated.PostMongoResponseObject, error) {
	_, project, response, err := checkUserHasProjectPermission[generated.PostMongo401JSONResponse](server, ctx, int64(request.Body.ProjectId), authorization.Admin)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	db, err := server.DatabaseProvider.CreateMongoDatabase(ctx, queries.CreateMongoDatabaseParams{
		Host:         request.Body.Host,
		Username:     request.Body.Username,
		Password:     request.Body.Password,
		DatabaseName: request.Body.DatabaseName,
		Port:         int32(request.Body.Port),
		Version:      sql.NullString{Valid: false},
		ProjectID:    project.ID,
		SaltKey:      server.saltKey,
	})
	if err != nil {
		return nil, err
	}

	return generated.PostMongo201JSONResponse{
		Success: true,
		MongoDatabase: generated.MongoDatabase{
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

func (server *serverHandler) DeleteMongoId(ctx context.Context, request generated.DeleteMongoIdRequestObject) (generated.DeleteMongoIdResponseObject, error) {
	database, err := server.DatabaseProvider.GetMongoDatabase(ctx, queries.GetMongoDatabaseParams{
		ID:      request.Id,
		SaltKey: server.saltKey,
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting Mongo database: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.DeleteMongoId404JSONResponse{
			Success: false,
			Message: "Database not found",
		}, nil
	}

	_, project, response, err := checkUserHasProjectPermission[generated.DeleteMongoId401JSONResponse](server, ctx, int64(database.ProjectID), authorization.Admin)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}
	if project.ID != database.ProjectID {
		return generated.DeleteMongoId401JSONResponse{
			Success: false,
			Message: "Database not found in project",
		}, nil
	}

	err = server.DatabaseProvider.DeleteMongoDatabase(ctx, database.ID)
	if err != nil {
		return nil, err
	}

	return generated.DeleteMongoId204JSONResponse{
		Success: true,
	}, nil
}
