-- name: GetDockerScannedLayersForProject :many
SELECT
    layer_hash
FROM
    project_docker_scanned_layers
WHERE
    project_id = $1;

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

