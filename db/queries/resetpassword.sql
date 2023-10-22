-- name: CreateResetPasswordToken :one
INSERT INTO reset_password_tokens(id, user_id)
    VALUES ($1, $2)
RETURNING
    *;

-- name: GetResetPasswordToken :one
SELECT
    *
FROM
    reset_password_tokens
WHERE
    id = $1
LIMIT 1;

-- name: InvalidateResetPasswordToken :exec
UPDATE
    reset_password_tokens
SET
    valid = FALSE
WHERE
    id = $1;

