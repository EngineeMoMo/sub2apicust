package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type customStockGenerationCapture struct {
	service.AdminService
	input *service.GenerateRedeemCodesInput
}

func (capture *customStockGenerationCapture) GenerateRedeemCodes(_ context.Context, input *service.GenerateRedeemCodesInput) ([]service.RedeemCode, error) {
	capture.input = input
	return []service.RedeemCode{}, nil
}
func TestCustomRedeemHandlerForwardsStockPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)
	capture := &customStockGenerationCapture{}
	handler := NewRedeemHandler(capture, nil)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/redeem-codes/generate", strings.NewReader(`{"count":1,"type":"subscription","value":0,"group_id":10,"validity_days":7,"plan_id":22}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler.Generate(ctx)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotNil(t, capture.input)
	require.Equal(t, int64(22), capture.input.PlanID)
	require.Equal(t, int64(10), *capture.input.GroupID)
}
func TestCustomRedeemFixedSubscriptionAllowsZeroValue(t *testing.T) {
	status := postCreateAndRedeemValidation(t, newCreateAndRedeemHandler(), map[string]any{"code": "fixed-sub-zero", "type": "subscription", "user_id": 1, "group_id": 10, "validity_days": 7, "plan_id": 22})
	require.NotEqual(t, http.StatusBadRequest, status)
}
func TestCustomRedeemBatchDTOForwardsStockPlan(t *testing.T) {
	planID := int64(22)
	result := redeemBatchUpdateFieldsFromDTO(dto.BatchUpdateRedeemCodeFields{PlanID: dto.NullableInt64Field{Set: true, Value: &planID}})
	require.True(t, result.PlanID.Set)
	require.Equal(t, planID, *result.PlanID.Value)
}
