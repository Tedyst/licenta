package saver

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/nvd"
	"github.com/tedyst/licenta/scanner"
)

type BaseQuerier interface {
	GetScanGroup(ctx context.Context, id int64) (*queries.ScanGroup, error)
	GetScan(ctx context.Context, id int64) (*queries.GetScanRow, error)
	UpdateScanStatus(ctx context.Context, params queries.UpdateScanStatusParams) error
	CreateScanResult(ctx context.Context, params queries.CreateScanResultParams) (*queries.ScanResult, error)
	CreateScanBruteforceResult(ctx context.Context, arg queries.CreateScanBruteforceResultParams) (*queries.ScanBruteforceResult, error)
	UpdateScanBruteforceResult(ctx context.Context, params queries.UpdateScanBruteforceResultParams) error

	GetCvesByProductAndVersion(ctx context.Context, arg queries.GetCvesByProductAndVersionParams) ([]*queries.GetCvesByProductAndVersionRow, error)
}

type baseSaver struct {
	queries            BaseQuerier
	bruteforceProvider bruteforce.BruteforceProvider

	logger *slog.Logger

	bruteforceResults map[scanner.User]int64

	scan    *queries.Scan
	scanner scanner.Scanner

	runAfterScan func(ctx context.Context) error
}

func (saver *baseSaver) insertResults(ctx context.Context, results []scanner.ScanResult) error {
	for _, result := range results {
		if _, err := saver.queries.CreateScanResult(ctx, queries.CreateScanResultParams{
			ScanID:     saver.scan.ID,
			Severity:   int32(result.Severity()),
			Message:    result.Detail(),
			ScanSource: int32(saver.scanner.GetScannerID()),
		}); err != nil {
			return fmt.Errorf("could not insert scan result: %w", err)
		}
	}
	return nil
}

func (saver *baseSaver) bruteforceUpdateStatus(ctx context.Context) func(status map[scanner.User]bruteforce.BruteforceUserStatus) error {
	return func(status map[scanner.User]bruteforce.BruteforceUserStatus) error {
		for user, entry := range status {
			if _, ok := saver.bruteforceResults[user]; !ok {
				username, err := user.GetUsername()
				if err != nil {
					return fmt.Errorf("could not get username: %w", err)
				}
				bfuser, err := saver.queries.CreateScanBruteforceResult(ctx, queries.CreateScanBruteforceResultParams{
					ScanID:   saver.scan.ID,
					ScanType: int32(saver.scanner.GetScannerID()),
					Username: username,
					Password: sql.NullString{String: entry.FoundPassword, Valid: entry.FoundPassword != ""},
					Tried:    int32(entry.Tried),
					Total:    int32(entry.Total),
				})
				if err != nil {
					return fmt.Errorf("could not insert bruteforce result: %w", err)
				}
				saver.bruteforceResults[user] = bfuser.ID
				continue
			}

			if err := saver.queries.UpdateScanBruteforceResult(ctx, queries.UpdateScanBruteforceResultParams{
				ID:       saver.bruteforceResults[user],
				Password: sql.NullString{String: entry.FoundPassword, Valid: entry.FoundPassword != ""},
				Tried:    int32(entry.Tried),
				Total:    int32(entry.Total),
			}); err != nil {
				return fmt.Errorf("could not update bruteforce result: %w", err)
			}
		}
		return nil
	}
}

func createBaseSaver(q BaseQuerier, bruteforceProvider bruteforce.BruteforceProvider, logger *slog.Logger, scan *queries.Scan, sc scanner.Scanner) *baseSaver {
	return &baseSaver{
		queries:            q,
		bruteforceProvider: bruteforceProvider,
		logger:             logger,
		scan:               scan,
		scanner:            sc,
		bruteforceResults:  map[scanner.User]int64{},
	}
}

func (runner *baseSaver) failScan(ctx context.Context, err error) error {
	if err2 := runner.queries.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
		ID:     runner.scan.ID,
		Status: models.SCAN_FINISHED,
		Error:  sql.NullString{String: err.Error(), Valid: true},
		EndedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}); err2 != nil {
		return fmt.Errorf("could not update scan status: %w; %w", err, err2)
	}
	return err
}

