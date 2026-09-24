package handler

import (
	"commercial-diving-decompression-control/backend/internal/audit"
	"commercial-diving-decompression-control/backend/internal/dto"
	"commercial-diving-decompression-control/backend/internal/service"
	"commercial-diving-decompression-control/backend/internal/util"
	"github.com/gin-gonic/gin"
)

type DiverProfileHandler struct{ service *service.DiverProfileService }

func NewDiverProfileHandler(service *service.DiverProfileService) *DiverProfileHandler {
	return &DiverProfileHandler{service: service}
}

func auditActor(c *gin.Context) audit.Entry {
	return audit.Entry{RequestID: util.RequestID(c), ActorID: util.UserID(c), ActorUsername: util.Username(c)}
}

func (h *DiverProfileHandler) List(c *gin.Context) {
	page, size := util.QueryPage(c)
	items, total, err := h.service.List(c.Request.Context(), util.SearchTerm(c), c.Query("status"), page, size)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, util.Page{Items: items, Total: total, Page: page, Size: size})
}

func (h *DiverProfileHandler) Get(c *gin.Context) {
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

func (h *DiverProfileHandler) Create(c *gin.Context) {
	var req dto.CreateDiverProfileRequest
	if !util.BindJSON(c, &req) {
		return
	}
	item, err := h.service.Create(c.Request.Context(), req, auditActor(c))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.Created(c, item)
}

func (h *DiverProfileHandler) Update(c *gin.Context) {
	id, ok := util.ParamID(c)
	if !ok {
		return
	}
	var req dto.UpdateDiverProfileRequest
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

func (h *DiverProfileHandler) Plans(c *gin.Context) {
	id, ok := util.ParamID(c)
	if !ok {
		return
	}
	page, size := util.QueryPage(c)
	items, total, err := h.service.Plans(c.Request.Context(), id, page, size)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, util.Page{Items: items, Total: total, Page: page, Size: size})
}
