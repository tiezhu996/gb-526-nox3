package dto

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"commercial-diving-decompression-control/backend/internal/constants"
	"commercial-diving-decompression-control/backend/internal/decompression"
	"commercial-diving-decompression-control/backend/internal/model"
)

var planCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9-]{2,39}$`)

type CreateDivePlanRequest struct {
	PlanCode            string               `json:"plan_code" binding:"required,min=3,max=40"`
	DiverProfileID      uint                 `json:"diver_profile_id" binding:"required,min=1"`
	WorksitePressureBar float64              `json:"worksite_pressure_bar" binding:"required,gte=0.7,lte=1.5"`
	BreathingMix        decompression.GasMix `json:"breathing_mix" binding:"required"`
	PlannedAt           time.Time            `json:"planned_at" binding:"required"`
}

type DivePlanResponse struct {
	ID                  uint                 `json:"id"`
	PlanCode            string               `json:"plan_code"`
	DiverProfileID      uint                 `json:"diver_profile_id"`
	DiverProfileCode    string               `json:"diver_profile_code,omitempty"`
	WorksitePressureBar float64              `json:"worksite_pressure_bar"`
	BreathingMix        decompression.GasMix `json:"breathing_mix"`
	PlanStatus          constants.PlanStatus `json:"plan_status"`
	CreatedBy           uint                 `json:"created_by"`
	ReviewedBy          *uint                `json:"reviewed_by"`
	Version             uint                 `json:"version"`
	PlannedAt           time.Time            `json:"planned_at"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

type TransitionPlanRequest struct {
	TargetStatus constants.PlanStatus `json:"target_status" binding:"required"`
	Version      uint                 `json:"version" binding:"required,min=1"`
	Reason       string               `json:"reason" binding:"required,min=3,max=300"`
}

func (r CreateDivePlanRequest) ValidateBusiness() error {
	code := strings.ToUpper(strings.TrimSpace(r.PlanCode))
	if !planCodePattern.MatchString(code) {
		return fmt.Errorf("plan_code must contain 3-40 uppercase letters, digits, or hyphens")
	}
	if r.PlannedAt.IsZero() {
		return fmt.Errorf("planned_at is required")
	}
	if err := r.BreathingMix.Validate(); err != nil {
		return fmt.Errorf("breathing_mix: %w", err)
	}
	return nil
}

func NewDivePlanResponse(item model.DivePlan, profileCode string) (DivePlanResponse, error) {
	mix, err := decompression.DecodeGasMix(item.BreathingMixJSON)
	if err != nil {
		return DivePlanResponse{}, fmt.Errorf("decode plan %d breathing mix: %w", item.ID, err)
	}
	return DivePlanResponse{
		ID: item.ID, PlanCode: item.PlanCode, DiverProfileID: item.DiverProfileID, DiverProfileCode: profileCode,
		WorksitePressureBar: item.WorksitePressureBar, BreathingMix: mix, PlanStatus: item.PlanStatus,
		CreatedBy: item.CreatedBy, ReviewedBy: item.ReviewedBy, Version: item.Version,
		PlannedAt: item.PlannedAt, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}, nil
}
