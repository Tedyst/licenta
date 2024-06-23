-- name: CreateWebauthnCredential :one
INSERT INTO webauthn_credentials(user_id, credential_id, name, public_key, attestation_type, transport, user_present, user_verified, backup_eligible, backup_state, aa_guid, sign_count, clone_warning, attachment)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING
    *;

-- name: GetWebauthnCredentialsByUserID :many
SELECT
    *
FROM
    webauthn_credentials
WHERE
    user_id = $1;

-- name: GetUserByWebauthnCredentialID :one
SELECT
    sqlc.embed(users)
FROM
    webauthn_credentials
    JOIN users ON webauthn_credentials.user_id = users.id
WHERE
    webauthn_credentials.credential_id = $1;

-- name: UpdateWebauthnCredential :one
UPDATE
    webauthn_credentials
SET
    user_id = $1,
    credential_id = $2,
    public_key = $3,
    attestation_type = $4,
    transport = $5,
    user_present = $6,
    user_verified = $7,
    backup_eligible = $8,
    backup_state = $9,
    aa_guid = $10,
    sign_count = $11,
    clone_warning = $12,
    attachment = $13
WHERE
    id = $14
RETURNING
    *;

