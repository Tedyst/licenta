CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users(
    id bigserial PRIMARY KEY,
    username text NOT NULL UNIQUE,
    password TEXT NOT NULL,
    email text NOT NULL UNIQUE,
    recovery_codes text,
    totp_secret text,
    recover_selector text UNIQUE,
    recover_verifier text UNIQUE,
    recover_expiry timestamp with time zone,
    login_attempt_count integer NOT NULL DEFAULT 0,
    login_last_attempt timestamp with time zone,
    locked timestamp with time zone,
    confirm_selector text UNIQUE,
    confirm_verifier text UNIQUE,
    confirmed boolean NOT NULL DEFAULT FALSE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX users_username_idx ON users(username);

CREATE INDEX users_email_idx ON users(email);

CREATE TABLE reset_password_tokens(
    id uuid PRIMARY KEY,
    user_id bigint REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    valid boolean NOT NULL DEFAULT TRUE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE totp_secret_tokens(
    id bigserial PRIMARY KEY,
    user_id bigint REFERENCES users(id) ON DELETE CASCADE NOT NULL UNIQUE,
    valid boolean NOT NULL DEFAULT TRUE,
    totp_secret text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE organizations(
    id bigserial PRIMARY KEY,
    name text NOT NULL UNIQUE,
    encryption_key bytea NOT NULL DEFAULT gen_random_bytes(64),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX organizations_name_idx ON organizations(name);

CREATE TABLE organization_members(
    id bigserial PRIMARY KEY,
    organization_id bigint REFERENCES organizations(id) ON DELETE CASCADE NOT NULL,
    user_id bigint REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    role integer NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX organization_members_organization_id_idx ON organization_members(organization_id);

CREATE TABLE projects(
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    organization_id bigint NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    remote boolean NOT NULL DEFAULT FALSE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX projects_organization_id_idx ON projects(organization_id);

CREATE INDEX projects_name_orgianization_id_idx ON projects(name, organization_id);

CREATE TABLE project_members(
    id bigserial PRIMARY KEY,
    project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id bigint REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    role smallint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE git_repositories(
    id bigserial PRIMARY KEY,
    project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    git_repository text NOT NULL,
    username text NOT NULL,
    password text NOT NULL,
    private_key text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE git_commits(
    id bigserial PRIMARY KEY,
    repository_id bigint NOT NULL REFERENCES git_repositories(id) ON DELETE CASCADE,
    commit_hash text NOT NULL UNIQUE,
    author text,
    author_email text,
    commit_date timestamp with time zone,
    description text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE git_results(
    id bigserial PRIMARY KEY,
    commit bigint REFERENCES git_commits(id) ON DELETE CASCADE NOT NULL,
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

CREATE TABLE docker_images(
    id bigserial PRIMARY KEY,
    project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    docker_image text NOT NULL,
    username text NOT NULL,
    password text NOT NULL,
    min_probability float,
    probability_decrease_multiplier float,
    probability_increase_multiplier float,
    entropy_threshold float,
    logistic_growth_rate float,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE docker_layers(
    id bigserial PRIMARY KEY,
    image_id bigint REFERENCES docker_images(id) ON DELETE CASCADE NOT NULL,
    layer_hash text NOT NULL UNIQUE,
    scanned_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE docker_results(
    id bigserial PRIMARY KEY,
    layer_id bigint REFERENCES docker_layers(id) ON DELETE CASCADE NOT NULL,
    name text NOT NULL,
    line text NOT NULL,
    line_number integer NOT NULL,
    previous_lines text NOT NULL,
    match text NOT NULL,
    probability float NOT NULL,
    username text,
    password text,
    filename text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE nvd_cpes(
    id bigserial PRIMARY KEY,
    cpe text NOT NULL UNIQUE,
    database_type int NOT NULL,
    version text NOT NULL,
    last_modified timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX nvd_cpes_database_version_idx ON nvd_cpes(database_type, version);

CREATE TABLE nvd_cves(
    id bigserial PRIMARY KEY,
    cve_id text NOT NULL UNIQUE,
    description text NOT NULL,
    published timestamp with time zone NOT NULL,
    last_modified timestamp with time zone NOT NULL,
    score float NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE nvd_cve_cpes(
    id bigserial PRIMARY KEY,
    cve_id bigint REFERENCES nvd_cves(id) ON DELETE CASCADE NOT NULL,
    cpe_id bigint REFERENCES nvd_cpes(id) ON DELETE CASCADE NOT NULL
);

CREATE INDEX nvd_cve_cpes_cve_id_idx ON nvd_cve_cpes(cve_id);

CREATE INDEX nvd_cve_cpes_cpe_id_idx ON nvd_cve_cpes(cpe_id);

CREATE TABLE default_bruteforce_passwords(
    id bigserial PRIMARY KEY,
    password text NOT NULL UNIQUE
);

CREATE TABLE postgres_databases(
    id bigserial PRIMARY KEY,
    project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    host text NOT NULL,
    port integer NOT NULL,
    database_name text NOT NULL,
    username text NOT NULL,
    password text NOT NULL,
    version text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE scan_groups(
    id bigserial PRIMARY KEY,
    project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    created_by bigint REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE workers(
    id bigserial PRIMARY KEY,
    token text NOT NULL UNIQUE,
    organization bigint NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE scans(
    id bigserial PRIMARY KEY,
    scan_group_id bigint NOT NULL REFERENCES scan_groups(id) ON DELETE CASCADE,
    scan_type integer NOT NULL,
    status integer NOT NULL,
    error text,
    worker_id bigint REFERENCES workers(id) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ended_at timestamp with time zone
);

CREATE TABLE git_scans(
    id bigserial PRIMARY KEY,
    scan_id bigint NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    repository_id bigint NOT NULL REFERENCES git_repositories(id) ON DELETE CASCADE
);

CREATE TABLE docker_scans(
    id bigserial PRIMARY KEY,
    scan_id bigint NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    image_id bigint NOT NULL REFERENCES docker_images(id) ON DELETE CASCADE
);

CREATE TABLE postgres_scans(
    id bigserial PRIMARY KEY,
    scan_id bigint NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    database_id bigint NOT NULL REFERENCES postgres_databases(id) ON DELETE CASCADE
);

CREATE TABLE scan_results(
    id bigserial PRIMARY KEY,
    scan_id bigint NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    severity integer NOT NULL,
    message text NOT NULL,
    scan_source integer NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX scan_results_scan_id_idx ON scan_results(scan_id);

CREATE TABLE scan_bruteforce_results(
    id bigserial PRIMARY KEY,
    scan_id bigint NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    scan_type integer NOT NULL,
    username text NOT NULL,
    password text,
    total integer NOT NULL,
    tried integer NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE bruteforced_passwords(
    id bigserial PRIMARY KEY,
    hash text NOT NULL,
    username text NOT NULL,
    password text,
    last_bruteforce_id bigint,
    project_id bigint REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE (hash, username, project_id)
);

CREATE TABLE worker_projects(
    id bigserial PRIMARY KEY,
    project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    worker_id bigint NOT NULL REFERENCES workers(id) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE remember_me_tokens(
    id bigserial PRIMARY KEY,
    user_id bigint REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    token text NOT NULL UNIQUE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE webauthn_credentials(
    id bigserial PRIMARY KEY,
    user_id bigint REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    credential_id bytea NOT NULL UNIQUE,
    public_key bytea NOT NULL,
    attestation_type text NOT NULL,
    transport text[] NOT NULL,
    user_present boolean NOT NULL,
    user_verified boolean NOT NULL,
    backup_eligible boolean NOT NULL,
    backup_state boolean NOT NULL,
    aa_guid bytea NOT NULL,
    sign_count integer NOT NULL,
    clone_warning boolean NOT NULL,
    attachment text NOT NULL
);

CREATE TABLE mysql_databases(
    id bigserial PRIMARY KEY,
    project_id bigint NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    host text NOT NULL,
    port integer NOT NULL,
    database_name text NOT NULL,
    username text NOT NULL,
    password text NOT NULL,
    version text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE mysql_scans(
    id bigserial PRIMARY KEY,
    scan_id bigint NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    database_id bigint NOT NULL REFERENCES mysql_databases(id) ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION encrypt_data(project_id bigint, salt_key text, data text)
    RETURNS text
    AS $$
    SELECT
        encode(pgp_sym_encrypt(data,(
                    SELECT
                        CONCAT(organizations.encryption_key, salt_key)
                    FROM organizations
                    INNER JOIN projects ON projects.organization_id = organizations.id
                    WHERE
                        projects.id = project_id)), 'hex')
$$
LANGUAGE sql;

CREATE OR REPLACE FUNCTION decrypt_data(project_id bigint, salt_key text, data text)
    RETURNS text
    AS $$
    SELECT
        pgp_sym_decrypt(decode(data, 'hex'),(
                SELECT
                    CONCAT(organizations.encryption_key, salt_key)
                FROM organizations
                INNER JOIN projects ON projects.organization_id = organizations.id
                WHERE
                    projects.id = project_id))
$$
LANGUAGE sql;

