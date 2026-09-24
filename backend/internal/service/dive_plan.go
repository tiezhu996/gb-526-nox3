package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"commercial-diving-decompression-control/backend/internal/audit"
	"commercial-diving-decompression-control/backend/internal/constants"
	"commercial-diving-decompression-control/backend/internal/decompression"
	"commercial-diving-decompression-control/backend/internal/dto"
	"commercial-diving-decompression-control/backend/internal/model"
	"commercial-diving-decompression-control/backend/internal/repository"
	"commercial-diving-decompression-control/backend/internal/util"
)

type DivePlanService struct {
	plans    *repository.DivePlanRepository
	profiles *repository.DiverProfileRepository
}

func NewDivePlanService(plans *repository.DivePlanRepository, profiles *repository.DiverProfileRepository) *DivePlanService {
	return &DivePlanService{plans: plans, profiles: profiles}
}

func (s *DivePlanService) List(ctx context.Context, search, status, profileRaw string, page, size int) ([]dto.DivePlanResponse, int64, error) {
	profileID := uint(0)
	if profileRaw != "" {
		parsed, err := strconv.ParseUint(profileRaw, 10, 64)
		if err != nil {
			return nil, 0, util.BadRequest("INVALID_PROFILE_ID", "diver_profile_id must be an integer", err)
		}
		profileID = uint(parsed)
	}
	items, total, err := s.plans.List(ctx, search, status, profileID, page, size)
	if err != nil {
		return nil, 0, err
	}
	responses := make([]dto.DivePlanResponse, 0, len(items))
	profileCodes := map[uint]string{}
	for _, item := range items {
		code, exists := profileCodes[item.DiverProfileID]
		if !exists {
			profile, profileErr := s.profiles.Get(ctx, item.DiverProfileID)
			if profileErr != nil {
				return nil, 0, profileErr
			}
			code = profile.ProfileCode
			profileCodes[item.DiverProfileID] = code
		}
		response, decodeErr := dto.NewDivePlanResponse(item, code)
		if decodeErr != nil {
			return nil, 0, decodeErr
		}
		responses = append(responses, response)
	}
	return responses, total, nil
}

func (s *DivePlanService) Get(ctx context.Context, id uint) (dto.DivePlanResponse, error) {
	item, err := s.plans.Get(ctx, id)
	if err != nil {
		return dto.DivePlanResponse{}, err
	}
	profile, err := s.profiles.Get(ctx, item.DiverProfileID)
	if err != nil {
		return dto.DivePlanResponse{}, err
	}
	return dto.NewDivePlanResponse(item, profile.ProfileCode)
}

func (s *DivePlanService) Create(ctx context.Context, req dto.CreateDivePlanRequest, actor audit.Entry) (dto.DivePlanResponse, error) {
	if err := req.ValidateBusiness(); err != nil {
		return dto.DivePlanResponse{}, util.Unprocessable("INVALID_PLAN", err.Error(), err)
	}
	profile, err := s.profiles.Get(ctx, req.DiverProfileID)
	if err != nil {
		return dto.DivePlanResponse{}, err
	}
	if profile.ProfileStatus != "active" {
		return dto.DivePlanResponse{}, util.Unprocessable("PROFILE_NOT_ACTIVE", "only an active training profile can be linked to a new plan", nil)
	}
	mixJSON, err := decompression.EncodeGasMix(req.BreathingMix)
	if err != nil {
		return dto.DivePlanResponse{}, util.Unprocessable("INVALID_GAS_MIX", err.Error(), err)
	}
	item := model.DivePlan{
		PlanCode: strings.ToUpper(strings.TrimSpace(req.PlanCode)), DiverProfileID: req.DiverProfileID,
		WorksitePressureBar: req.WorksitePressureBar, BreathingMixJSON: mixJSON,
		PlanStatus: constants.PlanDraft, CreatedBy: actor.ActorID, Version: 1, PlannedAt: req.PlannedAt.UTC(),
	}
	actor.Action = "dive_plan.create"
	actor.EntityType = "dive_plan"
	actor.AfterSummary = fmt.Sprintf("draft profile=%d worksite_pressure=%.2f gas=%s", req.DiverProfileID, req.WorksitePressureBar, mixJSON)
	if err := s.plans.Create(ctx, &item, actor); err != nil {
		return dto.DivePlanResponse{}, err
	}
	return dto.NewDivePlanResponse(item, profile.ProfileCode)
}

func (s *DivePlanService) Archive(ctx context.Context, id uint, req dto.TransitionPlanRequest, actor audit.Entry) (dto.DivePlanResponse, error) {
	if req.TargetStatus != constants.PlanArchived {
		return dto.DivePlanResponse{}, util.Unprocessable("INVALID_PLAN_TRANSITION", "archive endpoint only accepts archived target_status", nil)
	}
	current, err := s.plans.Get(ctx, id)
	if err != nil {
		return dto.DivePlanResponse{}, err
	}
	if current.Version != req.Version {
		return dto.DivePlanResponse{}, util.Conflict("PLAN_VERSION_CONFLICT", "dive plan was changed by another user", nil)
	}
	if !constants.CanTransitionPlan(current.PlanStatus, req.TargetStatus) {
		return dto.DivePlanResponse{}, util.Unprocessable("INVALID_PLAN_TRANSITION", fmt.Sprintf("cannot transition from %s to %s", current.PlanStatus, req.TargetStatus), nil)
	}
	actor.Action = "dive_plan.archive"
	actor.EntityType = "dive_plan"
	actor.BeforeSummary = string(current.PlanStatus)
	actor.AfterSummary = fmt.Sprintf("%s reason=%s", req.TargetStatus, strings.TrimSpace(req.Reason))
	if err := s.plans.Transition(ctx, current, req.TargetStatus, &actor.ActorID, actor); err != nil {
		return dto.DivePlanResponse{}, err
	}
	return s.Get(ctx, id)
}
