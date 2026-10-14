ALTER TABLE custom_auth_ip_blacklist_config
    ADD COLUMN IF NOT EXISTS block_seconds INTEGER NOT NULL DEFAULT 900
    CHECK (block_seconds BETWEEN 60 AND 86400 AND block_seconds % 60 = 0);

ALTER TABLE custom_auth_ip_blacklist
    ADD COLUMN IF NOT EXISTS blocked_until TIMESTAMPTZ;

UPDATE custom_auth_ip_blacklist_config SET version = version + 1, updated_at = NOW() WHERE id = 1;
