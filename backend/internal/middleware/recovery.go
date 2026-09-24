package middleware

import (
	"fmt"
	"log/slog"
	"runtime/debug"

	"commercial-diving-decompression-control/backend/internal/util"
	"github.com/gin-gonic/gin"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("request panic recovered", "request_id", util.RequestID(c), "error", fmt.Sprint(recovered), "stack", string(debug.Stack()))
				c.Abort()
				util.Fail(c, util.Internal(fmt.Errorf("panic recovered: %v", recovered)))
			}
		}()
		c.Next()
	}
}
