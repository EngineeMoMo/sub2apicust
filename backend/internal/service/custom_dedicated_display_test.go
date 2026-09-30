//go:build unit

package service

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestCustomDedicatedAdminNames(t *testing.T) {
	database, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = database.Close() }()
	service := &CustomDedicatedService{db: database}
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "account_id", "group_id", "label", "expires_at", "revoked_at", "updated_at", "user_name", "account_name", "group_name", "platform", "users"}).
		AddRow(1, 11, 22, 33, "包号", now, nil, now, "用户甲", "Claude 账号", "专属组", "anthropic", `[{"id":11,"name":"用户甲"},{"id":12,"name":"用户乙"}]`).
		AddRow(2, 12, 23, 34, "已删除关联", now, nil, now, "", "", "", "", `[{"id":12,"name":""}]`)
	mock.ExpectQuery(regexp.QuoteMeta(customDedicatedAdminListSQL)).WithArgs(50).WillReturnRows(rows)
	views, err := service.AdminList(context.Background(), 2)
	require.NoError(t, err)
	require.Len(t, views, 2)
	require.Equal(t, "用户甲", views[0].UserName)
	require.Equal(t, []int64{11, 12}, views[0].UserIDs)
	require.Equal(t, "用户乙", views[0].Users[1].Name)
	require.Equal(t, "Claude 账号", views[0].AccountName)
	require.Equal(t, "专属组", views[0].GroupName)
	require.Equal(t, int64(34), views[1].GroupID)
	require.Empty(t, views[1].GroupName)
	_, err = service.AdminList(context.Background(), 0)
	require.ErrorIs(t, err, ErrDedicatedInput)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomDedicatedUserGroupNames(t *testing.T) {
	database, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = database.Close() }()
	service := &CustomDedicatedService{db: database}
	views := []CustomDedicatedView{{ID: 1, GroupID: 33}, {ID: 2, GroupID: 34}}
	mock.ExpectQuery(`SELECT d.id, .*WHERE d.deleted_at IS NULL AND .*d.user_ids.* AND d.id IN \(\$2,\$3\)`).WithArgs(int64(11), int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "专属组"))
	require.NoError(t, service.FillGroupNames(context.Background(), 11, views))
	require.Equal(t, "专属组", views[0].GroupName)
	require.Empty(t, views[1].GroupName)
	body, err := json.Marshal(views[0])
	require.NoError(t, err)
	require.NotContains(t, string(body), "account_name")
	require.NotContains(t, string(body), "user_name")
	require.NoError(t, service.FillGroupNames(context.Background(), 11, nil))
	require.ErrorIs(t, service.FillGroupNames(context.Background(), 0, views), ErrDedicatedAccess)
	mock.ExpectQuery(`SELECT d.id,`).WillReturnError(errors.New("unavailable"))
	require.Error(t, service.FillGroupNames(context.Background(), 11, views))
	require.NoError(t, mock.ExpectationsWereMet())
}
