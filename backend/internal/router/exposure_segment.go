package router

import (
	"commercial-diving-decompression-control/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterExposureSegmentRoutes(api *gin.RouterGroup, h *handler.ExposureSegmentHandler, write gin.HandlerFunc) {
	api.GET("/plans/:id/segments", h.ListByPlan)
	api.POST("/plans/:id/segments", write, h.Create)
	api.PUT("/plans/:id/segments/order", write, h.Reorder)
	segments := api.Group("/segments")
	segments.GET("/:id", h.Get)
	segments.PUT("/:id", write, h.Update)
}
