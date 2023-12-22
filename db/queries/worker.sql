-- name: GetWorkersForProject :many
SELECT
    sqlc.embed(workers)
FROM
    workers
    INNER JOIN worker_projects ON workers.id = worker_projects.worker_id
WHERE
    worker_projects.project_id = $1;