func (runner *baseSaver) Scan(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "run")
	defer span.End()
	runner.logger.DebugContext(ctx, "Starting scan")

	if err := runner.queries.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
		ID:     runner.scan.ID,
		Status: models.SCAN_RUNNING,
	}); err != nil {
		return fmt.Errorf("could not update scan status: %w", err)
	}

	if _, err := runner.queries.CreateScanResult(ctx, queries.CreateScanResultParams{
		ScanID:     runner.scan.ID,
		Severity:   int32(scanner.SEVERITY_INFORMATIONAL),
		Message:    "Started scanning",
		ScanSource: int32(runner.scanner.GetScannerID()),
	}); err != nil {
		return fmt.Errorf("could not insert scan result: %w", err)
	}

	if err := runner.runScanner(ctx); err != nil {
		return runner.failScan(ctx, err)
	}

	if runner.runAfterScan != nil {
		if err := runner.runAfterScan(ctx); err != nil {
			return runner.failScan(ctx, fmt.Errorf("could not run function after scan: %w", err))
		}
	}

	if err := runner.queries.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
		ID:     runner.scan.ID,
		Status: models.SCAN_FINISHED,
		EndedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}); err != nil {
		return fmt.Errorf("could not update scan status: %w", err)
	}

	runner.logger.DebugContext(ctx, "Finished scan")

	if _, err := runner.queries.CreateScanResult(ctx, queries.CreateScanResultParams{
		ScanID:     runner.scan.ID,
		Severity:   int32(scanner.SEVERITY_INFORMATIONAL),
		Message:    "Finished the scan",
		ScanSource: int32(runner.scanner.GetScannerID()),
	}); err != nil {
		return fmt.Errorf("could not insert scan result: %w", err)
	}

	return nil
}

func (runner *baseSaver) runScanner(ctx context.Context) error {
	if err := runner.scanner.Ping(ctx); err != nil && err != scanner.ErrPingNotSupported {
		return fmt.Errorf("could not ping database: %w", err)
	}

	runner.logger.DebugContext(ctx, "Pinged database")

	if err := runner.scanner.CheckPermissions(ctx); err != nil && err != scanner.ErrCheckPermissionsNotSupported {
		return fmt.Errorf("could not check permissions: %w", err)
	}

	runner.logger.DebugContext(ctx, "Checked permissions")

	results, err := runner.scanner.ScanConfig(ctx)
	if err != nil && err != scanner.ErrScanConfigNotSupported {
		return fmt.Errorf("could not scan config: %w", err)
	}
	if err := runner.insertResults(ctx, results); err != nil {
		return fmt.Errorf("could not insert scan results: %w", err)
	}

	runner.logger.DebugContext(ctx, "Scanned config")

	_, err = runner.scanner.GetUsers(ctx)
	if err != nil && err != scanner.ErrGetUsersNotSupported {
		return fmt.Errorf("could not get users: %w", err)
	}

	runner.logger.DebugContext(ctx, "Got users")

	version, err := runner.scanner.GetVersion(ctx)
	if err != nil && err != scanner.ErrVersionNotSupported {
		return fmt.Errorf("could not get version: %w", err)
	}

	if err != scanner.ErrVersionNotSupported && runner.scanner.GetNvdProductType() != nvd.PRODUCT_UNKNOWN {
		runner.logger.DebugContext(ctx, "Got version")

		cves, err := runner.queries.GetCvesByProductAndVersion(ctx, queries.GetCvesByProductAndVersionParams{
			DatabaseType: int32(runner.scanner.GetNvdProductType()),
			Version:      version,
		})
		if err != nil {
			return fmt.Errorf("could not get cves: %w", err)
		}

		for _, cve := range cves {
			if _, err := runner.queries.CreateScanResult(ctx, queries.CreateScanResultParams{
				ScanID:     runner.scan.ID,
				Severity:   int32(scanner.SEVERITY_HIGH),
				Message:    fmt.Sprintf("Vulnerability %s found. Please update to the latest version", cve.NvdCfe.CveID),
				ScanSource: int32(runner.scanner.GetScannerID()),
			}); err != nil {
				return fmt.Errorf("could not insert scan result: %w", err)
			}
		}

		runner.logger.DebugContext(ctx, "Verified version for CVEs")
	}

	if err := runner.bruteforce(ctx); err != nil {
		return fmt.Errorf("could not bruteforce: %w", err)
	}

	return nil
}

