-- Extra profile fields (avatar + age), requested alongside the settings tab.
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS avatar_url TEXT NOT NULL DEFAULT '';
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS age INT;

-- One user can favorite either a restaurant (franchise-service) or a dish
-- (menu-service) — no FK, both services own their own data.
CREATE TABLE IF NOT EXISTS favorites (
    id         UUID PRIMARY KEY,
    user_id    UUID        NOT NULL,
    kind       TEXT        NOT NULL CHECK (kind IN ('restaurant', 'dish')),
    target_id  UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, kind, target_id)
);

CREATE INDEX IF NOT EXISTS idx_favorites_user ON favorites (user_id);

CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id      UUID PRIMARY KEY,
    email_orders BOOLEAN     NOT NULL DEFAULT TRUE,
    email_promos BOOLEAN     NOT NULL DEFAULT TRUE,
    sms_orders   BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
