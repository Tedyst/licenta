package local

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/email"
	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/tasks"
)

type localRunner struct {
	SaverRunner
	NvdRunner
	GitRunner
	emailRunner
	DockerRunner

	queries db.TransactionQuerier
}

func NewLocalRunner(debug bool, emailSender email.EmailSender, queries db.TransactionQuerier, exchange messages.Exchange, bruteforceProvider bruteforce.BruteforceProvider) *localRunner {
	return &localRunner{
		NvdRunner:    *NewNVDRunner(queries),
		GitRunner:    *NewGitRunner(queries),
		emailRunner:  *NewEmailRunner(emailSender),
		DockerRunner: *NewDockerRunner(queries),
		queries:      queries,
		SaverRunner:  *NewSaverRunner(queries, exchange, bruteforceProvider),
	}
}

func (runner *localRunner) ScheduleFullRun(ctx context.Context, project *queries.Project, scanGroup *queries.ScanGroup, sourceType string, scanType string) error {
	if err := runner.ScheduleSourceRun(ctx, project, scanGroup, sourceType); err != nil {
		return fmt.Errorf("failed to schedule source run: %w", err)
	}

	scans, err := runner.queries.GetScansForScanGroup(ctx, scanGroup.ID)
	if err != nil {
		return fmt.Errorf("failed to get scans for scan group: %w", err)
	}

	for _, scan := range scans {
		if err := runner.ScheduleSaverRun(ctx, scan, scanType); err != nil {
			return fmt.Errorf("failed to schedule saver run: %w", err)
		}
	}

	return nil
}

func (source *localRunner) ScheduleSourceRun(ctx context.Context, project *queries.Project, scanGroup *queries.ScanGroup, sourceType string) error {
	var runGit bool
	var runDocker bool
	switch sourceType {
	case "all":
		runGit = true
		runDocker = true
	case "git":
		runGit = true
	case "docker":
		runDocker = true
	}

	if runGit {
		slog.DebugContext(ctx, "Scheduling git scan", "project", project.ID)
		repos, err := source.queries.GetGitRepositoriesForProject(ctx, project.ID)
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("failed to get git repositories for project: %w", err)
		}

		for _, repo := range repos {
			slog.DebugContext(ctx, "Scheduling git scan", "repo", repo.ID, "url", repo.GitRepository)
			if err := source.ScanGitRepository(ctx, repo); err != nil {
				return fmt.Errorf("failed to scan git repository: %w", err)
			}
			slog.DebugContext(ctx, "Finished scanning git repo", "repo", repo.ID, "url", repo.GitRepository)
		}

		slog.DebugContext(ctx, "Finished scanning git repositories", "project", project.ID)
	}

	if runDocker {
		slog.DebugContext(ctx, "Scheduling docker scan", "project", project.ID)
		images, err := source.queries.GetDockerImagesForProject(ctx, project.ID)
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("failed to get docker images for project: %w", err)
		}

		for _, image := range images {
			slog.DebugContext(ctx, "Scheduling docker scan", "image", image.ID, "name", image.DockerImage)
			if err := source.ScanDockerRepository(ctx, image); err != nil {
				return fmt.Errorf("failed to scan docker repository: %w", err)
			}
		}

		slog.DebugContext(ctx, "Finished scanning docker images", "project", project.ID)
	}

	return nil
}

var _ tasks.TaskRunner = (*localRunner)(nil)
