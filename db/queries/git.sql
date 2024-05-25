-- name: GetGitRepositoriesForProject :many
SELECT
    id,
    project_id,
    git_repository,
    decrypt_data(project_id, sqlc.arg(salt_key), username) AS username,
    decrypt_data(project_id, sqlc.arg(salt_key), PASSWORD) AS PASSWORD,
    decrypt_data(project_id, sqlc.arg(salt_key), private_key) AS private_key
FROM
    git_repositories
WHERE
    project_id = $1;

-- name: GetGitRepository :one
SELECT
    id,
    project_id,
    git_repository,
    decrypt_data(project_id, sqlc.arg(salt_key), username) AS username,
    decrypt_data(project_id, sqlc.arg(salt_key), PASSWORD) AS PASSWORD,
    decrypt_data(project_id, sqlc.arg(salt_key), private_key) AS private_key
FROM
    git_repositories
WHERE
    id = $1;

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
    *
FROM ((
        SELECT
            git_commits.id AS commit_id,
            git_commits.repository_id,
            git_commits.commit_hash,
            git_commits.author,
            git_commits.author_email,
            git_commits.commit_date,
            git_commits.description,
            git_commits.created_at AS commit_created_at,
            git_results.*
        FROM
            git_commits
        LEFT JOIN git_results ON git_commits.id = git_results.commit
    WHERE
        git_commits.repository_id = $1
        AND git_results.id IS NULL
    ORDER BY
        git_commits.commit_date DESC
    LIMIT 25)
UNION (
    SELECT
        git_commits.id AS commit_id,
        git_commits.repository_id,
        git_commits.commit_hash,
        git_commits.author,
        git_commits.author_email,
        git_commits.commit_date,
        git_commits.description,
        git_commits.created_at AS commit_created_at,
        git_results.*
    FROM
        git_commits
    LEFT JOIN git_results ON git_commits.id = git_results.commit
WHERE
    git_commits.repository_id = $1
    AND git_results.id IS NOT NULL)) AS asd
ORDER BY
    commit_date DESC;

-- name: UpdateGitRepository :one
UPDATE
    git_repositories
SET
    git_repository = $2,
    username = encrypt_data(sqlc.arg(project_id), sqlc.arg(salt_key), sqlc.arg(username)),
    PASSWORD = encrypt_data(sqlc.arg(project_id), sqlc.arg(salt_key), sqlc.arg(PASSWORD)),
    private_key = encrypt_data(sqlc.arg(project_id), sqlc.arg(salt_key), sqlc.arg(private_key))
WHERE
    id = $1
RETURNING
    *;

-- name: CreateGitRepository :one
INSERT INTO git_repositories(project_id, git_repository, username, PASSWORD, private_key)
    VALUES ($1, $2, encrypt_data($1, sqlc.arg(salt_key), sqlc.arg(username)), encrypt_data($1, sqlc.arg(salt_key), sqlc.arg(PASSWORD)), encrypt_data($1, sqlc.arg(salt_key), sqlc.arg(private_key)))
RETURNING
    *;

-- name: CreateGitScan :one
INSERT INTO git_scans(repository_id, scan_id)
    VALUES ($1, $2)
RETURNING
    *;

-- name: GetGitScanByScan :many
SELECT
    *
FROM
    git_scans
WHERE
    scan_id = $1;

-- name: GetGitScanByScanAndRepo :one
SELECT
    sqlc.embed(git_scans),
    sqlc.embed(scans)
FROM
    git_scans
    INNER JOIN scans ON scans.id = git_scans.scan_id
WHERE
    scans.scan_group_id = $1
    AND repository_id = $2;

