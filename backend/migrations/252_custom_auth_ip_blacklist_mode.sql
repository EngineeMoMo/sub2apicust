ALTER TABLE custom_auth_ip_blacklist_config
    ADD COLUMN IF NOT EXISTS block_mode TEXT NOT NULL DEFAULT 'timed'
    CHECK (block_mode IN ('timed', 'permanent'));

UPDATE custom_auth_ip_blacklist_config SET version = version + 1, updated_at = NOW() WHERE id = 1;
