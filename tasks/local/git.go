package local

import (
	"context"
	"database/sql"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/extractors/file"
	"github.com/tedyst/licenta/extractors/git"
	"github.com/tedyst/licenta/models"
)

type gitQuerier interface {
	GetGitScannedCommitsForProjectBatch(ctx context.Context, params queries.GetGitScannedCommitsForProjectBatchParams) ([]string, error)
	CreateGitCommitForProject(ctx context.Context, params queries.CreateGitCommitForProjectParams) (*queries.ProjectGitScannedCommit, error)
	CreateGitResultForCommit(ctx context.Context, params []queries.CreateGitResultForCommitParams) (int64, error)
}

type gitRunner struct {
	queries gitQuerier

	FileScannerProvider func(opts ...file.Option) (*file.FileScanner, error)
	GitScannerProvider  func(repoUrl string, fileScanner git.FileScanner, options ...git.Option) (*git.GitScan, error)
}

func NewGitRunner(queries gitQuerier) *gitRunner {
	return &gitRunner{
		queries:            queries,
		GitScannerProvider: git.New,
	}
}

func (r *gitRunner) ScanGitRepository(ctx context.Context, repo *models.ProjectGitRepository) error {
	if repo == nil {
		return errors.New("repo is nil")
	}

	options := []git.Option{}
	if repo.Username.Valid && repo.Password.Valid {
		options = append(options, git.WithCredentials(&http.BasicAuth{
			Username: repo.Username.String,
			Password: repo.Password.String,
		}))
	}
	if repo.Username.Valid && repo.PrivateKey.Valid {
		key, err := ssh.NewPublicKeysFromFile(repo.Username.String, repo.PrivateKey.String, "")
		if err != nil {
			return err
		}
		options = append(options, git.WithCredentials(key))
	}

	options = append(options, git.WithSkipCommitFunc(func(batch []git.BatchItem) ([]git.BatchItem, error) {
		commits := []string{}
		commitsMap := map[string]git.BatchItem{}
		for _, item := range batch {
			commits = append(commits, item.Commit.Hash.String())
			commitsMap[item.Commit.Hash.String()] = item
		}
		newBatch, err := r.queries.GetGitScannedCommitsForProjectBatch(ctx, queries.GetGitScannedCommitsForProjectBatchParams{
			ProjectID:    repo.ProjectID,
			CommitHashes: commits,
		})
		if err != nil {
			return nil, err
		}
		var result []git.BatchItem
		for _, item := range newBatch {
			result = append(result, commitsMap[item])
		}
		return result, nil
	}))

	options = append(options, git.WithCallbackResult(func(ctx context.Context, scanner *git.GitScan, result *git.GitResult) error {
		commit, err := r.queries.CreateGitCommitForProject(ctx, queries.CreateGitCommitForProjectParams{
			ProjectID:  repo.ProjectID,
			CommitHash: result.CommitHash,
		})
		if err != nil {
			return err
		}

		results := []queries.CreateGitResultForCommitParams{}
		for _, item := range result.Results {
			results = append(results, queries.CreateGitResultForCommitParams{
				ProjectID:   repo.ProjectID,
				Commit:      commit.ID,
				Name:        item.Name,
				Line:        item.Line,
				LineNumber:  int32(item.LineNumber),
				Match:       item.Match,
				Probability: item.Probability,
				Username:    sql.NullString{String: item.Username, Valid: true},
				Password:    sql.NullString{String: item.Password, Valid: true},
				Filename:    item.FileName,
			})
		}
		_, err = r.queries.CreateGitResultForCommit(ctx, results)
		if err != nil {
			return err
		}
		return nil
	}))

	fileScanner, err := r.FileScannerProvider()
	if err != nil {
		return err
	}

	scanner, err := r.GitScannerProvider(repo.GitRepository, fileScanner, options...)
	if err != nil {
		return err
	}

	err = scanner.Scan(ctx)
	if err != nil {
		return err
	}
	return nil
}
