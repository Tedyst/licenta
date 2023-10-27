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

-- name: GetGitScannedCommitsForProject :many
SELECT
    commit_hash
FROM
    project_git_scanned_commits
WHERE
    project_id = $1;

