package saver

import (
	"context"
	"errors"

	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
)

type Saver interface {
	ScanForPublicAccessOnly(context.Context) error
	Scan(context.Context) error
}

type CreateSaverFunc func(context.Context, BaseQuerier, bruteforce.BruteforceProvider, *queries.Scan) (Saver, error)

var ErrSaverNotNeeded = errors.New("saver not needed")

var savers = map[string]CreateSaverFunc{
	"postgres": NewPostgresSaver,
}

func NewSaver(ctx context.Context, queries BaseQuerier, bruteforceProvider bruteforce.BruteforceProvider, scan *queries.Scan, scanType string) (Saver, error) {
	if scanType == "all" {
		for _, createSaver := range savers {
			saver, err := createSaver(ctx, queries, bruteforceProvider, scan)
			if err != nil && !errors.Is(err, ErrSaverNotNeeded) {
				return nil, err
			}
			if err == nil {
				return saver, nil
			}
		}
	}
	createSaver, ok := savers[scanType]
	if !ok {
		return nil, errors.New("invalid scan type")
	}
	return createSaver(ctx, queries, bruteforceProvider, scan)
}
