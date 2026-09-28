//go:build unit

package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

type publicCatalogSettings struct {
	service.SettingRepository
	values map[string]string
	err    error
}

func (s publicCatalogSettings) GetMultiple(context.Context, []string) (map[string]string, error) {
	return s.values, s.err
}

func TestPublicPlansProjection(t *testing.T) {
	quota := 12.5
	plans := []*dbent.SubscriptionPlan{
		{ID: 9, GroupID: 42, Name: "在售", Price: 25, Currency: "USD", ForSale: true, Features: `["权益"]`, ValidityDays: 1, ValidityUnit: "months"},
		{ID: 3, Name: "下架", ForSale: false},
		{ID: 7, Name: "缺少权益", ForSale: true, Features: "invalid"},
	}
	result := publicPlans(plans, map[int64]service.PlanGroupInfo{42: {DailyLimitUSD: &quota, ModelScopes: []string{"test"}}})
	require.Len(t, result, 2)
	require.Equal(t, int64(9), result[0].ID)
	require.Equal(t, int64(7), result[1].ID)
	require.Equal(t, []string{"权益"}, result[0].Features)
	require.Equal(t, &quota, result[0].DailyLimitUSD)
	require.Nil(t, result[1].DailyLimitUSD)
	require.NotNil(t, result[1].Features)
	require.NotNil(t, result[1].ModelScopes)
	body, err := json.Marshal(result[0])
	require.NoError(t, err)
	var fields map[string]any
	require.NoError(t, json.Unmarshal(body, &fields))
	allowed := map[string]bool{"id": true, "name": true, "description": true, "price": true, "original_price": true, "currency": true, "validity_days": true, "validity_unit": true, "features": true, "daily_limit_usd": true, "weekly_limit_usd": true, "monthly_limit_usd": true, "supported_model_scopes": true}
	for key := range fields {
		require.True(t, allowed[key], "unexpected field: %s", key)
	}
	require.Equal(t, []publicPlan{}, publicPlans(nil, nil))
}

func TestPublicPlansRejectDisabledOrUnavailableSettings(t *testing.T) {
	for _, test := range []struct {
		name   string
		values map[string]string
		err    error
		status int
	}{
		{"backend", map[string]string{service.SettingKeyBackendModeEnabled: "true"}, nil, 404},
		{"subscription", map[string]string{service.SettingKeySubscriptionEnabled: "false"}, nil, 404},
		{"settings failure", nil, errors.New("settings unavailable"), 500},
	} {
		t.Run(test.name, func(t *testing.T) {
			settings := service.NewSettingService(publicCatalogSettings{values: test.values, err: test.err}, &config.Config{})
			handler := &PaymentHandler{}
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest("GET", "/api/v1/payment/public/plans", nil)
			handler.GetPublicPlans(settings)(ctx)
			require.Equal(t, test.status, recorder.Code)
			require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
		})
	}
}

func TestPublicPlansAnonymousCatalog(t *testing.T) {
	database, err := sql.Open("sqlite", "file:custom_public_catalog?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(entsql.OpenDB(dialect.SQLite, database))))
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	for _, item := range []struct {
		name string
		sort int
		sale bool
	}{
		{"后排套餐", 20, true},
		{"下架套餐", 0, false},
		{"首排套餐", 10, true},
	} {
		_, err := client.SubscriptionPlan.Create().SetGroupID(42).SetName(item.name).
			SetPrice(25).SetSortOrder(item.sort).SetForSale(item.sale).Save(ctx)
		require.NoError(t, err)
	}
	for _, enabled := range []string{"true", "false"} {
		t.Run("payment_"+enabled, func(t *testing.T) {
			settings := service.NewSettingService(publicCatalogSettings{values: map[string]string{
				service.SettingKeySubscriptionEnabled: "true",
				service.SettingPaymentEnabled:         enabled,
			}}, &config.Config{})
			handler := NewPaymentHandler(nil, service.NewPaymentConfigService(client, nil, nil))
			router := gin.New()
			router.GET("/api/v1/payment/public/plans", handler.GetPublicPlans(settings))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/v1/payment/public/plans", nil))
			require.Equal(t, 200, recorder.Code)
			var result struct {
				Data struct {
					Plans           []publicPlan `json:"plans"`
					PurchaseEnabled bool         `json:"purchase_enabled"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &result))
			require.Len(t, result.Data.Plans, 2)
			require.Equal(t, "首排套餐", result.Data.Plans[0].Name)
			require.Equal(t, "后排套餐", result.Data.Plans[1].Name)
			require.Equal(t, enabled == "true", result.Data.PurchaseEnabled)
		})
	}
}
