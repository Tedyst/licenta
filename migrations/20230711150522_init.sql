
-- +migrate Up
CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  password TEXT NOT NULL,
  email TEXT NOT NULL,
  admin BOOLEAN NOT NULL DEFAULT FALSE
);

-- +migrate Down
DROP TABLE users;
