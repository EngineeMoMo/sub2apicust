ALTER TABLE custom_auth_ip_blacklist
    ADD COLUMN source TEXT NOT NULL DEFAULT 'auto' CHECK (source IN ('auto', 'manual')),
    ADD COLUMN reason TEXT NOT NULL DEFAULT '' CHECK (char_length(reason) <= 500),
    ALTER COLUMN first_event_at DROP NOT NULL,
    ALTER COLUMN last_event_at DROP NOT NULL,
    DROP CONSTRAINT custom_auth_ip_blacklist_trigger_count_check,
    DROP CONSTRAINT custom_auth_ip_blacklist_window_seconds_check;

ALTER TABLE custom_auth_ip_blacklist ADD CONSTRAINT custom_auth_ip_blacklist_source_stats_check CHECK (
    (source = 'auto' AND trigger_count > 0 AND window_seconds > 0
        AND first_event_at IS NOT NULL AND last_event_at IS NOT NULL)
    OR (source = 'manual' AND trigger_count = 0 AND window_seconds = 0
        AND first_event_at IS NULL AND last_event_at IS NULL AND rule_version = 0
        AND trigger_path = '' AND char_length(btrim(reason)) > 0)
);
