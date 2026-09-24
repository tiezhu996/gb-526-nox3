package middleware

import (
	"strings"

	"commercial-diving-decompression-control/backend/internal/auth"
	"commercial-diving-decompression-control/backend/internal/util"
	"github.com/gin-gonic/gin"
)

func Auth(service *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			c.Abort()
			util.Fail(c, util.Unauthorized("bearer token is required"))
			return
		}
		claims, err := service.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			c.Abort()
			util.Fail(c, err)
			return
		}
		// Re-read the account so deactivation and role changes take effect
		// immediately instead of waiting for the JWT to expire.
		user, err := service.CurrentPrincipal(c.Request.Context(), claims.UserID)
		if err != nil {
			c.Abort()
			util.Fail(c, err)
			return
		}
		c.Set(util.ContextUserID, user.ID)
		c.Set(util.ContextUsername, user.Username)
		c.Set(util.ContextRole, user.Role)
		c.Next()
	}
}
