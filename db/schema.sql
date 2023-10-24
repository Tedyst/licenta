CREATE TABLE users(
  id bigserial PRIMARY KEY,
  username text NOT NULL UNIQUE,
  password TEXT NOT NULL,
  email text NOT NULL,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE sessions(
  id uuid PRIMARY KEY,
  user_id bigint REFERENCES users(id),
  scope text[] NOT NULL,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE reset_password_tokens(
  id uuid PRIMARY KEY,
  user_id bigint REFERENCES users(id),
  valid boolean NOT NULL DEFAULT TRUE,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE totp_secret_tokens(
  id bigserial PRIMARY KEY,
  user_id bigint REFERENCES users(id) NOT NULL UNIQUE,
  valid boolean NOT NULL DEFAULT TRUE,
  totp_secret text NOT NULL,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE organizations(
  id bigserial PRIMARY KEY,
  name text NOT NULL UNIQUE,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE organization_members(
  id bigserial PRIMARY KEY,
  organization_id bigint REFERENCES organizations(id) NOT NULL,
  user_id bigint REFERENCES users(id) NOT NULL,
  role integer NOT NULL,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE projects(
  id bigserial PRIMARY KEY,
  name text NOT NULL,
  organization_id bigint REFERENCES organizations(id) NOT NULL,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE project_members(
  id bigserial PRIMARY KEY,
  project_id bigint REFERENCES projects(id) NOT NULL,
  user_id bigint REFERENCES users(id) NOT NULL,
  role smallint NOT NULL,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

