package routes

import (
	"log"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RegisterAuthIPBlacklistRoutes(v1 *gin.RouterGroup, handlers *handler.Handlers,
	adminAuth middleware.AdminAuthMiddleware, auditLog middleware.AuditLogMiddleware,
	settings *service.SettingService, limiter *middleware.PanelRateLimiter, redisClient *redis.Client,
) gin.HandlerFunc {
	db, err := repository.ProvideSQLDB(handlers.Auth.AuthIPBlacklistEntClient())
	if err != nil {
		log.Printf("[AuthIPBlacklist] 数据库初始化失败: %v", err)
	}
	guard := service.NewAuthIPBlacklistService(repository.NewAuthIPBlacklistRepository(db), repository.NewAuthIPWindowCounter(redisClient))
	RegisterAuthIPBlacklistAdminRoutes(v1, admin.NewAuthIPBlacklistHandler(guard), adminAuth, auditLog, settings, limiter)
	return middleware.NewAuthIPBlacklistGuard(guard)
}

func RegisterAuthIPBlacklistAdminRoutes(v1 *gin.RouterGroup, handler *admin.AuthIPBlacklistHandler,
	adminAuth middleware.AdminAuthMiddleware, auditLog middleware.AuditLogMiddleware,
	settings *service.SettingService, limiter *middleware.PanelRateLimiter,
) {
	group := v1.Group("/admin/ip-blacklist")
	group.Use(gin.HandlerFunc(adminAuth), limiter.Global(), gin.HandlerFunc(auditLog), middleware.AdminComplianceGuard(settings))
	group.GET("/config", handler.Config)
	group.PUT("/config", handler.UpdateConfig)
	group.GET("/entries", handler.List)
	group.POST("/entries/:id/unblock", handler.Unblock)
}
