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
  organization_id bigint NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX projects_organization_id_idx ON projects(organization_id);

CREATE INDEX projects_name_orgianization_id_idx ON projects(name, organization_id);

CREATE TABLE project_members(
  id bigserial PRIMARY KEY,
  project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  user_id bigint REFERENCES users(id) NOT NULL,
  role smallint NOT NULL,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE project_git_repositories(
  id bigserial PRIMARY KEY,
  project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  git_repository text NOT NULL,
  username text,
  password text,
  private_key text,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE project_git_scanned_commits(
  id bigserial PRIMARY KEY,
  project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  commit_hash text NOT NULL UNIQUE,
  scanned_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE project_git_results(
  id bigserial PRIMARY KEY,
  project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  commit bigint REFERENCES project_git_scanned_commits(id) NOT NULL,
  result jsonb NOT NULL,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE project_docker_images(
  id bigserial PRIMARY KEY,
  project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  docker_image text NOT NULL,
  username text,
  password text,
  min_probability float,
  use_default_words_reduce_probability boolean DEFAULT TRUE NOT NULL,
  use_default_words_increase_probability boolean DEFAULT TRUE NOT NULL,
  use_default_passwords_completely_ignore boolean DEFAULT TRUE NOT NULL,
  use_default_usernames_completely_ignore boolean DEFAULT TRUE NOT NULL,
  probaility_decrease_multiplier float,
  probability_increase_multiplier float,
  entropy_threshold integer,
  logistic_growth_rate float,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE project_docker_layer_scans(
  id bigserial PRIMARY KEY,
  project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  docker_image bigint REFERENCES project_docker_images(id) ON DELETE CASCADE NOT NULL,
  finished boolean NOT NULL DEFAULT FALSE,
  scanned_layers integer NOT NULL DEFAULT 0,
  layers_to_scan integer NOT NULL DEFAULT 0,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE project_docker_scanned_layers(
  id bigserial PRIMARY KEY,
  project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  layer_hash text NOT NULL UNIQUE,
  scanned_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE project_docker_layer_results(
  id bigserial PRIMARY KEY,
  project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  layer bigint REFERENCES project_docker_scanned_layers(id) ON DELETE CASCADE NOT NULL,
  name text NOT NULL,
  line text NOT NULL,
  line_number integer NOT NULL,
  match text NOT NULL,
  probability float NOT NULL,
  username text,
  password text,
  filename text NOT NULL,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

