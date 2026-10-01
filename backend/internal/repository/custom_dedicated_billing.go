package repository

import (
	"context"
	"database/sql"
	"encoding/hex"
	"math"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// 包号准入凭证在DB验证，异步结算不重新按当前到期／成员状态决定是否扣余额。
func applyCustomDedicatedBilling(ctx context.Context, tx *sql.Tx, cmd *service.UsageBillingCommand) error {
	// 服务端执行ID必须属于此准入凭证；WS另带上游轮次哈希，不能用客户端ID覆盖。
	prefix := "dedicated:" + cmd.DedicatedLeaseID
	validRequestID := cmd.RequestID == prefix
	if suffix, ok := strings.CutPrefix(cmd.RequestID, prefix+":"); ok && len(suffix) == 64 {
		_, err := hex.DecodeString(suffix)
		validRequestID = err == nil
	}
	if cmd.DedicatedLeaseID == "" || cmd.DedicatedBindingID <= 0 || cmd.DedicatedGroupID <= 0 || !validRequestID || math.IsNaN(cmd.DedicatedReferenceCost) || math.IsInf(cmd.DedicatedReferenceCost, 0) || cmd.DedicatedReferenceCost < 0 {
		return service.ErrDedicatedAccess
	}
	var valid bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM custom_dedicated_request_leases WHERE id=$1 AND binding_id=$2 AND account_id=$3 AND user_id=$4 AND api_key_id=$5 AND group_id=$6)`, cmd.DedicatedLeaseID, cmd.DedicatedBindingID, cmd.AccountID, cmd.UserID, cmd.APIKeyID, cmd.DedicatedGroupID).Scan(&valid); err != nil {
		return err
	}
	if !valid || cmd.BillingType != service.BillingTypeDedicated || cmd.BalanceCost != 0 || cmd.SubscriptionCost != 0 {
		return service.ErrDedicatedAccess
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO custom_dedicated_billing_usage(request_id,api_key_id,lease_id,binding_id,user_id,account_id,group_id,model,reference_cost,input_tokens,output_tokens) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, cmd.RequestID, cmd.APIKeyID, cmd.DedicatedLeaseID, cmd.DedicatedBindingID, cmd.UserID, cmd.AccountID, cmd.DedicatedGroupID, cmd.Model, cmd.DedicatedReferenceCost, cmd.InputTokens, cmd.OutputTokens)
	return err
}
