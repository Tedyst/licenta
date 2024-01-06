package local

import (
	"context"
	"database/sql"
	errs "errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/models"
)

type allScannerQuerier interface {
	scanQuerier
	postgresQuerier

	GetProject(ctx context.Context, id int64) (*queries.Project, error)
	GetPostgresScanByScanID(ctx context.Context, scanID int64) (*queries.PostgresScan, error)

	GetWorkersForProject(ctx context.Context, projectID int64) ([]*queries.Worker, error)
}

type allScannerRunner struct {
	queries         allScannerQuerier
	messageExchange messages.Exchange

	postgresScanRunner postgresScanRunner
}

func NewAllScannerRunner(queries allScannerQuerier, messageExchange messages.Exchange, bruteforceProvider bruteforce.BruteforceProvider) *allScannerRunner {
	return &allScannerRunner{
		queries:            queries,
		messageExchange:    messageExchange,
		postgresScanRunner: *NewPostgresScanRunner(queries, bruteforceProvider),
	}
}

func (r *allScannerRunner) RunAllScanners(ctx context.Context, scan *models.Scan, runningRemote bool) error {
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
		return err
	}

	return nil
}

func (r *allScannerRunner) runAllScanners(ctx context.Context, scan *models.Scan, runningRemote bool) error {
	project, err := r.queries.GetProject(ctx, scan.ProjectID)
	if err != nil {
		return err
	}

	postgresScan, err := r.queries.GetPostgresScanByScanID(ctx, scan.ID)
	if err != nil {
		return err
	}
	if err != pgx.ErrNoRows {
		err := r.postgresScanRunner.scanPostgresDB(ctx, postgresScan, scan, project.Remote && !runningRemote)
		if err != nil {
			return err
		}
	}

	if project.Remote && !runningRemote {
		workers, err := r.queries.GetWorkersForProject(ctx, scan.ProjectID)
		if err != nil {
			return errors.Wrap(err, "could not get workers for project")
		}

		if len(workers) == 0 {
			return errors.New("no workers available")
		}

		message := messages.GetStartScanMessage(scan)

		for _, worker := range workers {
			err := r.messageExchange.PublishSendScanToWorkerMessage(ctx, worker, &message)
			if err != nil {
				return errors.Wrap(err, "could not publish message")
			}
		}

		return nil
	}

	if err := r.queries.UpdateScanStatus(ctx, queries.UpdateScanStatusParams{
		ID:     postgresScan.ScanID,
		Status: models.SCAN_FINISHED,
	}); err != nil {
		return errors.Wrap(err, "could not update scan status")
	}

	return nil
}
