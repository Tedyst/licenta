package tasks

import (
	"context"

	"github.com/tedyst/licenta/extractors/docker"
)

type TaskRunner interface {
	EmailTasksRunner

	ExtractDockerImage(
		ctx context.Context,
		imageName string,
		opts ...docker.Option,
	)
}

type EmailTasksRunner interface {
	SendResetEmail(ctx context.Context, address string, subject string, html string, text string)
}
