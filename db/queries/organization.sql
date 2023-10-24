-- name: GetOrganizationByName :one
SELECT
    *
FROM
    organizations
WHERE
    name = $1
LIMIT 1;

-- name: GetOrganizationMembers :many
SELECT
    *
FROM
    organization_members
WHERE
    organization_id = $1;

-- name: GetOrganizationsByUser :many
SELECT
    organizations.*
FROM
    organizations
    INNER JOIN organization_members ON organizations.id = organization_members.organization_id
WHERE
    organization_members.user_id = $1;

-- name: GetOrganizationUser :one
SELECT
    *
FROM
    organization_members
WHERE
    organization_id = $1
    AND user_id = $2
LIMIT 1;

-- name: GetOrganizationPermissionsForUser :one
SELECT
    organization_members.role::smallint AS role
FROM
    organization_members
WHERE
    organization_id = $1
    AND user_id = $2;

