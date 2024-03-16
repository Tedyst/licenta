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

type BaseCreater interface {
	CreateScan(ctx context.Context, params queries.CreateScanParams) (*queries.Scan, error)
}

type CreateSaverFunc func(context.Context, BaseQuerier, bruteforce.BruteforceProvider, *queries.Scan) (Saver, error)
type CreaterFunc func(ctx context.Context, baseCreater BaseCreater, projectID int64, scanGroupID int64) ([]*queries.Scan, error)

var ErrSaverNotNeeded = errors.New("saver not needed")

var savers = map[string]CreateSaverFunc{}
var creaters = map[string]CreaterFunc{}

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
		return nil, errors.New("invalid scan type")
	}
	createSaver, ok := savers[scanType]
	if !ok {
		return nil, errors.New("invalid scan type")
	}
	return createSaver(ctx, queries, bruteforceProvider, scan)
}

func CreateScans(ctx context.Context, baseCreater BaseCreater, projectID int64, scanGroupID int64, scanType string) ([]*queries.Scan, error) {
	if scanType == "all" {
		var scans []*queries.Scan
		for _, createScanner := range creaters {
			s, err := createScanner(ctx, baseCreater, projectID, scanGroupID)
			if err != nil {
				return nil, err
			}
			scans = append(scans, s...)
		}
		return scans, nil
	}
	createScanner, ok := creaters[scanType]
	if !ok {
		return nil, errors.New("invalid scan type")
	}
	return createScanner(ctx, baseCreater, projectID, scanGroupID)
}
