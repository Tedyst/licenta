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

-- name: GetProjectByID :one
SELECT
    *
FROM
    projects
WHERE
    id = $1
LIMIT 1;

