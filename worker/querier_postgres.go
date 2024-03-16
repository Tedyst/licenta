package worker

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
)

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

func (q *remoteQuerier) GetPostgresScanByScanID(ctx context.Context, scanID int64) (*queries.PostgresScan, error) {
	return &queries.PostgresScan{
		ID:         scanID,
		DatabaseID: q.postgresScan.DatabaseID,
	}, nil
}
