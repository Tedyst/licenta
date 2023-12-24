package handlers

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
)

const bruteforcePasswordsPerPage = 10000

func (server *serverHandler) PostProjectProjectidScannerPostgres(ctx context.Context, request generated.PostProjectProjectidScannerPostgresRequestObject) (generated.PostProjectProjectidScannerPostgresResponseObject, error) {
	scan, err := server.Queries.CreatePostgresScan(ctx, queries.CreatePostgresScanParams{
		PostgresDatabaseID: 1,
		Status:             int32(models.SCAN_QUEUED),
	})
	if err != nil {
		return nil, err
	}

	go func() {
		err := server.TaskRunner.SchedulePostgresScan(ctx, scan)
		if err != nil {
			slog.Error("Error scheduling postgres scan", "error", err)
		}
	}()

	endedTime := ""
	if scan.EndedAt.Valid {
		endedTime = scan.EndedAt.Time.Format(time.RFC3339)
	}
	return generated.PostProjectProjectidScannerPostgres200JSONResponse{
		Success: true,
		Scan: &generated.PostgresScan{
			CreatedAt: scan.CreatedAt.Time.Format(time.RFC3339),
			EndedAt:   endedTime,
			Error:     scan.Error.String,
			Id:        int(scan.ID),
			Status:    int(scan.Status),
		},
	}, nil
}

func (server *serverHandler) GetProjectProjectidScannerPostgresScanid(ctx context.Context, request generated.GetProjectProjectidScannerPostgresScanidRequestObject) (generated.GetProjectProjectidScannerPostgresScanidResponseObject, error) {
	scan, err := server.Queries.GetPostgresScan(ctx, request.Scanid)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		return generated.GetProjectProjectidScannerPostgresScanid404JSONResponse{
			Success: false,
			Message: "Scan not found",
		}, nil
	}

	scanResults, err := server.Queries.GetPostgresScanResults(ctx, scan.ID)
	if err != nil {
		return nil, err
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

	return generated.GetProjectProjectidScannerPostgresScanid200JSONResponse{
		Success: true,
		Scan: generated.PostgresScan{
			CreatedAt: scan.CreatedAt.Time.Format(time.RFC3339),
			EndedAt:   scan.EndedAt.Time.Format(time.RFC3339),
			Error:     scan.Error.String,
			Id:        int(scan.ID),
			Status:    int(scan.Status),
		},
		Results: results,
	}, nil
}

func (server *serverHandler) PatchProjectProjectidScannerPostgresScanid(ctx context.Context, request generated.PatchProjectProjectidScannerPostgresScanidRequestObject) (generated.PatchProjectProjectidScannerPostgresScanidResponseObject, error) {
	if request.Body == nil {
		return generated.PatchProjectProjectidScannerPostgresScanid401JSONResponse{
			Success: false,
			Message: "Invalid request",
		}, nil
	}

	scan, err := server.Queries.GetPostgresScan(ctx, request.Scanid)
	if err != nil {
		return nil, err
	}

	database, err := server.Queries.GetPostgresDatabase(ctx, scan.PostgresDatabaseID)
	if err != nil {
		return nil, err
	}

	if database.ProjectID != int64(request.Projectid) {
		return generated.PatchProjectProjectidScannerPostgresScanid401JSONResponse{
			Success: false,
			Message: "Invalid request",
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

	return generated.PatchProjectProjectidScannerPostgresScanid200JSONResponse{
		Success: true,
		Scan: &generated.PostgresScan{
			CreatedAt: scan.CreatedAt.Time.Format(time.RFC3339),
			EndedAt:   scan.EndedAt.Time.Format(time.RFC3339),
			Error:     scan.Error.String,
			Id:        int(scan.ID),
			Status:    int(scan.Status),
		},
	}, nil
}

func (server *serverHandler) PostProjectProjectidScannerPostgresScanidResult(ctx context.Context, request generated.PostProjectProjectidScannerPostgresScanidResultRequestObject) (generated.PostProjectProjectidScannerPostgresScanidResultResponseObject, error) {
	if request.Body == nil {
		return generated.PostProjectProjectidScannerPostgresScanidResult400JSONResponse{
			Success: false,
			Message: "Invalid request",
		}, nil
	}

	scan, err := server.Queries.GetPostgresScan(ctx, request.Scanid)
	if err != nil {
		return nil, err
	}

	database, err := server.Queries.GetPostgresDatabase(ctx, scan.PostgresDatabaseID)
	if err != nil {
		return nil, err
	}

	if database.ProjectID != int64(request.Projectid) {
		return generated.PostProjectProjectidScannerPostgresScanidResult400JSONResponse{
			Success: false,
			Message: "Invalid request",
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

	return generated.PostProjectProjectidScannerPostgresScanidResult200JSONResponse{
		Success: true,
		Scan: &generated.PostgresScanResult{
			CreatedAt: scanresult.CreatedAt.Time.Format(time.RFC3339),
			Id:        int(scanresult.ID),
			Message:   scanresult.Message,
			Severity:  int(scanresult.Severity),
		},
	}, nil
}
