-- name: GetProjectByOrganizationAndName :one
SELECT
    *
FROM
    projects
WHERE
    organization_id = $1
    AND name = $2
LIMIT 1;

-- name: GetProjectMembers :many
SELECT
    *
FROM
    project_members
WHERE
    project_id = $1;

-- name: GetProjectsByOrganization :many
SELECT
    *
FROM
    projects
WHERE
    organization_id = $1;

-- name: GetProjectUser :one
SELECT
    *
FROM
    project_members
WHERE
    project_id = $1
    AND user_id = $2
LIMIT 1;

-- name: GetProjectPermissionsForUser :one
SELECT
    MIN(ROLE)::smallint AS role
FROM (
    SELECT
        project_members.role AS role
    FROM
        project_members
    WHERE
        project_members.project_id = $1
        AND project_members.user_id = $2
    UNION
    SELECT
        organization_members.role AS role
    FROM
        organization_members
    WHERE
        organization_id = $3
        AND user_id = $2) AS role;

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

