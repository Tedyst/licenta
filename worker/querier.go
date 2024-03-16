package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/nvd"
	"github.com/tedyst/licenta/saver"
)

type remoteQuerier struct {
	client    generated.ClientWithResponsesInterface
	scan      *queries.Scan
	scanGroup *queries.ScanGroup
}

func (*remoteQuerier) UpdateBruteforcedPassword(ctx context.Context, arg queries.UpdateBruteforcedPasswordParams) (*queries.BruteforcedPassword, error) {
	return nil, nil
}

var _ saver.BaseQuerier = (*remoteQuerier)(nil)

func (q *remoteQuerier) GetScan(ctx context.Context, id int64) (*queries.GetScanRow, error) {
	return &queries.GetScanRow{
		Scan: *q.scan,
	}, nil
}

func (q *remoteQuerier) GetScanGroup(ctx context.Context, id int64) (*queries.ScanGroup, error) {
	return q.scanGroup, nil
}

func (q *remoteQuerier) UpdateScanStatus(ctx context.Context, params queries.UpdateScanStatusParams) error {
	slog.InfoContext(ctx, "Updating scan status", "params", params, "scan", q.scan.ID)

	response, err := q.client.PatchScanIdWithResponse(ctx, params.ID, generated.PatchScan{
		Status:  int(params.Status),
		EndedAt: params.EndedAt.Time.Format(time.RFC3339Nano),
		Error:   params.Error.String,
	})
	if err != nil {
		return fmt.Errorf("cannot update scan status: %w", err)
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body), "endpoint", "UpdateScanStatus")

	switch response.StatusCode() {
	case http.StatusOK:
		slog.InfoContext(ctx, "Received response", "status", response.StatusCode(), "body", response.JSON200, "endpoint", "UpdateScanStatus")
		return nil
	default:
		slog.ErrorContext(ctx, "Received response", "status", response.StatusCode(), "body", response.Body, "endpoint", "UpdateScanStatus")
		return errors.New("error updating scan status")
	}
}

func (q *remoteQuerier) CreateScanResult(ctx context.Context, params queries.CreateScanResultParams) (*queries.ScanResult, error) {
	slog.InfoContext(ctx, "Creating scan result", "params", params, "endpoint", "CreateScanResult")

	response, err := q.client.PostScanIdResultWithResponse(ctx, params.ScanID, generated.CreateScanResult{
		Message:  params.Message,
		Severity: int(params.Severity),
	})
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body), "endpoint", "CreateScanResult")

	switch response.StatusCode() {
	case http.StatusOK:
		slog.InfoContext(ctx, "Received response", "status", response.StatusCode(), "body", response.JSON200, "endpoint", "CreateScanResult")
		return &queries.ScanResult{
			ID:        int64(response.JSON200.Scan.Id),
			Severity:  int32(response.JSON200.Scan.Severity),
			Message:   response.JSON200.Scan.Message,
			CreatedAt: pgtype.Timestamptz{Time: time.Now()},
		}, nil
	default:
		slog.ErrorContext(ctx, "Received response", "status", response.StatusCode(), "body", response.Body, "endpoint", "CreateScanResult")
		return nil, errors.New("error creating scan result")
	}
}

