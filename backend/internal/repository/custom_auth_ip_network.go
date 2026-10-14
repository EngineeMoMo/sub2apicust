package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func lockAuthIPPolicy(ctx context.Context, tx *sql.Tx, exclusive bool) error {
	query := `SELECT pg_advisory_xact_lock_shared(874231, 254)`
	if exclusive {
		query = `SELECT pg_advisory_xact_lock(874231, 254)`
	}
	_, err := tx.ExecContext(ctx, query)
	return err
}

func authIPNetworkState(ctx context.Context, tx *sql.Tx, ip string) (string, error) {
	var state string
	err := tx.QueryRowContext(ctx, `SELECT COALESCE(string_agg(id::text || ':' || generation::text || ':' ||
CASE WHEN blocked AND (blocked_until IS NULL OR blocked_until > statement_timestamp()) THEN '1' ELSE '0' END, ',' ORDER BY id), '')
FROM custom_auth_ip_blacklist WHERE ip && $1::inet AND ip != $1::inet`, ip).Scan(&state)
	if err != nil || state == "" {
		return "", err
	}
	sum := sha256.Sum256([]byte(state))
	return hex.EncodeToString(sum[:]), nil
}

func activeAuthIPEntry(ctx context.Context, tx *sql.Tx, target string) (*service.AuthIPBlacklistEntry, error) {
	entry, err := scanAuthIPEntry(tx.QueryRowContext(ctx, `SELECT `+authIPEntryColumns+`
FROM custom_auth_ip_blacklist entry WHERE entry.ip >>= $1::inet AND entry.blocked
AND (entry.blocked_until IS NULL OR entry.blocked_until > statement_timestamp())
ORDER BY masklen(entry.ip) DESC, entry.id DESC LIMIT 1`, target))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return entry, err
}
