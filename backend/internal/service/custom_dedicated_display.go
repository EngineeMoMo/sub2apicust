package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type CustomDedicatedAdminView struct {
	CustomDedicatedBinding
	ConfigStatus string                  `json:"config_status"`
	Users        []CustomDedicatedMember `json:"users"`
	Platform     string                  `json:"platform"`
	UserName     string                  `json:"user_name"`
	AccountName  string                  `json:"account_name"`
	GroupName    string                  `json:"group_name"`
}

type CustomDedicatedMember struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

const customDedicatedAdminListSQL = `SELECT d.id, d.user_id, d.account_id, d.group_id, d.label, d.expires_at, d.revoked_at, d.updated_at,
 COALESCE(NULLIF(u.username, ''), u.email, ''), COALESCE(a.name, ''), COALESCE(g.name, ''), COALESCE(a.platform, ''),
 (SELECT jsonb_agg(jsonb_build_object('id', member.id::bigint, 'name', COALESCE(NULLIF(person.username, ''), person.email, '')) ORDER BY member.position)
 FROM jsonb_array_elements_text(CASE WHEN d.user_ids='[]'::jsonb THEN jsonb_build_array(d.user_id) ELSE d.user_ids END) WITH ORDINALITY member(id, position)
 LEFT JOIN users person ON person.id=member.id::bigint AND person.deleted_at IS NULL)
 FROM custom_dedicated_accounts d
 LEFT JOIN users u ON u.id=d.user_id AND u.deleted_at IS NULL
 LEFT JOIN accounts a ON a.id=d.account_id AND a.deleted_at IS NULL
 LEFT JOIN groups g ON g.id=d.group_id AND g.deleted_at IS NULL
 WHERE d.deleted_at IS NULL
 ORDER BY d.id DESC LIMIT 50 OFFSET $1`

func (s *CustomDedicatedService) AdminList(ctx context.Context, page int) ([]CustomDedicatedAdminView, error) {
	if page < 1 || page > 1000000 {
		return nil, ErrDedicatedInput
	}
	rows, err := s.db.QueryContext(ctx, customDedicatedAdminListSQL, (page-1)*50)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make([]CustomDedicatedAdminView, 0)
	for rows.Next() {
		var item CustomDedicatedAdminView
		var members []byte
		if err := rows.Scan(&item.ID, &item.UserID, &item.AccountID, &item.GroupID, &item.Label, &item.ExpiresAt, &item.RevokedAt, &item.UpdatedAt, &item.UserName, &item.AccountName, &item.GroupName, &item.Platform, &members); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(members, &item.Users); err != nil {
			return nil, err
		}
		for _, member := range item.Users {
			item.UserIDs = append(item.UserIDs, member.ID)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for index := range result {
		item := &result[index]
		members, _ := json.Marshal(customDedicatedMembers(item.CustomDedicatedBinding))
		var structure, membership bool
		err := s.db.QueryRowContext(ctx, customDedicatedHealthSQL, item.AccountID, item.GroupID, string(members)).Scan(&structure, &membership)
		if err != nil {
			return nil, err
		}
		item.ConfigStatus = "valid"
		if !structure || s.simpleMode {
			item.ConfigStatus = "invalid_structure"
		} else if !membership {
			item.ConfigStatus = "member_warning"
		}
	}
	return result, nil
}

const customDedicatedHealthSQL = `SELECT (` + customDedicatedStructureSQL + `), (` + customDedicatedIntegritySQL + `)`

func (s *CustomDedicatedService) FillGroupNames(ctx context.Context, userID int64, views []CustomDedicatedView) error {
	if userID <= 0 {
		return ErrDedicatedAccess
	}
	if len(views) == 0 {
		return nil
	}
	arguments := []any{userID}
	placeholders := make([]string, len(views))
	for index, view := range views {
		arguments = append(arguments, view.ID)
		placeholders[index] = fmt.Sprintf("$%d", index+2)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT d.id, COALESCE(g.name, '') FROM custom_dedicated_accounts d
 LEFT JOIN groups g ON g.id=d.group_id AND g.deleted_at IS NULL
 WHERE d.deleted_at IS NULL AND (d.user_ids @> jsonb_build_array($1::bigint) OR (d.user_ids='[]'::jsonb AND d.user_id=$1)) AND d.id IN (`+strings.Join(placeholders, ",")+`)`, arguments...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	names := make(map[int64]string, len(views))
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		names[id] = name
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for index := range views {
		views[index].GroupName = names[views[index].ID]
	}
	return nil
}
