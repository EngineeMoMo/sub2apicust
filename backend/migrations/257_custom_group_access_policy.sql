CREATE TABLE IF NOT EXISTS custom_group_access_policies (
    group_id BIGINT PRIMARY KEY REFERENCES groups(id) ON DELETE CASCADE,
    monthly_recharge_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    monthly_recharge_min NUMERIC(20,2) NOT NULL DEFAULT 0,
    balance_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    balance_min NUMERIC(20,2) NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ,
    fallback_group_id BIGINT REFERENCES groups(id) ON DELETE RESTRICT,
    rate_schedule_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    rate_schedule JSONB NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(rate_schedule) = 'array'),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    CHECK (monthly_recharge_min >= 0 AND (NOT monthly_recharge_enabled OR monthly_recharge_min > 0)),
    CHECK (balance_min >= 0 AND (NOT balance_enabled OR balance_min > 0)),
    CHECK (fallback_group_id IS NULL OR (expires_at IS NOT NULL AND fallback_group_id <> group_id))
);

CREATE INDEX IF NOT EXISTS idx_custom_group_access_fallback ON custom_group_access_policies(fallback_group_id)
    WHERE fallback_group_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_custom_group_access_recharge ON payment_orders(user_id, paid_at)
    WHERE order_type = 'balance' AND status IN ('COMPLETED', 'PARTIALLY_REFUNDED');

CREATE FUNCTION custom_group_access_notify() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    previous JSONB := CASE WHEN TG_OP IN ('UPDATE','DELETE') THEN to_jsonb(OLD) ELSE '{}'::jsonb END;
    current_row JSONB := CASE WHEN TG_OP IN ('INSERT','UPDATE') THEN to_jsonb(NEW) ELSE '{}'::jsonb END;
    event_kind TEXT;
    current_id BIGINT;
    previous_id BIGINT;
BEGIN
    IF TG_OP = 'TRUNCATE' THEN
        PERFORM pg_notify('custom_group_access_changed', '{"kind":"reset"}');
        RETURN NULL;
    END IF;
    IF TG_TABLE_NAME = 'custom_group_access_policies' THEN
        PERFORM pg_notify('custom_group_access_changed', '{"kind":"policies"}');
        RETURN NULL;
    ELSIF TG_TABLE_NAME = 'payment_orders' THEN
        IF current_row->>'order_type' IS DISTINCT FROM 'balance' AND previous->>'order_type' IS DISTINCT FROM 'balance' THEN RETURN NULL; END IF;
        IF TG_OP = 'UPDATE' AND
           jsonb_build_array(current_row->'user_id',current_row->'order_type',current_row->'amount',current_row->'bonus_amount',current_row->'refund_amount',current_row->'status',current_row->'paid_at') =
           jsonb_build_array(previous->'user_id',previous->'order_type',previous->'amount',previous->'bonus_amount',previous->'refund_amount',previous->'status',previous->'paid_at') THEN RETURN NULL; END IF;
        event_kind := 'monthly';
        current_id := (current_row->>'user_id')::bigint;
        previous_id := (previous->>'user_id')::bigint;
    ELSIF TG_TABLE_NAME = 'groups' THEN
        event_kind := 'group';
        current_id := (current_row->>'id')::bigint;
        previous_id := (previous->>'id')::bigint;
    ELSIF TG_TABLE_NAME = 'users' THEN
        current_id := (COALESCE(current_row->>'id',previous->>'id'))::bigint;
        IF TG_OP = 'UPDATE' AND
           current_row - ARRAY['balance','frozen_balance','updated_at','last_used_at','last_active_at','total_recharged'] =
           previous - ARRAY['balance','frozen_balance','updated_at','last_used_at','last_active_at','total_recharged'] THEN
            IF current_row->'balance' IS DISTINCT FROM previous->'balance' THEN
                PERFORM pg_notify('custom_group_access_changed', jsonb_build_object('kind','balance','id',current_id,'balance',current_row->'balance')::text);
            END IF;
            RETURN NULL;
        END IF;
        event_kind := 'user';
    ELSE
        IF TG_TABLE_NAME = 'user_subscriptions' AND TG_OP = 'UPDATE' AND
           jsonb_build_array(current_row->'user_id',current_row->'group_id',current_row->'status',current_row->'starts_at',current_row->'expires_at',current_row->'deleted_at') =
           jsonb_build_array(previous->'user_id',previous->'group_id',previous->'status',previous->'starts_at',previous->'expires_at',previous->'deleted_at') THEN RETURN NULL; END IF;
        event_kind := CASE WHEN TG_TABLE_NAME = 'user_subscriptions' THEN 'subscription' ELSE 'user' END;
        current_id := (current_row->>'user_id')::bigint;
        previous_id := (previous->>'user_id')::bigint;
    END IF;
    IF current_id IS NOT NULL THEN
        PERFORM pg_notify('custom_group_access_changed', jsonb_build_object('kind',event_kind,'id',current_id)::text);
    END IF;
    IF previous_id IS NOT NULL AND previous_id IS DISTINCT FROM current_id THEN
        PERFORM pg_notify('custom_group_access_changed', jsonb_build_object('kind',event_kind,'id',previous_id)::text);
    END IF;
    RETURN NULL;
END;
$$;

DO $$
DECLARE table_name TEXT;
BEGIN
    FOREACH table_name IN ARRAY ARRAY['custom_group_access_policies','payment_orders','users','groups','user_allowed_groups','user_group_rate_multipliers','user_subscriptions'] LOOP
        IF to_regclass(table_name) IS NOT NULL THEN
            EXECUTE format('CREATE TRIGGER custom_group_access_change AFTER INSERT OR UPDATE OR DELETE ON %I FOR EACH ROW EXECUTE FUNCTION custom_group_access_notify()', table_name);
            EXECUTE format('CREATE TRIGGER custom_group_access_truncate AFTER TRUNCATE ON %I FOR EACH STATEMENT EXECUTE FUNCTION custom_group_access_notify()', table_name);
        END IF;
    END LOOP;
END;
$$;
