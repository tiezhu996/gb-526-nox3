package handler

import (
	"strconv"

	"commercial-diving-decompression-control/backend/internal/dto"
	"commercial-diving-decompression-control/backend/internal/service"
	"commercial-diving-decompression-control/backend/internal/util"
	"github.com/gin-gonic/gin"
)

type DecompressionAssessmentHandler struct {
	service *service.DecompressionAssessmentService
}

func NewDecompressionAssessmentHandler(service *service.DecompressionAssessmentService) *DecompressionAssessmentHandler {
	return &DecompressionAssessmentHandler{service: service}
}

func queryUint(c *gin.Context, name string) (uint, bool) {
	raw := c.Query(name)
	if raw == "" {
		return 0, true
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		util.Fail(c, util.BadRequest("INVALID_QUERY_ID", name+" must be a positive integer", err))
		return 0, false
	}
	return uint(value), true
}

func (h *DecompressionAssessmentHandler) List(c *gin.Context) {
	page, size := util.QueryPage(c)
	planID, ok := queryUint(c, "plan_id")
	if !ok {
		return
	}
	items, total, err := h.service.List(c.Request.Context(), planID, c.Query("status"), page, size)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, util.Page{Items: items, Total: total, Page: page, Size: size})
}

func (h *DecompressionAssessmentHandler) Get(c *gin.Context) {
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

func (h *DecompressionAssessmentHandler) Run(c *gin.Context) {
	planID, ok := util.ParamID(c)
	if !ok {
		return
	}
	var req dto.RunAssessmentRequest
	if !util.BindJSON(c, &req) {
		return
	}
	item, err := h.service.Run(c.Request.Context(), planID, req, auditActor(c))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.Created(c, item)
}

func (h *DecompressionAssessmentHandler) Submit(c *gin.Context) {
	id, ok := util.ParamID(c)
	if !ok {
		return
	}
	var req dto.TransitionPlanRequest
	if !util.BindJSON(c, &req) {
		return
	}
	item, err := h.service.Submit(c.Request.Context(), id, req, auditActor(c))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, item)
}

func (h *DecompressionAssessmentHandler) Approve(c *gin.Context) {
	id, ok := util.ParamID(c)
	if !ok {
		return
	}
	var req dto.TransitionPlanRequest
	if !util.BindJSON(c, &req) {
		return
	}
	item, err := h.service.Approve(c.Request.Context(), id, req, auditActor(c))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, item)
}

func (h *DecompressionAssessmentHandler) Compare(c *gin.Context) {
	leftID, ok := util.ParamID(c)
	if !ok {
		return
	}
	rightID, ok := queryUint(c, "other_id")
	if !ok {
		return
	}
	if rightID == 0 {
		util.Fail(c, util.BadRequest("OTHER_ID_REQUIRED", "other_id query parameter is required", nil))
		return
	}
	comparison, err := h.service.Compare(c.Request.Context(), leftID, rightID)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, comparison)
}
