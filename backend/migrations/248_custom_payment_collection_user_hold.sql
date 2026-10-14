-- [CUSTOM] 收款预占按用户封顶，防止单个账号反复下单再取消占满全站额度。
-- 247 已发布不可改；此处整体替换同名触发函数，新增复查节流列。
LOCK TABLE payment_orders IN SHARE ROW EXCLUSIVE MODE;
ALTER TABLE custom_payment_collection_ledger ADD COLUMN IF NOT EXISTS release_checked_at TIMESTAMPTZ;

CREATE OR REPLACE FUNCTION custom_payment_collection_guard() RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
 policy JSONB;
 used NUMERIC;
 zone TEXT;
 user_cap NUMERIC;
 user_held NUMERIC;
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
   -- 同一用户未付且未释放的预占合计不超过 user_hold_max；旧配置缺字段时按单笔上限的3倍、且不超过每日上限。
   IF NEW.user_id IS NOT NULL AND NEW.pay_amount>0 AND COALESCE(NEW.provider_snapshot->>'currency','CNY')='CNY' THEN
    user_cap:=NULLIF(policy->>'user_hold_max','')::numeric;
    IF user_cap IS NULL OR user_cap<=0 THEN
     user_cap:=LEAST((policy->>'single_max')::numeric*3,(policy->>'daily_max')::numeric);
    END IF;
    SELECT COALESCE(sum(l.amount),0) INTO user_held FROM custom_payment_collection_ledger l
     JOIN payment_orders o ON o.id=l.order_id
     WHERE o.user_id=NEW.user_id AND l.paid_at IS NULL AND NOT l.released;
    IF user_held+NEW.pay_amount>user_cap THEN RAISE EXCEPTION 'CUSTOM_COLLECTION_USER_HOLD_LIMIT'; END IF;
   END IF;
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
