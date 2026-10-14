package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (repo *authIPBlacklistRepository) ManualBlock(ctx context.Context, input service.AuthIPManualBlock) (*service.AuthIPBlacklistEntry, error) {
	if err := repo.ready(); err != nil {
		return nil, err
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if err = lockAuthIP(ctx, tx, input.IP); err != nil {
		return nil, err
	}
	var id int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM custom_auth_ip_blacklist
WHERE ip = $1::inet AND blocked AND (blocked_until IS NULL OR blocked_until > clock_timestamp())`, input.IP).Scan(&id)
	if err == nil {
		return nil, service.ErrAuthIPAlreadyBlocked
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	entry, err := scanAuthIPEntry(tx.QueryRowContext(ctx, `INSERT INTO custom_auth_ip_blacklist AS entry
(ip, blocked, generation, blocked_at, trigger_path, trigger_count, window_seconds, first_event_at, last_event_at, rule_version, blocked_until, source, reason)
VALUES ($1::inet, TRUE, 1, statement_timestamp(), '', 0, 0, NULL, NULL, 0,
CASE WHEN $2 = 'permanent' THEN NULL ELSE statement_timestamp() + $3 * INTERVAL '1 second' END, 'manual', $4)
ON CONFLICT (ip) DO UPDATE SET blocked = TRUE, generation = entry.generation + 1,
blocked_at = EXCLUDED.blocked_at, trigger_path = '', trigger_count = 0, window_seconds = 0,
first_event_at = NULL, last_event_at = NULL, rule_version = 0, blocked_until = EXCLUDED.blocked_until,
source = 'manual', reason = EXCLUDED.reason, unblocked_at = NULL, updated_at = statement_timestamp()
RETURNING `+authIPEntryColumns, input.IP, input.BlockMode, input.BlockSeconds, input.Reason))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return entry, nil
}
