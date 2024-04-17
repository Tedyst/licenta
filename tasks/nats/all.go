package nats

import (
	"context"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/tedyst/licenta/tasks"
	"github.com/tedyst/licenta/tasks/local"
)

type allTasksRunner struct {
	dockerScannerTaskRunner
	gitScannerTaskRunner
	emailSenderTaskRunner
	nvdScannerTaskRunner
	natsScannerTaskRunner
}

type AllTasksQuerier interface {
	DockerQuerier
	GitQuerier
	local.NvdQuerier
	local.SaverQuerier
}

func NewAllTasksRunner(conn *nats.Conn, localRunner tasks.TaskRunner, querier AllTasksQuerier, concurrency int) *allTasksRunner {
	return &allTasksRunner{
		dockerScannerTaskRunner: *NewDockerScannerTaskRunner(conn, localRunner, querier, concurrency),
		gitScannerTaskRunner:    *NewGitScannerTaskRunner(conn, localRunner, querier, concurrency),
		emailSenderTaskRunner:   *NewEmailSenderTaskRunner(conn, localRunner, concurrency),
		nvdScannerTaskRunner:    *NewNVDScannerTaskRunner(conn, localRunner, querier, concurrency),
		natsScannerTaskRunner:   *NewNatsScannerTaskRunner(conn, localRunner, querier, concurrency),
	}
}

func (r *allTasksRunner) RunAll(ctx context.Context) error {
	wg := sync.WaitGroup{}

	err := r.dockerScannerTaskRunner.Run(ctx, &wg)
	if err != nil {
		return err
	}
	err = r.gitScannerTaskRunner.Run(ctx, &wg)
	if err != nil {
		return err
	}
	err = r.emailSenderTaskRunner.Run(ctx, &wg)
	if err != nil {
		return err
	}
	err = r.nvdScannerTaskRunner.Run(ctx, &wg)
	if err != nil {
		return err
	}
	err = r.natsScannerTaskRunner.Run(ctx, &wg)
	if err != nil {
		return err
	}

	return nil
}
