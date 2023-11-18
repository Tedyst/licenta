package tasks

import (
	"context"

	"github.com/tedyst/licenta/models"
)

type TaskRunner interface {
	EmailTasksRunner
	DockerTasksRunner
	GitTasksRunner
}

type EmailTasksRunner interface {
	SendResetEmail(ctx context.Context, address string, subject string, html string, text string)
}

type DockerTasksRunner interface {
	ScanDockerRepository(ctx context.Context, image *models.ProjectDockerImage) error
}

type GitTasksRunner interface {
	ScanGitRepository(ctx context.Context, repo *models.ProjectGitRepository) error
}
