-- name: InsertBruteforcePasswords :exec
INSERT INTO default_bruteforce_passwords(PASSWORD)
    VALUES (unnest(sqlc.arg(passwords)::text[]))
ON CONFLICT
    DO NOTHING;

-- name: CreateBruteforcedPassword :one
INSERT INTO bruteforced_passwords(hash, username, PASSWORD, last_bruteforce_id, project_id)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: UpdateBruteforcedPassword :one
UPDATE
    bruteforced_passwords
SET
    last_bruteforce_id = $2,
    PASSWORD = $3
WHERE
    id = $1
RETURNING
    *;

-- name: GetBruteforcedPasswords :one
SELECT
    *
FROM
    bruteforced_passwords
WHERE
    hash = $1
    AND username = $2
    AND (project_id = $3
        OR project_id = NULL)
LIMIT 1;

-- name: GetBruteforcePasswordsForProjectCount :one
SELECT
    SUM(count)
FROM (
    SELECT
        COUNT(*)
    FROM
        default_bruteforce_passwords
    UNION ALL
    SELECT
        COUNT(*)
    FROM
        docker_results
        INNER JOIN docker_layers ON docker_results.layer_id = docker_layers.id
        INNER JOIN docker_images ON docker_layers.image_id = docker_images.id
    WHERE
        docker_images.project_id = $1
    UNION ALL
    SELECT
        COUNT(*)
    FROM
        git_results
        INNER JOIN git_commits ON git_results.commit = git_commits.id
        INNER JOIN git_repositories ON git_commits.repository_id = git_repositories.id
    WHERE
        git_repositories.project_id = $1) AS count;

-- name: GetBruteforcePasswordsSpecificForProject :many
SELECT
    docker_results.PASSWORD
FROM
    docker_results
    INNER JOIN docker_layers ON docker_results.layer_id = docker_layers.id
    INNER JOIN docker_images ON docker_layers.image_id = docker_images.id
WHERE
    docker_images.project_id = $1
UNION ALL
SELECT
    PASSWORD
FROM
    git_results
    INNER JOIN git_commits ON git_results.commit = git_commits.id
    INNER JOIN git_repositories ON git_commits.repository_id = git_repositories.id
WHERE
    git_repositories.project_id = $1;

-- name: GetBruteforcePasswordsPaginated :many
(
    SELECT
        default_bruteforce_passwords.id,
        PASSWORD
    FROM
        default_bruteforce_passwords
    WHERE
        default_bruteforce_passwords.id > sqlc.arg('last_id')
    LIMIT sqlc.arg('limit'))
UNION (
    SELECT
        -1,
        docker_results.PASSWORD
    FROM
        docker_results
        INNER JOIN docker_layers ON docker_results.layer_id = docker_layers.id
        INNER JOIN docker_images ON docker_layers.image_id = docker_images.id
    WHERE
        docker_images.project_id = sqlc.arg('project_id')
        AND docker_results.PASSWORD IS NOT NULL
        AND sqlc.arg('last_id') = - 1)
UNION (
    SELECT
        -1,
        git_results.PASSWORD
    FROM
        git_results
        INNER JOIN git_commits ON git_results.commit = git_commits.id
        INNER JOIN git_repositories ON git_commits.repository_id = git_repositories.id
    WHERE
        git_repositories.project_id = sqlc.arg('project_id')
        AND git_results.PASSWORD IS NOT NULL
        AND sqlc.arg('last_id') = - 1);

-- name: GetSpecificBruteforcePasswordID :one
SELECT
    subq.id
FROM (
    SELECT
        id
    FROM
        default_bruteforce_passwords
    WHERE
        default_bruteforce_passwords.PASSWORD = $1
    UNION ALL
    SELECT
        -1
    FROM
        docker_results
        INNER JOIN docker_layers ON docker_results.layer_id = docker_layers.id
        INNER JOIN docker_images ON docker_layers.image_id = docker_images.id
    WHERE
        docker_images.project_id = $2
        AND docker_results.PASSWORD = $1
    UNION ALL
    SELECT
        -1
    FROM
        git_results
        INNER JOIN git_commits ON git_results.commit = git_commits.id
        INNER JOIN git_repositories ON git_commits.repository_id = git_repositories.id
    WHERE
        git_repositories.project_id = $2
        AND git_results.PASSWORD = $1) AS subq
LIMIT 1;

