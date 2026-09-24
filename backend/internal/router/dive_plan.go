package router

import (
	"commercial-diving-decompression-control/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterDivePlanRoutes(api *gin.RouterGroup, h *handler.DivePlanHandler, write, review gin.HandlerFunc) {
	plans := api.Group("/plans")
	plans.GET("", h.List)
	plans.GET("/:id", h.Get)
	plans.POST("", write, h.Create)
	plans.POST("/:id/archive", review, h.Archive)
}
