-- name: CreatePostgresScan :one
INSERT INTO postgres_scan(postgres_database_id, status, worker_id)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: CreatePostgresScanResult :one
INSERT INTO postgres_scan_results(postgres_scan_id, severity, message)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: UpdatePostgresScanStatus :exec
UPDATE
    postgres_scan
SET
    status = $2,
    error = $3,
    ended_at = $4
WHERE
    id = $1;

-- name: GetPostgresScanResults :many
SELECT
    *
FROM
    postgres_scan_results
WHERE
    postgres_scan_id = $1;

-- name: GetPostgresScan :one
SELECT
    *
FROM
    postgres_scan
WHERE
    id = $1;

-- name: GetPostgresDatabase :one
SELECT
    *
FROM
    postgres_databases
WHERE
    id = $1;

-- name: CreatePostgresScanBruteforceResult :one
INSERT INTO postgres_scan_bruteforce_results(postgres_scan_id, username, PASSWORD, tried, total)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: UpdatePostgresScanBruteforceResult :exec
UPDATE
    postgres_scan_bruteforce_results
SET
    PASSWORD = $2,
    tried = $3,
    total = $4
WHERE
    id = $1;

