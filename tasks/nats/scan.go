package nats

import (
	"context"
	"log/slog"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/tasks"
	"github.com/tedyst/licenta/tasks/local"
	"golang.org/x/sync/semaphore"
)

type natsScannerTaskSender struct {
	conn *nats.Conn
}

func NewNatsScannerTaskSender(conn *nats.Conn) *natsScannerTaskSender {
	return &natsScannerTaskSender{conn: conn}
}

const runServerRemoteQueue = "run-saver-remote"

func (ns *natsScannerTaskSender) RunSaverRemote(ctx context.Context, scan *queries.Scan, scanType string) error {
	return publishMessage(ctx, ns.conn, runServerRemoteQueue, &RunSaverRemoteMessage{
		ScanId:   scan.ID,
		ScanType: scanType,
	}, 0)
}

const runServerForPublicRemoteQueue = "run-saver-for-public-remote"

func (ns *natsScannerTaskSender) RunSaverForPublic(ctx context.Context, scan *queries.Scan, scanType string) error {
	return publishMessage(ctx, ns.conn, runServerForPublicRemoteQueue, &RunSaverForPublicMessage{
		ScanId:   scan.ID,
		ScanType: scanType,
	}, 0)
}

const scheduleSaverRunRemoteQueue = "schedule-saver-run-remote"

func (ns *natsScannerTaskSender) ScheduleSaverRun(ctx context.Context, scan *queries.Scan, scanType string) error {
	return publishMessage(ctx, ns.conn, scheduleSaverRunRemoteQueue, &ScheduleSaverRunMessage{
		ScanId:   scan.ID,
		ScanType: scanType,
	}, 0)
}

const scheduleFullRun = "schedule-full-run-remote"

func (ns *natsScannerTaskSender) ScheduleFullRun(ctx context.Context, project *queries.Project, scan *queries.ScanGroup, sourceType string, scanType string) error {
	return publishMessage(ctx, ns.conn, scheduleFullRun, &ScheduleFullRunMessage{
		ScanGroupId: scan.ID,
		ProjectId:   project.ID,
		SourceType:  sourceType,
		ScanType:    scanType,
	}, 0)
}

const scheduleSourceRun = "schedule-source-run-remote"

func (ns *natsScannerTaskSender) ScheduleSourceRun(ctx context.Context, project *queries.Project, scanGroup *queries.ScanGroup, sourceType string) error {
	return publishMessage(ctx, ns.conn, scheduleSourceRun, &ScheduleSourceRunMessage{
		ScanGroupId: scanGroup.ID,
		ProjectId:   project.ID,
		SourceType:  sourceType,
	}, 0)
}

type natsScannerTaskRunner struct {
	conn        *nats.Conn
	localRunner tasks.TaskRunner
	semaphore   *semaphore.Weighted
	querier     local.SaverQuerier
}

func NewNatsScannerTaskRunner(conn *nats.Conn, localRunner tasks.TaskRunner, querier local.SaverQuerier, concurrency int) *natsScannerTaskRunner {
	return &natsScannerTaskRunner{
		conn:        conn,
		localRunner: localRunner,
		semaphore:   semaphore.NewWeighted(int64(concurrency)),
		querier:     querier,
	}
}

func (ns *natsScannerTaskRunner) Run(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := receiveMessage(ctx, ns.conn, ns.semaphore, runServerRemoteQueue, func(ctx context.Context, message *RunSaverRemoteMessage) error {
			scan, err := ns.querier.GetScan(ctx, message.ScanId)
			if err != nil {
				return nil
			}

			err = ns.localRunner.RunSaverRemote(ctx, &scan.Scan, message.ScanType)
			if err != nil {
				return nil
			}

			return nil
		})

		if err != nil {
			slog.Error("failed to receive message", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := receiveMessage(ctx, ns.conn, ns.semaphore, runServerForPublicRemoteQueue, func(ctx context.Context, message *RunSaverForPublicMessage) error {
			scan, err := ns.querier.GetScan(ctx, message.ScanId)
			if err != nil {
				return nil
			}

			err = ns.localRunner.RunSaverForPublic(ctx, &scan.Scan, message.ScanType)
			if err != nil {
				return nil
			}

			return nil
		})

		if err != nil {
			slog.Error("failed to receive message", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := receiveMessage(ctx, ns.conn, ns.semaphore, scheduleSaverRunRemoteQueue, func(ctx context.Context, message *ScheduleSaverRunMessage) error {
			scan, err := ns.querier.GetScan(ctx, message.ScanId)
			if err != nil {
				return nil
			}

			err = ns.localRunner.ScheduleSaverRun(ctx, &scan.Scan, message.ScanType)
			if err != nil {
				return nil
			}

			return nil
		})

		if err != nil {
			slog.Error("failed to receive message", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := receiveMessage(ctx, ns.conn, ns.semaphore, scheduleFullRun, func(ctx context.Context, message *ScheduleFullRunMessage) error {
			project, err := ns.querier.GetProject(ctx, message.ProjectId)
			if err != nil {
				return nil
			}

			scan, err := ns.querier.GetScanGroup(ctx, message.ScanGroupId)
			if err != nil {
				return nil
			}

			err = ns.localRunner.ScheduleFullRun(ctx, project, scan, message.SourceType, message.ScanType)
			if err != nil {
				return nil
			}

			return nil
		})

		if err != nil {
			slog.Error("failed to receive message", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := receiveMessage(ctx, ns.conn, ns.semaphore, scheduleSourceRun, func(ctx context.Context, message *ScheduleSourceRunMessage) error {
			project, err := ns.querier.GetProject(ctx, message.ProjectId)
			if err != nil {
				return nil
			}

			scan, err := ns.querier.GetScanGroup(ctx, message.ScanGroupId)
			if err != nil {
				return nil
			}

			err = ns.localRunner.ScheduleSourceRun(ctx, project, scan, message.SourceType)
			if err != nil {
				return nil
			}

			return nil
		})

		if err != nil {
			slog.Error("failed to receive message", "error", err)
		}
	}()

	return nil
}
