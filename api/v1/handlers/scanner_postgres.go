package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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
			CreatedAt:    database.PostgresDatabase.CreatedAt.Time.Format(time.RFC3339),
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
			CreatedAt:    database.PostgresDatabase.CreatedAt.Time.Format(time.RFC3339),
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
