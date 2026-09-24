package handler

import (
	"commercial-diving-decompression-control/backend/internal/dto"
	"commercial-diving-decompression-control/backend/internal/service"
	"commercial-diving-decompression-control/backend/internal/util"
	"github.com/gin-gonic/gin"
)

type DivePlanHandler struct{ service *service.DivePlanService }

func NewDivePlanHandler(service *service.DivePlanService) *DivePlanHandler {
	return &DivePlanHandler{service: service}
}

func (h *DivePlanHandler) List(c *gin.Context) {
	page, size := util.QueryPage(c)
	items, total, err := h.service.List(c.Request.Context(), util.SearchTerm(c), c.Query("status"), c.Query("diver_profile_id"), page, size)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, util.Page{Items: items, Total: total, Page: page, Size: size})
}

func (h *DivePlanHandler) Get(c *gin.Context) {
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

func (h *DivePlanHandler) Create(c *gin.Context) {
	var req dto.CreateDivePlanRequest
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

func (h *DivePlanHandler) Archive(c *gin.Context) {
	id, ok := util.ParamID(c)
	if !ok {
		return
	}
	var req dto.TransitionPlanRequest
	if !util.BindJSON(c, &req) {
		return
	}
	item, err := h.service.Archive(c.Request.Context(), id, req, auditActor(c))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, item)
}
