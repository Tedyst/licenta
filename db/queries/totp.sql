-- name: GetTOTPSecretForUser :one
SELECT
    *
FROM
    totp_secret_tokens
WHERE
    user_id = $1
    AND valid = TRUE
LIMIT 1;

-- name: CreateTOTPSecretForUser :one
INSERT INTO totp_secret_tokens(user_id, totp_secret, valid)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: InvalidateTOTPSecretForUser :exec
UPDATE
    totp_secret_tokens
SET
    valid = FALSE
WHERE
    user_id = $1
    AND valid = TRUE;

-- name: GetInvalidTOTPSecretForUser :one
SELECT
    *
FROM
    totp_secret_tokens
WHERE
    user_id = $1
    AND valid = FALSE
LIMIT 1;

-- name: ValidateTOTPSecretForUser :exec
UPDATE
    totp_secret_tokens
SET
    valid = TRUE
WHERE
    user_id = $1
    AND valid = FALSE;

