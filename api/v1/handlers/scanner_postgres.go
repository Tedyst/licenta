package handlers

import (
	"context"
	"database/sql"
	errs "errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
)

const bruteforcePasswordsPerPage = 10000

func (server *serverHandler) PostScannerPostgresDatabasePostgresDatabaseId(ctx context.Context, request generated.PostScannerPostgresDatabasePostgresDatabaseIdRequestObject) (generated.PostScannerPostgresDatabasePostgresDatabaseIdResponseObject, error) {
	w := server.workerauth.GetWorker(ctx)
	if w == nil {
		return generated.PostScannerPostgresDatabasePostgresDatabaseId401JSONResponse{
			Success: false,
			Message: "Unauthorized",
		}, nil
	}

	_, err := server.Queries.GetPostgresDatabase(ctx, request.PostgresDatabaseId)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if err == pgx.ErrNoRows {
		return generated.PostScannerPostgresDatabasePostgresDatabaseId404JSONResponse{
			Success: false,
			Message: "Database not found",
		}, nil
	}

	scan, err := server.Queries.CreatePostgresScan(ctx, queries.CreatePostgresScanParams{
		PostgresDatabaseID: request.PostgresDatabaseId,
		Status:             int32(models.SCAN_QUEUED),
	})
	if err != nil {
		return nil, err
	}

	go func() {
		ctx := context.WithoutCancel(ctx)
		err := server.TaskRunner.SchedulePostgresScan(ctx, scan)
		if err != nil {
			slog.Error("Error scheduling postgres scan", "error", err)
		}
	}()

	endedTime := ""
	if scan.EndedAt.Valid {
		endedTime = scan.EndedAt.Time.Format(time.RFC3339)
	}
	return generated.PostScannerPostgresDatabasePostgresDatabaseId200JSONResponse{
		Success: true,
		Scan: &generated.PostgresScan{
			CreatedAt:       scan.CreatedAt.Time.Format(time.RFC3339),
			EndedAt:         endedTime,
			Error:           scan.Error.String,
			Id:              int(scan.ID),
			Status:          int(scan.Status),
			MaximumSeverity: 0,
		},
	}, nil
}

func (server *serverHandler) GetScannerPostgresScanScanid(ctx context.Context, request generated.GetScannerPostgresScanScanidRequestObject) (response generated.GetScannerPostgresScanScanidResponseObject, err error) {
	transaction, err := server.Queries.StartTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errs.Join(err, transaction.EndTransaction(ctx, err))
	}()

	scan, err := transaction.GetPostgresScan(ctx, request.Scanid)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "GetScannerPostgresScanScanid: error getting postgres scan")
	}
	if err == sql.ErrNoRows {
		return generated.GetScannerPostgresScanScanid404JSONResponse{
			Success: false,
			Message: "Scan not found",
		}, nil
	}

	scanResults, err := transaction.GetPostgresScanResults(ctx, scan.PostgresScan.ID)
	if err != nil {
		return nil, errors.Wrap(err, "GetScannerPostgresScanScanid: error getting postgres scan results")
	}

	results := make([]generated.PostgresScanResult, len(scanResults))
	for i, scanResult := range scanResults {
		results[i] = generated.PostgresScanResult{
			CreatedAt: scanResult.CreatedAt.Time.Format(time.RFC3339),
			Id:        int(scanResult.ID),
			Message:   scanResult.Message,
			Severity:  int(scanResult.Severity),
		}
	}

	return generated.GetScannerPostgresScanScanid200JSONResponse{
		Success: true,
		Scan: generated.PostgresScan{
			CreatedAt:       scan.PostgresScan.CreatedAt.Time.Format(time.RFC3339),
			EndedAt:         scan.PostgresScan.EndedAt.Time.Format(time.RFC3339),
			Error:           scan.PostgresScan.Error.String,
			Id:              int(scan.PostgresScan.ID),
			Status:          int(scan.PostgresScan.Status),
			MaximumSeverity: int(scan.MaximumSeverity),
		},
		Results: results,
	}, nil
}

