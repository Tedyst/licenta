package tasks

import (
	"context"

	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/nvd"
)

type TaskRunner interface {
	EmailTasksRunner
	DockerTasksRunner
	GitTasksRunner
	VulnerabilityTasksRunner
}

type EmailTasksRunner interface {
	SendResetEmail(ctx context.Context, address string, subject string, html string, text string) error
	SendCVEVulnerabilityEmail(ctx context.Context, project *models.Project) error
	SendCVEMailsToAllProjectMembers(ctx context.Context, projectID int64) error
	SendCVEMailsToAllProjects(ctx context.Context) error
}

type DockerTasksRunner interface {
	ScanDockerRepository(ctx context.Context, image *models.ProjectDockerImage) error
}

type GitTasksRunner interface {
	ScanGitRepository(ctx context.Context, repo *models.ProjectGitRepository) error
}

type VulnerabilityTasksRunner interface {
	UpdateNVDVulnerabilitiesForProduct(ctx context.Context, product nvd.Product) error
}
