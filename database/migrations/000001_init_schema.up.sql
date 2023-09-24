CREATE TABLE IF NOT EXISTS users
(
    id         SERIAL PRIMARY KEY,
    email      VARCHAR(125) UNIQUE,
    name      VARCHAR(125),
    username   VARCHAR(30) UNIQUE,
    password   VARCHAR(2048) NOT NULL,
    is_active  BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tokens
(
    id            VARCHAR(36) PRIMARY KEY,
    user_id       INT,
    refresh_token VARCHAR(2048),
    platform_id   INT       NOT NULL,
    is_blocked    BOOLEAN   NOT NULL DEFAULT FALSE,
    expires_at    TIMESTAMP NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP
);

CREATE TABLE IF NOT EXISTS otps
(
    id         SERIAL PRIMARY KEY,
    type       VARCHAR(255),
    receiver    VARCHAR(255),
    code       VARCHAR(2048),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);