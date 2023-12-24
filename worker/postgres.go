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
	localmessages "github.com/tedyst/licenta/messages/local"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/tasks/local"
)

type remotePostgresQuerier struct {
	client generated.ClientWithResponsesInterface
	task   Task
}

func (q *remotePostgresQuerier) GetPostgresDatabase(ctx context.Context, id int64) (*queries.GetPostgresDatabaseRow, error) {
	return &queries.GetPostgresDatabaseRow{
		PostgresDatabase: q.task.PostgresScan.Database,
		ScanCount:        0,
	}, nil
}

func (q *remotePostgresQuerier) UpdatePostgresScanStatus(ctx context.Context, params queries.UpdatePostgresScanStatusParams) error {
	slog.InfoContext(ctx, "Updating scan status", "params", params)

	response, err := q.client.PatchScannerPostgresScanScanidWithResponse(ctx, params.ID, generated.PatchPostgresScan{
		Status:  int(params.Status),
		EndedAt: params.EndedAt.Time.Format(time.RFC3339),
		Error:   params.Error.String,
	})
	if err != nil {
		return err
	}

	switch response.StatusCode() {
	case http.StatusOK:
		slog.InfoContext(ctx, "Received response", "status", response.StatusCode(), "body", response.JSON200)
		return nil
	default:
		slog.ErrorContext(ctx, "Received response", "status", response.StatusCode(), "body", response.Body)
		return errors.New("error updating scan status")
	}
}

func (q *remotePostgresQuerier) CreatePostgresScanResult(ctx context.Context, params queries.CreatePostgresScanResultParams) (*models.PostgresScanResult, error) {
	slog.InfoContext(ctx, "Creating scan result", "params", params)

	response, err := q.client.PostScannerPostgresScanScanidResultWithResponse(ctx, params.PostgresScanID, generated.CreatePostgresScanResult{
		Message:  params.Message,
		Severity: int(params.Severity),
	})
	if err != nil {
		return nil, err
	}

	switch response.StatusCode() {
	case http.StatusOK:
		slog.InfoContext(ctx, "Received response", "status", response.StatusCode(), "body", response.JSON200)
		return &models.PostgresScanResult{
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

func (q *remotePostgresQuerier) CreatePostgresScanBruteforceResult(ctx context.Context, arg queries.CreatePostgresScanBruteforceResultParams) (*models.PostgresScanBruteforceResult, error) {
	slog.InfoContext(ctx, "Creating bruteforce result", "params", arg)
	return &queries.PostgresScanBruteforceResult{
		ID:             1,
		PostgresScanID: 1,
		Username:       "asdasd",
		Password:       sql.NullString{String: "asdasd", Valid: true},
		Total:          1,
		Tried:          1,
	}, nil
}

func (q *remotePostgresQuerier) UpdatePostgresScanBruteforceResult(ctx context.Context, params queries.UpdatePostgresScanBruteforceResultParams) error {
	return nil
}

func (q *remotePostgresQuerier) GetWorkersForProject(ctx context.Context, projectID int64) ([]*queries.GetWorkersForProjectRow, error) {
	return []*queries.GetWorkersForProjectRow{}, nil
}

func (q *remotePostgresQuerier) GetCvesByProductAndVersion(ctx context.Context, arg queries.GetCvesByProductAndVersionParams) ([]*queries.GetCvesByProductAndVersionRow, error) {
	return []*queries.GetCvesByProductAndVersionRow{}, nil
}

func ScanPostgresDB(ctx context.Context, client generated.ClientWithResponsesInterface, task Task) error {
	localexchange := localmessages.NewLocalExchange()
	runner := local.NewScannerRunner(&remotePostgresQuerier{
		client: client,
		task:   task,
	}, &remoteBruteforceProvider{
		client: client,
		task:   task,
	}, localexchange)
	return runner.ScanPostgresDB(ctx, &task.PostgresScan.Scan)
}
