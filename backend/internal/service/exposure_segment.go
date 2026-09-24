package service

import (
	"context"
	"fmt"
	"strings"

	"commercial-diving-decompression-control/backend/internal/audit"
	"commercial-diving-decompression-control/backend/internal/decompression"
	"commercial-diving-decompression-control/backend/internal/dto"
	"commercial-diving-decompression-control/backend/internal/model"
	"commercial-diving-decompression-control/backend/internal/repository"
	"commercial-diving-decompression-control/backend/internal/util"
)

type ExposureSegmentService struct {
	segments *repository.ExposureSegmentRepository
	plans    *repository.DivePlanRepository
	max      int
}

func NewExposureSegmentService(segments *repository.ExposureSegmentRepository, plans *repository.DivePlanRepository, maxSegments int) *ExposureSegmentService {
	return &ExposureSegmentService{segments: segments, plans: plans, max: maxSegments}
}

func (s *ExposureSegmentService) ListByPlan(ctx context.Context, planID uint) ([]dto.ExposureSegmentResponse, error) {
	if _, err := s.plans.Get(ctx, planID); err != nil {
		return nil, err
	}
	items, err := s.segments.ListByPlan(ctx, planID)
	if err != nil {
		return nil, err
	}
	responses := make([]dto.ExposureSegmentResponse, 0, len(items))
	for _, item := range items {
		response, decodeErr := dto.NewExposureSegmentResponse(item)
		if decodeErr != nil {
			return nil, decodeErr
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func (s *ExposureSegmentService) Get(ctx context.Context, id uint) (dto.ExposureSegmentResponse, error) {
	item, err := s.segments.Get(ctx, id)
	if err != nil {
		return dto.ExposureSegmentResponse{}, err
	}
	return dto.NewExposureSegmentResponse(item)
}

func (s *ExposureSegmentService) Create(ctx context.Context, planID uint, req dto.CreateExposureSegmentRequest, actor audit.Entry) (dto.ExposureSegmentResponse, error) {
	if err := req.ValidateBusiness(); err != nil {
		return dto.ExposureSegmentResponse{}, util.Unprocessable("INVALID_SEGMENT", err.Error(), err)
	}
	current, err := s.segments.ListByPlan(ctx, planID)
	if err != nil {
		return dto.ExposureSegmentResponse{}, err
	}
	if len(current) >= s.max {
		return dto.ExposureSegmentResponse{}, util.Unprocessable("SEGMENT_LIMIT", fmt.Sprintf("plan cannot exceed %d segments", s.max), nil)
	}
	mixJSON, err := decompression.EncodeGasMix(req.GasMix)
	if err != nil {
		return dto.ExposureSegmentResponse{}, util.Unprocessable("INVALID_GAS_MIX", err.Error(), err)
	}
	item := model.ExposureSegment{PlanID: planID, SequenceNo: req.SequenceNo, DepthM: req.DepthM, DurationMin: req.DurationMin, AscentRateMMin: req.AscentRateMMin, GasMixJSON: mixJSON, SegmentType: req.SegmentType, Notes: strings.TrimSpace(req.Notes)}
	actor.Action = "exposure_segment.create"
	actor.EntityType = "exposure_segment"
	actor.AfterSummary = fmt.Sprintf("plan=%d sequence=%d type=%s depth=%.1f duration=%.1f gas=%s", planID, req.SequenceNo, req.SegmentType, req.DepthM, req.DurationMin, mixJSON)
	if err := s.segments.Create(ctx, &item, req.PlanVersion, actor); err != nil {
		return dto.ExposureSegmentResponse{}, err
	}
	return dto.NewExposureSegmentResponse(item)
}

func (s *ExposureSegmentService) Update(ctx context.Context, id uint, req dto.UpdateExposureSegmentRequest, actor audit.Entry) (dto.ExposureSegmentResponse, error) {
	if err := req.ValidateBusiness(); err != nil {
		return dto.ExposureSegmentResponse{}, util.Unprocessable("INVALID_SEGMENT", err.Error(), err)
	}
	current, err := s.segments.Get(ctx, id)
	if err != nil {
		return dto.ExposureSegmentResponse{}, err
	}
	mixJSON, err := decompression.EncodeGasMix(req.GasMix)
	if err != nil {
		return dto.ExposureSegmentResponse{}, util.Unprocessable("INVALID_GAS_MIX", err.Error(), err)
	}
	changes := map[string]any{"depth_m": req.DepthM, "duration_min": req.DurationMin, "ascent_rate_mmin": req.AscentRateMMin, "gas_mix_json": mixJSON, "segment_type": req.SegmentType, "notes": strings.TrimSpace(req.Notes)}
	actor.Action = "exposure_segment.update"
	actor.EntityType = "exposure_segment"
	actor.BeforeSummary = fmt.Sprintf("plan=%d sequence=%d type=%s depth=%.1f duration=%.1f", current.PlanID, current.SequenceNo, current.SegmentType, current.DepthM, current.DurationMin)
	actor.AfterSummary = fmt.Sprintf("type=%s depth=%.1f duration=%.1f gas=%s", req.SegmentType, req.DepthM, req.DurationMin, mixJSON)
	if err := s.segments.Update(ctx, current, req.PlanVersion, changes, actor); err != nil {
		return dto.ExposureSegmentResponse{}, err
	}
	return s.Get(ctx, id)
}

func (s *ExposureSegmentService) Reorder(ctx context.Context, planID uint, req dto.ReorderExposureSegmentsRequest, actor audit.Entry) ([]dto.ExposureSegmentResponse, error) {
	actor.Action = "exposure_segment.reorder"
	actor.EntityType = "dive_plan"
	actor.BeforeSummary = fmt.Sprintf("plan=%d input_version=%d", planID, req.Version)
	actor.AfterSummary = fmt.Sprintf("ordered_ids=%v", req.OrderedIDs)
	if err := s.segments.Reorder(ctx, planID, req.Version, req.OrderedIDs, actor); err != nil {
		return nil, err
	}
	return s.ListByPlan(ctx, planID)
}
