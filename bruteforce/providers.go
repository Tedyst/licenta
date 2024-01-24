package bruteforce

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
)

type PasswordProvider interface {
	GetCount() (int, error)
	GetSpecificPassword(password string) (int64, bool, error)
	Next() bool
	Error() error
	Current() (int64, string, error)
	Start(index int64) error
	Close()

	SavePasswordHash(username, hash, password string, maxInternalID int64) error
	GetPasswordByHash(username, hash string) (string, int64, error)
}

type databasePasswordProvider struct {
	projectID int64

	count    int
	rows     pgx.Rows
	database db.TransactionQuerier

	context context.Context
}

func (d *databasePasswordProvider) GetCount() (int, error) {
	return d.count, nil
}

func (d *databasePasswordProvider) GetSpecificPassword(password string) (int64, bool, error) {
	id, err := d.database.GetSpecificBruteforcePasswordID(d.context, queries.GetSpecificBruteforcePasswordIDParams{
		Password:  password,
		ProjectID: d.projectID,
	})
	if err == pgx.ErrNoRows {
		return -1, false, nil
	}
	return id, true, err
}

func (d *databasePasswordProvider) Next() bool {
	if d.rows == nil {
		return false
	}
	return d.rows.Next()
}

func (d *databasePasswordProvider) Error() error {
	if d.rows == nil {
		return nil
	}
	return d.rows.Err()
}

func (d *databasePasswordProvider) Current() (int64, string, error) {
	if d.rows == nil {
		return -1, "", nil
	}
	var password string
	var id int64
	err := d.rows.Scan(&id, &password)
	return id, password, err
}

func (d *databasePasswordProvider) Start(index int64) error {
	rows, err := d.database.GetRawPool().Query(context.Background(), "SELECT id, password FROM default_bruteforce_passwords WHERE id > $1 UNION all SELECT -1, password FROM project_docker_layer_results WHERE project_id = $2 UNION all SELECT -1, password FROM project_git_results WHERE project_id = $2 ORDER BY id ASC", index, d.projectID)
	if err != nil {
		return err
	}
	d.rows = rows
	return nil
}

func (d *databasePasswordProvider) Close() {
	if d.rows == nil {
		return
	}
	d.rows.Close()
}

func (d *databasePasswordProvider) SavePasswordHash(username, hash, password string, maxInternalID int64) error {
	_, err := d.database.CreateBruteforcedPassword(d.context, queries.CreateBruteforcedPasswordParams{
		Username: username,
		Hash:     hash,
		Password: sql.NullString{String: password, Valid: password != ""},
		LastBruteforceID: sql.NullInt64{
			Int64: maxInternalID,
			Valid: true,
		},
		ProjectID: sql.NullInt64{Valid: false},
	})
	return err
}

func (d *databasePasswordProvider) GetPasswordByHash(username, hash string) (string, int64, error) {
	p, err := d.database.GetBruteforcedPasswords(d.context, queries.GetBruteforcedPasswordsParams{
		Username: username,
		Hash:     hash,
		ProjectID: sql.NullInt64{
			Int64: d.projectID,
			Valid: true,
		},
	})
	if err == pgx.ErrNoRows {
		return "", 0, nil
	}
	return p.Password.String, p.LastBruteforceID.Int64, err
}

func NewDatabasePasswordProvider(ctx context.Context, database db.TransactionQuerier, projectID int64) (*databasePasswordProvider, error) {
	count := 0
	err := database.GetRawPool().QueryRow(ctx, "select SUM(count) from (SELECT COUNT(*) FROM default_bruteforce_passwords union all select COUNT(*) from project_docker_layer_results where project_id = $1 UNION all SELECT COUNT(*) FROM project_git_results WHERE project_id = $1) as count;", projectID).Scan(&count)
	if err != nil {
		return nil, err
	}

	return &databasePasswordProvider{
		projectID: projectID,
		count:     count,
		database:  database,
		context:   ctx,
	}, nil
}

var _ PasswordProvider = (*databasePasswordProvider)(nil)

type passwordHashes struct {
	hash     string
	username string
	password string
}

type passwordListProvider struct {
	passwordHashes []passwordHashes
	passwords      []string
	index          int
}

func NewPasswordListIterator(passwords []string) *passwordListProvider {
	return &passwordListProvider{
		passwords: passwords,
		index:     0,
	}
}

func (p *passwordListProvider) GetCount() (int, error) {
	return len(p.passwords), nil
}

func (p *passwordListProvider) GetSpecificPassword(password string) (int64, bool, error) {
	for i, pass := range p.passwords {
		if pass == password {
			return int64(i), true, nil
		}
	}
	return 0, false, nil
}

func (p *passwordListProvider) Next() bool {
	if p.index >= len(p.passwords) {
		return false
	}
	p.index++
	return true
}

func (p *passwordListProvider) Error() error {
	return nil
}

func (p *passwordListProvider) Current() (int64, string, error) {
	return int64(p.index - 1), p.passwords[p.index-1], nil
}

func (p *passwordListProvider) Start(index int64) error {
	if index < 0 || index >= int64(len(p.passwords)) {
		return nil
	}
	p.index = int(index)
	return nil
}

func (p *passwordListProvider) Close() {

}

func (p *passwordListProvider) SavePasswordHash(username, hash, password string, maxInternalID int64) error {
	p.passwordHashes = append(p.passwordHashes, passwordHashes{
		hash:     hash,
		username: username,
		password: password,
	})
	return nil
}

func (p *passwordListProvider) GetPasswordByHash(username, hash string) (string, int64, error) {
	for _, h := range p.passwordHashes {
		if h.hash == hash && h.username == username {
			return h.password, 0, nil
		}
	}
	return "", 0, nil
}

var _ PasswordProvider = (*passwordListProvider)(nil)
