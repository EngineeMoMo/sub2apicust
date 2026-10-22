package middleware

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func RegistrationIPContext() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(service.WithRegistrationIP(ctx.Request.Context(), ip.GetTrustedClientIP(ctx)))
		ctx.Next()
	}
}
