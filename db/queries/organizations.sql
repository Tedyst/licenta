-- name: GetOrganizationsForUser :many
SELECT
    *
FROM
    organizations
WHERE
    id IN (
        SELECT
            organization_id
        FROM
            organization_members
        WHERE
            user_id = $1);

-- name: GetOrganization :one
SELECT
    *
FROM
    organizations
WHERE
    id = $1;

-- name: GetOrganizationProjects :many
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
    organization_id = $1;

-- name: GetAllOrganizationProjectsForUser :many
SELECT
    *
FROM
    projects
WHERE
    organization_id IN (
        SELECT
            id
        FROM
            organizations
        WHERE
            id IN (
                SELECT
                    organization_id
                FROM
                    organization_members
                WHERE
                    user_id = $1));

-- name: GetAllOrganizationMembersForOrganizationsThatContainUser :many
SELECT
    *
FROM
    organization_members
    INNER JOIN users ON organization_members.user_id = users.id
WHERE
    organization_id IN (
        SELECT
            id
        FROM
            organizations
        WHERE
            id IN (
                SELECT
                    organization_id
                FROM
                    organization_members
                WHERE
                    organization_members.user_id = $1));

-- name: DeleteOrganization :exec
DELETE FROM organizations
WHERE id = $1;

-- name: CreateOrganization :one
INSERT INTO organizations(name)
    VALUES ($1)
RETURNING
    *;

-- name: AddUserToOrganization :exec
INSERT INTO organization_members(organization_id, user_id, ROLE)
    VALUES ($1, $2, $3);

