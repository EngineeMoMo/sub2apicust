package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type customGroupAccessSQLStore struct{ db *sql.DB }

const customGroupAccessColumns = `group_id, monthly_recharge_enabled, monthly_recharge_min,
 balance_enabled, balance_min, expires_at, fallback_group_id, rate_schedule_enabled, rate_schedule, updated_at`

type customGroupAccessScanner interface{ Scan(...any) error }

func scanCustomGroupAccess(row customGroupAccessScanner) (CustomGroupAccessPolicy, error) {
	var result CustomGroupAccessPolicy
	var schedule []byte
	err := row.Scan(&result.GroupID, &result.MonthlyRechargeEnabled, &result.MonthlyRechargeMin,
		&result.BalanceEnabled, &result.BalanceMin, &result.ExpiresAt, &result.FallbackGroupID, &result.RateScheduleEnabled, &schedule, &result.UpdatedAt)
	if err == nil {
		err = json.Unmarshal(schedule, &result.RateSchedule)
	}
	return result, err
}

type customGroupAccessQuerier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func listCustomGroupAccess(ctx context.Context, db customGroupAccessQuerier) (map[int64]CustomGroupAccessPolicy, error) {
	rows, err := db.QueryContext(ctx, "SELECT "+customGroupAccessColumns+" FROM custom_group_access_policies")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make(map[int64]CustomGroupAccessPolicy)
	for rows.Next() {
		policy, err := scanCustomGroupAccess(rows)
		if err != nil {
			return nil, err
		}
		result[policy.GroupID] = policy
	}
	return result, rows.Err()
}

func (s *customGroupAccessSQLStore) List(ctx context.Context) (map[int64]CustomGroupAccessPolicy, error) {
	return listCustomGroupAccess(ctx, s.db)
}

const customGroupAccessFactsSQL = `SELECT u.balance, COALESCE((
 SELECT SUM(GREATEST(o.amount-o.bonus_amount,0) * CASE
  WHEN o.status='PARTIALLY_REFUNDED' THEN GREATEST(0,1-o.refund_amount/NULLIF(o.amount,0))
  ELSE 1 END)
 FROM payment_orders o WHERE o.user_id=u.id AND o.order_type='balance'
 AND o.status IN ('COMPLETED','PARTIALLY_REFUNDED')
 AND o.paid_at >= $2 AND o.paid_at < $3
),0) FROM users u WHERE u.id=$1 AND u.deleted_at IS NULL`

func (s *customGroupAccessSQLStore) Facts(ctx context.Context, userID int64, start, end time.Time) (customGroupAccessFacts, error) {
	var result customGroupAccessFacts
	err := s.db.QueryRowContext(ctx, customGroupAccessFactsSQL, userID, start, end).Scan(&result.Balance, &result.MonthlyRecharge)
	return result, err
}

func validateCustomGroupAccessGraph(policies map[int64]CustomGroupAccessPolicy) error {
	for groupID := range policies {
		seen := make(map[int64]bool)
		for current := groupID; current != 0; {
			if seen[current] || len(seen) >= 32 {
				return infraerrors.BadRequest("GROUP_ACCESS_FALLBACK_CYCLE", "到期后分组不能形成循环，且切换链不能超过32组")
			}
			seen[current] = true
			policy := policies[current]
			if policy.FallbackGroupID == nil {
				break
			}
			current = *policy.FallbackGroupID
		}
	}
	return nil
}

func (s *customGroupAccessSQLStore) Save(ctx context.Context, policy CustomGroupAccessPolicy) (*CustomGroupAccessPolicy, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(25720261015)"); err != nil {
		return nil, err
	}
	var groupID int64
	if err := tx.QueryRowContext(ctx, "SELECT id FROM groups WHERE id=$1 AND deleted_at IS NULL FOR SHARE", policy.GroupID).Scan(&groupID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	policies, err := listCustomGroupAccess(ctx, tx)
	if err != nil {
		return nil, err
	}
	current := policies[policy.GroupID]
	if (current.UpdatedAt == nil) != (policy.UpdatedAt == nil) || (current.UpdatedAt != nil && !current.UpdatedAt.Equal(*policy.UpdatedAt)) {
		return nil, ErrGroupAccessPolicyStale
	}
	policies[policy.GroupID] = policy
	if err := validateCustomGroupAccessGraph(policies); err != nil {
		return nil, err
	}
	schedule := policy.RateSchedule
	if schedule == nil {
		schedule = []CustomGroupRatePeriod{}
	}
	scheduleJSON, err := json.Marshal(schedule)
	if err != nil {
		return nil, err
	}
	saved, err := scanCustomGroupAccess(tx.QueryRowContext(ctx, `INSERT INTO custom_group_access_policies
 (`+customGroupAccessColumns+`) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,clock_timestamp())
 ON CONFLICT(group_id) DO UPDATE SET monthly_recharge_enabled=EXCLUDED.monthly_recharge_enabled,
 monthly_recharge_min=EXCLUDED.monthly_recharge_min, balance_enabled=EXCLUDED.balance_enabled,
 balance_min=EXCLUDED.balance_min, expires_at=EXCLUDED.expires_at,
 fallback_group_id=EXCLUDED.fallback_group_id, rate_schedule_enabled=EXCLUDED.rate_schedule_enabled,
 rate_schedule=EXCLUDED.rate_schedule, updated_at=clock_timestamp()
 RETURNING `+customGroupAccessColumns,
		policy.GroupID, policy.MonthlyRechargeEnabled, policy.MonthlyRechargeMin,
		policy.BalanceEnabled, policy.BalanceMin, policy.ExpiresAt, policy.FallbackGroupID, policy.RateScheduleEnabled, string(scheduleJSON)))
	if err != nil {
		return nil, fmt.Errorf("save group access policy: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &saved, nil
}
