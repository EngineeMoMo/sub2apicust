package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

func (s *CustomDedicatedService) AdmitBillingRequest(ctx context.Context, key *APIKey, websocket bool) error {
	if !key.IsCustomDedicatedPrepaid() {
		return nil
	}
	grant := key.customDedicatedBilling
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ErrDedicatedAccess
	}
	defer func() { _ = tx.Rollback() }()
	binding, err := scanCustomDedicated(tx.QueryRowContext(ctx, "SELECT "+customDedicatedColumns+` FROM custom_dedicated_accounts WHERE id=$1 AND deleted_at IS NULL FOR UPDATE`, grant.BindingID))
	if err != nil {
		return ErrDedicatedAccess
	}
	subject := customDedicatedSubject{UserID: grant.UserID, GroupID: grant.GroupID}
	requestBinding := binding
	requestBinding.GroupID = grant.GroupID
	var acceptedAt time.Time
	if err := tx.QueryRowContext(ctx, "SELECT clock_timestamp()").Scan(&acceptedAt); err != nil {
		return ErrDedicatedAccess
	}
	if binding.AccountID != grant.AccountID || !customDedicatedAllowed(requestBinding, subject, nil, acceptedAt) {
		return ErrDedicatedAccess
	}
	valid, err := customDedicatedRequestAccess(ctx, tx, binding, subject)
	if err != nil || !valid {
		return ErrDedicatedAccess
	}
	var random [24]byte
	if _, err := rand.Read(random[:]); err != nil {
		return ErrDedicatedAccess
	}
	receiptID := hex.EncodeToString(random[:])
	if _, err := tx.ExecContext(ctx, `INSERT INTO custom_dedicated_request_leases(id,binding_id,account_id,user_id,api_key_id,group_id,expires_at,accepted_at,finished_at) VALUES($1,$2,$3,$4,$5,$6,$7,$7,$7)`, receiptID, grant.BindingID, grant.AccountID, grant.UserID, grant.APIKeyID, grant.GroupID, acceptedAt); err != nil {
		return ErrDedicatedAccess
	}
	if err := tx.Commit(); err != nil {
		return ErrDedicatedAccess
	}
	grant.LeaseID = receiptID
	grant.WebSocket = websocket
	return nil
}

func (s *APIKeyService) AdmitCustomDedicatedRequest(ctx context.Context, key *APIKey, websocket bool) error {
	if !key.IsCustomDedicatedPrepaid() {
		return nil
	}
	if s.customDedicated == nil {
		return ErrDedicatedAccess
	}
	return s.customDedicated.AdmitBillingRequest(ctx, key, websocket)
}
