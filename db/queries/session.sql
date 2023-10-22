-- name: GetSession :one
SELECT
  *
FROM
  sessions
WHERE
  id = $1
LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions(id, user_id, scope)
  VALUES ($1, $2, $3)
RETURNING
  *;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions
WHERE user_id = $1;

-- name: UpdateSession :exec
UPDATE
  sessions
SET
  user_id = coalesce(sqlc.narg(user_id), user_id),
  scope = coalesce(sqlc.narg(scope), scope)
WHERE
  id = $1;