func (q *remoteQuerier) CreateScanBruteforceResult(ctx context.Context, arg queries.CreateScanBruteforceResultParams) (*queries.ScanBruteforceResult, error) {
	slog.InfoContext(ctx, "Creating bruteforce result", "params", arg, "endpoint", "CreateScanBruteforceResult")

	response, err := q.client.PostScanIdBruteforceresultsWithResponse(ctx, arg.ScanID, generated.CreateBruteforceScanResult{
		Password: arg.Password.String,
		Total:    int(arg.Total),
		Tried:    int(arg.Tried),
		Username: arg.Username,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create scan bruteforce result: %w", err)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		slog.InfoContext(ctx, "Received response", "status", response.StatusCode(), "body", response.JSON200, "endpoint", "CreateScanBruteforceResult")
		return &queries.ScanBruteforceResult{
			ID:        int64(response.JSON200.Bruteforcescanresult.Id),
			ScanID:    arg.ScanID,
			ScanType:  arg.ScanType,
			Username:  arg.Username,
			Password:  arg.Password,
			Total:     arg.Total,
			Tried:     arg.Tried,
			CreatedAt: q.scan.CreatedAt,
		}, nil
	default:
		slog.ErrorContext(ctx, "Received response", "status", response.StatusCode(), "body", response.Body, "endpoint", "CreateScanBruteforceResult")
		return nil, errors.New("error creating scan result")
	}
}

func (q *remoteQuerier) UpdateScanBruteforceResult(ctx context.Context, params queries.UpdateScanBruteforceResultParams) error {
	slog.InfoContext(ctx, "Updating bruteforce result", "params", params, "endpoint", "UpdateScanBruteforceResult")

	response, err := q.client.PatchBruteforceresultsIdWithResponse(ctx, params.ID, generated.PatchBruteforceScanResult{
		Password: params.Password.String,
		Total:    int(params.Total),
		Tried:    int(params.Tried),
	})
	if err != nil {
		return fmt.Errorf("cannot update scan bruteforce result: %w", err)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		slog.InfoContext(ctx, "Received response", "status", response.StatusCode(), "body", response.JSON200, "endpoint", "UpdateScanBruteforceResult")
		return nil
	default:
		slog.ErrorContext(ctx, "Received response", "status", response.StatusCode(), "body", response.Body, "endpoint", "UpdateScanBruteforceResult")
		return errors.New("error creating scan result")
	}
}

func (q *remoteQuerier) GetCvesByProductAndVersion(ctx context.Context, arg queries.GetCvesByProductAndVersionParams) ([]*queries.GetCvesByProductAndVersionRow, error) {
	response, err := q.client.GetCvesDbTypeVersionWithResponse(ctx, nvd.GetNvdProductName(nvd.Product(arg.DatabaseType)), arg.Version)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body), "endpoint", "GetCvesByProductAndVersion")

	switch response.StatusCode() {
	case http.StatusOK:
		var result []*queries.GetCvesByProductAndVersionRow
		for _, cve := range response.JSON200.Cves {
			result = append(result, &queries.GetCvesByProductAndVersionRow{
				NvdCfe: queries.NvdCfe{
					CveID:       cve.CveId,
					Description: cve.Description,
					ID:          int64(cve.Id),
					LastModified: pgtype.Timestamptz{
						Time:  time.Now(),
						Valid: true,
					},
					Published: pgtype.Timestamptz{
						Time:  time.Now(),
						Valid: true,
					},
				},
			})
		}
		return result, nil
	default:
		return nil, errors.New("error getting cves")
	}
}

func (q *remoteQuerier) GetProject(ctx context.Context, id int64) (*queries.Project, error) {
	response, err := q.client.GetProjectIdWithResponse(ctx, id)
	if err != nil {
		return nil, errors.New("cannot get project")
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body), "endpoint", "GetProject")

	switch response.StatusCode() {
	case http.StatusOK:
		return &queries.Project{
			ID:             response.JSON200.Project.Id,
			Name:           response.JSON200.Project.Name,
			OrganizationID: response.JSON200.Project.OrganizationId,
			Remote:         response.JSON200.Project.Remote,
		}, nil
	default:
		return nil, errors.New("error getting project")
	}
}

func (q *remoteQuerier) GetWorkersForProject(ctx context.Context, projectID int64) ([]*queries.Worker, error) {
	return []*queries.Worker{}, nil
}

func (q *remoteQuerier) CreateBruteforcedPassword(ctx context.Context, arg queries.CreateBruteforcedPasswordParams) (*queries.BruteforcedPassword, error) {
	response, err := q.client.PostProjectIdBruteforcedPasswordWithResponse(ctx, q.scanGroup.ProjectID, generated.CreateBruteforcedPassword{
		Password:         arg.Password.String,
		Hash:             arg.Hash,
		LastBruteforceId: int(arg.LastBruteforceID.Int64),
		Username:         arg.Username,
	})
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body), "endpoint", "CreateBruteforcedPassword")

	switch response.StatusCode() {
	case http.StatusOK:
		return &queries.BruteforcedPassword{
			ID: int64(response.JSON200.BruteforcedPassword.Id),
			Password: sql.NullString{
				String: response.JSON200.BruteforcedPassword.Password,
				Valid:  response.JSON200.BruteforcedPassword.Password != "",
			},
			Hash: response.JSON200.BruteforcedPassword.Hash,
			LastBruteforceID: sql.NullInt64{
				Int64: int64(response.JSON200.BruteforcedPassword.LastBruteforceId),
				Valid: response.JSON200.BruteforcedPassword.LastBruteforceId != 0,
			},
			ProjectID: sql.NullInt64{
				Int64: int64(response.JSON200.BruteforcedPassword.ProjectId),
				Valid: response.JSON200.BruteforcedPassword.ProjectId != 0,
			},
			Username: response.JSON200.BruteforcedPassword.Username,
		}, nil
	default:
		return nil, errors.New("error creating bruteforced password")
	}
}

