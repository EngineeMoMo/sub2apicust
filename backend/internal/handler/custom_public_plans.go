package handler

import (
	"encoding/json"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type publicPlan struct {
	StockRemaining  int      `json:"stock_remaining"`
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Price           float64  `json:"price"`
	OriginalPrice   *float64 `json:"original_price,omitempty"`
	Currency        string   `json:"currency,omitempty"`
	ValidityDays    int      `json:"validity_days"`
	ValidityUnit    string   `json:"validity_unit"`
	Features        []string `json:"features"`
	DailyLimitUSD   *float64 `json:"daily_limit_usd,omitempty"`
	WeeklyLimitUSD  *float64 `json:"weekly_limit_usd,omitempty"`
	MonthlyLimitUSD *float64 `json:"monthly_limit_usd,omitempty"`
	ModelScopes     []string `json:"supported_model_scopes"`
}

func publicPlans(plans []*dbent.SubscriptionPlan, groups map[int64]service.PlanGroupInfo) []publicPlan {
	result := make([]publicPlan, 0, len(plans))
	for _, plan := range plans {
		if !plan.ForSale {
			continue
		}
		features := []string{}
		if err := json.Unmarshal([]byte(plan.Features), &features); err != nil || features == nil {
			features = []string{}
		}
		info := groups[plan.GroupID]
		scopes := info.ModelScopes
		if scopes == nil {
			scopes = []string{}
		}
		result = append(result, publicPlan{
			StockRemaining: service.CustomPlanStockRemaining(plan),
			ID:             plan.ID, Name: plan.Name, Description: plan.Description,
			Price: plan.Price, OriginalPrice: plan.OriginalPrice, Currency: plan.Currency,
			ValidityDays: plan.ValidityDays, ValidityUnit: plan.ValidityUnit, Features: features,
			DailyLimitUSD: info.DailyLimitUSD, WeeklyLimitUSD: info.WeeklyLimitUSD,
			MonthlyLimitUSD: info.MonthlyLimitUSD, ModelScopes: scopes,
		})
	}
	return result
}

func (h *PaymentHandler) GetPublicPlans(settings *service.SettingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		if settings == nil {
			response.Error(c, 503, "Public settings unavailable")
			return
		}
		config, err := settings.GetPublicSettings(c.Request.Context())
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		if config.BackendModeEnabled || !config.SubscriptionEnabled {
			response.NotFound(c, "Subscription catalog unavailable")
			return
		}
		plans, err := h.configService.ListPlansForSale(c.Request.Context())
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, gin.H{
			"plans":            publicPlans(plans, h.configService.GetGroupInfoMap(c.Request.Context(), plans)),
			"purchase_enabled": config.PaymentEnabled,
		})
	}
}
