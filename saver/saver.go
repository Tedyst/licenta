package saver

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/scanner"
)

type Saver interface {
	ScanForPublicAccessOnly(context.Context) error
	Scan(context.Context) error
}

type BaseCreater interface {
	CreateScan(ctx context.Context, params queries.CreateScanParams) (*queries.Scan, error)
}

type CreateSaverFunc func(context.Context, BaseQuerier, bruteforce.BruteforceProvider, *queries.Scan, bool, string) (Saver, error)
type CreaterFunc func(ctx context.Context, baseCreater BaseCreater, projectID int64, scanGroupID int64) ([]*queries.Scan, error)

var ErrSaverNotNeeded = errors.New("saver not needed")

var savers = map[string]CreateSaverFunc{}
var creaters = map[string]CreaterFunc{}

func NewSaver(ctx context.Context, q BaseQuerier, bruteforceProvider bruteforce.BruteforceProvider, scan *queries.Scan, scanType string, projectIsRemote bool, saltKey string) (Saver, error) {
	if scanType == "all" {
		for _, createSaver := range savers {
			saver, err := createSaver(ctx, q, bruteforceProvider, scan, projectIsRemote, saltKey)
			if err != nil && !errors.Is(err, ErrSaverNotNeeded) {
				_, err2 := q.CreateScanResult(ctx, queries.CreateScanResultParams{
					ScanID:     scan.ID,
					Severity:   int32(scanner.SEVERITY_HIGH),
					Message:    "Failed to run scan. Please try again later. Error message: " + err.Error(),
					ScanSource: -1,
				})
				if err2 != nil {
					return nil, errors.Join(err, err2)
				}
				err2 = q.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
					ID:      scan.ID,
					Status:  models.SCAN_FINISHED,
					Error:   sql.NullString{String: err.Error(), Valid: true},
					EndedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				})
				if err2 != nil {
					return nil, errors.Join(err, err2)
				}
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
	return createSaver(ctx, q, bruteforceProvider, scan, projectIsRemote, saltKey)
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
