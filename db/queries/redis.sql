-- name: CreateRedisScan :one
INSERT INTO redis_scans(scan_id, database_id)
    VALUES ($1, $2)
RETURNING
    *;

-- name: GetRedisScan :one
SELECT
    *
FROM
    redis_scans
WHERE
    id = $1
LIMIT 1;

-- name: GetRedisScanByScanID :one
SELECT
    *
FROM
    redis_scans
WHERE
    scan_id = $1
LIMIT 1;

-- name: UpdateRedisVersion :exec
UPDATE
    redis_databases
SET
    version = $2
WHERE
    id = $1;

-- name: UpdateRedisDatabase :exec
UPDATE
    redis_databases
SET
    host = $2,
    port = $3,
    username = encrypt_data(sqlc.arg(project_id), sqlc.arg(salt_key), sqlc.arg(username)),
    PASSWORD = encrypt_data(sqlc.arg(project_id), sqlc.arg(salt_key), sqlc.arg(PASSWORD)),
    version = $4
WHERE
    id = $1;

-- name: GetRedisDatabasesForProject :many
SELECT
    id,
    project_id,
    host,
    port,
    decrypt_data(project_id, sqlc.arg(salt_key), username) AS username,
    decrypt_data(project_id, sqlc.arg(salt_key), PASSWORD) AS PASSWORD,
    version,
    created_at
FROM
    redis_databases
WHERE
    project_id = $1;

-- name: GetRedisDatabase :one
SELECT
    id,
    project_id,
    host,
    port,
    decrypt_data(project_id, sqlc.arg(salt_key), username) AS username,
    decrypt_data(project_id, sqlc.arg(salt_key), PASSWORD) AS PASSWORD,
    version,
    created_at,
(
        SELECT
            COUNT(*)
        FROM
            redis_scans
        WHERE
            redis_scans.database_id = redis_databases.id) AS scan_count
FROM
    redis_databases
WHERE
    redis_databases.id = $1;

-- name: GetProjectInfoForRedisScanByScanID :one
SELECT
    sqlc.embed(projects),
    redis_databases.id AS database_id,
    redis_databases.project_id AS database_project_id,
    redis_databases.host AS database_host,
    redis_databases.port AS database_port,
    decrypt_data(redis_databases.project_id, sqlc.arg(salt_key), redis_databases.username) AS database_username,
    decrypt_data(redis_databases.project_id, sqlc.arg(salt_key), redis_databases.PASSWORD) AS database_PASSWORD,
    redis_databases.version AS database_version,
    redis_databases.created_at AS database_created_at,
    sqlc.embed(redis_scans)
FROM
    projects
    JOIN redis_databases ON redis_databases.project_id = projects.id
    JOIN redis_scans ON redis_scans.database_id = redis_databases.id
WHERE
    redis_scans.scan_id = $1;

-- name: CreateRedisDatabase :one
INSERT INTO redis_databases(project_id, host, port, username, PASSWORD, version)
    VALUES ($1, $2, $3, encrypt_data($1, sqlc.arg(salt_key), sqlc.arg(username)), encrypt_data($1, sqlc.arg(salt_key), sqlc.arg(PASSWORD)), $4)
RETURNING
    *;

-- name: DeleteRedisDatabase :exec
DELETE FROM redis_databases
WHERE id = $1;