func (server *serverHandler) PatchScannerPostgresScanScanid(ctx context.Context, request generated.PatchScannerPostgresScanScanidRequestObject) (generated.PatchScannerPostgresScanScanidResponseObject, error) {
	if request.Body == nil {
		return generated.PatchScannerPostgresScanScanid400JSONResponse{
			Success: false,
			Message: "Invalid request",
		}, nil
	}

	scan, err := server.Queries.GetPostgresScan(ctx, request.Scanid)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if err == pgx.ErrNoRows {
		return generated.PatchScannerPostgresScanScanid404JSONResponse{
			Success: false,
			Message: "Scan not found",
		}, nil
	}

	t, err := time.Parse(time.RFC3339, request.Body.EndedAt)
	if err != nil {
		return nil, err
	}

	err = server.Queries.UpdatePostgresScanStatus(ctx, queries.UpdatePostgresScanStatusParams{
		ID:     int64(request.Scanid),
		Status: int32(request.Body.Status),
		Error:  sql.NullString{String: request.Body.Error, Valid: request.Body.Error != ""},
		EndedAt: pgtype.Timestamptz{
			Time:  t,
			Valid: true,
		},
	})

	if err != nil {
		return nil, err
	}

	return generated.PatchScannerPostgresScanScanid200JSONResponse{
		Success: true,
		Scan: &generated.PostgresScan{
			CreatedAt:       scan.PostgresScan.CreatedAt.Time.Format(time.RFC3339),
			EndedAt:         scan.PostgresScan.EndedAt.Time.Format(time.RFC3339),
			Error:           scan.PostgresScan.Error.String,
			Id:              int(scan.PostgresScan.ID),
			Status:          int(scan.PostgresScan.Status),
			MaximumSeverity: int(scan.MaximumSeverity),
		},
	}, nil
}

func (server *serverHandler) PostScannerPostgresScanScanidResult(ctx context.Context, request generated.PostScannerPostgresScanScanidResultRequestObject) (generated.PostScannerPostgresScanScanidResultResponseObject, error) {
	if request.Body == nil {
		return generated.PostScannerPostgresScanScanidResult400JSONResponse{
			Success: false,
			Message: "Invalid request",
		}, nil
	}

	_, err := server.Queries.GetPostgresScan(ctx, request.Scanid)
	if err != nil {
		return nil, err
	}
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if err == pgx.ErrNoRows {
		return generated.PostScannerPostgresScanScanidResult404JSONResponse{
			Success: false,
			Message: "Scan not found",
		}, nil
	}

	scanresult, err := server.Queries.CreatePostgresScanResult(ctx, queries.CreatePostgresScanResultParams{
		PostgresScanID: int64(request.Scanid),
		Severity:       int32(request.Body.Severity),
		Message:        request.Body.Message,
	})
	if err != nil {
		return nil, err
	}

	return generated.PostScannerPostgresScanScanidResult200JSONResponse{
		Success: true,
		Scan: &generated.PostgresScanResult{
			CreatedAt: scanresult.CreatedAt.Time.Format(time.RFC3339),
			Id:        int(scanresult.ID),
			Message:   scanresult.Message,
			Severity:  int(scanresult.Severity),
		},
	}, nil
}

