package worker

import (
	"context"
	"database/sql"
	"log/slog"

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
	return nil
}

func (q *remotePostgresQuerier) CreatePostgresScanResult(ctx context.Context, params queries.CreatePostgresScanResultParams) (*models.PostgresScanResult, error) {
	slog.InfoContext(ctx, "Creating scan result", "params", params)
	return &queries.PostgresScanResult{
		ID:             1,
		PostgresScanID: 1,
		Severity:       1,
		Message:        "asdasd",
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