func (q *remoteQuerier) GetBruteforcePasswordsForProjectCount(ctx context.Context, projectID int64) (int64, error) {
	response, err := q.client.GetProjectIdBruteforcePasswordsWithResponse(ctx, projectID, &generated.GetProjectIdBruteforcePasswordsParams{})
	if err != nil {
		return 0, err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body), "endpoint", "GetBruteforcePasswordsForProjectCount")

	switch response.StatusCode() {
	case http.StatusOK:
		return int64(response.JSON200.Count), nil
	default:
		return 0, errors.New("error getting bruteforce passwords")
	}
}

func (q *remoteQuerier) GetBruteforcePasswordsPaginated(ctx context.Context, arg queries.GetBruteforcePasswordsPaginatedParams) ([]*queries.DefaultBruteforcePassword, error) {
	lastId := int32(arg.LastID)
	response, err := q.client.GetProjectIdBruteforcePasswordsWithResponse(ctx, q.scanGroup.ProjectID, &generated.GetProjectIdBruteforcePasswordsParams{
		LastPasswordId: &lastId,
	})
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body), "endpoint", "GetBruteforcePasswordsPaginated")

	switch response.StatusCode() {
	case http.StatusOK:
		var result []*queries.DefaultBruteforcePassword
		for _, password := range response.JSON200.Results {
			result = append(result, &queries.DefaultBruteforcePassword{
				ID:       int64(password.Id),
				Password: password.Password,
			})
		}
		return result, nil
	default:
		return nil, errors.New("error getting bruteforce passwords")
	}
}

func (q *remoteQuerier) GetBruteforcedPasswords(ctx context.Context, arg queries.GetBruteforcedPasswordsParams) (*queries.BruteforcedPassword, error) {
	response, err := q.client.GetProjectIdBruteforcedPasswordWithResponse(ctx, q.scanGroup.ProjectID, &generated.GetProjectIdBruteforcedPasswordParams{
		Hash:     arg.Hash,
		Username: arg.Username,
	})
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body), "endpoint", "GetBruteforcedPasswords")

	switch response.StatusCode() {
	case http.StatusOK:
		return &queries.BruteforcedPassword{
			ID: int64(response.JSON200.BruteforcedPassword.Id),
			Password: sql.NullString{
				String: response.JSON200.BruteforcedPassword.Password,
				Valid:  response.JSON200.BruteforcedPassword.Password != "",
			},
			Hash: response.JSON200.BruteforcedPassword.Hash,
			LastBruteforceID: sql.NullInt64{
				Int64: int64(response.JSON200.BruteforcedPassword.LastBruteforceId),
				Valid: response.JSON200.BruteforcedPassword.LastBruteforceId != 0,
			},
			ProjectID: sql.NullInt64{
				Int64: int64(response.JSON200.BruteforcedPassword.ProjectId),
				Valid: response.JSON200.BruteforcedPassword.ProjectId != 0,
			},
			Username: response.JSON200.BruteforcedPassword.Username,
		}, nil
	case http.StatusNotFound:
		return nil, pgx.ErrNoRows
	default:
		return nil, errors.New("error getting bruteforced password")
	}
}

func (q *remoteQuerier) GetSpecificBruteforcePasswordID(ctx context.Context, arg queries.GetSpecificBruteforcePasswordIDParams) (int64, error) {
	response, err := q.client.GetProjectIdBruteforcePasswordsWithResponse(ctx, arg.ProjectID, &generated.GetProjectIdBruteforcePasswordsParams{
		Password: &arg.Password,
	})
	if err != nil {
		return 0, err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body), "endpoint", "GetSpecificBruteforcePasswordID")

	switch response.StatusCode() {
	case http.StatusOK:
		return int64(response.JSON200.Results[0].Id), nil
	case http.StatusNotFound:
		return 0, pgx.ErrNoRows
	default:
		return 0, errors.New("error getting bruteforced password")
	}
}
