-- name: GetDockerScannedLayersForImage :many
SELECT
    layer_hash
FROM
    docker_layers
WHERE
    image_id = $1;

-- name: CreateDockerScannedLayerForProject :one
INSERT INTO docker_layers(layer_hash, image_id)
    VALUES ($1, $2)
RETURNING
    *;

-- name: GetDockerImage :one
SELECT
    id,
    project_id,
    docker_image,
    decrypt_data(project_id, sqlc.arg(salt_key), username) AS username,
    decrypt_data(project_id, sqlc.arg(salt_key), PASSWORD) AS PASSWORD
FROM
    docker_images
WHERE
    id = $1;

-- name: GetDockerImagesForProject :many
SELECT
    id,
    project_id,
    docker_image,
    decrypt_data(project_id, sqlc.arg(salt_key), username) AS username,
    decrypt_data(project_id, sqlc.arg(salt_key), PASSWORD) AS PASSWORD
FROM
    docker_images
WHERE
    project_id = $1;

-- name: CreateDockerLayerResultsForProject :copyfrom
INSERT INTO docker_results(project_id, layer_id, name, line, line_number, MATCH, probability, username, PASSWORD, filename, previous_lines)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: DeleteDockerImage :exec
DELETE FROM docker_images
WHERE id = $1;

-- name: GetDockerLayersAndResultsForImage :many
SELECT
    *
FROM ((
        SELECT
            docker_layers.id AS lid,
            docker_layers.image_id,
            docker_layers.layer_hash,
            docker_layers.scanned_at,
            docker_results.*
        FROM
            docker_layers
        LEFT JOIN docker_results ON docker_layers.id = docker_results.layer_id
    WHERE
        docker_layers.image_id = $1
        AND docker_layers.id IS NULL
    ORDER BY
        docker_layers.scanned_at DESC
    LIMIT 25)
UNION (
    SELECT
        docker_layers.id AS lid,
        docker_layers.image_id,
        docker_layers.layer_hash,
        docker_layers.scanned_at,
        docker_results.*
    FROM
        docker_layers
    LEFT JOIN docker_results ON docker_layers.id = docker_results.layer_id
WHERE
    docker_layers.image_id = $1
    AND docker_layers.id IS NOT NULL)) AS asd
ORDER BY
    scanned_at DESC;

-- name: UpdateDockerImage :one
UPDATE
    docker_images
SET
    docker_image = $2,
    username = encrypt_data(sqlc.arg(project_id), sqlc.arg(salt_key), sqlc.arg(username)),
    PASSWORD = encrypt_data(sqlc.arg(project_id), sqlc.arg(salt_key), sqlc.arg(PASSWORD))
WHERE
    id = $1
RETURNING
    *;

-- name: CreateDockerImage :one
INSERT INTO docker_images(project_id, docker_image, username, PASSWORD)
    VALUES ($1, $2, encrypt_data($1, sqlc.arg(salt_key), sqlc.arg(username)), encrypt_data($1, sqlc.arg(salt_key), sqlc.arg(PASSWORD)))
RETURNING
    *;

-- name: CreateDockerScan :one
INSERT INTO docker_scans(image_id, scan_id)
    VALUES ($1, $2)
RETURNING
    *;

-- name: GetDockerScanByScanAndRepo :one
SELECT
    sqlc.embed(docker_scans),
    sqlc.embed(scans)
FROM
    docker_scans
    INNER JOIN scans ON scans.id = docker_scans.scan_id
WHERE
    scans.scan_group_id = $1
    AND image_id = $2;

