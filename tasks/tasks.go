package tasks

import (
	"context"

	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/nvd"
)

type TaskRunner interface {
	EmailTasksRunner
	DockerTasksRunner
	GitTasksRunner
	VulnerabilityTasksRunner
	ScannerTaskRunner
	SourceTaskRunner
	ScannerSourceTaskRunner
}

type EmailTasksRunner interface {
	SendEmail(ctx context.Context, address string, subject string, html string, text string) error
}

type DockerTasksRunner interface {
	ScanDockerRepository(ctx context.Context, image *queries.DockerImage, scan *queries.Scan) error
}

type GitTasksRunner interface {
	ScanGitRepository(ctx context.Context, repo *queries.GitRepository, scan *queries.Scan) error
}

type VulnerabilityTasksRunner interface {
	UpdateNVDVulnerabilitiesForProduct(ctx context.Context, product nvd.Product) error
}

type ScannerTaskRunner interface {
	RunSaverRemote(ctx context.Context, scan *queries.Scan, scanType string) error
	RunSaverForPublic(ctx context.Context, scan *queries.Scan, scanType string) error
	ScheduleSaverRun(ctx context.Context, scan *queries.Scan, scanType string) error
}

type SourceTaskRunner interface {
	ScheduleSourceRun(ctx context.Context, project *queries.Project, scanGroup *queries.ScanGroup, sourceType string) error
}

type ScannerSourceTaskRunner interface {
	ScannerTaskRunner
	SourceTaskRunner
	ScheduleFullRun(ctx context.Context, project *queries.Project, scanGroup *queries.ScanGroup, sourceType string, scanType string) error
}
