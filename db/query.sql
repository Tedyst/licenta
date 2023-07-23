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

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = $1 LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions (
  id, user_id, totp_key
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: UpdateSession :exec
UPDATE sessions SET
    user_id = $2,
    totp_key = $3
WHERE id = $1;

-- name: GetUserAndSessionBySessionID :one
SELECT sqlc.embed(users), sqlc.embed(sessions) FROM users
INNER JOIN sessions ON sessions.user_id = users.id
WHERE sessions.id = $1 LIMIT 1;

-- name: CreateResetPasswordToken :one
INSERT INTO reset_password_tokens (
  id, user_id
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetResetPasswordToken :one
SELECT * FROM reset_password_tokens
WHERE id = $1 LIMIT 1;

-- name: InvalidateResetPasswordToken :exec
UPDATE reset_password_tokens SET
    valid = FALSE
WHERE id = $1;