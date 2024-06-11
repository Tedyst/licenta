-- name: CreateMongoScan :one
INSERT INTO mongo_scans(scan_id, database_id)
    VALUES ($1, $2)
RETURNING
    *;

-- name: GetMongoScan :one
SELECT
    *
FROM
    mongo_scans
WHERE
    id = $1
LIMIT 1;

-- name: GetMongoScanByScanID :one
SELECT
    *
FROM
    mongo_scans
WHERE
    scan_id = $1
LIMIT 1;

-- name: UpdateMongoVersion :exec
UPDATE
    mongo_databases
SET
    version = $2
WHERE
    id = $1;

-- name: UpdateMongoDatabase :exec
UPDATE
    mongo_databases
SET
    database_name = $2,
    host = $3,
    port = $4,
    username = encrypt_data(sqlc.arg(project_id), sqlc.arg(salt_key), sqlc.arg(username)),
    PASSWORD = encrypt_data(sqlc.arg(project_id), sqlc.arg(salt_key), sqlc.arg(PASSWORD)),
    version = $5
WHERE
    id = $1;

-- name: GetMongoDatabasesForProject :many
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
    mongo_databases
WHERE
    project_id = $1;

-- name: GetMongoDatabase :one
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
            mongo_scans
        WHERE
            mongo_scans.database_id = mongo_databases.id) AS scan_count
FROM
    mongo_databases
WHERE
    mongo_databases.id = $1;

-- name: GetProjectInfoForMongoScanByScanID :one
SELECT
    sqlc.embed(projects),
    mongo_databases.id AS database_id,
    mongo_databases.project_id AS database_project_id,
    mongo_databases.host AS database_host,
    mongo_databases.port AS database_port,
    mongo_databases.database_name AS database_database_name,
    decrypt_data(mongo_databases.project_id, sqlc.arg(salt_key), mongo_databases.username) AS database_username,
    decrypt_data(mongo_databases.project_id, sqlc.arg(salt_key), mongo_databases.PASSWORD) AS database_PASSWORD,
    mongo_databases.version AS database_version,
    mongo_databases.created_at AS database_created_at,
    sqlc.embed(mongo_scans)
FROM
    projects
    JOIN mongo_databases ON mongo_databases.project_id = projects.id
    JOIN mongo_scans ON mongo_scans.database_id = mongo_databases.id
WHERE
    mongo_scans.scan_id = $1;

-- name: CreateMongoDatabase :one
INSERT INTO mongo_databases(project_id, database_name, host, port, username, PASSWORD, version)
    VALUES ($1, $2, $3, $4, encrypt_data($1, sqlc.arg(salt_key), sqlc.arg(username)), encrypt_data($1, sqlc.arg(salt_key), sqlc.arg(PASSWORD)), $5)
RETURNING
    *;

-- name: DeleteMongoDatabase :exec
DELETE FROM mongo_databases
WHERE id = $1;

