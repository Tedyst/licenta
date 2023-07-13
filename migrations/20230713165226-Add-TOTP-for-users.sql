
-- +migrate Up
ALTER TABLE users ADD COLUMN totp_secret TEXT;

-- +migrate Down
ALTER TABLE users DROP COLUMN totp_secret;
