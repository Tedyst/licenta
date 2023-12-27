-- name: CreateWebauthnCredential :one
INSERT INTO webauthn_credentials(user_id, credential_id, public_key, attestation_type, transport, user_present, user_verified, backup_eligible, backup_state, aa_guid, sign_count, clone_warning, attachment)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING
    *;

-- name: GetWebauthnCredentialsByUserID :many
SELECT
    *
FROM
    webauthn_credentials
WHERE
    user_id = $1;

