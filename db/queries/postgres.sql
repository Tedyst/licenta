-- name: CreatePostgresScan :one
INSERT INTO postgres_scans(scan_id, database_id)
    VALUES ($1, $2)
RETURNING
    *;

-- name: GetPostgresScan :one
SELECT
    *
FROM
    postgres_scans
WHERE
    id = $1
LIMIT 1;

-- name: GetPostgresScanByScanID :one
SELECT
    *
FROM
    postgres_scans
WHERE
    scan_id = $1
LIMIT 1;

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
    version = $7
WHERE
    id = $1;

-- name: GetPostgresDatabasesForProject :many
SELECT
    *
FROM
    postgres_databases
WHERE
    project_id = $1;

-- name: GetPostgresDatabase :one
SELECT
    sqlc.embed(postgres_databases),
(
        SELECT
            COUNT(*)
        FROM
            postgres_scans
        WHERE
            postgres_scans.database_id = postgres_databases.id) AS scan_count
FROM
    postgres_databases
WHERE
    postgres_databases.id = $1;

-- name: GetProjectInfoForPostgresScanByScanID :one
SELECT
    sqlc.embed(projects),
    sqlc.embed(postgres_databases),
    sqlc.embed(postgres_scans)
FROM
    projects
    JOIN postgres_databases ON postgres_databases.project_id = projects.id
    JOIN postgres_scans ON postgres_scans.database_id = postgres_databases.id
WHERE
    postgres_scans.scan_id = $1;

-- name: CreatePostgresDatabase :one
INSERT INTO postgres_databases(project_id, database_name, host, port, username, PASSWORD, version)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    *;

