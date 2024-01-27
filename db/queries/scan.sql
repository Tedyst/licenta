-- name: CreateScanGroup :one
INSERT INTO scan_groups(project_id, created_by)
    VALUES ($1, $2)
RETURNING
    *;

-- name: CreateScan :one
INSERT INTO scans(status, worker_id, scan_group_id)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: GetScanGroup :one
SELECT
    *
FROM
    scan_groups
WHERE
    id = $1;

-- name: CreateScanResult :one
INSERT INTO scan_results(scan_id, severity, message, scan_source)
    VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: UpdateScanStatus :exec
UPDATE
    scans
SET
    status = $2,
    error = $3,
    ended_at = $4
WHERE
    id = $1;

-- name: GetScanResults :many
SELECT
    *
FROM
    scan_results
WHERE
    scan_id = $1;

-- name: GetScanResultsByScanIdAndScanSource :many
SELECT
    *
FROM
    scan_results
WHERE
    scan_id = $1
    AND scan_source = $2;

-- name: GetScan :one
SELECT
    sqlc.embed(scans),
(
        SELECT
            COALESCE(MAX(scan_results.severity), 0)::integer
        FROM
            scan_results
        WHERE
            scan_id = scans.id) AS maximum_severity,
(
        SELECT
            COALESCE(id, 0)::bigint
        FROM
            postgres_scans
        WHERE
            postgres_scans.scan_id = scans.id
        LIMIT 1) AS postgres_scan
FROM
    scans
WHERE
    scans.id = $1;

-- name: CreateScanBruteforceResult :one
INSERT INTO scan_bruteforce_results(scan_id, scan_type, username, PASSWORD, tried, total)
    VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: UpdateScanBruteforceResult :exec
UPDATE
    scan_bruteforce_results
SET
    PASSWORD = $2,
    tried = $3,
    total = $4
WHERE
    id = $1;

-- name: GetScanBruteforceResults :many
SELECT
    *
FROM
    scan_bruteforce_results
WHERE
    scan_id = $1;

-- name: GetScansForProject :many
SELECT
    sqlc.embed(scans),
(
        SELECT
            MAX(scan_results.severity)::integer
        FROM
            scan_results
        WHERE
            scan_id = scans.id) AS maximum_severity,
(
        SELECT
            id
        FROM
            postgres_scans
        WHERE
            postgres_scans.scan_id = scans.id
        LIMIT 1) AS postgres_scan
FROM
    scans
    INNER JOIN scan_groups ON scans.scan_group_id = scan_groups.id
WHERE
    scan_groups.project_id = $1
ORDER BY
    scans.id DESC;

