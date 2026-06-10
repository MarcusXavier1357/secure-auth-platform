CREATE TABLE roles (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    role_id BIGINT REFERENCES roles(id),

    name VARCHAR(255) NOT NULL,

    email VARCHAR(255) NOT NULL UNIQUE,

    password_hash TEXT NOT NULL,

    active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE permissions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    code VARCHAR(255) NOT NULL UNIQUE,

    description TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE user_permissions (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,

    PRIMARY KEY (
        user_id,
        permission_id
    )
);

CREATE TABLE sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    user_id BIGINT NOT NULL REFERENCES users(id),

    refresh_token_hash TEXT NOT NULL,

    expires_at TIMESTAMP NOT NULL,

    revoked BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE audit_logs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    user_id BIGINT REFERENCES users(id),

    action VARCHAR(100) NOT NULL,

    entity VARCHAR(100) NOT NULL,

    entity_id BIGINT,

    old_data JSONB,

    new_data JSONB,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role_id ON users(role_id);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at) WHERE revoked = FALSE;
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity, entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
