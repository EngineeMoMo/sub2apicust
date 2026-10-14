-- [CUSTOM] 付费工具实际调用量与价格快照；旧账单保持 NULL，不补扣历史费用。
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS tool_usage JSONB;
