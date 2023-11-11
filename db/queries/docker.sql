-- name: GetDockerScannedLayersForProject :many
SELECT
    layer_hash
FROM
    project_docker_scanned_layers
WHERE
    project_id = $1;

-- name: CreateDockerScannedLayerForProject :one
INSERT INTO project_docker_scanned_layers(project_id, layer_hash)
    VALUES ($1, $2)
RETURNING
    *;

-- name: GetDockerImagesForProject :many
SELECT
    *
FROM
    project_docker_images
WHERE
    project_id = $1;

-- name: CreateDockerImageForProject :one
INSERT INTO project_docker_images(project_id, docker_image, username, PASSWORD)
    VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: DeleteDockerImageForProject :exec
DELETE FROM project_docker_images
WHERE project_id = $1
    AND docker_image = $2;

-- name: CreateDockerLayerResultsForProject :copyfrom
INSERT INTO project_docker_layer_results(project_id, layer, name, line, line_number, MATCH, probability, username, PASSWORD, filename)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: CreateDockerLayerScanForProject :one
INSERT INTO project_docker_layer_scans(project_id, docker_image, layers_to_scan)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: UpdateDockerLayerScanForProject :one
UPDATE
    project_docker_layer_scans
SET
    finished = $3,
    scanned_layers = $4
WHERE
    project_id = $1
    AND docker_image = $2
RETURNING
    *;

-- name: GetDockerLayerScanForProject :one
SELECT
    *
FROM
    project_docker_layer_scans
WHERE
    project_id = $1
    AND docker_image = $2;

