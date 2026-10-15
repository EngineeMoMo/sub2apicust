CREATE FUNCTION custom_group_active_binding_guard() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.group_id IS NOT NULL THEN
        PERFORM id FROM groups WHERE id=NEW.group_id AND deleted_at IS NULL FOR SHARE;
        IF NOT FOUND THEN
            RAISE EXCEPTION 'cannot bind to deleted or missing group %', NEW.group_id USING ERRCODE='23503';
        END IF;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER custom_api_key_active_group
    BEFORE INSERT OR UPDATE OF group_id ON api_keys
    FOR EACH ROW EXECUTE FUNCTION custom_group_active_binding_guard();

CREATE TRIGGER custom_user_allowed_active_group
    BEFORE INSERT OR UPDATE OF group_id ON user_allowed_groups
    FOR EACH ROW EXECUTE FUNCTION custom_group_active_binding_guard();
