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

func (server *serverHandler) GetRedisId(ctx context.Context, request generated.GetRedisIdRequestObject) (generated.GetRedisIdResponseObject, error) {
	database, err := server.DatabaseProvider.GetRedisDatabase(ctx, queries.GetRedisDatabaseParams{
		ID:      request.Id,
		SaltKey: server.saltKey,
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting Redis database: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.GetRedisId404JSONResponse{
			Success: false,
			Message: "Database not found",
		}, nil
	}

	return generated.GetRedisId200JSONResponse{
		Success: true,
		RedisDatabase: generated.RedisDatabase{
			CreatedAt: database.CreatedAt.Time.Format(time.RFC3339Nano),
			Host:      database.Host,
			Password:  database.Password,
			Id:        int(database.ID),
			Port:      int(database.Port),
			Username:  database.Username,
			Version:   database.Version.String,
			ProjectId: int(database.ProjectID),
		},
	}, nil
}

func (server *serverHandler) PatchRedisId(ctx context.Context, request generated.PatchRedisIdRequestObject) (generated.PatchRedisIdResponseObject, error) {
	database, err := server.DatabaseProvider.GetRedisDatabase(ctx, queries.GetRedisDatabaseParams{
		ID:      request.Id,
		SaltKey: server.saltKey,
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting Redis database: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.PatchRedisId404JSONResponse{
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
	port := database.Port
	if request.Body.Port != nil {
		port = int32(*request.Body.Port)
	}
	version := database.Version
	if request.Body.Version != nil {
		version = sql.NullString{String: *request.Body.Version, Valid: true}
	}

	err = server.DatabaseProvider.UpdateRedisDatabase(ctx, queries.UpdateRedisDatabaseParams{
		ID:        int64(request.Id),
		Host:      host,
		Username:  username,
		Password:  password,
		Port:      port,
		Version:   version,
		ProjectID: database.ProjectID,
		SaltKey:   server.saltKey,
	})
	if err != nil {
		return nil, err
	}

	return generated.PatchRedisId200JSONResponse{
		Success: true,
		RedisDatabase: generated.RedisDatabase{
			CreatedAt: database.CreatedAt.Time.Format(time.RFC3339Nano),
			Host:      host,
			Password:  password,
			Id:        int(database.ID),
			Port:      int(port),
			Username:  username,
			ProjectId: int(database.ProjectID),
			Version:   version.String,
		},
	}, nil
}

func (server *serverHandler) GetRedisScans(ctx context.Context, request generated.GetRedisScansRequestObject) (generated.GetRedisScansResponseObject, error) {
	worker, err := server.workerauth.GetWorker(ctx)
	if err != nil {
		return nil, err
	}

	RedisScan, err := server.DatabaseProvider.GetProjectInfoForRedisScanByScanID(ctx, queries.GetProjectInfoForRedisScanByScanIDParams{
		ScanID:  request.Params.Scan,
		SaltKey: server.saltKey,
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting Redis scan: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.GetRedisScans404JSONResponse{
			Success: false,
			Message: "Scan not found",
		}, nil
	}

	hasPerm, err := server.authorization.WorkerHasPermissionForProject(ctx, &RedisScan.Project, worker, authorization.Worker)
	if err != nil {
		return nil, err
	}
	if !hasPerm {
		return generated.GetRedisScans401JSONResponse{
			Success: false,
			Message: "Worker does not have permission for project",
		}, nil
	}

	return generated.GetRedisScans200JSONResponse{
		Success: true,
		Scans: []generated.RedisScan{{
			DatabaseId: int(RedisScan.RedisScan.DatabaseID),
			Id:         int(RedisScan.RedisScan.ID),
		}},
	}, nil
}

func (server *serverHandler) GetRedis(ctx context.Context, request generated.GetRedisRequestObject) (generated.GetRedisResponseObject, error) {
	_, project, response, err := checkUserHasProjectPermission[generated.GetRedis401JSONResponse](server, ctx, int64(request.Params.Project), authorization.Viewer)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	databases, err := server.DatabaseProvider.GetRedisDatabasesForProject(ctx, queries.GetRedisDatabasesForProjectParams{
		ProjectID: project.ID,
		SaltKey:   server.saltKey,
	})
	if err != nil {
		return nil, err
	}

	RedisDatabases := make([]generated.RedisDatabase, len(databases))
	for i, db := range databases {
		RedisDatabases[i] = generated.RedisDatabase{
			CreatedAt: db.CreatedAt.Time.Format(time.RFC3339Nano),
			Host:      db.Host,
			Id:        int(db.ID),
			Port:      int(db.Port),
			Username:  db.Username,
			Version:   db.Version.String,
			ProjectId: int(db.ProjectID),
		}
	}

	return generated.GetRedis200JSONResponse{
		Success:        true,
		RedisDatabases: RedisDatabases,
	}, nil
}

func (server *serverHandler) PostRedis(ctx context.Context, request generated.PostRedisRequestObject) (generated.PostRedisResponseObject, error) {
	_, project, response, err := checkUserHasProjectPermission[generated.PostRedis401JSONResponse](server, ctx, int64(request.Body.ProjectId), authorization.Admin)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	db, err := server.DatabaseProvider.CreateRedisDatabase(ctx, queries.CreateRedisDatabaseParams{
		Host:      request.Body.Host,
		Username:  request.Body.Username,
		Password:  request.Body.Password,
		Port:      int32(request.Body.Port),
		Version:   sql.NullString{Valid: false},
		ProjectID: project.ID,
		SaltKey:   server.saltKey,
	})
	if err != nil {
		return nil, err
	}

	return generated.PostRedis201JSONResponse{
		Success: true,
		RedisDatabase: generated.RedisDatabase{
			CreatedAt: time.Now().Format(time.RFC3339Nano),
			Host:      db.Host,
			Id:        int(db.ID),
			Port:      int(db.Port),
			Username:  db.Username,
			Version:   db.Version.String,
			ProjectId: int(db.ProjectID),
		},
	}, nil
}

func (server *serverHandler) DeleteRedisId(ctx context.Context, request generated.DeleteRedisIdRequestObject) (generated.DeleteRedisIdResponseObject, error) {
	database, err := server.DatabaseProvider.GetRedisDatabase(ctx, queries.GetRedisDatabaseParams{
		ID:      request.Id,
		SaltKey: server.saltKey,
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting Redis database: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.DeleteRedisId404JSONResponse{
			Success: false,
			Message: "Database not found",
		}, nil
	}

	_, project, response, err := checkUserHasProjectPermission[generated.DeleteRedisId401JSONResponse](server, ctx, int64(database.ProjectID), authorization.Admin)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}
	if project.ID != database.ProjectID {
		return generated.DeleteRedisId401JSONResponse{
			Success: false,
			Message: "Database not found in project",
		}, nil
	}

	err = server.DatabaseProvider.DeleteRedisDatabase(ctx, database.ID)
	if err != nil {
		return nil, err
	}

	return generated.DeleteRedisId204JSONResponse{
		Success: true,
	}, nil
}
