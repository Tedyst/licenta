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
    sqlc.embed(postgres_scan),
(
        SELECT
            COALESCE(MAX(postgres_scan_results.severity), 0)::integer
        FROM
            postgres_scan_results
        WHERE
            postgres_scan_id = postgres_scan.id) AS maximum_severity
FROM
    postgres_scan
WHERE
    postgres_scan.id = $1;

-- name: GetPostgresDatabase :one
SELECT
    sqlc.embed(postgres_databases),
(
        SELECT
            COUNT(*)
        FROM
            postgres_scan
        WHERE
            postgres_scan.postgres_database_id = postgres_databases.id) AS scan_count
FROM
    postgres_databases
WHERE
    postgres_databases.id = $1;

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

-- name: GetPostgresScansForProject :many
SELECT
    sqlc.embed(postgres_scan),
(
        SELECT
            MAX(postgres_scan_results.severity)::integer
        FROM
            postgres_scan_results
        WHERE
            postgres_scan_id = postgres_scan.id) AS maximum_severity
FROM
    postgres_scan
WHERE
    postgres_scan.postgres_database_id IN (
        SELECT
            postgres_databases.id
        FROM
            postgres_databases
        WHERE
            postgres_databases.project_id = $1)
ORDER BY
    postgres_scan.id DESC;

-- name: GetPostgresScansForDatabase :many
SELECT
    sqlc.embed(postgres_scan),
(
        SELECT
            COALESCE(MAX(postgres_scan_results.severity), 0)::integer
        FROM
            postgres_scan_results
        WHERE
            postgres_scan_id = postgres_scan.id) AS maximum_severity
FROM
    postgres_scan
WHERE
    postgres_scan.postgres_database_id = $1
ORDER BY
    postgres_scan.id DESC;

-- name: UpdatePostgresVersion :exec
UPDATE
    postgres_databases
SET
    version = $2
WHERE
    id = $1;

-- name: UpdatePostgresDatabase :exec
UPDATE
    postgres_databases
SET
    database_name = $2,
    host = $3,
    port = $4,
    username = $5,
    PASSWORD = $6,
    remote = $7,
    version = $8
WHERE
    id = $1;

-- name: GetPostgresDatabasesForProject :many
SELECT
    *
FROM
    postgres_databases
WHERE
    project_id = $1;

