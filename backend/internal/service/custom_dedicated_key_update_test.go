//go:build unit

package service

import (
	"context"
	"database/sql"
)

type customDedicatedKeyUpdateRepo struct {
	APIKeyRepository
	database *sql.DB
}

func (repo *customDedicatedKeyUpdateRepo) GetByID(ctx context.Context, id int64) (*APIKey, error) {
	groupID := int64(33)
	key := &APIKey{ID: id, UserID: 12, GroupID: &groupID}
	err := repo.database.QueryRowContext(ctx, "SELECT status FROM api_keys WHERE user_id=12 AND group_id=33").Scan(&key.Status)
	return key, err
}

func (repo *customDedicatedKeyUpdateRepo) Update(ctx context.Context, key *APIKey, fields APIKeyUpdateFields) error {
	_, err := repo.database.ExecContext(ctx, "UPDATE api_keys SET status=$1 WHERE user_id=12 AND group_id=33", key.Status)
	return err
}
