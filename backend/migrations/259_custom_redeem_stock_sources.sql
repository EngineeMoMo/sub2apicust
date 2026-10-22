-- [CUSTOM] 兑换码明确绑定库存来源；旧码无记录时保留既有归属／唯一套餐回退。
CREATE TABLE IF NOT EXISTS custom_redeem_stock_sources (
    redeem_code_id BIGINT PRIMARY KEY REFERENCES redeem_codes(id) ON DELETE CASCADE,
    plan_id BIGINT NOT NULL CHECK (plan_id > 0)
);
