
-- +migrate Up
CREATE TABLE sessions (
  id UUID PRIMARY KEY,
  user_id BIGINT REFERENCES users(id),
  totp_key TEXT,
  waiting_2fa BIGINT REFERENCES users(id),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE sessions;
