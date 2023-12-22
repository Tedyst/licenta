-- name: GetWorkersForProject :many
SELECT
    sqlc.embed(workers)
FROM
    workers
    INNER JOIN worker_projects ON workers.id = worker_projects.worker_id
WHERE
    worker_projects.project_id = $1;

-- name: BindPostgresScanToWorker :exec
UPDATE
    postgres_scan
SET
    worker_id = $2
WHERE
    id = $1;

-- name: GetWorkerForPostgresScan :one
SELECT
    sqlc.embed(workers)
FROM
    workers
    INNER JOIN postgres_scan ON workers.id = postgres_scan.worker_id
WHERE
    postgres_scan.id = $1;

