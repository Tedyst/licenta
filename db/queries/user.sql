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
  username = coalesce(sqlc.narg(username), username),
  PASSWORD = coalesce(sqlc.narg(PASSWORD), PASSWORD),
  email = coalesce(sqlc.narg(email), email)
WHERE
  id = sqlc.arg(id);

-- name: GetUserByUsernameOrEmail :one
SELECT
  *
FROM
  users
WHERE
  username = $1
  OR email = $1
LIMIT 1;

-- name: UpdateUserPassword :exec
UPDATE
  users
SET
  PASSWORD = $2
WHERE
  id = $1;

