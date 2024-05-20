package bruteforce

import (
	"context"
	"fmt"

	"github.com/tedyst/licenta/scanner"
)

type BruteforceProvider interface {
	NewBruteforcer(ctx context.Context, sc scanner.Scanner, statusFunc StatusFunc, projectID int64) (Bruteforcer, error)
}

type Bruteforcer interface {
	BruteforcePasswordAllUsers(ctx context.Context) ([]scanner.ScanResult, error)
}

type databaseBruteforceProvider struct {
	queries DatabasePasswordProviderInterface
}

var _ BruteforceProvider = (*databaseBruteforceProvider)(nil)

func NewDatabaseBruteforceProvider(queries DatabasePasswordProviderInterface) *databaseBruteforceProvider {
	return &databaseBruteforceProvider{
		queries: queries,
	}
}

func (d *databaseBruteforceProvider) NewBruteforcer(ctx context.Context, sc scanner.Scanner, statusFunc StatusFunc, projectID int64) (Bruteforcer, error) {
	passProvider, err := NewDatabasePasswordProvider(ctx, d.queries, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create password provider: %w", err)
	}
	defer passProvider.Close()

	return NewBruteforcer(passProvider, sc, statusFunc), nil
}
