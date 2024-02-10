package local

import (
	"context"
	"database/sql"
	errs "errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/models"
)

type AllScannerQuerier interface {
	ScanQuerier
	PostgresQuerier

	GetProject(ctx context.Context, id int64) (*queries.Project, error)
	GetPostgresScanByScanID(ctx context.Context, scanID int64) (*queries.PostgresScan, error)

	GetWorkersForProject(ctx context.Context, projectID int64) ([]*queries.Worker, error)
}

type allScannerRunner struct {
	queries         AllScannerQuerier
	messageExchange messages.Exchange

	postgresScanRunner postgresScanRunner
}

func NewAllScannerRunner(queries AllScannerQuerier, messageExchange messages.Exchange, bruteforceProvider bruteforce.BruteforceProvider) *allScannerRunner {
	return &allScannerRunner{
		queries:            queries,
		messageExchange:    messageExchange,
		postgresScanRunner: *NewPostgresScanRunner(queries, bruteforceProvider),
	}
}

func (r *allScannerRunner) RunAllScanners(ctx context.Context, scan *queries.Scan, runningRemote bool) error {
	if err := r.runAllScanners(ctx, scan, runningRemote); err != nil {
		if err2 := r.queries.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
			ID:     scan.ID,
			Status: models.SCAN_FINISHED,
			Error:  sql.NullString{String: err.Error(), Valid: true},
			EndedAt: pgtype.Timestamptz{
				Time:  time.Now(),
				Valid: true,
			},
		}); err2 != nil {
			return errs.Join(err, errors.Wrap(err2, "could not update scan status"))
		}
		return errors.Wrap(err, "could not run all scanners")
	}

	return nil
}

func (r *allScannerRunner) runAllScanners(ctx context.Context, scan *queries.Scan, runningRemote bool) error {
	scanGroup, err := r.queries.GetScanGroup(ctx, scan.ScanGroupID)
	if err != nil {
		return errors.Wrap(err, "cannot get scan group")
	}
	project, err := r.queries.GetProject(ctx, scanGroup.ProjectID)
	if err != nil {
		return errors.Wrap(err, "cannot get project")
	}

	postgresScan, err := r.queries.GetPostgresScanByScanID(ctx, scan.ID)
	if err != nil {
		return errors.Wrap(err, "cannot get postgres scan")
	}
	if err != pgx.ErrNoRows {
		err := r.postgresScanRunner.scanPostgresDB(ctx, postgresScan, scan, project.Remote && !runningRemote)
		if err != nil {
			return errors.Wrap(err, "cannot run postgres scanner")
		}
	}

	if project.Remote && !runningRemote {
		slog.DebugContext(ctx, "Sending task to remote workers", "scan", scan.ID)

		scangroup, err := r.queries.GetScanGroup(ctx, scan.ScanGroupID)
		if err != nil {
			return errors.Wrap(err, "could not get scan group")
		}
		workers, err := r.queries.GetWorkersForProject(ctx, scangroup.ProjectID)
		if err != nil {
			return errors.Wrap(err, "could not get workers for project")
		}

		if len(workers) == 0 {
			return errors.New("no workers available")
		}

		message := messages.GetStartScanMessage(scan)

		for _, worker := range workers {
			err := r.messageExchange.PublishSendScanToWorkerMessage(ctx, worker, message)
			if err != nil {
				return errors.Wrap(err, "could not publish message")
			}
		}

		return nil
	}

	if err := r.queries.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
		ID:      scan.ID,
		Status:  models.SCAN_FINISHED,
		EndedAt: pgtype.Timestamptz{Time: time.Now(), Valid: runningRemote || !project.Remote},
	}); err != nil {
		return errors.Wrap(err, "could not update scan status")
	}

	return nil
}
