-- name: InsertBruteforcePasswords :exec
INSERT INTO default_bruteforce_passwords(PASSWORD)
    VALUES (unnest(sqlc.arg(passwords)::text[]))
ON CONFLICT
    DO NOTHING;

-- name: InsertBruteforcedPassword :exec
INSERT INTO bruteforced_passwords(hash, username, PASSWORD, last_bruteforce_id)
    VALUES ($1, $2, $3, $4)
ON CONFLICT
    DO NOTHING;

-- name: GetBruteforcedPasswordByHashAndUsername :one
SELECT
    *
FROM
    bruteforced_passwords
WHERE
    hash = $1
    AND username = $2
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
        project_docker_layer_results
    WHERE
        project_docker_layer_results.project_id = $1
    UNION ALL
    SELECT
        COUNT(*)
    FROM
        project_git_results
    WHERE
        project_git_results.project_id = $1) AS count;

-- name: GetBruteforcePasswordsSpecificForProject :many
SELECT
    PASSWORD
FROM
    project_docker_layer_results
WHERE
    project_docker_layer_results.project_id = $1
UNION ALL
SELECT
    PASSWORD
FROM
    project_git_results
WHERE
    project_git_results.project_id = $1;

-- name: GetBruteforcePasswordsPaginated :many
SELECT
    id,
    PASSWORD
FROM
    default_bruteforce_passwords
WHERE
    id > sqlc.arg('last_id')
LIMIT sqlc.arg('limit');

