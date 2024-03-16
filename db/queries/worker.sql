-- name: GetWorkersForProject :many
SELECT
    workers.*
FROM
    workers
    INNER JOIN worker_projects ON workers.id = worker_projects.worker_id
WHERE
    worker_projects.project_id = $1;

-- name: BindScanToWorker :one
UPDATE
    scans
SET
    worker_id = $2
WHERE
    id = $1
    AND worker_id IS NULL
RETURNING
    *;

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

-- name: GetWorkerForProject :one
SELECT
    workers.*
FROM
    workers
    INNER JOIN worker_projects ON workers.id = worker_projects.worker_id
WHERE
    worker_projects.project_id = $1
    AND workers.token = $2;

