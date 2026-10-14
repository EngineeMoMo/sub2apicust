package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthIPBlacklistHandler struct {
	guard *service.AuthIPBlacklistService
}

func NewAuthIPBlacklistHandler(guard *service.AuthIPBlacklistService) *AuthIPBlacklistHandler {
	return &AuthIPBlacklistHandler{guard: guard}
}

func (handler *AuthIPBlacklistHandler) Config(ctx *gin.Context) {
	config, err := handler.guard.Config(ctx.Request.Context())
	if response.ErrorFrom(ctx, err) {
		return
	}
	response.Success(ctx, config)
}

func (handler *AuthIPBlacklistHandler) UpdateConfig(ctx *gin.Context) {
	ctx.Set("audit_action", "admin.ip_blacklist.config.update")
	var config service.AuthIPBlacklistConfig
	if err := ctx.ShouldBindJSON(&config); err != nil {
		response.BadRequest(ctx, "规则参数无效")
		return
	}
	updated, err := handler.guard.UpdateConfig(ctx.Request.Context(), config)
	if response.ErrorFrom(ctx, err) {
		return
	}
	ctx.Set("audit_extra", map[string]any{"enabled": updated.Enabled, "threshold": updated.Threshold,
		"window_seconds": updated.WindowSeconds, "block_seconds": updated.BlockSeconds, "block_mode": updated.BlockMode, "config_version": updated.Version})
	response.Success(ctx, updated)
}

func (handler *AuthIPBlacklistHandler) List(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		response.BadRequest(ctx, "分页参数无效")
		return
	}
	pageSize, err := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	if err != nil {
		response.BadRequest(ctx, "分页参数无效")
		return
	}
	items, total, err := handler.guard.List(ctx.Request.Context(), strings.TrimSpace(ctx.Query("ip")), page, pageSize)
	if response.ErrorFrom(ctx, err) {
		return
	}
	response.Paginated(ctx, items, total, page, pageSize)
}

func (handler *AuthIPBlacklistHandler) Unblock(ctx *gin.Context) {
	ctx.Set("audit_action", "admin.ip_blacklist.unblock")
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "记录编号无效")
		return
	}
	var request struct {
		Generation int64 `json:"generation"`
	}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.BadRequest(ctx, "解封参数无效")
		return
	}
	entry, err := handler.guard.Unblock(ctx.Request.Context(), id, request.Generation)
	if response.ErrorFrom(ctx, err) {
		return
	}
	ctx.Set("audit_extra", map[string]any{"blacklist_id": entry.ID, "target_ip": entry.IP, "generation": entry.Generation})
	response.Success(ctx, entry)
}

func (handler *AuthIPBlacklistHandler) ManualBlock(ctx *gin.Context) {
	ctx.Set("audit_action", "admin.ip_blacklist.manual_block")
	var request service.AuthIPManualBlock
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.BadRequest(ctx, "手动封禁参数无效")
		return
	}
	entry, err := handler.guard.ManualBlock(ctx.Request.Context(), request)
	if response.ErrorFrom(ctx, err) {
		return
	}
	ctx.Set("audit_extra", map[string]any{"blacklist_id": entry.ID, "target_ip": entry.IP,
		"generation": entry.Generation, "source": entry.Source, "reason": entry.Reason,
		"block_mode": request.BlockMode, "blocked_until": entry.BlockedUntil})
	response.Success(ctx, entry)
}
