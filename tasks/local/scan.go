package local

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/saver"
)

type saverRunner struct {
	queries            SaverQuerier
	messageExchange    messages.Exchange
	bruteforceProvider bruteforce.BruteforceProvider
}

type SaverQuerier interface {
	saver.BaseQuerier
	GetProject(ctx context.Context, id int64) (*queries.Project, error)
	GetWorkersForProject(ctx context.Context, projectID int64) ([]*queries.Worker, error)
}

func NewSaverRunner(queries SaverQuerier, messageExchange messages.Exchange, bruteforceProvider bruteforce.BruteforceProvider) *saverRunner {
	return &saverRunner{
		queries:            queries,
		messageExchange:    messageExchange,
		bruteforceProvider: bruteforceProvider,
	}
}

func (r *saverRunner) RunSaverRemote(ctx context.Context, scan *queries.Scan, scanType string) error {
	saver, err := saver.NewSaver(ctx, r.queries, r.bruteforceProvider, scan, scanType)
	if err != nil {
		return err
	}

	return saver.Scan(ctx)
}

func (r *saverRunner) RunSaverForPublic(ctx context.Context, scan *queries.Scan, scanType string) error {
	saver, err := saver.NewSaver(ctx, r.queries, r.bruteforceProvider, scan, scanType)
	if err != nil {
		return err
	}

	return saver.ScanForPublicAccessOnly(ctx)
}

func (r *saverRunner) ScheduleSaverRun(ctx context.Context, scan *queries.Scan, scanType string) error {
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
	}

	return nil
}
