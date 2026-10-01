package service

import (
	"context"
	"database/sql"
	"errors"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var ErrDedicatedDelete = infraerrors.BadRequest("DEDICATED_ACCOUNT_DELETE", "请先撤销包号，再删除记录")

const customDedicatedViewSQL = "SELECT " + customDedicatedColumns + " FROM custom_dedicated_accounts WHERE id=$1 AND deleted_at IS NULL AND (user_ids @> jsonb_build_array($2::bigint) OR (user_ids='[]'::jsonb AND user_id=$2))"

func customDedicatedMembers(binding CustomDedicatedBinding) []int64 {
	if len(binding.UserIDs) > 0 {
		return binding.UserIDs
	}
	return []int64{binding.UserID}
}

func validateCustomDedicatedMembers(members []int64) error {
	if len(members) == 0 || len(members) > 100 {
		return ErrDedicatedInput
	}
	seen := make(map[int64]bool, len(members))
	for _, id := range members {
		if id <= 0 || seen[id] {
			return ErrDedicatedInput
		}
		seen[id] = true
	}
	return nil
}

func customDedicatedHasMember(binding CustomDedicatedBinding, userID int64) bool {
	for _, id := range customDedicatedMembers(binding) {
		if id == userID {
			return true
		}
	}
	return false
}

func customDedicatedRemoveMembers(ctx context.Context, tx *sql.Tx, previous, next CustomDedicatedBinding) ([]int64, error) {
	removed := make([]int64, 0)
	for _, userID := range customDedicatedMembers(previous) {
		if customDedicatedHasMember(next, userID) {
			continue
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM user_allowed_groups WHERE (group_id=$1 OR group_id IN (SELECT ag.group_id FROM account_groups ag JOIN groups g ON g.id=ag.group_id WHERE ag.account_id=$3 AND g.deleted_at IS NULL AND g.is_exclusive=TRUE)) AND user_id=$2", previous.GroupID, userID, previous.AccountID); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, "UPDATE api_keys SET status='inactive', updated_at=NOW() WHERE (group_id=$1 OR group_id IN (SELECT ag.group_id FROM account_groups ag JOIN groups g ON g.id=ag.group_id WHERE ag.account_id=$3 AND g.deleted_at IS NULL AND g.is_exclusive=TRUE)) AND user_id=$2 AND deleted_at IS NULL AND status <> 'inactive'", previous.GroupID, userID, previous.AccountID); err != nil {
			return nil, err
		}
		removed = append(removed, userID)
	}
	return removed, nil
}

func (s *CustomDedicatedService) Delete(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	binding, err := scanCustomDedicated(tx.QueryRowContext(ctx, "SELECT "+customDedicatedColumns+" FROM custom_dedicated_accounts WHERE id=$1 AND deleted_at IS NULL FOR UPDATE", id))
	if errors.Is(err, sql.ErrNoRows) {
		return ErrDedicatedNotFound
	}
	if err != nil {
		return err
	}
	if binding.RevokedAt == nil {
		return ErrDedicatedDelete
	}
	if _, err := tx.ExecContext(ctx, "UPDATE custom_dedicated_accounts SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1", id); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.apiKeys.InvalidateAuthCacheByGroupID(ctx, binding.GroupID)
	return nil
}
