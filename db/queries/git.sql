-- name: GetGitRepositoriesForProject :many
SELECT
    *
FROM
    git_repositories
WHERE
    project_id = $1;

-- name: GetGitRepository :one
SELECT
    *
FROM
    git_repositories
WHERE
    id = $1;

-- name: CreateGitRepositoryForProject :one
INSERT INTO git_repositories(project_id, git_repository, username, PASSWORD)
    VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: DeleteGitRepositoryForProject :exec
DELETE FROM git_repositories
WHERE project_id = $1
    AND git_repository = $2;

-- name: CreateGitCommitForProject :one
INSERT INTO git_commits(repository_id, commit_hash, author, author_email, description, commit_date)
    VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: GetGitScannedCommitsForProject :many
SELECT
    commit_hash
FROM
    git_commits
    INNER JOIN git_repositories ON git_repositories.id = git_commits.repository_id
WHERE
    git_repositories.project_id = $1;

-- name: GetGitScannedCommitsForProjectBatch :many
SELECT
    commit_hash
FROM
    git_commits
    INNER JOIN git_repositories ON git_repositories.id = git_commits.repository_id
WHERE
    project_id = $1
    AND commit_hash = ANY (sqlc.arg(commit_hashes)::text[]);

-- name: CreateGitResultForCommit :copyfrom
INSERT INTO git_results(
COMMIT, name, line, line_number, MATCH, probability, username, PASSWORD, filename)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: DeleteGitRepository :exec
DELETE FROM git_repositories
WHERE id = $1;

-- name: GetGitCommitsWithResults :many
SELECT
    sqlc.embed(git_commits),
    sqlc.embed(git_results)
FROM
    git_commits
    LEFT JOIN git_results ON git_commits.commit_hash = git_results.commit
WHERE
    git_commits.repository_id = $1;

-- name: UpdateGitRepository :one
UPDATE
    git_repositories
SET
    git_repository = $2,
    username = $3,
    PASSWORD = $4,
    private_key = $5
WHERE
    id = $1
RETURNING
    *;

-- name: CreateGitRepository :one
INSERT INTO git_repositories(project_id, git_repository, username, PASSWORD, private_key)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