func (server *serverHandler) GetScannerPostgresDatabasePostgresDatabaseId(ctx context.Context, request generated.GetScannerPostgresDatabasePostgresDatabaseIdRequestObject) (generated.GetScannerPostgresDatabasePostgresDatabaseIdResponseObject, error) {
	database, err := server.Queries.GetPostgresDatabase(ctx, request.PostgresDatabaseId)
	if err != nil && err != pgx.ErrNoRows {
		return nil, errors.Wrap(err, "error getting postgres database")
	}
	if err == pgx.ErrNoRows {
		return generated.GetScannerPostgresDatabasePostgresDatabaseId404JSONResponse{
			Success: false,
			Message: "Database not found",
		}, nil
	}

	scans, err := server.Queries.GetPostgresScansForDatabase(ctx, database.PostgresDatabase.ID)
	if err != nil {
		return nil, errors.Wrap(err, "error getting postgres scans for database")
	}

	results := make([]generated.PostgresScan, len(scans))
	for i, scan := range scans {
		endedTime := ""
		if scan.PostgresScan.EndedAt.Valid {
			endedTime = scan.PostgresScan.EndedAt.Time.Format(time.RFC3339)
		}
		results[i] = generated.PostgresScan{
			CreatedAt:       scan.PostgresScan.CreatedAt.Time.Format(time.RFC3339),
			EndedAt:         endedTime,
			Error:           scan.PostgresScan.Error.String,
			Id:              int(scan.PostgresScan.ID),
			Status:          int(scan.PostgresScan.Status),
			MaximumSeverity: int(scan.MaximumSeverity),
		}
	}

	return generated.GetScannerPostgresDatabasePostgresDatabaseId200JSONResponse{
		Success: true,
		Database: generated.PostgresDatabase{
			CreatedAt:    database.PostgresDatabase.CreatedAt.Time.Format(time.RFC3339),
			Host:         database.PostgresDatabase.Host,
			DatabaseName: database.PostgresDatabase.DatabaseName,
			Password:     database.PostgresDatabase.Password,
			Remote:       database.PostgresDatabase.Remote,
			Id:           int(database.PostgresDatabase.ID),
			Port:         int(database.PostgresDatabase.Port),
			Username:     database.PostgresDatabase.Username,
			Version:      database.PostgresDatabase.Version.String,
			ProjectId:    int(database.PostgresDatabase.ProjectID),
		},
		Scans:     results,
		ScanCount: int(database.ScanCount),
	}, nil
}

func (server *serverHandler) PatchScannerPostgresDatabasePostgresDatabaseId(ctx context.Context, request generated.PatchScannerPostgresDatabasePostgresDatabaseIdRequestObject) (generated.PatchScannerPostgresDatabasePostgresDatabaseIdResponseObject, error) {
	database, err := server.Queries.GetPostgresDatabase(ctx, request.PostgresDatabaseId)
	if err != nil && err != pgx.ErrNoRows {
		return nil, errors.Wrap(err, "error getting postgres database")
	}
	if err == pgx.ErrNoRows {
		return generated.PatchScannerPostgresDatabasePostgresDatabaseId404JSONResponse{
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
	remote := database.PostgresDatabase.Remote
	if request.Body.Remote != nil {
		remote = *request.Body.Remote
	}
	version := database.PostgresDatabase.Version
	if request.Body.Version != nil {
		version = sql.NullString{String: *request.Body.Version, Valid: true}
	}

	err = server.Queries.UpdatePostgresDatabase(ctx, queries.UpdatePostgresDatabaseParams{
		ID:           int64(request.PostgresDatabaseId),
		Host:         host,
		Username:     username,
		Password:     password,
		DatabaseName: databaseName,
		Port:         port,
		Remote:       remote,
		Version:      version,
	})
	if err != nil {
		return nil, err
	}

	return generated.PatchScannerPostgresDatabasePostgresDatabaseId200JSONResponse{
		Success: true,
		Database: &generated.PostgresDatabase{
			CreatedAt:    database.PostgresDatabase.CreatedAt.Time.Format(time.RFC3339),
			Host:         host,
			DatabaseName: databaseName,
			Password:     password,
			Remote:       remote,
			Id:           int(database.PostgresDatabase.ID),
			Port:         int(port),
			Username:     username,
			ProjectId:    int(database.PostgresDatabase.ProjectID),
			Version:      version.String,
		},
	}, nil
}
