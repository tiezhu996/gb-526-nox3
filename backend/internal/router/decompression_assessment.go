package router

import (
	"commercial-diving-decompression-control/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterDecompressionAssessmentRoutes(api *gin.RouterGroup, h *handler.DecompressionAssessmentHandler, write, review gin.HandlerFunc) {
	api.POST("/plans/:id/assessments/run", write, h.Run)
	assessments := api.Group("/assessments")
	assessments.GET("", h.List)
	assessments.GET("/:id", h.Get)
	assessments.GET("/:id/compare", h.Compare)
	assessments.POST("/:id/submit", write, h.Submit)
	assessments.POST("/:id/approve", review, h.Approve)
}
