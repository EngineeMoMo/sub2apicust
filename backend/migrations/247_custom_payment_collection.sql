-- [CUSTOM] 收款账本独立于订单保留；退款、删除不恢复已收款额度。
LOCK TABLE payment_orders IN SHARE ROW EXCLUSIVE MODE;
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS sales_mode VARCHAR(20) NOT NULL DEFAULT 'online'
 CHECK(sales_mode IN ('online','contact_admin'));
INSERT INTO settings(key,value) VALUES('CUSTOM_PAYMENT_COLLECTION','{"enabled":false,"single_max":50,"daily_max":1000,"timezone":"Asia/Shanghai","quick_amounts":[10,20,30,40],"allow_custom_amount":true,"contact_text":""}') ON CONFLICT(key) DO NOTHING;
CREATE TABLE IF NOT EXISTS custom_payment_collection_ledger (
 id BIGSERIAL PRIMARY KEY,
 order_id BIGINT UNIQUE,
 reference VARCHAR(100) UNIQUE,
 amount NUMERIC(20,2) NOT NULL CHECK(amount>0),
 paid_at TIMESTAMPTZ,
 released BOOLEAN NOT NULL DEFAULT false
);
CREATE INDEX IF NOT EXISTS custom_collection_paid ON custom_payment_collection_ledger(paid_at);
CREATE INDEX IF NOT EXISTS custom_collection_held ON custom_payment_collection_ledger(order_id) WHERE paid_at IS NULL AND NOT released;
-- 旧无币种快照按原支付宝／微信／易支付人民币语义回填；未确认关单的旧单保守占用。
INSERT INTO custom_payment_collection_ledger(order_id,amount,paid_at)
 SELECT id,pay_amount,paid_at FROM payment_orders
 WHERE pay_amount>0 AND COALESCE(provider_snapshot->>'currency','CNY')='CNY'
 AND (paid_at IS NOT NULL OR status='PENDING')
 ON CONFLICT(order_id) DO NOTHING;

CREATE OR REPLACE FUNCTION custom_payment_collection_guard() RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
 policy JSONB;
 used NUMERIC;
 zone TEXT;
BEGIN
 IF TG_OP='INSERT' THEN
  IF NEW.order_type='subscription' AND EXISTS(SELECT 1 FROM subscription_plans WHERE id=NEW.plan_id AND sales_mode='contact_admin') THEN
   RAISE EXCEPTION 'CUSTOM_PLAN_CONTACT_ADMIN';
  END IF;
  -- 锁由所有新订单与配置更新共用；之后的聚合读取已提交的前一笔预占。
  SELECT value::jsonb INTO policy FROM settings WHERE key='CUSTOM_PAYMENT_COLLECTION' FOR UPDATE;
  IF COALESCE((policy->>'enabled')::boolean,false) THEN
   IF COALESCE(NEW.provider_snapshot->>'currency','CNY')<>'CNY' THEN RAISE EXCEPTION 'CUSTOM_COLLECTION_CURRENCY'; END IF;
   IF NEW.pay_amount>(policy->>'single_max')::numeric THEN RAISE EXCEPTION 'CUSTOM_COLLECTION_SINGLE_LIMIT'; END IF;
   zone:=policy->>'timezone';
   SELECT COALESCE(sum(amount),0) INTO used FROM custom_payment_collection_ledger
    WHERE (paid_at IS NULL AND NOT released) OR
      (paid_at >= (date_trunc('day',now() AT TIME ZONE zone) AT TIME ZONE zone)
       AND paid_at < ((date_trunc('day',now() AT TIME ZONE zone)+interval '1 day') AT TIME ZONE zone));
   IF used+NEW.pay_amount>(policy->>'daily_max')::numeric THEN RAISE EXCEPTION 'CUSTOM_COLLECTION_DAILY_LIMIT'; END IF;
  END IF;
  IF NEW.pay_amount>0 AND COALESCE(NEW.provider_snapshot->>'currency','CNY')='CNY' THEN
   INSERT INTO custom_payment_collection_ledger(order_id,amount,paid_at) VALUES(NEW.id,NEW.pay_amount,NEW.paid_at);
  END IF;
 ELSIF NEW.paid_at IS NOT NULL AND OLD.paid_at IS NULL THEN
  -- 实际付款不能因收款限额拒绝记账；跨日回调按本站确认日计入，确认前一直预占。
  UPDATE custom_payment_collection_ledger SET paid_at=NEW.paid_at,released=false WHERE order_id=NEW.id;
 END IF;
 RETURN NEW;
END $$;
DROP TRIGGER IF EXISTS custom_payment_collection_guard ON payment_orders;
CREATE TRIGGER custom_payment_collection_guard AFTER INSERT OR UPDATE ON payment_orders
 FOR EACH ROW EXECUTE FUNCTION custom_payment_collection_guard();
