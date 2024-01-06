package bruteforce

import (
	"context"

	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/scanner"
)

type BruteforceProvider interface {
	NewBruteforcer(ctx context.Context, sc scanner.Scanner, statusFunc StatusFunc, projectID int64) (Bruteforcer, error)
}

type Bruteforcer interface {
	BruteforcePasswordAllUsers(ctx context.Context) ([]scanner.ScanResult, error)
}

type databaseBruteforceProvider struct {
	queries db.TransactionQuerier
}

var _ BruteforceProvider = (*databaseBruteforceProvider)(nil)

func NewDatabaseBruteforceProvider(queries db.TransactionQuerier) *databaseBruteforceProvider {
	return &databaseBruteforceProvider{
		queries: queries,
	}
}

func (d *databaseBruteforceProvider) NewBruteforcer(ctx context.Context, sc scanner.Scanner, statusFunc StatusFunc, projectID int64) (Bruteforcer, error) {
	passProvider, err := NewDatabasePasswordProvider(ctx, d.queries, projectID)
	if err != nil {
		return nil, err
	}
	defer passProvider.Close()

	return NewBruteforcer(passProvider, sc, statusFunc), nil
}
