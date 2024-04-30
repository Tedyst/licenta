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
    username = $5,
    PASSWORD = $6,
    version = $7
WHERE
    id = $1;

-- name: GetMysqlDatabasesForProject :many
SELECT
    *
FROM
    mysql_databases
WHERE
    project_id = $1;

-- name: GetMysqlDatabase :one
SELECT
    sqlc.embed(mysql_databases),
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
    sqlc.embed(mysql_databases),
    sqlc.embed(mysql_scans)
FROM
    projects
    JOIN mysql_databases ON mysql_databases.project_id = projects.id
    JOIN mysql_scans ON mysql_scans.database_id = mysql_databases.id
WHERE
    mysql_scans.scan_id = $1;

-- name: CreateMysqlDatabase :one
INSERT INTO mysql_databases(project_id, database_name, host, port, username, PASSWORD, version)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    *;

