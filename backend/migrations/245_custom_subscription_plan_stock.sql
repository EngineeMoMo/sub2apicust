-- [CUSTOM] 套餐售卖份数。-1 不限量；stock_used 包含待付预占及历史已付。
-- 原子条件更新串行化同套餐购买；与订单写入同事务，回滚不扣库存。
-- 迁移执行器包裹事务；先锁订单写入，避免回填与安装触发器之间遗漏在途下单。
LOCK TABLE payment_orders IN SHARE ROW EXCLUSIVE MODE;
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS stock_limit INTEGER NOT NULL DEFAULT -1
    CHECK (stock_limit >= -1);

CREATE OR REPLACE FUNCTION custom_plan_stock_occupies(order_status TEXT, payment_time TIMESTAMPTZ)
RETURNS BOOLEAN LANGUAGE SQL IMMUTABLE AS $$
    SELECT payment_time IS NOT NULL OR order_status IN
        ('PENDING', 'PAID', 'RECHARGING', 'COMPLETED', 'REFUNDING', 'REFUNDED', 'REFUND_FAILED', 'REFUND_REQUESTED');
$$;

-- 仅首次安装回填，重复迁移不可重置已售计数（已付订单删除也不补货）。
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
        WHERE table_schema = current_schema() AND table_name = 'subscription_plans' AND column_name = 'stock_used') THEN
        ALTER TABLE subscription_plans ADD COLUMN stock_used INTEGER NOT NULL DEFAULT 0 CHECK (stock_used >= 0);
        UPDATE subscription_plans p SET stock_used = (
            SELECT count(*) FROM payment_orders o WHERE o.plan_id = p.id AND o.order_type = 'subscription'
            AND custom_plan_stock_occupies(o.status, o.paid_at)
        );
    END IF;
END $$;

CREATE OR REPLACE FUNCTION custom_subscription_plan_stock_guard()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
    was_held BOOLEAN := FALSE;
    now_held BOOLEAN := FALSE;
BEGIN
    IF TG_OP <> 'INSERT' THEN
        was_held := OLD.order_type = 'subscription' AND OLD.plan_id IS NOT NULL
            AND custom_plan_stock_occupies(OLD.status, OLD.paid_at);
    END IF;
    IF TG_OP = 'DELETE' THEN
        -- 已付款／已履约删除不自动补货，仅释放未付预占。
        IF was_held AND OLD.paid_at IS NULL AND OLD.status = 'PENDING' THEN
            UPDATE subscription_plans SET stock_used = stock_used - 1 WHERE id = OLD.plan_id;
        END IF;
        RETURN OLD;
    END IF;
    IF TG_OP = 'UPDATE' AND OLD.order_type = 'subscription' AND OLD.plan_id IS NOT NULL THEN
        IF NEW.plan_id IS DISTINCT FROM OLD.plan_id OR NEW.order_type IS DISTINCT FROM OLD.order_type THEN
            RAISE EXCEPTION 'CUSTOM_PLAN_STOCK_ORDER_IMMUTABLE';
        END IF;
        IF OLD.paid_at IS NOT NULL AND NEW.paid_at IS NULL THEN
            RAISE EXCEPTION 'CUSTOM_PLAN_STOCK_PAYMENT_IMMUTABLE';
        END IF;
    END IF;
    now_held := NEW.order_type = 'subscription' AND NEW.plan_id IS NOT NULL
        AND custom_plan_stock_occupies(NEW.status, NEW.paid_at);
    IF now_held AND NOT was_held THEN
        UPDATE subscription_plans SET stock_used = stock_used + 1
        WHERE id = NEW.plan_id AND (stock_limit = -1 OR stock_used < stock_limit)
            AND (TG_OP <> 'INSERT' OR for_sale);
        IF NOT FOUND THEN
            RAISE EXCEPTION 'CUSTOM_PLAN_OUT_OF_STOCK' USING ERRCODE = 'P0001';
        END IF;
    ELSIF was_held AND NOT now_held THEN
        UPDATE subscription_plans SET stock_used = stock_used - 1 WHERE id = OLD.plan_id;
    END IF;
    RETURN NEW;
END $$;

DROP TRIGGER IF EXISTS custom_subscription_plan_stock_guard ON payment_orders;
CREATE TRIGGER custom_subscription_plan_stock_guard
BEFORE INSERT OR UPDATE OR DELETE ON payment_orders
FOR EACH ROW EXECUTE FUNCTION custom_subscription_plan_stock_guard();
