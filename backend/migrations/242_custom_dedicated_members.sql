-- [CUSTOM] 多用户包号与逻辑删除；删除记录仍保留为旧密钥隔离标记。
ALTER TABLE custom_dedicated_accounts ADD COLUMN IF NOT EXISTS user_ids JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE custom_dedicated_accounts ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE custom_dedicated_accounts DROP CONSTRAINT IF EXISTS custom_dedicated_accounts_account_id_key;
ALTER TABLE custom_dedicated_accounts DROP CONSTRAINT IF EXISTS custom_dedicated_accounts_group_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_custom_dedicated_live_account ON custom_dedicated_accounts(account_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_custom_dedicated_live_group ON custom_dedicated_accounts(group_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_custom_dedicated_members ON custom_dedicated_accounts USING GIN(user_ids);
