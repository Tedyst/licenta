package local

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/saver"
)

type SaverRunner struct {
	queries            SaverQuerier
	messageExchange    messages.Exchange
	bruteforceProvider bruteforce.BruteforceProvider

	saltKey string
}

type SaverQuerier interface {
	saver.BaseQuerier
	GetProject(ctx context.Context, id int64) (*queries.Project, error)
	GetWorkersForProject(ctx context.Context, projectID int64) ([]*queries.Worker, error)
}

func NewSaverRunner(queries SaverQuerier, messageExchange messages.Exchange, bruteforceProvider bruteforce.BruteforceProvider, saltKey string) *SaverRunner {
	return &SaverRunner{
		queries:            queries,
		messageExchange:    messageExchange,
		bruteforceProvider: bruteforceProvider,
		saltKey:            saltKey,
	}
}

func (r *SaverRunner) RunSaverRemote(ctx context.Context, scan *queries.Scan, scanType string) error {
	saver, err := saver.NewSaver(ctx, r.queries, r.bruteforceProvider, scan, scanType, false, r.saltKey)
	if err != nil {
		return err
	}

	return saver.Scan(ctx)
}

func (r *SaverRunner) RunSaverForPublic(ctx context.Context, scan *queries.Scan, scanType string) error {
	saver, err := saver.NewSaver(ctx, r.queries, r.bruteforceProvider, scan, scanType, true, r.saltKey)
	if err != nil {
		return err
	}

	return saver.ScanForPublicAccessOnly(ctx)
}

func (r *SaverRunner) ScheduleSaverRun(ctx context.Context, scan *queries.Scan, scanType string) error {
	if scan.ScanType == models.SCAN_DOCKER || scan.ScanType == models.SCAN_GIT {
		return nil
	}

	scanGroup, err := r.queries.GetScanGroup(ctx, scan.ScanGroupID)
	if err != nil {
		return err
	}
	project, err := r.queries.GetProject(ctx, scanGroup.ProjectID)
	if err != nil {
		return err
	}

	if project.Remote {
		slog.DebugContext(ctx, "Sending task to remote workers", "scan", scan.ID)

		workers, err := r.queries.GetWorkersForProject(ctx, project.ID)
		if err != nil {
			return fmt.Errorf("could not get workers for project: %w", err)
		}

		if len(workers) == 0 {
			return errors.New("no workers available")
		}

		message := messages.GetStartScanMessage(scan)

		for _, worker := range workers {
			err := r.messageExchange.PublishSendScanToWorkerMessage(ctx, worker, message)
			if err != nil {
				return fmt.Errorf("could not publish message: %w", err)
			}
		}
	}

	err = r.RunSaverForPublic(ctx, scan, scanType)
	if err != nil {
		return err
	}

	if !project.Remote {
		err = r.RunSaverRemote(ctx, scan, scanType)
		if err != nil {
			return err
		}

		go func() {
			for {
				time.Sleep(5 * time.Second)
				scan, err := r.queries.GetScan(ctx, scan.ID)
				if err != nil {
					slog.ErrorContext(ctx, "failed to get scan", "error", err)
					return
				}

				if scan.Scan.WorkerID.Valid {
					slog.InfoContext(ctx, "Scan finished", "project", project.ID, "scan", scan.Scan.ID)
					return
				}

				err = r.RunSaverRemote(ctx, &scan.Scan, scanType)
				if err != nil {
					slog.ErrorContext(ctx, "failed to re-schedule saver run", "error", err)
				}

				slog.InfoContext(ctx, "Rescheduled scan", "project", project.ID, "scan", scan.Scan.ID, "worker", scan.Scan.WorkerID.Int64)
			}
		}()
	}

	return nil
}
