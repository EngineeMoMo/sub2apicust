package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *accountRepository) UpdateKiroOAuthCredentialsIfUnchanged(ctx context.Context, id int64, expectedCredentials map[string]any, expectedProxyID *int64, credentials map[string]any) (bool, error) {
	if r == nil || r.sql == nil {
		return false, errors.New("Kiro 账号 SQL 执行器未配置")
	}
	expectedJSON, err := json.Marshal(normalizeJSONMap(expectedCredentials))
	if err != nil {
		return false, err
	}
	credentialsJSON, err := json.Marshal(normalizeJSONMap(credentials))
	if err != nil {
		return false, err
	}
	result, err := r.sql.ExecContext(ctx, `
 WITH updated AS (
  UPDATE accounts AS a SET credentials=$1::jsonb, updated_at=NOW()
  WHERE a.id=$2 AND a.deleted_at IS NULL AND a.platform=$3 AND a.type=$4
   AND a.credentials=$5::jsonb AND a.proxy_id IS NOT DISTINCT FROM $6
  RETURNING a.id
 )
 INSERT INTO scheduler_outbox(event_type,account_id,group_id,payload)
 SELECT $7,updated.id,NULL,NULL FROM updated
 `, string(credentialsJSON), id, service.PlatformKiro, service.AccountTypeOAuth, string(expectedJSON), expectedProxyID, service.SchedulerOutboxEventAccountChanged)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	r.syncSchedulerAccountSnapshotDetached(ctx, id)
	return true, nil
}
