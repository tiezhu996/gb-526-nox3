package service

import (
	"context"
	"fmt"
	"strings"

	"commercial-diving-decompression-control/backend/internal/audit"
	"commercial-diving-decompression-control/backend/internal/dto"
	"commercial-diving-decompression-control/backend/internal/model"
	"commercial-diving-decompression-control/backend/internal/repository"
	"commercial-diving-decompression-control/backend/internal/util"
)

type DiverProfileService struct {
	profiles *repository.DiverProfileRepository
	plans    *repository.DivePlanRepository
}

func NewDiverProfileService(profiles *repository.DiverProfileRepository, plans *repository.DivePlanRepository) *DiverProfileService {
	return &DiverProfileService{profiles: profiles, plans: plans}
}

func (s *DiverProfileService) List(ctx context.Context, search, status string, page, size int) ([]dto.DiverProfileResponse, int64, error) {
	items, total, err := s.profiles.List(ctx, search, status, page, size)
	if err != nil {
		return nil, 0, err
	}
	responses := make([]dto.DiverProfileResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, dto.NewDiverProfileResponse(item))
	}
	return responses, total, nil
}

func (s *DiverProfileService) Get(ctx context.Context, id uint) (dto.DiverProfileResponse, error) {
	item, err := s.profiles.Get(ctx, id)
	if err != nil {
		return dto.DiverProfileResponse{}, err
	}
	return dto.NewDiverProfileResponse(item), nil
}

func (s *DiverProfileService) Create(ctx context.Context, req dto.CreateDiverProfileRequest, actor audit.Entry) (dto.DiverProfileResponse, error) {
	if err := req.ValidateBusiness(); err != nil {
		return dto.DiverProfileResponse{}, util.Unprocessable("INVALID_PROFILE", err.Error(), err)
	}
	item := model.DiverProfile{
		ProfileCode: strings.ToUpper(strings.TrimSpace(req.ProfileCode)), DisplayName: strings.TrimSpace(req.DisplayName),
		QualificationLevel: req.QualificationLevel, DefaultO2Fraction: req.DefaultO2Fraction, DefaultHeFraction: req.DefaultHeFraction,
		ProfileStatus: req.ProfileStatus, LimitsNote: strings.TrimSpace(req.LimitsNote), Version: 1,
	}
	actor.Action = "diver_profile.create"
	actor.EntityType = "diver_profile"
	actor.AfterSummary = fmt.Sprintf("code=%s qualification=%s status=%s gas=o2:%.3f he:%.3f", item.ProfileCode, item.QualificationLevel, item.ProfileStatus, item.DefaultO2Fraction, item.DefaultHeFraction)
	if err := s.profiles.Create(ctx, &item, actor); err != nil {
		return dto.DiverProfileResponse{}, err
	}
	return dto.NewDiverProfileResponse(item), nil
}

func (s *DiverProfileService) Update(ctx context.Context, id uint, req dto.UpdateDiverProfileRequest, actor audit.Entry) (dto.DiverProfileResponse, error) {
	if err := req.ValidateBusiness(); err != nil {
		return dto.DiverProfileResponse{}, util.Unprocessable("INVALID_PROFILE", err.Error(), err)
	}
	current, err := s.profiles.Get(ctx, id)
	if err != nil {
		return dto.DiverProfileResponse{}, err
	}
	if current.Version != req.Version {
		return dto.DiverProfileResponse{}, util.Conflict("PROFILE_VERSION_CONFLICT", "diver profile was changed by another user", nil)
	}
	changes := map[string]any{
		"display_name": strings.TrimSpace(req.DisplayName), "qualification_level": req.QualificationLevel,
		"default_o2_fraction": req.DefaultO2Fraction, "default_he_fraction": req.DefaultHeFraction,
		"profile_status": req.ProfileStatus, "limits_note": strings.TrimSpace(req.LimitsNote),
	}
	actor.Action = "diver_profile.update"
	actor.EntityType = "diver_profile"
	actor.BeforeSummary = fmt.Sprintf("v%d status=%s qualification=%s", current.Version, current.ProfileStatus, current.QualificationLevel)
	actor.AfterSummary = fmt.Sprintf("v%d status=%s qualification=%s", current.Version+1, req.ProfileStatus, req.QualificationLevel)
	if err := s.profiles.Update(ctx, current, changes, actor); err != nil {
		return dto.DiverProfileResponse{}, err
	}
	return s.Get(ctx, id)
}

func (s *DiverProfileService) Plans(ctx context.Context, id uint, page, size int) ([]dto.DivePlanResponse, int64, error) {
	profile, err := s.profiles.Get(ctx, id)
	if err != nil {
		return nil, 0, err
	}
	items, total, err := s.plans.List(ctx, "", "", id, page, size)
	if err != nil {
		return nil, 0, err
	}
	responses := make([]dto.DivePlanResponse, 0, len(items))
	for _, item := range items {
		response, decodeErr := dto.NewDivePlanResponse(item, profile.ProfileCode)
		if decodeErr != nil {
			return nil, 0, decodeErr
		}
		responses = append(responses, response)
	}
	return responses, total, nil
}
