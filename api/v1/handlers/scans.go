package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tedyst/licenta/api/authorization"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
)

func (server *serverHandler) GetScanId(ctx context.Context, request generated.GetScanIdRequestObject) (generated.GetScanIdResponseObject, error) {
	scan, err := server.DatabaseProvider.GetScan(ctx, request.Id)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("GetScanId: error getting scan: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.GetScanId404JSONResponse{
			Success: false,
			Message: "Scan not found",
		}, nil
	}

	scanResultsQ, err := server.DatabaseProvider.GetScanResults(ctx, scan.Scan.ID)
	if err != nil {
		return nil, fmt.Errorf("GetScannerPostgresScanScanid: error getting scan results: %w", err)
	}

	scanResults := make([]generated.ScanResult, len(scanResultsQ))
	for i, scanResult := range scanResultsQ {
		scanResults[i] = generated.ScanResult{
			CreatedAt:  scanResult.CreatedAt.Time.Format(time.RFC3339Nano),
			Id:         int(scanResult.ID),
			Message:    scanResult.Message,
			Severity:   int(scanResult.Severity),
			ScanSource: int(scanResult.ScanSource),
		}
	}

	bruteforceScanResultsQ, err := server.DatabaseProvider.GetScanBruteforceResults(ctx, scan.Scan.ID)
	if err != nil {
		return nil, fmt.Errorf("GetScannerPostgresScanScanid: error getting bruteforce scan results: %w", err)
	}

	bruteforceResults := make([]generated.BruteforceScanResult, len(bruteforceScanResultsQ))
	for i, scanResult := range bruteforceScanResultsQ {
		bruteforceResults[i] = generated.BruteforceScanResult{
			Id:       int(scanResult.ID),
			Password: scanResult.Password.String,
			Total:    int(scanResult.Total),
			Tried:    int(scanResult.Tried),
			Username: scanResult.Username,
		}
	}

	return generated.GetScanId200JSONResponse{
		Success: true,
		Scan: generated.Scan{
			CreatedAt:       scan.Scan.CreatedAt.Time.Format(time.RFC3339Nano),
			EndedAt:         scan.Scan.EndedAt.Time.Format(time.RFC3339Nano),
			Error:           scan.Scan.Error.String,
			Id:              int(scan.Scan.ID),
			Status:          int(scan.Scan.Status),
			MaximumSeverity: int(scan.MaximumSeverity),
			ScanGroupId:     int(scan.Scan.ScanGroupID),
			ScanType:        int(scan.Scan.ScanType),
		},
		Results:           scanResults,
		BruteforceResults: bruteforceResults,
	}, nil
}

func (server *serverHandler) PatchScanId(ctx context.Context, request generated.PatchScanIdRequestObject) (generated.PatchScanIdResponseObject, error) {
	if request.Body == nil {
		return generated.PatchScanId400JSONResponse{
			Success: false,
			Message: "Invalid request",
		}, nil
	}

	scan, err := server.DatabaseProvider.GetScan(ctx, request.Id)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("PatchScanId: error getting scan: %w", err)
	}
	if err == pgx.ErrNoRows {
		return generated.PatchScanId404JSONResponse{
			Success: false,
			Message: "Scan not found",
		}, nil
	}

	t, err := time.Parse(time.RFC3339Nano, request.Body.EndedAt)
	if err != nil {
		return nil, err
	}

	err = server.DatabaseProvider.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
		ID:     int64(request.Id),
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

	return generated.PatchScanId200JSONResponse{
		Success: true,
		Scan: &generated.Scan{
			CreatedAt:       scan.Scan.CreatedAt.Time.Format(time.RFC3339Nano),
			EndedAt:         scan.Scan.EndedAt.Time.Format(time.RFC3339Nano),
			Error:           scan.Scan.Error.String,
			Id:              int(scan.Scan.ID),
			Status:          int(scan.Scan.Status),
			MaximumSeverity: int(scan.MaximumSeverity),
			ScanGroupId:     int(scan.Scan.ScanGroupID),
			ScanType:        int(scan.Scan.ScanType),
		},
	}, nil
}

func (server *serverHandler) PostScanIdResult(ctx context.Context, request generated.PostScanIdResultRequestObject) (generated.PostScanIdResultResponseObject, error) {
	if request.Body == nil {
		return generated.PostScanIdResult400JSONResponse{
			Success: false,
			Message: "Invalid request",
		}, nil
	}

	_, err := server.DatabaseProvider.GetScan(ctx, request.Id)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if err == pgx.ErrNoRows {
		return generated.PostScanIdResult404JSONResponse{
			Success: false,
			Message: "Scan not found",
		}, nil
	}

	scanresult, err := server.DatabaseProvider.CreateScanResult(ctx, queries.CreateScanResultParams{
		ScanID:   int64(request.Id),
		Severity: int32(request.Body.Severity),
		Message:  request.Body.Message,
	})
	if err != nil {
		return nil, err
	}

	return generated.PostScanIdResult200JSONResponse{
		Success: true,
		Scan: &generated.ScanResult{
			CreatedAt: scanresult.CreatedAt.Time.Format(time.RFC3339Nano),
			Id:        int(scanresult.ID),
			Message:   scanresult.Message,
			Severity:  int(scanresult.Severity),
		},
	}, nil
}

func (server *serverHandler) GetScanGroups(ctx context.Context, request generated.GetScanGroupsRequestObject) (generated.GetScanGroupsResponseObject, error) {
	_, project, response, err := checkUserHasProjectPermission[generated.GetScanGroups401JSONResponse](server, ctx, int64(request.Params.Project), authorization.Viewer)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	scangroups, err := server.DatabaseProvider.GetScanGroupsForProject(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	scans := map[int64][]generated.Scan{}
	groups := []generated.ScanGroup{}

	for _, scangroup := range scangroups {
		if _, ok := scans[scangroup.ScanGroup.ID]; !ok {
			scans[scangroup.ScanGroup.ID] = []generated.Scan{}
			groups = append(groups, generated.ScanGroup{
				Id:        int(scangroup.ScanGroup.ID),
				ProjectId: int(scangroup.ScanGroup.ProjectID),
				Scans:     []generated.Scan{},
			})
		}

		scans[scangroup.ScanGroup.ID] = append(scans[scangroup.ScanGroup.ID], generated.Scan{
			CreatedAt:       scangroup.Scan.CreatedAt.Time.Format(time.RFC3339Nano),
			EndedAt:         scangroup.Scan.EndedAt.Time.Format(time.RFC3339Nano),
			Error:           scangroup.Scan.Error.String,
			Id:              int(scangroup.Scan.ID),
			Status:          int(scangroup.Scan.Status),
			ScanGroupId:     int(scangroup.ScanGroup.ID),
			MaximumSeverity: int(scangroup.MaximumSeverity),
			ScanType:        int(scangroup.Scan.ScanType),
		})
	}

	for i, group := range groups {
		groups[i].Scans = scans[int64(group.Id)]
	}

	return generated.GetScanGroups200JSONResponse{
		Success:    true,
		ScanGroups: groups,
	}, nil
}
