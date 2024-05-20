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
    WHERE
        docker_results.project_id = $1
    UNION ALL
    SELECT
        COUNT(*)
    FROM
        git_results
    WHERE
        git_results.project_id = $1) AS count;

-- name: GetBruteforcePasswordsSpecificForProject :many
SELECT
    PASSWORD
FROM
    docker_results
WHERE
    docker_results.project_id = $1
UNION ALL
SELECT
    PASSWORD
FROM
    git_results
WHERE
    git_results.project_id = $1;

-- name: GetBruteforcePasswordsPaginated :many
SELECT
    id,
    PASSWORD
FROM
    default_bruteforce_passwords
WHERE
    id > sqlc.arg('last_id')
LIMIT sqlc.arg('limit');

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
    WHERE
        docker_results.project_id = $2
        AND docker_results.PASSWORD = $1
    UNION ALL
    SELECT
        -1
    FROM
        git_results
    WHERE
        git_results.project_id = $2
        AND git_results.PASSWORD = $1) AS subq
LIMIT 1;

