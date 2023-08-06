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
LEFT JOIN sessions ON sessions.user_id = users.id
WHERE sessions.id = $1 LIMIT 1;
