package bruteforce

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/db/queries"
)

const MAX_PASSWORDS_PER_BATCH = 10000

type DatabasePasswordProviderInterface interface {
	GetBruteforcePasswordsPaginated(ctx context.Context, arg queries.GetBruteforcePasswordsPaginatedParams) ([]*queries.DefaultBruteforcePassword, error)
	GetSpecificBruteforcePasswordID(ctx context.Context, arg queries.GetSpecificBruteforcePasswordIDParams) (int64, error)
	CreateBruteforcedPassword(ctx context.Context, arg queries.CreateBruteforcedPasswordParams) (*queries.BruteforcedPassword, error)
	GetBruteforcedPasswords(ctx context.Context, arg queries.GetBruteforcedPasswordsParams) (*queries.BruteforcedPassword, error)
	GetBruteforcePasswordsForProjectCount(ctx context.Context, projectID int64) (int64, error)
	UpdateBruteforcedPassword(ctx context.Context, arg queries.UpdateBruteforcedPasswordParams) (*queries.BruteforcedPassword, error)
}

type PasswordProvider interface {
	GetCount() (int64, error)
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

	context  context.Context
	database DatabasePasswordProviderInterface
	error    error

	count        int64
	total        int64
	currentBatch []*queries.DefaultBruteforcePassword
	firstItem    bool
}

func (p *databasePasswordProvider) readBatch() error {
	if len(p.currentBatch) > 1 {
		return nil
	}

	lastID := int64(-1)
	if len(p.currentBatch) > 0 {
		lastID = p.currentBatch[len(p.currentBatch)-1].ID
	}
	response, err := p.database.GetBruteforcePasswordsPaginated(p.context, queries.GetBruteforcePasswordsPaginatedParams{
		LastID: lastID,
		Limit:  MAX_PASSWORDS_PER_BATCH,
	})
	if err != nil {
		return err
	}

	p.currentBatch = append(p.currentBatch, response...)
	return nil
}

func (d *databasePasswordProvider) GetCount() (int64, error) {
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
	err := d.readBatch()
	if err != nil {
		d.error = err
		return false
	}

	if len(d.currentBatch) == 0 {
		return false
	}
	if !d.firstItem {
		d.firstItem = true
		return true
	}
	d.currentBatch = d.currentBatch[1:]

	return len(d.currentBatch) != 0
}

func (d *databasePasswordProvider) Error() error {
	return d.error
}

func (d *databasePasswordProvider) Current() (int64, string, error) {
	return d.currentBatch[0].ID, d.currentBatch[0].Password, nil
}

func (d *databasePasswordProvider) Start(index int64) error {
	d.currentBatch = []*queries.DefaultBruteforcePassword{}
	d.firstItem = false
	d.count = 0
	return nil
}

func (d *databasePasswordProvider) Close() {

}

func (d *databasePasswordProvider) SavePasswordHash(username, hash, password string, maxInternalID int64) error {
	oldPW, err := d.database.GetBruteforcedPasswords(d.context, queries.GetBruteforcedPasswordsParams{
		Username: username,
		Hash:     hash,
		ProjectID: sql.NullInt64{
			Int64: d.projectID,
			Valid: d.projectID != 0,
		},
	})
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	if err != pgx.ErrNoRows {
		_, err = d.database.UpdateBruteforcedPassword(d.context, queries.UpdateBruteforcedPasswordParams{
			ID: oldPW.ID,
			LastBruteforceID: sql.NullInt64{
				Int64: maxInternalID,
				Valid: maxInternalID != 0,
			},
			Password: sql.NullString{String: password, Valid: password != ""},
		})
		return err
	}
	_, err = d.database.CreateBruteforcedPassword(d.context, queries.CreateBruteforcedPasswordParams{
		Username: username,
		Hash:     hash,
		Password: sql.NullString{String: password, Valid: password != ""},
		LastBruteforceID: sql.NullInt64{
			Int64: maxInternalID,
			Valid: maxInternalID != 0,
		},
		ProjectID: sql.NullInt64{
			Int64: d.projectID,
			Valid: d.projectID != 0,
		},
	})
	return err
}

func (d *databasePasswordProvider) GetPasswordByHash(username, hash string) (string, int64, error) {
	p, err := d.database.GetBruteforcedPasswords(d.context, queries.GetBruteforcedPasswordsParams{
		Username: username,
		Hash:     hash,
		ProjectID: sql.NullInt64{
			Int64: d.projectID,
			Valid: d.projectID != 0,
		},
	})
	if err == pgx.ErrNoRows {
		return "", 0, nil
	}
	return p.Password.String, p.LastBruteforceID.Int64, err
}

func NewDatabasePasswordProvider(ctx context.Context, database DatabasePasswordProviderInterface, projectID int64) (*databasePasswordProvider, error) {
	count, err := database.GetBruteforcePasswordsForProjectCount(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("could not get count of passwords: %w", err)
	}

	return &databasePasswordProvider{
		projectID: projectID,
		total:     count,
		database:  database,
		context:   ctx,
		count:     0,
		firstItem: false,
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

func (p *passwordListProvider) GetCount() (int64, error) {
	return int64(len(p.passwords)), nil
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
