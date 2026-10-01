-- [CUSTOM] 包号预付费：共享准入计数、可恢复并发租约及独立参考用量账本。
CREATE TABLE custom_dedicated_billing_policies (
    binding_id BIGINT PRIMARY KEY REFERENCES custom_dedicated_accounts(id),
    concurrency_limit INTEGER NOT NULL DEFAULT 2 CHECK (concurrency_limit BETWEEN 1 AND 200),
    rpm_limit INTEGER NOT NULL DEFAULT 30 CHECK (rpm_limit BETWEEN 1 AND 10000),
    daily_request_limit INTEGER NOT NULL DEFAULT 0 CHECK (daily_request_limit BETWEEN 0 AND 1000000),
    max_body_bytes BIGINT NOT NULL DEFAULT 2097152 CHECK (max_body_bytes BETWEEN 1024 AND 33554432),
    allow_images BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE custom_dedicated_request_counters (
    binding_id BIGINT PRIMARY KEY REFERENCES custom_dedicated_accounts(id),
    minute_start TIMESTAMPTZ NOT NULL,
    minute_requests INTEGER NOT NULL DEFAULT 0,
    day_start TIMESTAMPTZ NOT NULL,
    day_requests INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE custom_dedicated_request_leases (
    id VARCHAR(64) PRIMARY KEY,
    binding_id BIGINT NOT NULL REFERENCES custom_dedicated_accounts(id),
    account_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    api_key_id BIGINT NOT NULL,
    group_id BIGINT NOT NULL,
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ
);
CREATE INDEX idx_custom_dedicated_active_leases ON custom_dedicated_request_leases(binding_id, expires_at) WHERE finished_at IS NULL;
CREATE TABLE custom_dedicated_billing_usage (
    request_id TEXT NOT NULL,
    api_key_id BIGINT NOT NULL,
    lease_id VARCHAR(64) NOT NULL REFERENCES custom_dedicated_request_leases(id),
    binding_id BIGINT NOT NULL REFERENCES custom_dedicated_accounts(id),
    user_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    group_id BIGINT NOT NULL,
    model TEXT NOT NULL,
    reference_cost NUMERIC(20,8) NOT NULL CHECK (reference_cost >= 0),
    input_tokens BIGINT NOT NULL,
    output_tokens BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (request_id, api_key_id)
);
CREATE INDEX idx_custom_dedicated_billing_usage_binding ON custom_dedicated_billing_usage(binding_id, created_at);
