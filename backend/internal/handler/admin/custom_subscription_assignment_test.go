package admin

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type customAssignmentGroupRepo struct{ service.GroupRepository }

func (customAssignmentGroupRepo) GetByID(context.Context, int64) (*service.Group, error) {
	return &service.Group{ID: 10, SubscriptionType: service.SubscriptionTypeSubscription}, nil
}

type customAssignmentSubscriptionRepo struct {
	*bulkActionHandlerSubscriptionRepo
}

func (repo *customAssignmentSubscriptionRepo) ExistsByUserIDAndGroupID(_ context.Context, userID, groupID int64) (bool, error) {
	return repo.sub != nil && repo.sub.UserID == userID && repo.sub.GroupID == groupID, nil
}

func (repo *customAssignmentSubscriptionRepo) GetByUserIDAndGroupID(ctx context.Context, userID, groupID int64) (*service.UserSubscription, error) {
	if repo.sub == nil || repo.sub.UserID != userID || repo.sub.GroupID != groupID {
		return nil, service.ErrSubscriptionNotFound
	}
	return repo.GetByID(ctx, repo.sub.ID)
}

func TestCustomAdminSubscriptionAssignment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := service.DefaultIdempotencyConfig()
	cfg.ObserveOnly = false
	previous := service.DefaultIdempotencyCoordinator()
	service.SetDefaultIdempotencyCoordinator(service.NewIdempotencyCoordinator(newMemoryIdempotencyRepoStub(), cfg))
	t.Cleanup(func() { service.SetDefaultIdempotencyCoordinator(previous) })
	start := time.Now().Add(-time.Hour)
	expiry := start.AddDate(0, 0, 30)
	repo := &customAssignmentSubscriptionRepo{&bulkActionHandlerSubscriptionRepo{sub: &service.UserSubscription{
		ID: 1, UserID: 1, GroupID: 10, StartsAt: start, ExpiresAt: expiry, Status: service.SubscriptionStatusActive,
	}}}
	svc := service.NewSubscriptionService(customAssignmentGroupRepo{}, repo, nil, nil, nil)
	t.Cleanup(svc.Stop)
	handler := NewSubscriptionHandler(svc)
	router := gin.New()
	path := "/api/v1/admin/subscriptions/assign"
	router.POST(path, handler.Assign)
	body := `{"user_id":1,"group_id":10,"plan_id":71,"validity_days":3}`

	first := bulkActionHandlerRequest(router, path, body, "assign-3-days")
	require.Equal(t, http.StatusOK, first.Code, first.Body.String())
	require.Equal(t, expiry.AddDate(0, 0, 3), repo.sub.ExpiresAt)
	require.Equal(t, 1, repo.extendCalls)

	replay := bulkActionHandlerRequest(router, path, body, "assign-3-days")
	require.Equal(t, http.StatusOK, replay.Code, replay.Body.String())
	require.Equal(t, "true", replay.Header().Get("X-Idempotency-Replayed"))
	require.JSONEq(t, first.Body.String(), replay.Body.String())
	require.Equal(t, 1, repo.extendCalls)

	conflict := bulkActionHandlerRequest(router, path, strings.Replace(body, ":3}", ":7}", 1), "assign-3-days")
	require.Equal(t, http.StatusConflict, conflict.Code)
	require.Contains(t, conflict.Body.String(), "IDEMPOTENCY_KEY_CONFLICT")
	require.Equal(t, 1, repo.extendCalls)

	missing := bulkActionHandlerRequest(router, path, body, "")
	require.Equal(t, http.StatusBadRequest, missing.Code)
	require.Equal(t, 1, repo.extendCalls)

	distinct := bulkActionHandlerRequest(router, path, body, "assign-another-3-days")
	require.Equal(t, http.StatusOK, distinct.Code, distinct.Body.String())
	require.Equal(t, expiry.AddDate(0, 0, 6), repo.sub.ExpiresAt)
	require.Equal(t, 2, repo.extendCalls)
}
