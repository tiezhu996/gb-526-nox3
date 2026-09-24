package router

import (
	"commercial-diving-decompression-control/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterDiverProfileRoutes(api *gin.RouterGroup, h *handler.DiverProfileHandler, write gin.HandlerFunc) {
	divers := api.Group("/divers")
	divers.GET("", h.List)
	divers.GET("/:id", h.Get)
	divers.GET("/:id/plans", h.Plans)
	divers.POST("", write, h.Create)
	divers.PUT("/:id", write, h.Update)
}
