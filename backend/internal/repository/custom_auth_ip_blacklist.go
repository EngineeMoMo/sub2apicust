package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const authIPTargetColumn = `CASE WHEN masklen(entry.ip) = CASE family(entry.ip) WHEN 4 THEN 32 ELSE 128 END THEN host(entry.ip) ELSE entry.ip::text END`

const authIPEntryColumns = `entry.id, ` + authIPTargetColumn + `, entry.blocked, entry.generation,
entry.blocked_at, entry.trigger_path, entry.trigger_count, entry.window_seconds,
entry.first_event_at, entry.last_event_at, entry.rule_version, entry.blocked_until, entry.source, entry.reason`

type authIPBlacklistRepository struct {
	db *sql.DB
}

func NewAuthIPBlacklistRepository(db *sql.DB) service.AuthIPBlacklistRepository {
	return &authIPBlacklistRepository{db: db}
}

func (repo *authIPBlacklistRepository) ready() error {
	if repo.db == nil {
		return fmt.Errorf("认证黑名单数据库未配置")
	}
	return nil
}

type authIPRowScanner interface {
	Scan(...any) error
}

func scanAuthIPEntry(row authIPRowScanner, prefix ...any) (*service.AuthIPBlacklistEntry, error) {
	var id, generation, triggerCount, windowSeconds, ruleVersion sql.NullInt64
	var ip, triggerPath, source, reason sql.NullString
	var blocked sql.NullBool
	var blockedAt, firstEventAt, lastEventAt sql.NullTime
	var blockedUntil sql.NullTime
	targets := append(prefix, &id, &ip, &blocked, &generation, &blockedAt, &triggerPath,
		&triggerCount, &windowSeconds, &firstEventAt, &lastEventAt, &ruleVersion, &blockedUntil, &source, &reason)
	if err := row.Scan(targets...); err != nil {
		return nil, err
	}
	if !id.Valid {
		return nil, nil
	}
	entry := &service.AuthIPBlacklistEntry{
		ID: id.Int64, IP: ip.String, Blocked: blocked.Bool, Generation: generation.Int64,
		BlockedAt: blockedAt.Time.UTC(), TriggerPath: triggerPath.String, TriggerCount: int(triggerCount.Int64),
		WindowSeconds: int(windowSeconds.Int64), RuleVersion: ruleVersion.Int64,
		Source: source.String, Reason: reason.String,
	}
	if firstEventAt.Valid {
		instant := firstEventAt.Time.UTC()
		entry.FirstEventAt = &instant
	}
	if lastEventAt.Valid {
		instant := lastEventAt.Time.UTC()
		entry.LastEventAt = &instant
	}
	if blockedUntil.Valid {
		instant := blockedUntil.Time.UTC()
		entry.BlockedUntil = &instant
	}
	return entry, nil
}

