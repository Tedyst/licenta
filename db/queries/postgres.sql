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
    username = encrypt_data(sqlc.arg(project_id), sqlc.arg(salt_key), sqlc.arg(username)),
    PASSWORD = encrypt_data(sqlc.arg(project_id), sqlc.arg(salt_key), sqlc.arg(PASSWORD)),
    version = $5
WHERE
    id = $1;

-- name: GetPostgresDatabasesForProject :many
SELECT
    id,
    project_id,
    host,
    port,
    database_name,
    decrypt_data(project_id, sqlc.arg(salt_key), username) AS username,
    decrypt_data(project_id, sqlc.arg(salt_key), PASSWORD) AS PASSWORD,
    version,
    created_at
FROM
    postgres_databases
WHERE
    project_id = $1;

-- name: GetPostgresDatabase :one
SELECT
    id,
    project_id,
    host,
    port,
    database_name,
    decrypt_data(project_id, sqlc.arg(salt_key), username) AS username,
    decrypt_data(project_id, sqlc.arg(salt_key), PASSWORD) AS PASSWORD,
    version,
    created_at,
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
    postgres_databases.id AS database_id,
    postgres_databases.project_id AS database_project_id,
    postgres_databases.host AS database_host,
    postgres_databases.port AS database_port,
    postgres_databases.database_name AS database_database_name,
    decrypt_data(postgres_databases.project_id, sqlc.arg(salt_key), postgres_databases.username) AS database_username,
    decrypt_data(postgres_databases.project_id, sqlc.arg(salt_key), postgres_databases.PASSWORD) AS database_PASSWORD,
    postgres_databases.version AS database_version,
    postgres_databases.created_at AS database_created_at,
    sqlc.embed(postgres_scans)
FROM
    projects
    JOIN postgres_databases ON postgres_databases.project_id = projects.id
    JOIN postgres_scans ON postgres_scans.database_id = postgres_databases.id
WHERE
    postgres_scans.scan_id = $1;

-- name: CreatePostgresDatabase :one
INSERT INTO postgres_databases(project_id, database_name, host, port, username, PASSWORD, version)
    VALUES ($1, $2, $3, $4, encrypt_data($1, sqlc.arg(salt_key), sqlc.arg(username)), encrypt_data($1, sqlc.arg(salt_key), sqlc.arg(PASSWORD)), $5)
RETURNING
    *;

-- name: DeletePostgresDatabase :exec
DELETE FROM postgres_databases
WHERE id = $1;

