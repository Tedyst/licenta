-- name: GetNvdCPEsByDBType :many
SELECT
    *
FROM
    nvd_cpes
WHERE
    database_type = $1;

-- name: CreateNvdCPE :one
INSERT INTO nvd_cpes(cpe, database_type, version, last_modified)
    VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: UpdateNvdCPE :exec
UPDATE
    nvd_cpes
SET
    version = coalesce(sqlc.narg(version), version),
    last_modified = coalesce(sqlc.narg(last_modified), last_modified)
WHERE
    id = $1;

-- name: CreateNvdCve :one
INSERT INTO nvd_cves(cve_id, description, published, last_modified, score)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: CreateNvdCveCPE :one
INSERT INTO nvd_cve_cpes(cve_id, cpe_id)
    VALUES ($1, $2)
RETURNING
    *;

-- name: GetNvdCveByCveID :one
SELECT
    *
FROM
    nvd_cves
WHERE
    cve_id = $1;

-- name: GetCPEByProductAndVersion :one
SELECT
    *
FROM
    nvd_cpes
WHERE
    database_type = $1
    AND version = $2
LIMIT 1;

-- name: GetCveByCveID :one
SELECT
    *
FROM
    nvd_cves
WHERE
    cve_id = $1;

-- name: GetCveCpeByCveAndCpe :one
SELECT
    *
FROM
    nvd_cve_cpes
WHERE
    cve_id = $1
    AND cpe_id = $2;