func (repo *authIPBlacklistRepository) Snapshot(ctx context.Context, ip string) (*service.AuthIPBlacklistSnapshot, error) {
	if err := repo.ready(); err != nil {
		return nil, err
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if err = lockAuthIPPolicy(ctx, tx, false); err != nil {
		return nil, err
	}
	snapshot := &service.AuthIPBlacklistSnapshot{}
	err = tx.QueryRowContext(ctx, `SELECT enabled, threshold, window_seconds, version, block_seconds, block_mode,
aggregation_mode, network_ipv4_prefix, network_ipv6_prefix FROM custom_auth_ip_blacklist_config WHERE id = 1 FOR SHARE`).Scan(
		&snapshot.Config.Enabled, &snapshot.Config.Threshold, &snapshot.Config.WindowSeconds, &snapshot.Config.Version,
		&snapshot.Config.BlockSeconds, &snapshot.Config.BlockMode, &snapshot.Config.AggregationMode,
		&snapshot.Config.NetworkIPv4Prefix, &snapshot.Config.NetworkIPv6Prefix)
	if err != nil {
		return nil, err
	}
	target, err := service.AuthIPAggregationTarget(ip, snapshot.Config)
	if err != nil {
		return nil, err
	}
	if err = lockAuthIP(ctx, tx, target); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE custom_auth_ip_blacklist
SET blocked = FALSE, generation = generation + 1, unblocked_at = NOW(), updated_at = NOW()
WHERE ip = $1::inet AND blocked AND blocked_until IS NOT NULL AND blocked_until <= NOW()`, target); err != nil {
		return nil, err
	}
	entry, err := scanAuthIPEntry(tx.QueryRowContext(ctx, `SELECT `+authIPEntryColumns+`
FROM custom_auth_ip_blacklist entry WHERE entry.ip = $1::inet`, target))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	snapshot.Entry = entry
	snapshot.Config.NetworkState, err = authIPNetworkState(ctx, tx, target)
	if err != nil {
		return nil, err
	}
	active, err := activeAuthIPEntry(ctx, tx, ip)
	if err != nil {
		return nil, err
	}
	if active != nil {
		snapshot.Entry = active
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (repo *authIPBlacklistRepository) Config(ctx context.Context) (*service.AuthIPBlacklistConfig, error) {
	if err := repo.ready(); err != nil {
		return nil, err
	}
	config := &service.AuthIPBlacklistConfig{}
	err := repo.db.QueryRowContext(ctx, `SELECT enabled, threshold, window_seconds, version, block_seconds, block_mode,
aggregation_mode, network_ipv4_prefix, network_ipv6_prefix
FROM custom_auth_ip_blacklist_config WHERE id = 1`).Scan(&config.Enabled, &config.Threshold, &config.WindowSeconds, &config.Version,
		&config.BlockSeconds, &config.BlockMode, &config.AggregationMode, &config.NetworkIPv4Prefix, &config.NetworkIPv6Prefix)
	return config, err
}

func (repo *authIPBlacklistRepository) UpdateConfig(ctx context.Context, input service.AuthIPBlacklistConfig) (*service.AuthIPBlacklistConfig, error) {
	if err := repo.ready(); err != nil {
		return nil, err
	}
	config := &service.AuthIPBlacklistConfig{}
	err := repo.db.QueryRowContext(ctx, `UPDATE custom_auth_ip_blacklist_config
SET enabled = $1, threshold = $2, window_seconds = $3, version = version + 1, updated_at = NOW(), block_seconds = $5, block_mode = $6,
aggregation_mode = $7, network_ipv4_prefix = $8, network_ipv6_prefix = $9
WHERE id = 1 AND version = $4 RETURNING enabled, threshold, window_seconds, version, block_seconds, block_mode,
aggregation_mode, network_ipv4_prefix, network_ipv6_prefix`,
		input.Enabled, input.Threshold, input.WindowSeconds, input.Version, input.CooldownSeconds(), input.BlockMode,
		input.AggregationMode, input.NetworkIPv4Prefix, input.NetworkIPv6Prefix).
		Scan(&config.Enabled, &config.Threshold, &config.WindowSeconds, &config.Version, &config.BlockSeconds, &config.BlockMode,
			&config.AggregationMode, &config.NetworkIPv4Prefix, &config.NetworkIPv6Prefix)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrAuthIPBlacklistConflict
	}
	return config, err
}

func (repo *authIPBlacklistRepository) List(ctx context.Context, ip string, page, pageSize int) ([]service.AuthIPBlacklistEntry, int64, error) {
	if err := repo.ready(); err != nil {
		return nil, 0, err
	}
	tx, err := repo.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = tx.Rollback() }()
	var total int64
	filter := ` FROM custom_auth_ip_blacklist entry WHERE entry.blocked AND (entry.blocked_until IS NULL OR entry.blocked_until > NOW()) AND ($1 = '' OR entry.ip && NULLIF($1, '')::inet)`
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*)`+filter, ip).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT `+authIPEntryColumns+filter+` ORDER BY entry.blocked_at DESC, entry.id DESC LIMIT $2 OFFSET $3`,
		ip, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	entries := make([]service.AuthIPBlacklistEntry, 0)
	for rows.Next() {
		entry, scanErr := scanAuthIPEntry(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		entries = append(entries, *entry)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	if err = rows.Close(); err != nil {
		return nil, 0, err
	}
	if err = tx.Commit(); err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}

func lockAuthIP(ctx context.Context, tx *sql.Tx, ip string) error {
	_, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 874231))`, ip)
	return err
}

