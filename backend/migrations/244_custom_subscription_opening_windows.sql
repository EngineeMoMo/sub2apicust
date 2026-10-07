-- [CUSTOM] 周/月额度从开通时间计算。迁移旧的未启动或首次使用/手动重置偏移窗口。
-- 无法从聚合用量精确重建跨周期流水，因此迁移保留已用额度，不额外发放配额；
-- 偏移窗口先对齐到当前开通周期，下一次边界再按正常规则清零。
-- 已对齐窗口保持原值，避免阻止正常的过期窗口重置。订阅到期时间不变。
UPDATE user_subscriptions
SET weekly_window_start = CASE
      WHEN weekly_window_start IS NULL
        OR EXTRACT(EPOCH FROM (weekly_window_start - starts_at)) < 0
        OR MOD(EXTRACT(EPOCH FROM (weekly_window_start - starts_at)), 604800) <> 0
      THEN starts_at + GREATEST(0, FLOOR(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - starts_at)) / 604800)) * INTERVAL '168 hours'
      ELSE weekly_window_start END,
    monthly_window_start = CASE
      WHEN monthly_window_start IS NULL
        OR EXTRACT(EPOCH FROM (monthly_window_start - starts_at)) < 0
        OR MOD(EXTRACT(EPOCH FROM (monthly_window_start - starts_at)), 2592000) <> 0
      THEN starts_at + GREATEST(0, FLOOR(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - starts_at)) / 2592000)) * INTERVAL '720 hours'
      ELSE monthly_window_start END,
    daily_window_start = COALESCE(daily_window_start, starts_at)
WHERE deleted_at IS NULL AND expires_at > CURRENT_TIMESTAMP;
