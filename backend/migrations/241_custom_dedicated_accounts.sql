-- [CUSTOM] 包号绑定独立于上游账号结构；撤销后保留隔离记录，避免旧密钥回到共享池。
CREATE TABLE IF NOT EXISTS custom_dedicated_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    account_id BIGINT NOT NULL UNIQUE REFERENCES accounts(id),
    group_id BIGINT NOT NULL UNIQUE REFERENCES groups(id),
    label VARCHAR(80) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_custom_dedicated_accounts_user ON custom_dedicated_accounts(user_id, id);
