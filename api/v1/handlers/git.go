package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/api/authorization"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
)

func (server *serverHandler) DeleteGitId(ctx context.Context, request generated.DeleteGitIdRequestObject) (generated.DeleteGitIdResponseObject, error) {
	gitRepository, err := server.DatabaseProvider.GetGitRepository(ctx, request.Id)
	if err == pgx.ErrNoRows {
		return generated.DeleteGitId404JSONResponse{
			Success: false,
			Message: "Git repository not found",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("DeleteGitId: error getting git repository: %w", err)
	}

	_, _, response, err := checkUserHasProjectPermission[generated.DeleteGitId401JSONResponse](server, ctx, gitRepository.ProjectID, authorization.Admin)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	err = server.DatabaseProvider.DeleteGitRepository(ctx, gitRepository.ID)
	if err != nil {
		return nil, fmt.Errorf("DeleteGitId: error deleting git repository: %w", err)
	}

	return generated.DeleteGitId204JSONResponse{
		Success: true,
	}, nil
}

func (server *serverHandler) GetGit(ctx context.Context, request generated.GetGitRequestObject) (generated.GetGitResponseObject, error) {
	gitRepositories, err := server.DatabaseProvider.GetGitRepositoriesForProject(ctx, int64(request.Params.Project))
	if err == pgx.ErrNoRows {
		return generated.GetGit401JSONResponse{
			Success: false,
			Message: "Git repository not found",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetGit: error getting git repository: %w", err)
	}

	gitRepositoriesResponse := make([]generated.Git, len(gitRepositories))
	for i, gitRepository := range gitRepositories {
		gitRepositoriesResponse[i] = generated.Git{
			GitRepository: gitRepository.GitRepository,
		}
	}

	return generated.GetGit200JSONResponse{
		Success:         true,
		GitRepositories: gitRepositoriesResponse,
	}, nil
}

func (server *serverHandler) GetGitId(ctx context.Context, request generated.GetGitIdRequestObject) (generated.GetGitIdResponseObject, error) {
	gitRepository, err := server.DatabaseProvider.GetGitRepository(ctx, request.Id)
	if err == pgx.ErrNoRows {
		return generated.GetGitId404JSONResponse{
			Success: false,
			Message: "Git repository not found",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetGitId: error getting git repository: %w", err)
	}

	_, _, response, err := checkUserHasProjectPermission[generated.GetGitId401JSONResponse](server, ctx, gitRepository.ProjectID, authorization.Viewer)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	dbCommits, err := server.DatabaseProvider.GetGitCommitsWithResults(ctx, gitRepository.ID)
	if err != nil {
		return nil, fmt.Errorf("GetGitId: error getting git commits: %w", err)
	}

	commits := []generated.GitCommit{}
	commitResults := map[int64][]generated.GitResult{}

	for _, dbCommit := range dbCommits {
		if _, ok := commitResults[dbCommit.GitCommit.ID]; !ok {
			commitResults[dbCommit.GitCommit.ID] = []generated.GitResult{}
			commits = append(commits, generated.GitCommit{
				CommitHash:   dbCommit.GitCommit.CommitHash,
				CreatedAt:    dbCommit.GitCommit.CreatedAt.Time.Format(time.RFC3339Nano),
				Id:           int(dbCommit.GitCommit.ID),
				RepositoryId: int(dbCommit.GitCommit.RepositoryID),
				Results:      []generated.GitResult{},
			})
		}

		commitResults[dbCommit.GitCommit.ID] = append(commitResults[dbCommit.GitCommit.ID], generated.GitResult{
			Commit:      int(dbCommit.GitResult.Commit),
			Filename:    dbCommit.GitResult.Filename,
			Id:          int(dbCommit.GitResult.ID),
			Line:        dbCommit.GitResult.Line,
			LineNumber:  int(dbCommit.GitResult.LineNumber),
			Match:       dbCommit.GitResult.Match,
			Name:        dbCommit.GitResult.Name,
			Password:    dbCommit.GitResult.Password.String,
			Probability: float32(dbCommit.GitResult.Probability),
			Username:    dbCommit.GitResult.Username.String,
		})
	}

	for i, commit := range commits {
		commits[i].Results = commitResults[int64(commit.Id)]
	}

	return generated.GetGitId200JSONResponse{
		Success: true,
		Git:     generated.Git{GitRepository: gitRepository.GitRepository},
		Commits: commits,
	}, nil
}

func (server *serverHandler) PatchGitId(ctx context.Context, request generated.PatchGitIdRequestObject) (generated.PatchGitIdResponseObject, error) {
	gitRepository, err := server.DatabaseProvider.GetGitRepository(ctx, request.Id)
	if err == pgx.ErrNoRows {
		return generated.PatchGitId404JSONResponse{
			Success: false,
			Message: "Git repository not found",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("PatchGitId: error getting git repository: %w", err)
	}

	_, _, response, err := checkUserHasProjectPermission[generated.PatchGitId401JSONResponse](server, ctx, gitRepository.ProjectID, authorization.Admin)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	if request.Body.GitRepository != nil {
		gitRepository.GitRepository = *request.Body.GitRepository
	}
	if request.Body.Password != nil {
		gitRepository.Password = sql.NullString{String: *request.Body.Password, Valid: true}
	}
	if request.Body.Username != nil {
		gitRepository.Username = sql.NullString{String: *request.Body.Username, Valid: true}
	}
	if request.Body.PrivateKey != nil {
		gitRepository.PrivateKey = sql.NullString{String: *request.Body.PrivateKey, Valid: true}
	}

	repo, err := server.DatabaseProvider.UpdateGitRepository(ctx, queries.UpdateGitRepositoryParams{
		ID:            gitRepository.ID,
		GitRepository: gitRepository.GitRepository,
		Username:      gitRepository.Username,
		Password:      gitRepository.Password,
		PrivateKey:    gitRepository.PrivateKey,
	})

	if err != nil {
		return nil, fmt.Errorf("PatchGitId: error updating git repository: %w", err)
	}

	return generated.PatchGitId200JSONResponse{
		Success: true,
		Git: generated.Git{
			GitRepository: repo.GitRepository,
			HasSsh:        repo.PrivateKey.Valid,
			Id:            int(repo.ID),
			Password:      repo.Password.String,
			ProjectId:     int(repo.ProjectID),
			Username:      repo.Username.String,
		},
	}, nil
}

func (server *serverHandler) PostGit(ctx context.Context, request generated.PostGitRequestObject) (generated.PostGitResponseObject, error) {
	_, _, response, err := checkUserHasProjectPermission[generated.PostGit401JSONResponse](server, ctx, int64(request.Body.ProjectId), authorization.Admin)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	params := queries.CreateGitRepositoryParams{
		GitRepository: request.Body.GitRepository,
		ProjectID:     int64(request.Body.ProjectId),
	}
	if request.Body.Password != nil {
		params.Password = sql.NullString{String: *request.Body.Password, Valid: true}
	}
	if request.Body.Username != nil {
		params.Username = sql.NullString{String: *request.Body.Username, Valid: true}
	}
	if request.Body.PrivateKey != nil {
		params.PrivateKey = sql.NullString{String: *request.Body.PrivateKey, Valid: true}
	}

	gitRepository, err := server.DatabaseProvider.CreateGitRepository(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("PostGit: error creating git repository: %w", err)
	}

	return generated.PostGit201JSONResponse{
		Success: true,
		Git: generated.Git{
			GitRepository: gitRepository.GitRepository,
			HasSsh:        gitRepository.PrivateKey.Valid,
			Id:            int(gitRepository.ID),
			Password:      gitRepository.Password.String,
			ProjectId:     int(gitRepository.ProjectID),
			Username:      gitRepository.Username.String,
		},
	}, nil
}
