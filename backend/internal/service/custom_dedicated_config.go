package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const customDedicatedConfigSQL = `WITH selected_account AS (
 SELECT * FROM accounts WHERE id=$1 AND deleted_at IS NULL
), selected_group AS (
 SELECT * FROM groups WHERE id=$2 AND deleted_at IS NULL
), members AS (
 SELECT value::bigint AS id FROM jsonb_array_elements_text($3::jsonb)
), related_groups AS (
 SELECT g.* FROM account_groups ag JOIN groups g ON g.id=ag.group_id WHERE ag.account_id=$1 AND g.deleted_at IS NULL
), problems(priority, issue, resource_id) AS (
 SELECT 1, 'account_missing', $1::bigint WHERE NOT EXISTS (SELECT 1 FROM selected_account)
 UNION ALL SELECT 2, 'group_missing', $2::bigint WHERE NOT EXISTS (SELECT 1 FROM selected_group)
 UNION ALL SELECT 3, 'group_inactive', id FROM selected_group WHERE status IS DISTINCT FROM 'active'
 UNION ALL SELECT 4, 'group_not_exclusive', id FROM selected_group WHERE is_exclusive IS DISTINCT FROM TRUE
 UNION ALL SELECT 5, 'platform_mismatch', g.id FROM selected_group g CROSS JOIN selected_account a WHERE g.platform IS DISTINCT FROM a.platform
 UNION ALL SELECT 6, 'group_not_standard', id FROM selected_group WHERE subscription_type IS DISTINCT FROM 'standard'
 UNION ALL SELECT 7, 'fallback_route', id FROM selected_group WHERE fallback_group_id IS NOT NULL OR fallback_group_id_on_invalid_request IS NOT NULL
 UNION ALL SELECT 8, 'account_type', id FROM selected_account WHERE NOT COALESCE((platform='anthropic' AND type IN ('oauth','setup-token')) OR (platform='openai' AND type='oauth'), FALSE)
 UNION ALL SELECT 9, 'shadow_account', id FROM selected_account WHERE parent_account_id IS NOT NULL
 UNION ALL SELECT 10, 'account_has_shadows', child.id FROM accounts child WHERE child.parent_account_id=$1 AND child.deleted_at IS NULL
 UNION ALL SELECT 11, 'account_not_in_group', $1::bigint WHERE NOT EXISTS (SELECT 1 FROM account_groups WHERE account_id=$1 AND group_id=$2)
 UNION ALL SELECT 12, 'account_unsafe_groups', g.id FROM related_groups g CROSS JOIN selected_account a
 WHERE g.id<>$2 AND (g.status IS DISTINCT FROM 'active' OR g.is_exclusive IS DISTINCT FROM TRUE
 OR g.platform IS DISTINCT FROM a.platform OR g.subscription_type IS DISTINCT FROM 'standard'
 OR g.fallback_group_id IS NOT NULL OR g.fallback_group_id_on_invalid_request IS NOT NULL)
 UNION ALL SELECT 13, 'group_other_accounts', account_id FROM account_groups WHERE group_id=$2 AND account_id<>$1
 UNION ALL SELECT 14, 'no_members', $2::bigint WHERE NOT EXISTS (SELECT 1 FROM members)
 UNION ALL SELECT 15, 'other_authorized_users', ug.user_id FROM user_allowed_groups ug JOIN users u ON u.id=ug.user_id
 WHERE ug.group_id IN (SELECT id FROM related_groups) AND u.deleted_at IS NULL AND NOT $3::jsonb @> jsonb_build_array(ug.user_id)
 UNION ALL SELECT 16, 'group_subscriptions', group_id FROM user_subscriptions WHERE group_id IN (SELECT id FROM related_groups) AND deleted_at IS NULL
 UNION ALL SELECT 17, 'member_unavailable', member.id FROM members member WHERE NOT EXISTS (
 SELECT 1 FROM users u WHERE u.id=member.id AND u.deleted_at IS NULL AND u.status='active')
 UNION ALL SELECT 18, 'member_not_authorized', member.id FROM members member WHERE NOT EXISTS (
 SELECT 1 FROM user_allowed_groups WHERE user_id=member.id AND group_id=$2)
 UNION ALL SELECT 19, 'non_member_keys', k.user_id FROM api_keys k
 WHERE k.group_id IN (SELECT id FROM related_groups) AND k.deleted_at IS NULL AND k.status NOT IN ('disabled','inactive') AND NOT $3::jsonb @> jsonb_build_array(k.user_id)
 UNION ALL SELECT 20, 'related_group_other_accounts', ag.group_id FROM account_groups ag
 WHERE ag.group_id IN (SELECT id FROM related_groups WHERE id<>$2) AND ag.account_id<>$1
), first_problem AS (
 SELECT priority FROM problems ORDER BY priority LIMIT 1
), affected AS (
 SELECT DISTINCT issue, resource_id FROM problems WHERE priority=(SELECT priority FROM first_problem) ORDER BY issue, resource_id LIMIT 10
)
SELECT issue, string_agg(resource_id::text, ', ' ORDER BY resource_id) FROM affected GROUP BY issue`

