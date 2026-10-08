package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

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
