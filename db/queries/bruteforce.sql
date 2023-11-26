-- name: InsertBruteforcePasswords :exec
INSERT INTO default_bruteforce_passwords(PASSWORD)
    VALUES (unnest(sqlc.arg(passwords)::text[]))
ON CONFLICT
    DO NOTHING;

