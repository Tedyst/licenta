-- name: InsertBruteforcePasswords :exec
INSERT INTO default_bruteforce_passwords(PASSWORD)
    VALUES (unnest(sqlc.arg(passwords)::text[]))
ON CONFLICT
    DO NOTHING;

-- name: InsertBruteforcedPassword :exec
INSERT INTO bruteforced_passwords(hash, username, PASSWORD, last_bruteforce_id)
    VALUES ($1, $2, $3, $4)
ON CONFLICT
    DO NOTHING;

-- name: GetBruteforcedPasswordByHashAndUsername :one
SELECT
    *
FROM
    bruteforced_passwords
WHERE
    hash = $1
    AND username = $2
LIMIT 1;

