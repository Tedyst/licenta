-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY id;

-- name: CreateUser :one
INSERT INTO users (
  username, password, email
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: UpdateUser :exec
UPDATE users SET
    username = $2,
    password = $3,
    email = $4,
    admin = $5
WHERE id = $1;

-- name: GetUserByUsernameOrEmail :one
SELECT * FROM users
WHERE username = $1 OR email = $1 LIMIT 1;

-- name: UpdateUserTOTPSecret :exec
UPDATE users SET
    totp_secret = $2
WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users SET
    password = $2
WHERE id = $1;