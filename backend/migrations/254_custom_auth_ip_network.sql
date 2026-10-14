ALTER TABLE custom_auth_ip_blacklist
    DROP CONSTRAINT custom_auth_ip_blacklist_ip_check,
    ADD CONSTRAINT custom_auth_ip_blacklist_target_check CHECK (
        ip = network(ip) AND masklen(ip) > 0
    );

CREATE INDEX custom_auth_ip_blacklist_network_idx
    ON custom_auth_ip_blacklist USING gist (ip inet_ops);

ALTER TABLE custom_auth_ip_blacklist_config
    ADD COLUMN aggregation_mode TEXT NOT NULL DEFAULT 'network' CHECK (aggregation_mode IN ('ip', 'network')),
    ADD COLUMN network_ipv4_prefix INTEGER NOT NULL DEFAULT 24 CHECK (network_ipv4_prefix BETWEEN 1 AND 32),
    ADD COLUMN network_ipv6_prefix INTEGER NOT NULL DEFAULT 64 CHECK (network_ipv6_prefix BETWEEN 1 AND 128);

UPDATE custom_auth_ip_blacklist_config SET version = version + 1, updated_at = NOW();
