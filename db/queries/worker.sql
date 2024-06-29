-- name: GetWorkersForProject :many
SELECT
    workers.*
FROM
    workers
    INNER JOIN organizations ON workers.organization = organizations.id
    INNER JOIN projects ON organizations.id = projects.organization_id
WHERE
    projects.id = sqlc.arg(project_id);

-- name: BindScanToWorker :one
UPDATE
    scans
SET
    worker_id = $2
WHERE
    id = $1
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
    INNER JOIN organizations ON workers.organization = organizations.id
    INNER JOIN projects ON organizations.id = projects.organization_id
WHERE
    projects.id = sqlc.arg(project_id)
    AND workers.token = $1;

-- name: DeleteWorker :one
DELETE FROM workers
WHERE id = $1
RETURNING
    *;

-- name: GetWorker :one
SELECT
    *
FROM
    workers
WHERE
    id = $1;

-- name: GetWorkersByProject :many
SELECT
    workers.*
FROM
    workers
    INNER JOIN organizations ON workers.organization = organizations.id
    INNER JOIN projects ON organizations.id = projects.organization_id
WHERE
    projects.id = sqlc.arg(project_id);

-- name: CreateWorker :one
INSERT INTO workers(organization, name, token)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: GetWorkersForOrganization :many
SELECT
    workers.*
FROM
    workers
WHERE
    workers.organization = $1;

