-- name: CreateMysqlScan :one
INSERT INTO mysql_scans(scan_id, database_id)
    VALUES ($1, $2)
RETURNING
    *;

-- name: GetMysqlScan :one
SELECT
    *
FROM
    mysql_scans
WHERE
    id = $1
LIMIT 1;

-- name: GetMysqlScanByScanID :one
SELECT
    *
FROM
    mysql_scans
WHERE
    scan_id = $1
LIMIT 1;

-- name: UpdateMysqlVersion :exec
UPDATE
    mysql_databases
SET
    version = $2
WHERE
    id = $1;

-- name: UpdateMysqlDatabase :exec
UPDATE
    mysql_databases
SET
    database_name = $2,
    host = $3,
    port = $4,
    username = encrypt_data(sqlc.arg(project_id), sqlc.arg(salt_key), sqlc.arg(username)),
    PASSWORD = encrypt_data(sqlc.arg(project_id), sqlc.arg(salt_key), sqlc.arg(PASSWORD)),
    version = $5
WHERE
    id = $1;

-- name: GetMysqlDatabasesForProject :many
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
    mysql_databases
WHERE
    project_id = $1;

-- name: GetMysqlDatabase :one
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
            Mysql_scans
        WHERE
            Mysql_scans.database_id = mysql_databases.id) AS scan_count
FROM
    mysql_databases
WHERE
    mysql_databases.id = $1;

-- name: GetProjectInfoForMysqlScanByScanID :one
SELECT
    sqlc.embed(projects),
    mysql_databases.id AS database_id,
    mysql_databases.project_id AS database_project_id,
    mysql_databases.host AS database_host,
    mysql_databases.port AS database_port,
    mysql_databases.database_name AS database_database_name,
    decrypt_data(project_id, sqlc.arg(salt_key), mysql_databases.username) AS database_username,
    decrypt_data(project_id, sqlc.arg(salt_key), mysql_databases.PASSWORD) AS database_PASSWORD,
    mysql_databases.version AS database_version,
    mysql_databases.created_at AS database_created_at,
    sqlc.embed(mysql_scans)
FROM
    projects
    JOIN mysql_databases ON mysql_databases.project_id = projects.id
    JOIN mysql_scans ON mysql_scans.database_id = mysql_databases.id
WHERE
    mysql_scans.scan_id = $1;

-- name: CreateMysqlDatabase :one
INSERT INTO mysql_databases(project_id, database_name, host, port, username, PASSWORD, version)
    VALUES ($1, $2, $3, $4, encrypt_data($1, sqlc.arg(salt_key), sqlc.arg(username)), encrypt_data($1, sqlc.arg(salt_key), sqlc.arg(PASSWORD)), $5)
RETURNING
    *;

-- name: DeleteMysqlDatabase :exec
DELETE FROM mysql_databases
WHERE id = $1;

