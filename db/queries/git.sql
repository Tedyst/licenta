-- name: GetGitRepositoriesForProject :many
SELECT
    *
FROM
    project_git_repositories
WHERE
    project_id = $1;

-- name: CreateGitRepositoryForProject :one
INSERT INTO project_git_repositories(project_id, git_repository, username, PASSWORD)
    VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: DeleteGitRepositoryForProject :exec
DELETE FROM project_git_repositories
WHERE project_id = $1
    AND git_repository = $2;

-- name: CreateGitCommitForProject :one
INSERT INTO project_git_scanned_commits(project_id, commit_hash)
    VALUES ($1, $2)
RETURNING
    *;

-- name: GetGitScannedCommitsForProject :many
SELECT
    commit_hash
FROM
    project_git_scanned_commits
WHERE
    project_id = $1;

-- name: GetGitScannedCommitsForProjectBatch :many
SELECT
    commit_hash
FROM
    project_git_scanned_commits
WHERE
    project_id = $1
    AND commit_hash = ANY (sqlc.arg(commit_hashes)::string[]);

-- name: CreateGitResultForCommit :copyfrom
INSERT INTO project_git_results(project_id,
COMMIT, name, line, line_number, MATCH, probability, username, PASSWORD, filename)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

