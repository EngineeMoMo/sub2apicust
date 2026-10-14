CREATE TABLE IF NOT EXISTS custom_auth_ip_blacklist_config (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    threshold INTEGER NOT NULL DEFAULT 10 CHECK (threshold BETWEEN 1 AND 1000),
    window_seconds INTEGER NOT NULL DEFAULT 7200 CHECK (window_seconds BETWEEN 60 AND 86400 AND window_seconds % 60 = 0),
    version BIGINT NOT NULL DEFAULT 1 CHECK (version > 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO custom_auth_ip_blacklist_config (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS custom_auth_ip_blacklist (
    id BIGSERIAL PRIMARY KEY,
    ip INET NOT NULL UNIQUE CHECK (masklen(ip) = CASE family(ip) WHEN 4 THEN 32 ELSE 128 END),
    blocked BOOLEAN NOT NULL DEFAULT TRUE,
    generation BIGINT NOT NULL DEFAULT 1 CHECK (generation > 0),
    blocked_at TIMESTAMPTZ NOT NULL,
    trigger_path TEXT NOT NULL,
    trigger_count INTEGER NOT NULL CHECK (trigger_count > 0),
    window_seconds INTEGER NOT NULL CHECK (window_seconds > 0),
    first_event_at TIMESTAMPTZ NOT NULL,
    last_event_at TIMESTAMPTZ NOT NULL,
    rule_version BIGINT NOT NULL,
    unblocked_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS custom_auth_ip_blacklist_active_idx
    ON custom_auth_ip_blacklist (blocked_at DESC, id DESC) WHERE blocked;
