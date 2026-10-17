package service

import (
	"context"
	"fmt"
	"strconv"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

func (s *PaymentConfigService) customCheckProviderCollectionHolds(ctx context.Context, instanceID int64) error {
	rows, err := s.entClient.QueryContext(ctx, `SELECT count(*)
 FROM custom_payment_collection_ledger l
 JOIN payment_orders o ON o.id=l.order_id
 WHERE l.paid_at IS NULL AND NOT l.released
 AND (TRIM(COALESCE(o.provider_instance_id,''))=$1
  OR TRIM(COALESCE(o.provider_snapshot->>'provider_instance_id',''))=$1)`, strconv.FormatInt(instanceID, 10))
	if err != nil {
		return fmt.Errorf("check provider collection holds: %w", err)
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return fmt.Errorf("check provider collection holds: %w", err)
		}
		return fmt.Errorf("check provider collection holds: missing count")
	}
	var count int
	if err := rows.Scan(&count); err != nil {
		return fmt.Errorf("check provider collection holds: %w", err)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("check provider collection holds: %w", err)
	}
	if count > 0 {
		return infraerrors.Conflict("PROVIDER_COLLECTION_HOLDS",
			fmt.Sprintf("该支付通道仍有%d笔待确认收款占用，不能删除；请先在设置→支付→检查待确认订单中核账处理，或停用通道保留原配置", count)).
			WithMetadata(map[string]string{"count": strconv.Itoa(count)})
	}
	return nil
}
