-- name: GetUser :one
SELECT
  *
FROM
  users
WHERE
  id = $1
LIMIT 1;

-- name: ListUsers :many
SELECT
  *
FROM
  users
WHERE
  CASE WHEN @username::text = '' THEN
    TRUE
  ELSE
    username = @username::text
  END
  AND CASE WHEN @email::text = '' THEN
    TRUE
  ELSE
    email = @email::text
  END
  AND CASE WHEN @admin::text = '' THEN
    TRUE
  ELSE
    admin = @admin::boolean
  END
ORDER BY
  id;

-- name: ListUsersPaginated :many
SELECT
  *
FROM
  users
ORDER BY
  id
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT
  COUNT(*)
FROM
  users;

-- name: CreateUser :one
INSERT INTO users(username, PASSWORD, email)
  VALUES ($1, $2, $3)
RETURNING
  *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: UpdateUser :exec
UPDATE
  users
SET
  username = $2,
  PASSWORD = $3,
  email = $4,
  recovery_codes = $5,
  totp_secret = $6,
  recover_selector = $7,
  recover_verifier = $8,
  recover_expiry = $9,
  login_attempt_count = $10,
  login_last_attempt = $11,
  LOCKED = $12,
  confirm_selector = $13,
  confirm_verifier = $14,
  confirmed = $15
WHERE
  id = $1;

-- name: GetUserByUsernameOrEmail :one
SELECT
  *
FROM
  users
WHERE
  username = $1
  OR email = $2
LIMIT 1;

-- name: GetUserByRecoverSelector :one
SELECT
  *
FROM
  users
WHERE
  recover_selector = $1
LIMIT 1;

