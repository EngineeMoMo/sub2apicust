package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	dbent "github.com/Wei-Shaw/sub2api/ent"
)

func (s *SubscriptionService) AssignAdminSubscription(ctx context.Context, input *AssignSubscriptionInput) (*UserSubscription, error) {
	if input == nil {
		return nil, ErrSubscriptionNilInput
	}
	var subscription *UserSubscription
	err := s.withSubscriptionUpdateTx(ctx, func(txCtx context.Context) error {
		existing, err := s.userSubRepo.GetByUserIDAndGroupID(txCtx, input.UserID, input.GroupID)
		if err != nil && !errors.Is(err, ErrSubscriptionNotFound) {
			return fmt.Errorf("read subscription before assignment: %w", err)
		}
		if existing != nil {
			locked, lockErr := s.userSubRepo.GetByIDForUpdate(txCtx, existing.ID)
			if lockErr != nil {
				return fmt.Errorf("lock subscription before assignment: %w", lockErr)
			}
			if locked.Status == SubscriptionStatusSuspended {
				return ErrSubscriptionSuspended
			}
		}
		subscription, _, err = s.assignOrExtendSubscription(txCtx, input, true)
		return err
	})
	if err != nil {
		return nil, err
	}
	if dbent.TxFromContext(ctx) == nil {
		if cacheErr := s.invalidateSubscriptionCaches(input.UserID, input.GroupID); cacheErr != nil {
			log.Printf("Warning: invalidate admin subscription assignment caches: %v", cacheErr)
		}
	}
	return subscription, nil
}
