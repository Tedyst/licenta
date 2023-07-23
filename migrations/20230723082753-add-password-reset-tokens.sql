
-- +migrate Up
CREATE TABLE reset_password_tokens (
  id UUID PRIMARY KEY,
  user_id BIGINT REFERENCES users(id),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE reset_password_tokens;
