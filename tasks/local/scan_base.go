package local

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/nvd"
	"github.com/tedyst/licenta/scanner"
)

type scanQuerier interface {
	GetScan(ctx context.Context, id int64) (*queries.GetScanRow, error)
	UpdateScanStatus(ctx context.Context, params queries.UpdateScanStatusParams) error
	CreateScanResult(ctx context.Context, params queries.CreateScanResultParams) (*queries.ScanResult, error)
	CreateScanBruteforceResult(ctx context.Context, arg queries.CreateScanBruteforceResultParams) (*models.ScanBruteforceResult, error)
	UpdateScanBruteforceResult(ctx context.Context, params queries.UpdateScanBruteforceResultParams) error

	GetCvesByProductAndVersion(ctx context.Context, arg queries.GetCvesByProductAndVersionParams) ([]*queries.GetCvesByProductAndVersionRow, error)
}

type baseScanRunner struct {
	queries            scanQuerier
	bruteforceProvider bruteforce.BruteforceProvider

	logger *slog.Logger

	insertResults func([]scanner.ScanResult) error

	bruteforceResults      map[scanner.User]int64
	notifyBruteforceStatus func(status map[scanner.User]bruteforce.BruteforceUserStatus) error

	scan    *models.Scan
	scanner scanner.Scanner
}

func createScanner(ctx context.Context, q scanQuerier, bruteforceProvider bruteforce.BruteforceProvider, logger *slog.Logger, scan *models.Scan, sc scanner.Scanner) *baseScanRunner {
	runner := &baseScanRunner{
		queries:            q,
		bruteforceProvider: bruteforceProvider,
		logger:             logger,
		scan:               scan,
		scanner:            sc,
		bruteforceResults:  map[scanner.User]int64{},
	}

	runner.insertResults = func(results []scanner.ScanResult) error {
		for _, result := range results {
			if _, err := runner.queries.CreateScanResult(ctx, queries.CreateScanResultParams{
				ScanID:     scan.ID,
				Severity:   int32(result.Severity()),
				Message:    result.Detail(),
				ScanSource: int32(sc.GetScannerID()),
			}); err != nil {
				return errors.Wrap(err, "could not insert scan result")
			}
		}
		return nil
	}

	runner.notifyBruteforceStatus = func(status map[scanner.User]bruteforce.BruteforceUserStatus) error {
		for user, entry := range status {
			if _, ok := runner.bruteforceResults[user]; !ok {
				username, err := user.GetUsername()
				if err != nil {
					return errors.Wrap(err, "could not get username")
				}
				bfuser, err := runner.queries.CreateScanBruteforceResult(ctx, queries.CreateScanBruteforceResultParams{
					ScanID:   runner.scan.ID,
					Username: username,
					Password: sql.NullString{String: entry.FoundPassword, Valid: entry.FoundPassword != ""},
					Tried:    int32(entry.Tried),
					Total:    int32(entry.Total),
				})
				if err != nil {
					return errors.Wrap(err, "could not insert bruteforce result")
				}
				runner.bruteforceResults[user] = bfuser.ID
			} else {
				if err := runner.queries.UpdateScanBruteforceResult(ctx, queries.UpdateScanBruteforceResultParams{
					ID:       runner.bruteforceResults[user],
					Password: sql.NullString{String: entry.FoundPassword, Valid: entry.FoundPassword != ""},
					Tried:    int32(entry.Tried),
					Total:    int32(entry.Total),
				}); err != nil {
					return errors.Wrap(err, "could not update bruteforce result")
				}
			}
		}
		return nil
	}

	return runner
}

func (runner *baseScanRunner) run(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "run")
	defer span.End()

	runner.insertResults = func(results []scanner.ScanResult) error {
		for _, result := range results {
			if _, err := runner.queries.CreateScanResult(ctx, queries.CreateScanResultParams{
				ScanID:     runner.scan.ID,
				Severity:   int32(result.Severity()),
				Message:    result.Detail(),
				ScanSource: int32(runner.scanner.GetScannerID()),
			}); err != nil {
				return errors.Wrap(err, "could not insert scan result")
			}
		}
		return nil
	}

	runner.logger.DebugContext(ctx, "Starting scan")

	if err := runner.runScanner(ctx); err != nil {
		return err
	}

	runner.logger.DebugContext(ctx, "Finished scan")

	return nil
}

