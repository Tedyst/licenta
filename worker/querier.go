package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/nvd"
	"github.com/tedyst/licenta/tasks/local"
)

type remoteQuerier struct {
	client       generated.ClientWithResponsesInterface
	scan         *models.Scan
	postgresScan *models.PostgresScan
}

var _ local.AllScannerQuerier = (*remoteQuerier)(nil)

func (q *remoteQuerier) GetScan(ctx context.Context, id int64) (*queries.GetScanRow, error) {
	return &queries.GetScanRow{
		Scan:         *q.scan,
		PostgresScan: q.postgresScan.ID,
	}, nil
}

func (q *remoteQuerier) UpdateScanStatus(ctx context.Context, params queries.UpdateScanStatusParams) error {
	slog.InfoContext(ctx, "Updating scan status", "params", params, "scan", q.scan.ID)

	response, err := q.client.PatchScanIdWithResponse(ctx, params.ID, generated.PatchScan{
		Status:  int(params.Status),
		EndedAt: params.EndedAt.Time.Format(time.RFC3339),
		Error:   params.Error.String,
	})
	if err != nil {
		return errors.Wrap(err, "cannot update scan status")
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

	switch response.StatusCode() {
	case http.StatusOK:
		slog.InfoContext(ctx, "Received response", "status", response.StatusCode(), "body", response.JSON200)
		return nil
	default:
		slog.ErrorContext(ctx, "Received response", "status", response.StatusCode(), "body", response.Body)
		return errors.New("error updating scan status")
	}
}

func (q *remoteQuerier) CreateScanResult(ctx context.Context, params queries.CreateScanResultParams) (*queries.ScanResult, error) {
	slog.InfoContext(ctx, "Creating scan result", "params", params)

	response, err := q.client.PostScanIdResultWithResponse(ctx, params.ScanID, generated.CreateScanResult{
		Message:  params.Message,
		Severity: int(params.Severity),
	})
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

	switch response.StatusCode() {
	case http.StatusOK:
		slog.InfoContext(ctx, "Received response", "status", response.StatusCode(), "body", response.JSON200)
		return &models.ScanResult{
			ID:        int64(response.JSON200.Scan.Id),
			Severity:  int32(response.JSON200.Scan.Severity),
			Message:   response.JSON200.Scan.Message,
			CreatedAt: pgtype.Timestamptz{Time: time.Now()},
		}, nil
	default:
		slog.ErrorContext(ctx, "Received response", "status", response.StatusCode(), "body", response.Body)
		return nil, errors.New("error creating scan result")
	}
}

func (q *remoteQuerier) CreateScanBruteforceResult(ctx context.Context, arg queries.CreateScanBruteforceResultParams) (*models.ScanBruteforceResult, error) {
	slog.InfoContext(ctx, "Creating bruteforce result", "params", arg)

	response, err := q.client.PostScanIdBruteforceresultsWithResponse(ctx, arg.ScanID, generated.CreateBruteforceScanResult{
		Password: arg.Password.String,
		Total:    int(arg.Total),
		Tried:    int(arg.Tried),
		Username: arg.Username,
	})
	if err != nil {
		return nil, errors.Wrap(err, "cannot create scan bruteforce result")
	}

	switch response.StatusCode() {
	case http.StatusOK:
		slog.InfoContext(ctx, "Received response", "status", response.StatusCode(), "body", response.JSON200)
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
		slog.ErrorContext(ctx, "Received response", "status", response.StatusCode(), "body", response.Body)
		return nil, errors.New("error creating scan result")
	}
}

func (q *remoteQuerier) UpdateScanBruteforceResult(ctx context.Context, params queries.UpdateScanBruteforceResultParams) error {
	slog.InfoContext(ctx, "Updating bruteforce result", "params", params)

	response, err := q.client.PatchBruteforceresultsIdWithResponse(ctx, params.ID, generated.PatchBruteforceScanResult{
		Password: params.Password.String,
		Total:    int(params.Total),
		Tried:    int(params.Tried),
	})
	if err != nil {
		return errors.Wrap(err, "cannot update scan bruteforce result")
	}

	switch response.StatusCode() {
	case http.StatusOK:
		slog.InfoContext(ctx, "Received response", "status", response.StatusCode(), "body", response.JSON200)
		return nil
	default:
		slog.ErrorContext(ctx, "Received response", "status", response.StatusCode(), "body", response.Body)
		return errors.New("error creating scan result")
	}
}

func (q *remoteQuerier) GetCvesByProductAndVersion(ctx context.Context, arg queries.GetCvesByProductAndVersionParams) ([]*queries.GetCvesByProductAndVersionRow, error) {
	response, err := q.client.GetCvesDbTypeVersionWithResponse(ctx, nvd.GetNvdDatabaseName(nvd.Product(arg.DatabaseType)), arg.Version)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

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

func (q *remoteQuerier) GetPostgresDatabase(ctx context.Context, id int64) (*queries.GetPostgresDatabaseRow, error) {
	response, err := q.client.GetPostgresIdWithResponse(ctx, q.postgresScan.DatabaseID)
	if err != nil {
		return nil, errors.New("cannot get postgres database from server")
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

	switch response.StatusCode() {
	case http.StatusOK:
		return &queries.GetPostgresDatabaseRow{
			PostgresDatabase: queries.PostgresDatabase{
				ID:           int64(response.JSON200.PostgresDatabase.Id),
				ProjectID:    int64(response.JSON200.PostgresDatabase.ProjectId),
				Host:         response.JSON200.PostgresDatabase.Host,
				Port:         int32(response.JSON200.PostgresDatabase.Port),
				DatabaseName: response.JSON200.PostgresDatabase.DatabaseName,
				Username:     response.JSON200.PostgresDatabase.Username,
				Password:     response.JSON200.PostgresDatabase.Password,
				Version: sql.NullString{
					String: response.JSON200.PostgresDatabase.Version,
					Valid:  response.JSON200.PostgresDatabase.Version != "",
				},
			},
		}, nil
	default:
		return nil, errors.New("error getting postgres database")
	}
}

func (q *remoteQuerier) UpdatePostgresVersion(ctx context.Context, params queries.UpdatePostgresVersionParams) error {
	response, err := q.client.PatchPostgresIdWithResponse(ctx, q.postgresScan.DatabaseID, generated.PatchPostgresDatabase{
		Version: &params.Version.String,
	})
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

	switch response.StatusCode() {
	case http.StatusOK:
		return nil
	default:
		return errors.New("error updating postgres version")
	}
}

func (q *remoteQuerier) GetProject(ctx context.Context, id int64) (*queries.Project, error) {
	response, err := q.client.GetProjectIdWithResponse(ctx, id)
	if err != nil {
		return nil, errors.New("cannot get project")
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

	switch response.StatusCode() {
	case http.StatusOK:
		return &queries.Project{
			ID:             response.JSON200.Project.Id,
			Name:           response.JSON200.Project.Name,
			OrganizationID: response.JSON200.Project.OrganizationId,
			Remote:         response.JSON200.Project.Remote,
		}, nil
	default:
		return nil, errors.New("error updating postgres version")
	}
}

func (q *remoteQuerier) GetPostgresScanByScanID(ctx context.Context, scanID int64) (*queries.PostgresScan, error) {
	return &queries.PostgresScan{
		ID:         scanID,
		DatabaseID: q.postgresScan.DatabaseID,
	}, nil
}

func (q *remoteQuerier) GetWorkersForProject(ctx context.Context, projectID int64) ([]*queries.Worker, error) {
	return nil, nil
}
