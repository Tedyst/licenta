package local

import (
	"context"
	"database/sql"
	"fmt"

	"errors"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/extractors/file"
	"github.com/tedyst/licenta/extractors/git"
)

type GitQuerier interface {
	GetGitScannedCommitsForProjectBatch(ctx context.Context, params queries.GetGitScannedCommitsForProjectBatchParams) ([]string, error)
	CreateGitCommitForProject(ctx context.Context, params queries.CreateGitCommitForProjectParams) (*queries.GitCommit, error)
	CreateGitResultForCommit(ctx context.Context, params []queries.CreateGitResultForCommitParams) (int64, error)
}

type GitRunner struct {
	queries GitQuerier

	FileScannerProvider func(opts ...file.Option) (*file.FileScanner, error)
	GitScannerProvider  func(repoUrl string, fileScanner git.FileScanner, options ...git.Option) (*git.GitScan, error)
}

func NewGitRunner(queries GitQuerier) *GitRunner {
	return &GitRunner{
		queries:             queries,
		GitScannerProvider:  git.New,
		FileScannerProvider: file.NewScanner,
	}
}

func (r *GitRunner) ScanGitRepository(ctx context.Context, repo *queries.GitRepository) error {
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
		key, err := ssh.NewPublicKeys(repo.Username.String, []byte(repo.PrivateKey.String), "")
		if err != nil {
			return fmt.Errorf("error creating ssh key: %w", err)
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
			delete(commitsMap, item)
		}
		for _, item := range commitsMap {
			result = append(result, item)
		}
		return result, nil
	}))

	commitCache := map[string]*queries.GitCommit{}

	options = append(options, git.WithCallbackResult(func(ctx context.Context, scanner *git.GitScan, result *git.GitResult) error {
		var commit *queries.GitCommit
		if c, ok := commitCache[result.CommitHash]; ok {
			commit = c
		} else {
			var err error
			commit, err = r.queries.CreateGitCommitForProject(ctx, queries.CreateGitCommitForProjectParams{
				RepositoryID: repo.ID,
				CommitHash:   result.CommitHash,
			})
			if err != nil {
				return fmt.Errorf("error creating commit: %w", err)
			}
			commitCache[result.CommitHash] = commit
		}

		results := []queries.CreateGitResultForCommitParams{}
		for _, item := range result.Results {
			results = append(results, queries.CreateGitResultForCommitParams{
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
		_, err := r.queries.CreateGitResultForCommit(ctx, results)
		if err != nil {
			return fmt.Errorf("error creating result: %w", err)
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
