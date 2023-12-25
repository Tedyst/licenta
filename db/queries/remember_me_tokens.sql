-- name: CreateRememberMeToken :one
INSERT INTO remember_me_tokens(user_id, token)
    VALUES ($1, $2)
RETURNING
    *;

-- name: DeleteRememberMeTokensForUser :exec
DELETE FROM remember_me_tokens
WHERE user_id = $1;

-- name: DeleteRememberMeTokenByUserAndToken :exec
DELETE FROM remember_me_tokens
WHERE user_id = $1
    AND token = $2;

