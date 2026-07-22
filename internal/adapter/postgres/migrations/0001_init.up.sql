-- Profiles are keyed by the auth-service user id (no FK: services own their data).
CREATE TABLE IF NOT EXISTS profiles (
    user_id    UUID PRIMARY KEY,
    first_name TEXT        NOT NULL DEFAULT '',
    last_name  TEXT        NOT NULL DEFAULT '',
    phone      TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS addresses (
    id         UUID PRIMARY KEY,
    user_id    UUID        NOT NULL,
    label      TEXT        NOT NULL DEFAULT '',
    street     TEXT        NOT NULL,
    zip_code   TEXT        NOT NULL DEFAULT '',
    city       TEXT        NOT NULL,
    is_default BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_addresses_user ON addresses (user_id);
