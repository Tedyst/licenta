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

func (q *remoteQuerier) GetMysqlDatabase(ctx context.Context, r queries.GetMysqlDatabaseParams) (*queries.GetMysqlDatabaseRow, error) {
	response, err := q.client.GetMysqlIdWithResponse(ctx, r.ID)
	if err != nil {
		return nil, errors.New("cannot get Mysql database from server")
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

	switch response.StatusCode() {
	case http.StatusOK:
		return &queries.GetMysqlDatabaseRow{
			ID:           int64(response.JSON200.MysqlDatabase.Id),
			ProjectID:    int64(response.JSON200.MysqlDatabase.ProjectId),
			Host:         response.JSON200.MysqlDatabase.Host,
			Port:         int32(response.JSON200.MysqlDatabase.Port),
			DatabaseName: response.JSON200.MysqlDatabase.DatabaseName,
			Username:     response.JSON200.MysqlDatabase.Username,
			Password:     response.JSON200.MysqlDatabase.Password,
			Version: sql.NullString{
				String: response.JSON200.MysqlDatabase.Version,
				Valid:  response.JSON200.MysqlDatabase.Version != "",
			},
		}, nil
	default:
		return nil, errors.New("error getting Mysql database")
	}
}

func (q *remoteQuerier) UpdateMysqlVersion(ctx context.Context, params queries.UpdateMysqlVersionParams) error {
	response, err := q.client.PatchMysqlIdWithResponse(ctx, params.ID, generated.PatchMysqlDatabase{
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
		return errors.New("error updating Mysql version")
	}
}

func (q *remoteQuerier) GetMysqlScanByScanID(ctx context.Context, scanID int64) (*queries.MysqlScan, error) {
	response, err := q.client.GetMysqlScansWithResponse(ctx, &generated.GetMysqlScansParams{
		Scan: scanID,
	})
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Got response from server", "response", string(response.Body))

	switch response.StatusCode() {
	case http.StatusOK:
		return &queries.MysqlScan{
			ID:         int64(response.JSON200.Scans[0].Id),
			ScanID:     scanID,
			DatabaseID: int64(response.JSON200.Scans[0].DatabaseId),
		}, nil
	case http.StatusNotFound:
		return nil, pgx.ErrNoRows
	default:
		return nil, errors.New("error getting Mysql scan")
	}
}

var _ saver.MysqlQuerier = &remoteQuerier{}
