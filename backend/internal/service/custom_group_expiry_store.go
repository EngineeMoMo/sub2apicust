package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type customGroupExpiryMove struct {
	targetID int64
	users    int
	keys     int64
	authKeys []string
}

func (s *customGroupAccessSQLStore) dueGroupExpiries(ctx context.Context, now time.Time) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT p.group_id FROM custom_group_access_policies p
 JOIN groups g ON g.id=p.group_id WHERE g.deleted_at IS NULL AND p.expires_at <= $1 ORDER BY p.expires_at,p.group_id`, now)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []int64
	for rows.Next() {
		var groupID int64
		if err := rows.Scan(&groupID); err != nil {
			return nil, err
		}
		ids = append(ids, groupID)
	}
	return ids, rows.Err()
}

func (s *customGroupAccessSQLStore) moveExpiredGroup(ctx context.Context, sourceID int64, now time.Time) (*customGroupExpiryMove, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(25720261015)"); err != nil {
		return nil, err
	}
	policies, err := listCustomGroupAccess(ctx, tx)
	if err != nil {
		return nil, err
	}
	sourcePolicy := policies[sourceID]
	if !sourcePolicy.expired(now) {
		return nil, nil
	}
	seen := map[int64]bool{sourceID: true}
	policy := sourcePolicy
	targetID := sourceID
	for policy.expired(now) {
		if policy.FallbackGroupID == nil || seen[*policy.FallbackGroupID] || len(seen) >= 32 {
			return nil, ErrGroupAccessExpired
		}
		targetID = *policy.FallbackGroupID
		seen[targetID] = true
		policy = policies[targetID]
	}
	groups, err := lockCustomGroupExpiryGroups(ctx, tx, sourceID, targetID)
	if err != nil {
		return nil, err
	}
	source, target := groups[sourceID], groups[targetID]
	if source == nil {
		return nil, nil
	}
	if target == nil || !target.IsActive() || target.Platform != source.Platform {
		return nil, errors.New("平移目标须为同平台有效分组")
	}
	if source.IsSubscriptionType() || target.IsSubscriptionType() {
		return nil, errors.New("订阅权益不能自动删除或合并，请改用按量分组")
	}
	var dedicated bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM custom_dedicated_accounts WHERE group_id=ANY($1::bigint[]))`, pq.Array([]int64{sourceID, targetID})).Scan(&dedicated); err != nil {
		return nil, err
	}
	if dedicated {
		return nil, errors.New("包号分组不能通过福利到期任务删除或平移")
	}
	userIDs, err := customGroupExpiryUsers(ctx, tx, sourceID)
	if err != nil {
		return nil, err
	}
	move := &customGroupExpiryMove{targetID: targetID, users: len(userIDs)}
	rows, err := tx.QueryContext(ctx, `SELECT key FROM api_keys WHERE user_id=ANY($1::bigint[]) AND deleted_at IS NULL`, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			_ = rows.Close()
			return nil, err
		}
		move.authKeys = append(move.authKeys, key)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO user_allowed_groups(user_id,group_id)
 SELECT unnest($1::bigint[]),$2 ON CONFLICT DO NOTHING`, pq.Array(userIDs), targetID); err != nil {
		return nil, err
	}
	updated, err := tx.ExecContext(ctx, `UPDATE api_keys SET group_id=$2,updated_at=$3 WHERE group_id=$1 AND deleted_at IS NULL`, sourceID, targetID, now)
	if err != nil {
		return nil, err
	}
	move.keys, err = updated.RowsAffected()
	if err != nil {
		return nil, err
	}
	statements := []string{
		`DELETE FROM user_allowed_groups WHERE group_id=$1`,
		`DELETE FROM account_groups WHERE group_id=$1`,
		`UPDATE composite_model_routes SET deleted_at=NOW() WHERE group_id=$1 AND deleted_at IS NULL`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement, sourceID); err != nil {
			return nil, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE custom_group_access_policies SET fallback_group_id=$2,updated_at=clock_timestamp()
 WHERE group_id<>$1 AND fallback_group_id=$1`, sourceID, targetID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE groups SET
 fallback_group_id=CASE WHEN fallback_group_id=$1 THEN CASE WHEN id=$2 THEN NULL ELSE $2 END ELSE fallback_group_id END,
 fallback_group_id_on_invalid_request=CASE WHEN fallback_group_id_on_invalid_request=$1 THEN CASE WHEN id=$2 THEN NULL ELSE $2 END ELSE fallback_group_id_on_invalid_request END,updated_at=NOW()
 WHERE deleted_at IS NULL AND id<>$1 AND (fallback_group_id=$1 OR fallback_group_id_on_invalid_request=$1)`, sourceID, targetID); err != nil {
		return nil, err
	}
	if err := removeCustomGroupExpiryModeration(ctx, tx, sourceID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE groups SET deleted_at=$2,updated_at=$2 WHERE id=$1 AND deleted_at IS NULL`, sourceID, now); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO scheduler_outbox(event_type,group_id) VALUES($1,$2),($1,$3)`, SchedulerOutboxEventGroupChanged, sourceID, targetID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_notify('custom_group_access_changed','{"kind":"reset"}')`); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return move, nil
}