func (repo *authIPBlacklistRepository) Block(ctx context.Context, ip, path string, snapshot service.AuthIPBlacklistSnapshot, events service.AuthIPWindowCount) (*service.AuthIPBlacklistDecision, error) {
	if err := repo.ready(); err != nil {
		return nil, err
	}
	target, err := service.AuthIPAggregationTarget(ip, snapshot.Config)
	if err != nil {
		return nil, err
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if err = lockAuthIPPolicy(ctx, tx, strings.Contains(target, "/")); err != nil {
		return nil, err
	}
	var enabled bool
	var version int64
	if err = tx.QueryRowContext(ctx, `SELECT enabled, version FROM custom_auth_ip_blacklist_config WHERE id = 1 FOR SHARE`).Scan(&enabled, &version); err != nil {
		return nil, err
	}
	if !enabled || version != snapshot.Config.Version {
		return &service.AuthIPBlacklistDecision{}, nil
	}
	if err = lockAuthIP(ctx, tx, target); err != nil {
		return nil, err
	}
	active, err := activeAuthIPEntry(ctx, tx, ip)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return &service.AuthIPBlacklistDecision{Entry: active}, nil
	}
	state, err := authIPNetworkState(ctx, tx, target)
	if err != nil {
		return nil, err
	}
	if state != snapshot.Config.NetworkState {
		return &service.AuthIPBlacklistDecision{}, nil
	}
	current, err := scanAuthIPEntry(tx.QueryRowContext(ctx, `SELECT `+authIPEntryColumns+`
FROM custom_auth_ip_blacklist entry WHERE entry.ip = $1::inet FOR UPDATE`, target))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if current != nil && current.Blocked {
		return &service.AuthIPBlacklistDecision{}, nil
	}
	expectedGeneration := int64(0)
	if snapshot.Entry != nil {
		expectedGeneration = snapshot.Entry.Generation
	}
	if (current == nil && expectedGeneration != 0) || (current != nil && current.Generation != expectedGeneration) {
		return &service.AuthIPBlacklistDecision{}, nil
	}
	entry, err := scanAuthIPEntry(tx.QueryRowContext(ctx, `INSERT INTO custom_auth_ip_blacklist AS entry
(ip, blocked, generation, blocked_at, trigger_path, trigger_count, window_seconds, first_event_at, last_event_at, rule_version, blocked_until)
VALUES ($1::inet, TRUE, 1, $2, $3, $4, $5, $6, $7, $8, CASE WHEN $10 = 'permanent' THEN NULL ELSE NOW() + $9 * INTERVAL '1 second' END)
ON CONFLICT (ip) DO UPDATE SET blocked = TRUE, generation = entry.generation + 1,
blocked_at = EXCLUDED.blocked_at, trigger_path = EXCLUDED.trigger_path,
trigger_count = EXCLUDED.trigger_count, window_seconds = EXCLUDED.window_seconds,
first_event_at = EXCLUDED.first_event_at, last_event_at = EXCLUDED.last_event_at,
rule_version = EXCLUDED.rule_version, blocked_until = EXCLUDED.blocked_until, source = 'auto', reason = '', unblocked_at = NULL, updated_at = NOW()
RETURNING `+authIPEntryColumns,
		target, events.Now, path, events.Count, snapshot.Config.WindowSeconds, events.First, events.Last, snapshot.Config.Version, snapshot.Config.CooldownSeconds(), snapshot.Config.BlockMode))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &service.AuthIPBlacklistDecision{Entry: entry, NewlyBlocked: true}, nil
}

func (repo *authIPBlacklistRepository) Unblock(ctx context.Context, id, generation int64) (*service.AuthIPBlacklistEntry, error) {
	if err := repo.ready(); err != nil {
		return nil, err
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var ip string
	if err = tx.QueryRowContext(ctx, `SELECT `+authIPTargetColumn+` FROM custom_auth_ip_blacklist entry WHERE id = $1`, id).Scan(&ip); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrAuthIPBlacklistNotFound
		}
		return nil, err
	}
	if err = lockAuthIPPolicy(ctx, tx, true); err != nil {
		return nil, err
	}
	if err = lockAuthIP(ctx, tx, ip); err != nil {
		return nil, err
	}
	entry, err := scanAuthIPEntry(tx.QueryRowContext(ctx, `UPDATE custom_auth_ip_blacklist AS entry
SET blocked = FALSE, generation = generation + 1, unblocked_at = NOW(), updated_at = NOW()
WHERE id = $1 AND blocked AND generation = $2 AND (blocked_until IS NULL OR blocked_until > statement_timestamp()) RETURNING `+authIPEntryColumns, id, generation))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrAuthIPBlacklistConflict
	}
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return entry, nil
}
