package admin

import (
	"context"
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) RecheckCollection(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()
	released, err := h.paymentService.CustomRecheckCollectionHolds(ctx, getAdminIDFromContext(c))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"released": released})
}

func (h *PaymentHandler) CloseCollection(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "订单编号无效")
		return
	}
	if err = h.paymentService.CustomCloseCollection(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "渠道关单已确认，收款预占已释放"})
}

// ReleaseCollection 人工释放无法由渠道复查的收款预占：须写原因、确认，且过等待期；全程写审计。
func (h *PaymentHandler) ReleaseCollection(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "订单编号无效")
		return
	}
	var req struct {
		Reason    string `json:"reason"`
		Confirmed bool   `json:"confirmed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "释放参数无效")
		return
	}
	if err := h.paymentService.CustomForceReleaseCollection(c.Request.Context(), id, getAdminIDFromContext(c), req.Reason, req.Confirmed); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "已人工释放收款占用，并记录审计"})
}

// [CUSTOM] 仅管理员路由注册。
func (h *PaymentHandler) GetCollection(c *gin.Context) {
	cfg, err := h.configService.GetPaymentConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	usage, err := h.configService.CustomCollectionUsage(c.Request.Context(), cfg.Collection)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, usage)
}

func (h *PaymentHandler) RecordCollection(c *gin.Context) {
	var req struct {
		Amount    float64 `json:"amount"`
		Reference string  `json:"reference"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "收款登记格式无效")
		return
	}
	if err := h.configService.CustomRecordCollection(c.Request.Context(), req.Amount, req.Reference); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "已登记"})
}
