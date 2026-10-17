package service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
)

func customCollectionBoundInstance(order *dbent.PaymentOrder) (int64, string, bool) {
	instanceID := strings.TrimSpace(psStringValue(order.ProviderInstanceID))
	if snapshot := psOrderProviderSnapshot(order); snapshot != nil && snapshot.ProviderInstanceID != "" {
		if instanceID != "" && instanceID != snapshot.ProviderInstanceID {
			return 0, "", false
		}
		instanceID = snapshot.ProviderInstanceID
	}
	value, err := strconv.ParseInt(instanceID, 10, 64)
	return value, instanceID, err == nil && value > 0
}

func (s *PaymentService) customAutoReleaseMissingProviderHold(ctx context.Context, order *dbent.PaymentOrder, now time.Time, wait time.Duration, operator string) (bool, error) {
	if !customCollectionTerminalUnpaid(order) || now.Before(order.ExpiresAt.Add(wait+customCollectionReleaseMargin)) {
		return false, nil
	}
	instanceID, instanceRef, valid := customCollectionBoundInstance(order)
	if !valid {
		return false, nil
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.Client().ExecContext(ctx, `UPDATE custom_payment_collection_ledger SET released=true
 WHERE order_id=$1 AND paid_at IS NULL AND NOT released
 AND EXISTS (SELECT 1 FROM payment_orders o WHERE o.id=$1 AND o.paid_at IS NULL
  AND o.status IN ($2,$3,$4) AND o.expires_at <= $5
  AND (TRIM(COALESCE(o.provider_instance_id,''))='' OR TRIM(o.provider_instance_id)=$6)
  AND (TRIM(COALESCE(o.provider_snapshot->>'provider_instance_id',''))='' OR TRIM(o.provider_snapshot->>'provider_instance_id')=$6)
  AND (TRIM(COALESCE(o.provider_instance_id,''))=$6 OR TRIM(COALESCE(o.provider_snapshot->>'provider_instance_id',''))=$6))
 AND NOT EXISTS (SELECT 1 FROM payment_provider_instances WHERE id=$7)`,
		order.ID, OrderStatusExpired, OrderStatusCancelled, OrderStatusFailed,
		now.Add(-wait-customCollectionReleaseMargin), instanceRef, instanceID)
	if err != nil {
		return false, err
	}
	changed, err := result.RowsAffected()
	if err != nil || changed == 0 {
		return false, err
	}
	detail, err := json.Marshal(map[string]any{
		"reason":               "按管理员启用的自动清理规则，原支付通道已删除且已超过等待期，仅释放收款占用，未能核实实际收款",
		"provider_instance_id": instanceID,
		"expires_at":           order.ExpiresAt.UTC().Format(time.RFC3339),
		"release_after":        order.ExpiresAt.Add(wait + customCollectionReleaseMargin).UTC().Format(time.RFC3339),
	})
	if err != nil {
		return false, err
	}
	_, err = tx.PaymentAuditLog.Create().SetOrderID(strconv.FormatInt(order.ID, 10)).
		SetAction("COLLECTION_HOLD_AUTO_RELEASED").SetOperator(operator).SetDetail(string(detail)).Save(ctx)
	if err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
