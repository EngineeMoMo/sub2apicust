-- [CUSTOM] 限量套餐每个用户同时最多一笔待付订单，避免单账号用多笔待付把库存长时间占满。
-- 245 已发布不可改；此处整体替换同名触发函数，其余库存语义不变。不限量（-1）套餐不受影响。
LOCK TABLE payment_orders IN SHARE ROW EXCLUSIVE MODE;

CREATE OR REPLACE FUNCTION custom_subscription_plan_stock_guard()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
    was_held BOOLEAN := FALSE;
    now_held BOOLEAN := FALSE;
    limited BOOLEAN := FALSE;
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
            AND (TG_OP <> 'INSERT' OR for_sale)
        RETURNING stock_limit >= 0 INTO limited;
        IF NOT FOUND THEN
            RAISE EXCEPTION 'CUSTOM_PLAN_OUT_OF_STOCK' USING ERRCODE = 'P0001';
        END IF;
        -- 上面的条件更新已持有套餐行锁，同套餐的并发下单在此串行；读取到的是已提交的前一笔待付订单。
        IF TG_OP = 'INSERT' AND limited AND NEW.user_id IS NOT NULL AND NEW.status = 'PENDING' AND NEW.paid_at IS NULL
            AND EXISTS (SELECT 1 FROM payment_orders o WHERE o.user_id = NEW.user_id AND o.plan_id = NEW.plan_id
                AND o.order_type = 'subscription' AND o.status = 'PENDING' AND o.paid_at IS NULL) THEN
            RAISE EXCEPTION 'CUSTOM_PLAN_PENDING_EXISTS' USING ERRCODE = 'P0001';
        END IF;
    ELSIF was_held AND NOT now_held THEN
        UPDATE subscription_plans SET stock_used = stock_used - 1 WHERE id = OLD.plan_id;
    END IF;
    RETURN NEW;
END $$;
