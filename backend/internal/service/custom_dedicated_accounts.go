package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrDedicatedMode     = infraerrors.Forbidden("DEDICATED_ACCOUNT_MODE", "包号功能仅支持标准运行模式")
	ErrDedicatedAccess   = infraerrors.Forbidden("DEDICATED_ACCOUNT_UNAVAILABLE", "专属账号不可用或不属于当前用户，请联系管理员")
	ErrDedicatedConfig   = infraerrors.BadRequest("DEDICATED_ACCOUNT_CONFIG", "请使用仅包含该账号、仅授权所选用户、无备用路由的同平台标准专属分组；账号不能属于其他分组或拥有影子账号")
	ErrDedicatedInput    = infraerrors.BadRequest("DEDICATED_ACCOUNT_INPUT", "请选择用户、账号和分组，填写80字以内别名及未来的到期时间")
	ErrDedicatedConflict = infraerrors.Conflict("DEDICATED_ACCOUNT_CONFLICT", "该账号或分组已有包号记录，请修改原记录")
	ErrDedicatedNotFound = infraerrors.NotFound("DEDICATED_ACCOUNT_NOT_FOUND", "包号记录不存在")
)

type CustomDedicatedBinding struct {
	UserIDs   []int64    `json:"user_ids"`
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	AccountID int64      `json:"account_id"`
	GroupID   int64      `json:"group_id"`
	Label     string     `json:"label"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CustomDedicatedInput struct {
	UserIDs   []int64   `json:"user_ids"`
	UserID    int64     `json:"user_id"`
	AccountID int64     `json:"account_id"`
	GroupID   int64     `json:"group_id"`
	Label     string    `json:"label"`
	ExpiresAt time.Time `json:"expires_at"`
}

type CustomDedicatedWindow struct {
	Key              string     `json:"key"`
	RemainingPercent *float64   `json:"remaining_percent"`
	ResetsAt         *time.Time `json:"resets_at"`
	Stale            bool       `json:"stale"`
}

type CustomDedicatedView struct {
	GroupName  string                  `json:"group_name"`
	ID         int64                   `json:"id"`
	Label      string                  `json:"label"`
	Platform   string                  `json:"platform"`
	GroupID    int64                   `json:"group_id"`
	ExpiresAt  time.Time               `json:"expires_at"`
	Status     string                  `json:"status"`
	LastUsedAt *time.Time              `json:"last_used_at"`
	SampledAt  *time.Time              `json:"sampled_at"`
	CheckedAt  time.Time               `json:"checked_at"`
	QuotaState string                  `json:"quota_state"`
	Windows    []CustomDedicatedWindow `json:"windows"`
}

type CustomDedicatedService struct {
	simpleMode bool
	db         *sql.DB
	accounts   AccountRepository
	apiKeys    APIKeyAuthCacheInvalidator
}

type customDedicatedContextKey struct{}
type customDedicatedSubject struct{ UserID, GroupID int64 }

func NewCustomDedicatedService(db *sql.DB, accounts AccountRepository, apiKeys *APIKeyService, gateway *GatewayService, openai *OpenAIGatewayService) *CustomDedicatedService {
	result := &CustomDedicatedService{db: db, accounts: accounts, apiKeys: apiKeys, simpleMode: apiKeys.cfg != nil && apiKeys.cfg.RunMode == config.RunModeSimple}
	apiKeys.customDedicated = result
	gateway.customDedicated = result
	openai.customDedicated = result
	return result
}

const customDedicatedColumns = "id, user_id, account_id, group_id, label, expires_at, revoked_at, updated_at, user_ids"

type customDedicatedScanner interface{ Scan(...any) error }

func scanCustomDedicated(row customDedicatedScanner) (CustomDedicatedBinding, error) {
	var binding CustomDedicatedBinding
	var members []byte
	err := row.Scan(&binding.ID, &binding.UserID, &binding.AccountID, &binding.GroupID, &binding.Label, &binding.ExpiresAt, &binding.RevokedAt, &binding.UpdatedAt, &members)
	if err == nil {
		err = json.Unmarshal(members, &binding.UserIDs)
		if len(binding.UserIDs) == 0 {
			binding.UserIDs = []int64{binding.UserID}
		}
	}
	return binding, err
}

func (s *CustomDedicatedService) List(ctx context.Context, userID int64, page int) ([]CustomDedicatedBinding, error) {
	if userID < 0 || page < 1 || page > 1000000 {
		return nil, ErrDedicatedInput
	}
	rows, err := s.db.QueryContext(ctx, "SELECT "+customDedicatedColumns+" FROM custom_dedicated_accounts WHERE deleted_at IS NULL AND ($1::bigint = 0 OR user_ids @> jsonb_build_array($1::bigint) OR (user_ids = '[]'::jsonb AND user_id = $1)) ORDER BY id DESC LIMIT 50 OFFSET $2", userID, (page-1)*50)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	bindings := make([]CustomDedicatedBinding, 0)
	for rows.Next() {
		binding, err := scanCustomDedicated(rows)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, binding)
	}
	return bindings, rows.Err()
}

const customDedicatedIntegritySQL = `SELECT EXISTS (
 SELECT 1 FROM accounts a JOIN groups g ON g.id = $2
 WHERE a.id = $1 AND a.deleted_at IS NULL AND g.deleted_at IS NULL
 AND g.status = 'active' AND g.is_exclusive = TRUE
 AND g.platform = a.platform AND g.subscription_type = 'standard'
 AND g.fallback_group_id IS NULL AND g.fallback_group_id_on_invalid_request IS NULL
 AND a.platform IN ('anthropic', 'openai')
 AND ((a.platform = 'anthropic' AND a.type IN ('oauth','setup-token')) OR (a.platform = 'openai' AND a.type = 'oauth'))
 AND a.parent_account_id IS NULL
 AND NOT EXISTS (SELECT 1 FROM accounts child WHERE child.parent_account_id = a.id AND child.deleted_at IS NULL)
 AND EXISTS (SELECT 1 FROM account_groups ag WHERE ag.account_id = a.id AND ag.group_id = g.id)
 AND NOT EXISTS (SELECT 1 FROM account_groups ag WHERE (ag.account_id = a.id AND ag.group_id <> g.id) OR (ag.group_id = g.id AND ag.account_id <> a.id))
 AND jsonb_array_length($3::jsonb) > 0
 AND NOT EXISTS (SELECT 1 FROM jsonb_array_elements_text($3::jsonb) member(id)
   WHERE NOT EXISTS (SELECT 1 FROM users u JOIN user_allowed_groups ug ON ug.user_id=u.id AND ug.group_id=g.id
     WHERE u.id=member.id::bigint AND u.deleted_at IS NULL AND u.status='active'))
 AND NOT EXISTS (SELECT 1 FROM user_allowed_groups ug JOIN users other ON other.id=ug.user_id WHERE ug.group_id=g.id AND other.deleted_at IS NULL AND NOT $3::jsonb @> jsonb_build_array(ug.user_id))
 AND NOT EXISTS (SELECT 1 FROM api_keys k WHERE k.group_id=g.id AND k.deleted_at IS NULL AND NOT $3::jsonb @> jsonb_build_array(k.user_id))
 AND NOT EXISTS (SELECT 1 FROM user_subscriptions sub WHERE sub.group_id = g.id AND sub.deleted_at IS NULL)
)`

type customDedicatedQuery interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func customDedicatedIntegrity(ctx context.Context, query customDedicatedQuery, binding CustomDedicatedBinding) (bool, error) {
	var valid bool
	members, _ := json.Marshal(customDedicatedMembers(binding))
	err := query.QueryRowContext(ctx, customDedicatedIntegritySQL, binding.AccountID, binding.GroupID, string(members)).Scan(&valid)
	return valid, err
}

func (s *CustomDedicatedService) Save(ctx context.Context, id int64, input CustomDedicatedInput) (*CustomDedicatedBinding, error) {
	if s.simpleMode {
		return nil, ErrDedicatedMode
	}
	input.Label = strings.TrimSpace(input.Label)
	if input.UserIDs == nil {
		input.UserIDs = []int64{input.UserID}
	}
	if err := validateCustomDedicatedMembers(input.UserIDs); err != nil {
		return nil, err
	}
	input.UserID = input.UserIDs[0]
	if input.UserID <= 0 || input.AccountID <= 0 || input.GroupID <= 0 || input.Label == "" || utf8.RuneCountInString(input.Label) > 80 || !input.ExpiresAt.After(time.Now()) {
		return nil, ErrDedicatedInput
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	binding := CustomDedicatedBinding{ID: id, UserID: input.UserID, UserIDs: input.UserIDs, AccountID: input.AccountID, GroupID: input.GroupID, Label: input.Label, ExpiresAt: input.ExpiresAt}
	var previous CustomDedicatedBinding
	if id != 0 {
		previous, err = scanCustomDedicated(tx.QueryRowContext(ctx, "SELECT "+customDedicatedColumns+" FROM custom_dedicated_accounts WHERE id = $1 AND deleted_at IS NULL FOR UPDATE", id))
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDedicatedNotFound
		}
		if err != nil {
			return nil, err
		}
	}
	valid, err := customDedicatedIntegrity(ctx, tx, binding)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, ErrDedicatedConfig
	}
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM custom_dedicated_accounts WHERE (account_id = $1 OR group_id = $2) AND id <> $3 AND deleted_at IS NULL)", input.AccountID, input.GroupID, id).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDedicatedConflict
	}
	var row *sql.Row
	members, _ := json.Marshal(input.UserIDs)
	if id == 0 {
		row = tx.QueryRowContext(ctx, "INSERT INTO custom_dedicated_accounts (user_id, account_id, group_id, label, expires_at, user_ids) VALUES ($1,$2,$3,$4,$5,$6::jsonb) RETURNING "+customDedicatedColumns, input.UserID, input.AccountID, input.GroupID, input.Label, input.ExpiresAt, string(members))
	} else {
		row = tx.QueryRowContext(ctx, "UPDATE custom_dedicated_accounts SET user_id=$1, label=$2, expires_at=$3, revoked_at=NULL, updated_at=NOW(), account_id=$5, group_id=$6, user_ids=$7::jsonb WHERE id=$4 RETURNING "+customDedicatedColumns, input.UserID, input.Label, input.ExpiresAt, id, input.AccountID, input.GroupID, string(members))
	}
	binding, err = scanCustomDedicated(row)
	if err != nil {
		return nil, err
	}
	if id != 0 && (previous.AccountID != input.AccountID || previous.GroupID != input.GroupID) {
		oldMembers, _ := json.Marshal(customDedicatedMembers(previous))
		_, err = tx.ExecContext(ctx, "INSERT INTO custom_dedicated_accounts (user_id, account_id, group_id, label, expires_at, revoked_at, deleted_at, user_ids) VALUES ($1,$2,$3,$4,$5,NOW(),NOW(),$6::jsonb)", previous.UserID, previous.AccountID, previous.GroupID, previous.Label, previous.ExpiresAt, string(oldMembers))
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s.apiKeys.InvalidateAuthCacheByGroupID(ctx, input.GroupID)
	if previous.GroupID != 0 && previous.GroupID != input.GroupID {
		s.apiKeys.InvalidateAuthCacheByGroupID(ctx, previous.GroupID)
	}
	return &binding, nil
}

func (s *CustomDedicatedService) Revoke(ctx context.Context, id int64) error {
	var groupID int64
	err := s.db.QueryRowContext(ctx, "UPDATE custom_dedicated_accounts SET revoked_at=COALESCE(revoked_at,NOW()), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL RETURNING group_id", id).Scan(&groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrDedicatedNotFound
	}
	if err != nil {
		return err
	}
	s.apiKeys.InvalidateAuthCacheByGroupID(ctx, groupID)
	return nil
}

func (s *APIKeyService) CustomDedicatedContext(ctx context.Context, key *APIKey) (context.Context, error) {
	if s.customDedicated == nil {
		return ctx, nil
	}
	if key == nil || key.UserID <= 0 {
		return ctx, ErrDedicatedAccess
	}
	subject := customDedicatedSubject{UserID: key.UserID}
	if key.GroupID != nil {
		subject.GroupID = *key.GroupID
	}
	ctx = context.WithValue(ctx, customDedicatedContextKey{}, subject)
	if err := s.customDedicated.Check(ctx, nil, key.GroupID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func customDedicatedAllowed(binding CustomDedicatedBinding, subject customDedicatedSubject, account *Account, now time.Time) bool {
	if binding.RevokedAt != nil || !now.Before(binding.ExpiresAt) || !customDedicatedHasMember(binding, subject.UserID) || subject.GroupID != binding.GroupID {
		return false
	}
	return account == nil || account.ID == binding.AccountID
}

func (s *CustomDedicatedService) Check(ctx context.Context, account *Account, groupID *int64) error {
	if s == nil {
		return nil
	}
	subject, authenticated := ctx.Value(customDedicatedContextKey{}).(customDedicatedSubject)
	if !authenticated && groupID != nil {
		subject.GroupID = *groupID
	}
	accountID, parentID := int64(0), int64(0)
	if account != nil {
		accountID = account.ID
		if account.ParentAccountID != nil {
			parentID = *account.ParentAccountID
		}
	}
	rows, err := s.db.QueryContext(ctx, "SELECT "+customDedicatedColumns+" FROM custom_dedicated_accounts WHERE (group_id=$1 OR account_id=$2 OR account_id=$3 OR account_id=(SELECT parent_account_id FROM accounts WHERE id=$2)) AND (deleted_at IS NULL OR NOT EXISTS (SELECT 1 FROM custom_dedicated_accounts live WHERE live.deleted_at IS NULL AND live.group_id=$1 AND ($2::bigint=0 OR live.account_id=$2)))", subject.GroupID, accountID, parentID)
	if err != nil {
		return ErrDedicatedAccess
	}
	bindings := make([]CustomDedicatedBinding, 0, 2)
	for rows.Next() {
		binding, scanErr := scanCustomDedicated(rows)
		if scanErr != nil {
			_ = rows.Close()
			return ErrDedicatedAccess
		}
		bindings = append(bindings, binding)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return ErrDedicatedAccess
	}
	for _, binding := range bindings {
		if s.simpleMode {
			return ErrDedicatedAccess
		}
		if !authenticated || !customDedicatedAllowed(binding, subject, account, time.Now()) {
			return ErrDedicatedAccess
		}
		valid, err := customDedicatedIntegrity(ctx, s.db, binding)
		if err != nil || !valid {
			return ErrDedicatedAccess
		}
	}
	return nil
}

func (s *CustomDedicatedService) guardAccount(ctx context.Context, groupID *int64, account **Account, err *error) {
	if s == nil || *err != nil || *account == nil {
		return
	}
	if checkErr := s.Check(ctx, *account, groupID); checkErr != nil {
		*account = nil
		*err = checkErr
	}
}

func (s *CustomDedicatedService) guardSelection(ctx context.Context, groupID *int64, selection **AccountSelectionResult, err *error) {
	if s == nil || *err != nil || *selection == nil || (*selection).Account == nil {
		return
	}
	if checkErr := s.Check(ctx, (*selection).Account, groupID); checkErr != nil {
		if (*selection).ReleaseFunc != nil {
			(*selection).ReleaseFunc()
		}
		*selection = nil
		*err = checkErr
	}
}

func customDedicatedQuotaWindow(key string, window *UsageProgress, sampledAt *time.Time, now time.Time) CustomDedicatedWindow {
	result := CustomDedicatedWindow{Key: key}
	if window == nil {
		return result
	}
	result.ResetsAt = window.ResetsAt
	result.Stale = sampledAt == nil || now.Sub(*sampledAt) > 15*time.Minute || sampledAt.After(now.Add(time.Minute)) || (window.ResetsAt != nil && !now.Before(*window.ResetsAt))
	if result.Stale || math.IsNaN(window.Utilization) || math.IsInf(window.Utilization, 0) || window.Utilization < 0 {
		return result
	}
	remaining := math.Max(0, 100-window.Utilization)
	result.RemainingPercent = &remaining
	return result
}

func (s *CustomDedicatedService) View(ctx context.Context, userID, id int64, quota bool) (*CustomDedicatedView, error) {
	binding, err := scanCustomDedicated(s.db.QueryRowContext(ctx, customDedicatedViewSQL, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDedicatedNotFound
	}
	if err != nil {
		return nil, err
	}
	now := time.Now()
	view := &CustomDedicatedView{ID: binding.ID, Label: binding.Label, GroupID: binding.GroupID, ExpiresAt: binding.ExpiresAt, Status: "unavailable", CheckedAt: now, QuotaState: "unknown", Windows: []CustomDedicatedWindow{}}
	if binding.RevokedAt != nil {
		view.Status = "revoked"
		return view, nil
	}
	if !now.Before(binding.ExpiresAt) {
		view.Status = "expired"
		return view, nil
	}
	valid, err := customDedicatedIntegrity(ctx, s.db, binding)
	if err != nil {
		return nil, err
	}
	if !valid || s.simpleMode {
		view.Status = "configuration_error"
		return view, nil
	}
	account, err := s.accounts.GetByID(ctx, binding.AccountID)
	if err != nil {
		return view, nil
	}
	view.Platform = account.Platform
	if account.LastUsedAt != nil && !account.LastUsedAt.Before(binding.UpdatedAt) {
		view.LastUsedAt = account.LastUsedAt
	}
	if account.IsSchedulable() {
		view.Status = "available"
	} else if account.RateLimitResetAt != nil && now.Before(*account.RateLimitResetAt) {
		view.Status = "rate_limited"
	}
	if !quota {
		return view, nil
	}
	usage := customDedicatedPassiveUsage(account, now)
	view.SampledAt = usage.UpdatedAt
	for _, entry := range []struct {
		key    string
		window *UsageProgress
	}{{"five_hour", usage.FiveHour}, {"seven_day", usage.SevenDay}, {"seven_day_sonnet", usage.SevenDaySonnet}, {"seven_day_fable", usage.SevenDayFable}} {
		if entry.window == nil {
			continue
		}
		window := customDedicatedQuotaWindow(entry.key, entry.window, view.SampledAt, now)
		view.Windows = append(view.Windows, window)
		if window.RemainingPercent != nil {
			view.QuotaState = "available"
		} else if view.QuotaState != "available" && window.Stale {
			view.QuotaState = "stale"
		}
	}
	latest, err := scanCustomDedicated(s.db.QueryRowContext(ctx, customDedicatedViewSQL, id, userID))
	if err != nil || !latest.UpdatedAt.Equal(binding.UpdatedAt) || latest.RevokedAt != nil || !time.Now().Before(latest.ExpiresAt) {
		return nil, ErrDedicatedNotFound
	}
	return view, nil
}

func (s *CustomDedicatedService) UserList(ctx context.Context, userID int64, page int) ([]CustomDedicatedView, error) {
	if userID <= 0 {
		return nil, ErrDedicatedAccess
	}
	bindings, err := s.List(ctx, userID, page)
	if err != nil {
		return nil, err
	}
	views := make([]CustomDedicatedView, 0, len(bindings))
	for _, binding := range bindings {
		view, err := s.View(ctx, userID, binding.ID, true)
		if err != nil {
			return nil, fmt.Errorf("dedicated view: %w", err)
		}
		views = append(views, *view)
	}
	return views, nil
}
