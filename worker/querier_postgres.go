package worker

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/saver"
)

func (q *remoteQuerier) GetPostgresDatabase(ctx context.Context, r queries.GetPostgresDatabaseParams) (*queries.GetPostgresDatabaseRow, error) {
	response, err := q.client.GetPostgresIdWithResponse(ctx, r.ID)
	if err != nil {
		return nil, errors.New("cannot get postgres database from server")
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

	switch response.StatusCode() {
	case http.StatusOK:
		return &queries.GetPostgresDatabaseRow{
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
		}, nil
	default:
		return nil, errors.New("error getting postgres database")
	}
}

func (q *remoteQuerier) UpdatePostgresVersion(ctx context.Context, params queries.UpdatePostgresVersionParams) error {
	response, err := q.client.PatchPostgresIdWithResponse(ctx, params.ID, generated.PatchPostgresDatabase{
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

func (q *remoteQuerier) GetPostgresScanByScanID(ctx context.Context, scanID int64) (*queries.PostgresScan, error) {
	response, err := q.client.GetPostgresScansWithResponse(ctx, &generated.GetPostgresScansParams{
		Scan: scanID,
	})
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

	switch response.StatusCode() {
	case http.StatusOK:
		return &queries.PostgresScan{
			ID:         int64(response.JSON200.Scans[0].Id),
			ScanID:     scanID,
			DatabaseID: int64(response.JSON200.Scans[0].DatabaseId),
		}, nil
	case http.StatusNotFound:
		return nil, pgx.ErrNoRows
	default:
		return nil, errors.New("error getting postgres scan")
	}
}

var _ saver.PostgresQuerier = &remoteQuerier{}
