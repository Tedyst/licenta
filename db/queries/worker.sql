-- name: GetWorkersForProject :many
SELECT
    workers.*
FROM
    workers
    INNER JOIN worker_projects ON workers.id = worker_projects.worker_id
WHERE
    worker_projects.project_id = $1;

-- name: BindScanToWorker :exec
UPDATE
    scans
SET
    worker_id = $2
WHERE
    id = $1;

-- name: GetWorkerForScan :one
SELECT
    workers.*
FROM
    workers
    INNER JOIN scans ON workers.id = scans.worker_id
WHERE
    scans.id = $1;

-- name: GetWorkerByToken :one
SELECT
    *
FROM
    workers
WHERE
    workers.token = $1;