func (r *baseSaver) bruteforce(ctx context.Context) error {
	r.logger.DebugContext(ctx, "Bruteforcing passwords for all users")

	scangroup, err := r.queries.GetScanGroup(ctx, r.scan.ScanGroupID)
	if err != nil {
		return fmt.Errorf("could not get scan group: %w", err)
	}
	bruteforcer, err := r.bruteforceProvider.NewBruteforcer(ctx, r.scanner, r.bruteforceUpdateStatus(ctx), scangroup.ProjectID)
	if err != nil {
		return fmt.Errorf("could not create bruteforcer: %w", err)
	}

	bruteforceResult, err := bruteforcer.BruteforcePasswordAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("could not bruteforce passwords: %w", err)
	}
	if err := r.insertResults(ctx, bruteforceResult); err != nil {
		return fmt.Errorf("could not insert bruteforce results: %w", err)
	}

	r.logger.DebugContext(ctx, "Bruteforced passwords for all users")

	return nil
}

func (runner *baseSaver) ScanForPublicAccessOnly(ctx context.Context) error {
	if !runner.scanner.ShouldNotBePublic() {
		return nil
	}

	if _, err := runner.queries.CreateScanResult(ctx, queries.CreateScanResultParams{
		ScanID:     runner.scan.ID,
		Severity:   int32(scanner.SEVERITY_INFORMATIONAL),
		Message:    "Started checking for public access",
		ScanSource: int32(runner.scanner.GetScannerID()),
	}); err != nil {
		return fmt.Errorf("could not insert scan result: %w", err)
	}

	if err := runner.queries.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
		ID:     runner.scan.ID,
		Status: models.SCAN_CHECKING_PUBLIC_ACCESS,
	}); err != nil {
		return fmt.Errorf("could not update scan status: %w", err)
	}

	if err := runner.scanner.Ping(ctx); err != nil && err != scanner.ErrPingNotSupported {
		return nil
	}

	runner.logger.DebugContext(ctx, "Pinged database from public access")

	if _, err := runner.queries.CreateScanResult(ctx, queries.CreateScanResultParams{
		ScanID:     runner.scan.ID,
		Severity:   int32(scanner.SEVERITY_HIGH),
		Message:    "Database is accessible from public internet",
		ScanSource: int32(runner.scanner.GetScannerID()),
	}); err != nil {
		return fmt.Errorf("could not insert scan result: %w", err)
	}

	if _, err := runner.queries.CreateScanResult(ctx, queries.CreateScanResultParams{
		ScanID:     runner.scan.ID,
		Severity:   int32(scanner.SEVERITY_INFORMATIONAL),
		Message:    "Finished checking for public access. Proceeding with queuing to a project specific worker",
		ScanSource: int32(runner.scanner.GetScannerID()),
	}); err != nil {
		return fmt.Errorf("could not insert scan result: %w", err)
	}

	return nil
}

func CreateBaseScan[T any](
	getDatabasesQuerier func(q BaseCreater) (func(context.Context, int64) ([]T, error), error),
	createSpecificScan func(ctx context.Context, q BaseCreater, scanID int64, db T) (any, error),
	scanType int32,
) CreaterFunc {
	return func(ctx context.Context, q BaseCreater, projectID int64, scanGroupID int64) ([]*queries.Scan, error) {
		scans := []*queries.Scan{}

		getDatabases, err := getDatabasesQuerier(q)
		if err != nil {
			return nil, fmt.Errorf("could not get databases querier: %w", err)
		}
		databases, err := getDatabases(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("error getting Mysql databases for project: %w", err)
		}
		for _, db := range databases {
			scan, err := q.CreateScan(ctx, queries.CreateScanParams{
				Status:      models.SCAN_NOT_STARTED,
				ScanGroupID: scanGroupID,
				ScanType:    scanType,
			})
			if err != nil {
				return nil, fmt.Errorf("error creating Mysql scan: %w", err)
			}

			_, err = createSpecificScan(ctx, q, scan.ID, db)
			if err != nil {
				return nil, fmt.Errorf("error creating Mysql scan: %w", err)
			}

			scans = append(scans, scan)
		}
		return scans, nil
	}
}
