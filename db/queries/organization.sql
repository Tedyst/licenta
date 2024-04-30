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
    sqlc.embed(organizations),
(
        SELECT
            COUNT(*)
        FROM
            organization_members
        WHERE
            organization_id = organizations.id) AS users,
(
        SELECT
            COUNT(*)
        FROM
            projects
        WHERE
            organization_id = organizations.id) AS projects,
(
        SELECT
            COUNT(*)
        FROM
            scans
            INNER JOIN scan_groups ON scans.scan_group_id = scan_groups.id
            INNER JOIN projects ON scan_groups.project_id = projects.id
        WHERE
            projects.organization_id = organizations.id) AS scans,
(
        SELECT
            COALESCE(MAX(scan_results.severity), 0)::integer
        FROM
            scan_results
            INNER JOIN scans ON scan_results.scan_id = scans.id
            INNER JOIN scan_groups ON scans.scan_group_id = scan_groups.id
            INNER JOIN projects ON scan_groups.project_id = projects.id
        WHERE
            projects.organization_id = organizations.id) AS maximum_severity
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

-- name: GetOrganizationPermissionForUser :one
SELECT
    ROLE
FROM
    organization_members
WHERE
    organization_id = $1
    AND user_id = $2;

-- name: SetOrganizationPermissionsForUser :one
UPDATE
    organization_members
SET
    ROLE = $3
WHERE
    organization_id = $1
    AND user_id = $2
RETURNING
    *;

-- name: RemoveOrganizationUser :one
DELETE FROM organization_members
WHERE organization_id = $1
    AND user_id = $2
RETURNING
    *;

-- name: AddOrganizationUser :one
INSERT INTO organization_members(organization_id, user_id, ROLE)
    VALUES ($1, $2, $3)
RETURNING
    *;

