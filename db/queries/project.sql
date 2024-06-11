-- name: GetProjectByOrganizationAndName :one
SELECT
    *
FROM
    projects
WHERE
    organization_id = $1
    AND name = $2
LIMIT 1;

-- name: GetProjectsByOrganization :many
SELECT
    *
FROM
    projects
WHERE
    organization_id = $1;

-- name: GetProject :one
SELECT
    *
FROM
    projects
WHERE
    id = $1
LIMIT 1;

-- name: GetProjectWithStats :one
SELECT
    *,
(
        SELECT
            COUNT(*)
        FROM
            scans
            INNER JOIN scan_groups ON scans.scan_group_id = scan_groups.id
        WHERE
            scan_groups.project_id = projects.id) AS scans
FROM
    projects
WHERE
    projects.id = $1;

-- name: CreateProject :one
INSERT INTO projects(organization_id, name, remote)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: DeleteProject :one
DELETE FROM projects
WHERE id = $1
RETURNING
    *;

-- name: GetProjects :many
SELECT
    *
FROM
    projects;