func (s *customGroupAccessSQLStore) checkExpiryDedicatedGroups(ctx context.Context, sourceID, targetID int64) error {
	var dedicated bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM custom_dedicated_accounts WHERE group_id=ANY($1::bigint[]))`, pq.Array([]int64{sourceID, targetID})).Scan(&dedicated); err != nil {
		return err
	}
	if dedicated {
		return ErrGroupAccessExpiryDedicated
	}
	return nil
}

func lockCustomGroupExpiryGroups(ctx context.Context, tx *sql.Tx, sourceID, targetID int64) (map[int64]*Group, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,platform,status,subscription_type FROM groups
 WHERE id=ANY($1::bigint[]) AND deleted_at IS NULL ORDER BY id FOR UPDATE`, pq.Array([]int64{sourceID, targetID}))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	groups := make(map[int64]*Group)
	for rows.Next() {
		group := &Group{}
		if err := rows.Scan(&group.ID, &group.Platform, &group.Status, &group.SubscriptionType); err != nil {
			return nil, err
		}
		groups[group.ID] = group
	}
	return groups, rows.Err()
}

func customGroupExpiryUsers(ctx context.Context, tx *sql.Tx, groupID int64) ([]int64, error) {
	rows, err := tx.QueryContext(ctx, `SELECT user_id FROM api_keys WHERE group_id=$1 AND deleted_at IS NULL
 UNION SELECT user_id FROM user_allowed_groups WHERE group_id=$1 ORDER BY user_id`, groupID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	ids := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		ids = append(ids, userID)
	}
	return ids, rows.Err()
}

func removeCustomGroupExpiryModeration(ctx context.Context, tx *sql.Tx, groupID int64) error {
	var value string
	err := tx.QueryRowContext(ctx, "SELECT value FROM settings WHERE key=$1 FOR UPDATE", SettingKeyContentModerationConfig).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if !json.Valid([]byte(value)) {
		return nil
	}
	_, err = tx.ExecContext(ctx, `UPDATE settings SET value=jsonb_set(value::jsonb,'{group_ids}',
 (SELECT COALESCE(jsonb_agg(group_id ORDER BY ord),'[]'::jsonb)
 FROM jsonb_array_elements(value::jsonb->'group_ids') WITH ORDINALITY AS ids(group_id,ord)
 WHERE group_id<>to_jsonb($2::bigint)))::text,updated_at=NOW()
 WHERE key=$1 AND jsonb_typeof(value::jsonb->'group_ids')='array'
 AND value::jsonb->'group_ids' @> jsonb_build_array($2::bigint)`, SettingKeyContentModerationConfig, groupID)
	if err != nil {
		return fmt.Errorf("清理到期分组审核范围: %w", err)
	}
	return nil
}
