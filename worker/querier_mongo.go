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

func (q *remoteQuerier) GetMongoDatabase(ctx context.Context, r queries.GetMongoDatabaseParams) (*queries.GetMongoDatabaseRow, error) {
	response, err := q.client.GetMongoIdWithResponse(ctx, r.ID)
	if err != nil {
		return nil, errors.New("cannot get Mongo database from server")
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

	switch response.StatusCode() {
	case http.StatusOK:
		return &queries.GetMongoDatabaseRow{
			ID:           int64(response.JSON200.MongoDatabase.Id),
			ProjectID:    int64(response.JSON200.MongoDatabase.ProjectId),
			Host:         response.JSON200.MongoDatabase.Host,
			Port:         int32(response.JSON200.MongoDatabase.Port),
			DatabaseName: response.JSON200.MongoDatabase.DatabaseName,
			Username:     response.JSON200.MongoDatabase.Username,
			Password:     response.JSON200.MongoDatabase.Password,
			Version: sql.NullString{
				String: response.JSON200.MongoDatabase.Version,
				Valid:  response.JSON200.MongoDatabase.Version != "",
			},
		}, nil
	default:
		return nil, errors.New("error getting Mongo database")
	}
}

func (q *remoteQuerier) UpdateMongoVersion(ctx context.Context, params queries.UpdateMongoVersionParams) error {
	response, err := q.client.PatchMongoIdWithResponse(ctx, params.ID, generated.PatchMongoDatabase{
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
		return errors.New("error updating Mongo version")
	}
}

func (q *remoteQuerier) GetMongoScanByScanID(ctx context.Context, scanID int64) (*queries.MongoScan, error) {
	response, err := q.client.GetMongoScansWithResponse(ctx, &generated.GetMongoScansParams{
		Scan: scanID,
	})
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

	switch response.StatusCode() {
	case http.StatusOK:
		return &queries.MongoScan{
			ID:         int64(response.JSON200.Scans[0].Id),
			ScanID:     scanID,
			DatabaseID: int64(response.JSON200.Scans[0].DatabaseId),
		}, nil
	case http.StatusNotFound:
		return nil, pgx.ErrNoRows
	default:
		return nil, errors.New("error getting Mongo scan")
	}
}

var _ saver.MongoQuerier = &remoteQuerier{}
