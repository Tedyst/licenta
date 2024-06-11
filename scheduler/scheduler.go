package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/nvd"
	"github.com/tedyst/licenta/saver"
	"github.com/tedyst/licenta/tasks"
)

type scheduler struct {
	queries     db.TransactionQuerier
	tasksRunner tasks.TaskRunner

	saltKey string
}

func NewScheduler(queries db.TransactionQuerier, tasksRunner tasks.TaskRunner) *scheduler {
	return &scheduler{
		queries:     queries,
		tasksRunner: tasksRunner,
	}
}

func (s *scheduler) Run(ctx context.Context) error {
	slog.InfoContext(ctx, "Running scheduler")

	slog.InfoContext(ctx, "Updating NVD vulnerabilities for POSTGRESQL")
	err := s.tasksRunner.UpdateNVDVulnerabilitiesForProduct(ctx, nvd.POSTGRESQL)
	if err != nil {
		return fmt.Errorf("could not update nvd vulnerabilities for product: %w", err)
	}

	slog.InfoContext(ctx, "Updating NVD vulnerabilities for MONGODB")
	err = s.tasksRunner.UpdateNVDVulnerabilitiesForProduct(ctx, nvd.MYSQL)
	if err != nil {
		return fmt.Errorf("could not update nvd vulnerabilities for product: %w", err)
	}

	slog.InfoContext(ctx, "Updating NVD vulnerabilities for MONGODB")
	err = s.tasksRunner.UpdateNVDVulnerabilitiesForProduct(ctx, nvd.MONGODB)
	if err != nil {
		return fmt.Errorf("could not update nvd vulnerabilities for product: %w", err)
	}

	slog.InfoContext(ctx, "Updating NVD vulnerabilities for REDIS")
	err = s.tasksRunner.UpdateNVDVulnerabilitiesForProduct(ctx, nvd.REDIS)
	if err != nil {
		return fmt.Errorf("could not update nvd vulnerabilities for product: %w", err)
	}

	projects, err := s.queries.GetProjects(ctx)
	if err != nil {
		return fmt.Errorf("could not get projects: %w", err)
	}

	for _, project := range projects {
		slog.InfoContext(ctx, "Creating automatic scans for project", "project", project.ID)

		scanGroup, err := s.queries.CreateScanGroup(ctx, queries.CreateScanGroupParams{
			ProjectID: project.ID,
			CreatedBy: sql.NullInt64{Int64: models.AUTOMATIC_SCAN_USER_ID, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("error creating scan group: %w", err)
		}

		scans, err := saver.CreateScans(ctx, s.queries, project.ID, scanGroup.ID, "all")
		if err != nil {
			return fmt.Errorf("error creating scans: %w", err)
		}

		gitRepositories, err := s.queries.GetGitRepositoriesForProject(ctx, queries.GetGitRepositoriesForProjectParams{
			ProjectID: project.ID,
			SaltKey:   s.saltKey,
		})
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("error getting git repositories: %w", err)
		}

		for _, gitRepository := range gitRepositories {
			scan, err := s.queries.CreateScan(ctx, queries.CreateScanParams{
				Status:      models.SCAN_NOT_STARTED,
				ScanGroupID: scanGroup.ID,
				ScanType:    models.SCAN_GIT,
			})
			if err != nil {
				return fmt.Errorf("error creating scan: %w", err)
			}

			_, err = s.queries.CreateGitScan(ctx, queries.CreateGitScanParams{
				ScanID:       scan.ID,
				RepositoryID: gitRepository.ID,
			})
			if err != nil {
				return fmt.Errorf("error creating git scan: %w", err)
			}

			scans = append(scans, scan)
		}

		dockerImages, err := s.queries.GetDockerImagesForProject(ctx, queries.GetDockerImagesForProjectParams{
			ProjectID: project.ID,
			SaltKey:   s.saltKey,
		})
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("error getting docker images: %w", err)
		}

		for _, dockerImage := range dockerImages {
			scan, err := s.queries.CreateScan(ctx, queries.CreateScanParams{
				Status:      models.SCAN_NOT_STARTED,
				ScanGroupID: scanGroup.ID,
				ScanType:    models.SCAN_DOCKER,
			})

			if err != nil {
				return fmt.Errorf("error creating scan: %w", err)
			}

			_, err = s.queries.CreateDockerScan(ctx, queries.CreateDockerScanParams{
				ScanID:  scan.ID,
				ImageID: dockerImage.ID,
			})
			if err != nil {
				return fmt.Errorf("error creating docker scan: %w", err)
			}
		}

		slog.InfoContext(ctx, "Scheduling full run for project", "project", project.ID)

		err = s.tasksRunner.ScheduleFullRun(ctx, project, scanGroup, "all", "all")
		if err != nil {
			return fmt.Errorf("error scheduling full run: %w", err)
		}

		slog.InfoContext(ctx, "Scheduled full run for project", "project", project.ID)

	}

	return nil
}

func (s *scheduler) RunContinuous(ctx context.Context, duration time.Duration) error {
	for {
		err := s.Run(ctx)
		if err != nil {
			return err
		}

		ticker := time.NewTicker(duration)

		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "Stopping scheduler")
			return ctx.Err()
		case <-ticker.C:
			err := s.Run(ctx)
			if err != nil {
				return err
			}
			continue
		}
	}
}