func (runner *baseScanRunner) runScanner(ctx context.Context) error {
	if err := runner.queries.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
		ID:     runner.scan.ID,
		Status: models.SCAN_RUNNING,
	}); err != nil {
		return errors.Wrap(err, "could not update scan status")
	}

	if err := runner.scanner.Ping(ctx); err != nil && err != scanner.ErrPingNotSupported {
		return errors.Wrap(err, "could not ping database")
	}

	runner.logger.DebugContext(ctx, "Pinged database")

	if err := runner.scanner.CheckPermissions(ctx); err != nil && err != scanner.ErrCheckPermissionsNotSupported {
		return errors.Wrap(err, "could not check permissions")
	}

	runner.logger.DebugContext(ctx, "Checked permissions")

	results, err := runner.scanner.ScanConfig(ctx)
	if err != nil && err != scanner.ErrScanConfigNotSupported {
		return errors.Wrap(err, "could not scan config")
	}
	if err := runner.insertResults(results); err != nil {
		return errors.Wrap(err, "could not insert scan results")
	}

	runner.logger.DebugContext(ctx, "Scanned config")

	_, err = runner.scanner.GetUsers(ctx)
	if err != nil && err != scanner.ErrGetUsersNotSupported {
		return errors.Wrap(err, "could not get users")
	}

	runner.logger.DebugContext(ctx, "Got users")

	version, err := runner.scanner.GetVersion(ctx)
	if err != nil && err != scanner.ErrVersionNotSupported {
		return errors.Wrap(err, "could not get version")
	}

	if err != scanner.ErrVersionNotSupported && runner.scanner.GetNvdProductType() != nvd.PRODUCT_UNKNOWN {
		runner.logger.DebugContext(ctx, "Got version")

		cves, err := runner.queries.GetCvesByProductAndVersion(ctx, queries.GetCvesByProductAndVersionParams{
			DatabaseType: int32(runner.scanner.GetNvdProductType()),
			Version:      version,
		})
		if err != nil {
			return errors.Wrap(err, "could not get cves")
		}

		for _, cve := range cves {
			if _, err := runner.queries.CreateScanResult(ctx, queries.CreateScanResultParams{
				ScanID:     runner.scan.ID,
				Severity:   int32(scanner.SEVERITY_HIGH),
				Message:    fmt.Sprintf("Vulnerability %s found. Please update to the latest version", cve.NvdCfe.CveID),
				ScanSource: int32(runner.scanner.GetScannerID()),
			}); err != nil {
				return errors.Wrap(err, "could not insert scan result")
			}
		}

		runner.logger.DebugContext(ctx, "Verified version for CVEs")
	}

	if err := runner.bruteforce(ctx); err != nil {
		return errors.Wrap(err, "could not bruteforce")
	}

	return nil
}

func (r *baseScanRunner) bruteforce(ctx context.Context) error {
	r.logger.DebugContext(ctx, "Bruteforcing passwords for all users")

	bruteforcer, err := r.bruteforceProvider.NewBruteforcer(ctx, r.scanner, r.notifyBruteforceStatus, r.scan.ProjectID)
	if err != nil {
		return errors.Wrap(err, "could not create bruteforcer")
	}

	bruteforceResult, err := bruteforcer.BruteforcePasswordAllUsers(ctx)
	if err != nil {
		return errors.Wrap(err, "could not bruteforce passwords")
	}
	if err := r.insertResults(bruteforceResult); err != nil {
		return errors.Wrap(err, "could not insert bruteforce results")
	}

	r.logger.DebugContext(ctx, "Bruteforced passwords for all users")

	return nil
}

func (runner *baseScanRunner) scanForPublicAccess(ctx context.Context) error {
	if !runner.scanner.ShouldNotBePublic() {
		return nil
	}

	if err := runner.queries.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
		ID:     runner.scan.ID,
		Status: models.SCAN_CHECKING_PUBLIC_ACCESS,
	}); err != nil {
		return errors.Wrap(err, "could not update scan status")
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
		return errors.Wrap(err, "could not insert scan result")
	}

	return nil
}
