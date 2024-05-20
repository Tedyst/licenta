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
    *
FROM
    docker_images
WHERE
    id = $1;

-- name: GetDockerImagesForProject :many
SELECT
    *
FROM
    docker_images
WHERE
    project_id = $1;

-- name: CreateDockerImageForProject :one
INSERT INTO docker_images(project_id, docker_image, username, PASSWORD)
    VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: DeleteDockerImageForProject :exec
DELETE FROM docker_images
WHERE project_id = $1
    AND docker_image = $2;

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
    docker_layers.id IS NOT NULL)) AS asd
ORDER BY
    scanned_at DESC;

-- name: UpdateDockerImage :one
UPDATE
    docker_images
SET
    docker_image = $2,
    username = $3,
    PASSWORD = $4,
    min_probability = $5,
    probability_decrease_multiplier = $6,
    probability_increase_multiplier = $7,
    entropy_threshold = $8,
    logistic_growth_rate = $9
WHERE
    id = $1
RETURNING
    *;

-- name: CreateDockerImage :one
INSERT INTO docker_images(project_id, docker_image, username, PASSWORD, min_probability, probability_decrease_multiplier, probability_increase_multiplier, entropy_threshold, logistic_growth_rate)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING
    *;