var customDedicatedConfigMessages = map[string]string{
	"account_missing":              "所选账号不存在或已删除，账号编号",
	"group_missing":                "所选分组不存在或已删除，分组编号",
	"group_inactive":               "所选分组未启用，分组编号",
	"group_not_exclusive":          "所选分组未开启专属分组开关；名称中含专属不等于已开启，分组编号",
	"platform_mismatch":            "所选分组与账号的平台不一致，分组编号",
	"group_not_standard":           "所选分组不是标准分组，分组编号",
	"fallback_route":               "所选分组仍配置了备用路由或无效请求备用路由，分组编号",
	"account_type":                 "账号类型不支持包号；Claude需OAuth或Setup Token，ChatGPT/Codex需OAuth，账号编号",
	"shadow_account":               "所选账号是影子账号，账号编号",
	"account_has_shadows":          "所选账号仍有未删除的影子账号，影子账号编号",
	"account_not_in_group":         "所选账号尚未加入所选分组，账号编号",
	"account_unsafe_groups":        "账号关联的其他分组不符合标准专属分组要求，请检查专属开关、状态、平台和备用路由，分组编号",
	"related_group_other_accounts": "账号关联的其他专属分组还包含别的账号，分组编号",
	"group_other_accounts":         "所选分组还包含其他账号，其他账号编号",
	"no_members":                   "未选择包号成员，分组编号",
	"other_authorized_users":       "账号关联的专属分组还授权给包号名单以外的用户；同批共用者请全部加入同一条包号，用户编号",
	"group_subscriptions":          "所选分组仍有未删除的订阅记录，分组编号",
	"member_unavailable":           "包号成员未启用、不存在或已删除，用户编号",
	"member_not_authorized":        "包号成员尚未获得所选分组授权，用户编号",
	"non_member_keys":              "账号关联的专属分组仍有非包号成员的未停用密钥，密钥所属用户编号",
}

func customDedicatedConfigError(ctx context.Context, query customDedicatedQuery, binding CustomDedicatedBinding) error {
	members, _ := json.Marshal(customDedicatedMembers(binding))
	var issue, resourceIDs string
	err := query.QueryRowContext(ctx, customDedicatedConfigSQL, binding.AccountID, binding.GroupID, string(members)).Scan(&issue, &resourceIDs)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrDedicatedConfig
	}
	if err != nil {
		return err
	}
	message, known := customDedicatedConfigMessages[issue]
	if !known {
		return ErrDedicatedConfig
	}
	return infraerrors.BadRequest("DEDICATED_ACCOUNT_CONFIG", fmt.Sprintf("%s：%s（最多列出10项）。请调整配置后保存。", message, resourceIDs)).WithMetadata(map[string]string{
		"config_issue": issue,
		"resource_ids": resourceIDs,
		"account_id":   fmt.Sprint(binding.AccountID),
		"group_id":     fmt.Sprint(binding.GroupID),
	})
}
