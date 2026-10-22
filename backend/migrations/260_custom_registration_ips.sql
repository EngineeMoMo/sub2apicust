-- [CUSTOM] 成功建号与来源占额同事务；删号保留 IP，只有未完成 OAuth 建号的补偿回滚释放。
CREATE TABLE IF NOT EXISTS custom_registration_ips (
    ip TEXT PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_custom_registration_ips_user_id ON custom_registration_ips(user_id);
