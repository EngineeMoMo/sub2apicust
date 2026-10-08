-- [CUSTOM] 非购买分配与商品共用库存，记录归属以支持后续管理员续期。
CREATE TABLE IF NOT EXISTS custom_subscription_stock_allocations (
    id BIGSERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL,
    plan_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    group_id BIGINT NOT NULL,
    source TEXT NOT NULL,
    source_key TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS custom_subscription_stock_allocations_sub_idx
    ON custom_subscription_stock_allocations(subscription_id, id DESC);

-- 历史非购买订阅仅在分组唯一对应套餐、没有已付订单时才能确定归属。
-- 多套餐或混合来源不能猜测，保留待人工核对；重复迁移不得再扣一次。
LOCK TABLE user_subscriptions IN SHARE ROW EXCLUSIVE MODE;
LOCK TABLE payment_orders IN SHARE ROW EXCLUSIVE MODE;
WITH inserted AS (
    INSERT INTO custom_subscription_stock_allocations(subscription_id,plan_id,user_id,group_id,source,source_key)
    SELECT s.id,p.id,s.user_id,s.group_id,'legacy_manual','legacy-sub:' || s.id
    FROM user_subscriptions s
    JOIN subscription_plans p ON p.group_id=s.group_id
    WHERE (SELECT count(*) FROM subscription_plans candidates WHERE candidates.group_id=s.group_id)=1
      AND NOT EXISTS (SELECT 1 FROM payment_orders o
        WHERE o.user_id=s.user_id AND o.subscription_group_id=s.group_id AND o.order_type='subscription'
          AND (o.paid_at IS NOT NULL OR o.status IN ('PAID','RECHARGING','COMPLETED','REFUNDED','REFUNDING')))
      AND NOT EXISTS (SELECT 1 FROM custom_subscription_stock_allocations a WHERE a.subscription_id=s.id)
    ON CONFLICT (source_key) DO NOTHING
    RETURNING plan_id
), counts AS (SELECT plan_id,count(*) AS used FROM inserted GROUP BY plan_id)
UPDATE subscription_plans p SET stock_used=p.stock_used+counts.used FROM counts WHERE p.id=counts.plan_id;
