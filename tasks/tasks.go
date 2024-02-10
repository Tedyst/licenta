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
}

type EmailTasksRunner interface {
	SendResetEmail(ctx context.Context, address string, subject string, html string, text string) error
	SendCVEVulnerabilityEmail(ctx context.Context, project *queries.Project) error
	SendCVEMailsToAllProjectMembers(ctx context.Context, projectID int64) error
	SendCVEMailsToAllProjects(ctx context.Context) error
}

type DockerTasksRunner interface {
	ScanDockerRepository(ctx context.Context, image *queries.ProjectDockerImage) error
}

type GitTasksRunner interface {
	ScanGitRepository(ctx context.Context, repo *queries.ProjectGitRepository) error
}

type VulnerabilityTasksRunner interface {
	UpdateNVDVulnerabilitiesForProduct(ctx context.Context, product nvd.Product) error
}

type ScannerTaskRunner interface {
	AllScanTaskRunner
	PostgresTaskRunner
}

type PostgresTaskRunner interface {
	ScanPostgresDB(ctx context.Context, scan *queries.PostgresScan) error
}

type AllScanTaskRunner interface {
	RunAllScanners(ctx context.Context, scan *queries.Scan, runningRemote bool) error
}
