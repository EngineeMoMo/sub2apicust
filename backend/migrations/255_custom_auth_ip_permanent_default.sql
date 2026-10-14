ALTER TABLE custom_auth_ip_blacklist_config
    ALTER COLUMN block_mode SET DEFAULT 'permanent';

UPDATE custom_auth_ip_blacklist_config
SET block_mode = 'permanent', version = version + 1, updated_at = NOW()
WHERE id = 1;
