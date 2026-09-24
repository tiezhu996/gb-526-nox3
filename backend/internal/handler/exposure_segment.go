package handler

import (
	"commercial-diving-decompression-control/backend/internal/dto"
	"commercial-diving-decompression-control/backend/internal/service"
	"commercial-diving-decompression-control/backend/internal/util"
	"github.com/gin-gonic/gin"
)

type ExposureSegmentHandler struct {
	service *service.ExposureSegmentService
}

func NewExposureSegmentHandler(service *service.ExposureSegmentService) *ExposureSegmentHandler {
	return &ExposureSegmentHandler{service: service}
}

func (h *ExposureSegmentHandler) ListByPlan(c *gin.Context) {
	planID, ok := util.ParamID(c)
	if !ok {
		return
	}
	items, err := h.service.ListByPlan(c.Request.Context(), planID)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, items)
}

func (h *ExposureSegmentHandler) Get(c *gin.Context) {
	id, ok := util.ParamID(c)
	if !ok {
		return
	}
	item, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, item)
}

func (h *ExposureSegmentHandler) Create(c *gin.Context) {
	planID, ok := util.ParamID(c)
	if !ok {
		return
	}
	var req dto.CreateExposureSegmentRequest
	if !util.BindJSON(c, &req) {
		return
	}
	item, err := h.service.Create(c.Request.Context(), planID, req, auditActor(c))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.Created(c, item)
}

func (h *ExposureSegmentHandler) Update(c *gin.Context) {
	id, ok := util.ParamID(c)
	if !ok {
		return
	}
	var req dto.UpdateExposureSegmentRequest
	if !util.BindJSON(c, &req) {
		return
	}
	item, err := h.service.Update(c.Request.Context(), id, req, auditActor(c))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, item)
}

func (h *ExposureSegmentHandler) Reorder(c *gin.Context) {
	planID, ok := util.ParamID(c)
	if !ok {
		return
	}
	var req dto.ReorderExposureSegmentsRequest
	if !util.BindJSON(c, &req) {
		return
	}
	items, err := h.service.Reorder(c.Request.Context(), planID, req, auditActor(c))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, items)
}
