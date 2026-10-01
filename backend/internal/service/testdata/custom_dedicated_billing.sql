-- [CUSTOM] 仅隔离回归库使用的最小业务结构，绝不应用到业务数据库。
CREATE TABLE users(id BIGINT PRIMARY KEY,username TEXT,email TEXT,status TEXT DEFAULT 'active',deleted_at TIMESTAMPTZ,balance NUMERIC(20,8) DEFAULT 0,updated_at TIMESTAMPTZ DEFAULT NOW());
CREATE TABLE accounts(id BIGINT PRIMARY KEY,platform TEXT DEFAULT 'openai',type TEXT DEFAULT 'oauth',parent_account_id BIGINT,deleted_at TIMESTAMPTZ);
CREATE TABLE groups(id BIGINT PRIMARY KEY,status TEXT DEFAULT 'active',is_exclusive BOOLEAN DEFAULT TRUE,platform TEXT DEFAULT 'openai',subscription_type TEXT DEFAULT 'standard',fallback_group_id BIGINT,fallback_group_id_on_invalid_request BIGINT,deleted_at TIMESTAMPTZ);
CREATE TABLE account_groups(account_id BIGINT,group_id BIGINT);
CREATE TABLE user_allowed_groups(user_id BIGINT,group_id BIGINT);
CREATE TABLE api_keys(id BIGINT PRIMARY KEY,user_id BIGINT,group_id BIGINT,status TEXT DEFAULT 'active',quota NUMERIC DEFAULT 100,quota_used NUMERIC DEFAULT 0,usage_5h NUMERIC DEFAULT 0,usage_1d NUMERIC DEFAULT 0,usage_7d NUMERIC DEFAULT 0,window_5h_start TIMESTAMPTZ,window_1d_start TIMESTAMPTZ,window_7d_start TIMESTAMPTZ,updated_at TIMESTAMPTZ DEFAULT NOW(),deleted_at TIMESTAMPTZ);
CREATE TABLE user_subscriptions(group_id BIGINT,deleted_at TIMESTAMPTZ);
CREATE TABLE usage_billing_dedup(id BIGSERIAL PRIMARY KEY,request_id TEXT,api_key_id BIGINT,request_fingerprint TEXT,UNIQUE(request_id,api_key_id));
CREATE TABLE usage_billing_dedup_archive(request_id TEXT,api_key_id BIGINT,request_fingerprint TEXT,UNIQUE(request_id,api_key_id));
INSERT INTO users(id) VALUES(11),(12),(13);
INSERT INTO accounts(id) VALUES(22),(44);
INSERT INTO groups(id) VALUES(33),(55),(77);
INSERT INTO account_groups VALUES(22,33),(22,55),(44,77);
INSERT INTO user_allowed_groups VALUES(11,33),(12,33),(11,55),(12,55),(13,77);
INSERT INTO api_keys(id,user_id,group_id) VALUES(7,11,33),(8,12,55),(9,13,77);
