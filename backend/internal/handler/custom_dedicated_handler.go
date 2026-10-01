package handler

import (
	"errors"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type CustomDedicatedHandler struct {
	service *service.CustomDedicatedService
}

func NewCustomDedicatedHandler(dedicated *service.CustomDedicatedService) *CustomDedicatedHandler {
	return &CustomDedicatedHandler{service: dedicated}
}

func customDedicatedPage(c *gin.Context) (int, bool) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 || page > 1000000 {
		response.BadRequest(c, "无效页码")
		return 0, false
	}
	return page, true
}

func customDedicatedID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "无效的包号记录编号")
		return 0, false
	}
	return id, true
}

func customDedicatedError(c *gin.Context, err error) {
	for _, known := range []error{service.ErrDedicatedMode, service.ErrDedicatedInput, service.ErrDedicatedConfig, service.ErrDedicatedConflict, service.ErrDedicatedNotFound, service.ErrDedicatedAccess, service.ErrDedicatedDelete, service.ErrDedicatedStale, service.ErrDedicatedRestore, service.ErrDedicatedPolicy} {
		if errors.Is(err, known) {
			response.ErrorFrom(c, err)
			return
		}
	}
	response.InternalError(c, "包号操作失败，请稍后重试；若持续失败请联系管理员")
}

func (h *CustomDedicatedHandler) BillingPolicy(c *gin.Context) {
	id, valid := customDedicatedID(c)
	if !valid {
		return
	}
	var policy service.CustomDedicatedBillingPolicy
	var err error
	if c.Request.Method == "PUT" {
		if err := c.ShouldBindJSON(&policy); err != nil {
			response.BadRequest(c, "请检查包号使用限制格式")
			return
		}
		policy, err = h.service.UpdateBillingPolicy(c.Request.Context(), id, policy)
	} else {
		policy, err = h.service.BillingPolicy(c.Request.Context(), id)
	}
	if err != nil {
		customDedicatedError(c, err)
		return
	}
	response.Success(c, policy)
}

func (h *CustomDedicatedHandler) AdminList(c *gin.Context) {
	page, valid := customDedicatedPage(c)
	if !valid {
		return
	}
	bindings, err := h.service.AdminList(c.Request.Context(), page)
	if err != nil {
		customDedicatedError(c, err)
		return
	}
	response.Success(c, bindings)
}

func (h *CustomDedicatedHandler) Save(c *gin.Context) {
	id := int64(0)
	if c.Param("id") != "" {
		var valid bool
		id, valid = customDedicatedID(c)
		if !valid {
			return
		}
	}
	var input service.CustomDedicatedInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "请检查输入内容与到期时间格式")
		return
	}
	result, err := h.service.Save(c.Request.Context(), id, input)
	if err != nil {
		customDedicatedError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *CustomDedicatedHandler) Revoke(c *gin.Context) {
	id, valid := customDedicatedID(c)
	if !valid {
		return
	}
	if err := h.service.Revoke(c.Request.Context(), id); err != nil {
		customDedicatedError(c, err)
		return
	}
	response.Success(c, gin.H{"revoked": true})
}

func (h *CustomDedicatedHandler) Delete(c *gin.Context) {
	id, valid := customDedicatedID(c)
	if !valid {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		customDedicatedError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *CustomDedicatedHandler) UserList(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "请先登录")
		return
	}
	page, valid := customDedicatedPage(c)
	if !valid {
		return
	}
	views, err := h.service.UserList(c.Request.Context(), subject.UserID, page)
	if err == nil {
		err = h.service.FillGroupNames(c.Request.Context(), subject.UserID, views)
	}
	if err != nil {
		customDedicatedError(c, err)
		return
	}
	response.Success(c, views)
}

func (h *CustomDedicatedHandler) UserUsage(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "请先登录")
		return
	}
	id, valid := customDedicatedID(c)
	if !valid {
		return
	}
	view, err := h.service.View(c.Request.Context(), subject.UserID, id, true)
	if err != nil {
		customDedicatedError(c, err)
		return
	}
	views := []service.CustomDedicatedView{*view}
	if err := h.service.FillGroupNames(c.Request.Context(), subject.UserID, views); err != nil {
		customDedicatedError(c, err)
		return
	}
	response.Success(c, views[0])
}
