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

func (q *remoteQuerier) GetRedisDatabase(ctx context.Context, r queries.GetRedisDatabaseParams) (*queries.GetRedisDatabaseRow, error) {
	response, err := q.client.GetRedisIdWithResponse(ctx, r.ID)
	if err != nil {
		return nil, errors.New("cannot get Redis database from server")
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

	switch response.StatusCode() {
	case http.StatusOK:
		return &queries.GetRedisDatabaseRow{
			ID:        int64(response.JSON200.RedisDatabase.Id),
			ProjectID: int64(response.JSON200.RedisDatabase.ProjectId),
			Host:      response.JSON200.RedisDatabase.Host,
			Port:      int32(response.JSON200.RedisDatabase.Port),
			Username:  response.JSON200.RedisDatabase.Username,
			Password:  response.JSON200.RedisDatabase.Password,
			Version: sql.NullString{
				String: response.JSON200.RedisDatabase.Version,
				Valid:  response.JSON200.RedisDatabase.Version != "",
			},
		}, nil
	default:
		return nil, errors.New("error getting Redis database")
	}
}

func (q *remoteQuerier) UpdateRedisVersion(ctx context.Context, params queries.UpdateRedisVersionParams) error {
	response, err := q.client.PatchRedisIdWithResponse(ctx, params.ID, generated.PatchRedisDatabase{
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
		return errors.New("error updating Redis version")
	}
}

func (q *remoteQuerier) GetRedisScanByScanID(ctx context.Context, scanID int64) (*queries.RedisScan, error) {
	response, err := q.client.GetRedisScansWithResponse(ctx, &generated.GetRedisScansParams{
		Scan: scanID,
	})
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

	switch response.StatusCode() {
	case http.StatusOK:
		return &queries.RedisScan{
			ID:         int64(response.JSON200.Scans[0].Id),
			ScanID:     scanID,
			DatabaseID: int64(response.JSON200.Scans[0].DatabaseId),
		}, nil
	case http.StatusNotFound:
		return nil, pgx.ErrNoRows
	default:
		return nil, errors.New("error getting Redis scan")
	}
}

var _ saver.RedisQuerier = &remoteQuerier{}
