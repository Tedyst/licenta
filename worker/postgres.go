package worker

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	localmessages "github.com/tedyst/licenta/messages/local"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/scanner"
	"github.com/tedyst/licenta/tasks/local"
)

type remotePostgresQuerier struct {
	remoteURL string
	authToken string
	task      Task
}

func (q *remotePostgresQuerier) GetPostgresDatabase(ctx context.Context, id int64) (*models.PostgresDatabases, error) {
	return &q.task.PostgresScan.Database, nil
}

func (q *remotePostgresQuerier) UpdatePostgresScanStatus(ctx context.Context, params queries.UpdatePostgresScanStatusParams) error {
	slog.InfoContext(ctx, "Updating scan status", "params", params)

	remoteURL := q.remoteURL + "/api/v1/project/" + strconv.Itoa(int(q.task.PostgresScan.Database.ID)) + "/scanner/postgres/" + strconv.Itoa(int(params.ID)) + "/"

	req, err := http.NewRequest("PATCH", remoteURL, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	req.Header.Set("X-Worker-Token", q.authToken)

	body := generated.PatchProjectProjectidScannerPostgresScanidJSONRequestBody{
		Status:  int(params.Status),
		EndedAt: params.EndedAt.Time.Format(time.RFC3339),
	}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req.Body = io.NopCloser(bytes.NewReader(data))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	var response generated.PatchProjectProjectidScannerPostgresScanid200JSONResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	if !response.Success {
		return errors.New("error updating scan status")
	}

	slog.InfoContext(ctx, "Received response", "status", resp.StatusCode)

	return nil
}

func (q *remotePostgresQuerier) CreatePostgresScanResult(ctx context.Context, params queries.CreatePostgresScanResultParams) (*models.PostgresScanResult, error) {
	slog.InfoContext(ctx, "Creating scan result", "params", params)

	remoteURL := q.remoteURL + "/api/v1/project/" + strconv.Itoa(int(q.task.PostgresScan.Database.ID)) + "/scanner/postgres/" + strconv.Itoa(int(params.PostgresScanID)) + "/result"

	req, err := http.NewRequest("POST", remoteURL, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("X-Worker-Token", q.authToken)

	body := generated.PostProjectProjectidScannerPostgresScanidResultJSONRequestBody{
		Message:  params.Message,
		Severity: int(params.Severity),
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req.Body = io.NopCloser(bytes.NewReader(data))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var response generated.PostProjectProjectidScannerPostgresScanidResult200JSONResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, errors.New("error creating scan result")
	}

	slog.InfoContext(ctx, "Received response", "status", resp.StatusCode)

	t, err := time.Parse(time.RFC3339, response.Scan.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &models.PostgresScanResult{
		ID:             int64(response.Scan.Id),
		PostgresScanID: int64(response.Scan.PostgresScanId),
		Severity:       int32(response.Scan.Severity),
		Message:        response.Scan.Message,
		CreatedAt:      pgtype.Timestamptz{Time: t},
	}, nil
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

type remoteBruteforceProvider struct {
	remoteURL string
	authToken string
	task      Task
}

type remotePasswordProvider struct {
	remoteURL string
	authToken string
	task      Task
}

func (p *remotePasswordProvider) GetCount() (int, error) {
	return 0, nil
}

func (p *remotePasswordProvider) GetSpecificPassword(password string) (int64, bool, error) {
	return 0, false, nil
}

func (p *remotePasswordProvider) Next() bool {
	return false
}

func (p *remotePasswordProvider) Error() error {
	return nil
}

func (p *remotePasswordProvider) Current() (int64, string, error) {
	return 0, "", nil
}

func (p *remotePasswordProvider) Start(index int64) error {
	return nil
}

func (p *remotePasswordProvider) Close() {

}

func (p *remotePasswordProvider) SavePasswordHash(username, hash, password string, maxInternalID int64) error {
	return nil
}

func (p *remotePasswordProvider) GetPasswordByHash(username, hash string) (string, int64, error) {
	return "", 0, nil
}

func (p *remoteBruteforceProvider) NewBruteforcer(ctx context.Context, sc scanner.Scanner, statusFunc bruteforce.StatusFunc, projectID int) (bruteforce.Bruteforcer, error) {
	return bruteforce.NewBruteforcer(&remotePasswordProvider{
		remoteURL: p.remoteURL,
		authToken: p.authToken,
		task:      p.task,
	}, sc, statusFunc), nil
}

func ScanPostgresDB(ctx context.Context, remoteURL string, authToken string, task Task) error {
	localexchange := localmessages.NewLocalExchange()
	runner := local.NewScannerRunner(&remotePostgresQuerier{
		remoteURL: remoteURL,
		authToken: authToken,
		task:      task,
	}, &remoteBruteforceProvider{
		remoteURL: remoteURL,
		authToken: authToken,
		task:      task,
	}, localexchange)
	return runner.ScanPostgresDB(ctx, &task.PostgresScan.Scan)
}
