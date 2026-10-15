package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type CustomGroupAccessHandler struct {
	service *service.CustomGroupAccessService
	cfg     *config.Config
}

func NewCustomGroupAccessHandler(access *service.CustomGroupAccessService, cfg *config.Config) *CustomGroupAccessHandler {
	return &CustomGroupAccessHandler{service: access, cfg: cfg}
}

func (h *CustomGroupAccessHandler) supported(c *gin.Context) bool {
	if h.cfg != nil && h.cfg.RunMode == config.RunModeSimple {
		response.BadRequest(c, "分组开放规则仅支持标准运行模式")
		return false
	}
	return true
}

func (h *CustomGroupAccessHandler) Get(c *gin.Context) {
	if !h.supported(c) {
		return
	}
	groupID, valid := parsePositiveIDParam(c, "id")
	if !valid {
		return
	}
	policy, err := h.service.Get(c.Request.Context(), groupID)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, policy)
}

func (h *CustomGroupAccessHandler) Save(c *gin.Context) {
	if !h.supported(c) {
		return
	}
	groupID, valid := parsePositiveIDParam(c, "id")
	if !valid {
		return
	}
	var policy service.CustomGroupAccessPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		response.BadRequest(c, "开放规则格式无效")
		return
	}
	policy.GroupID = groupID
	saved, err := h.service.Save(c.Request.Context(), policy)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, saved)
}
