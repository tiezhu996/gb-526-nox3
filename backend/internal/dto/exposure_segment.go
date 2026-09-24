package dto

import (
	"fmt"
	"time"

	"commercial-diving-decompression-control/backend/internal/decompression"
	"commercial-diving-decompression-control/backend/internal/model"
)

type CreateExposureSegmentRequest struct {
	PlanVersion    uint                 `json:"plan_version" binding:"required,min=1"`
	SequenceNo     int                  `json:"sequence_no" binding:"required,min=1"`
	DepthM         float64              `json:"depth_m" binding:"gte=0,lte=120"`
	DurationMin    float64              `json:"duration_min" binding:"required,gt=0,lte=240"`
	AscentRateMMin float64              `json:"ascent_rate_mmin" binding:"gte=0,lte=18"`
	GasMix         decompression.GasMix `json:"gas_mix" binding:"required"`
	SegmentType    string               `json:"segment_type" binding:"required,oneof=descent bottom transit ascent surface"`
	Notes          string               `json:"notes" binding:"max=500"`
}

type UpdateExposureSegmentRequest struct {
	PlanVersion    uint                 `json:"plan_version" binding:"required,min=1"`
	DepthM         float64              `json:"depth_m" binding:"gte=0,lte=120"`
	DurationMin    float64              `json:"duration_min" binding:"required,gt=0,lte=240"`
	AscentRateMMin float64              `json:"ascent_rate_mmin" binding:"gte=0,lte=18"`
	GasMix         decompression.GasMix `json:"gas_mix" binding:"required"`
	SegmentType    string               `json:"segment_type" binding:"required,oneof=descent bottom transit ascent surface"`
	Notes          string               `json:"notes" binding:"max=500"`
}

type ReorderExposureSegmentsRequest struct {
	OrderedIDs []uint `json:"ordered_ids" binding:"required,min=1,dive,min=1"`
	Version    uint   `json:"version" binding:"required,min=1"`
}

type ExposureSegmentResponse struct {
	ID             uint                 `json:"id"`
	PlanID         uint                 `json:"plan_id"`
	SequenceNo     int                  `json:"sequence_no"`
	DepthM         float64              `json:"depth_m"`
	DurationMin    float64              `json:"duration_min"`
	AscentRateMMin float64              `json:"ascent_rate_mmin"`
	GasMix         decompression.GasMix `json:"gas_mix"`
	SegmentType    string               `json:"segment_type"`
	Notes          string               `json:"notes"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

func (r CreateExposureSegmentRequest) ValidateBusiness() error {
	if err := r.GasMix.Validate(); err != nil {
		return fmt.Errorf("gas_mix: %w", err)
	}
	if r.SegmentType == "surface" && r.DepthM != 0 {
		return fmt.Errorf("surface segment depth must be 0")
	}
	if r.SegmentType == "ascent" && r.AscentRateMMin <= 0 {
		return fmt.Errorf("ascent segment requires a positive ascent_rate_mmin")
	}
	return nil
}

func (r UpdateExposureSegmentRequest) ValidateBusiness() error {
	create := CreateExposureSegmentRequest{PlanVersion: r.PlanVersion, DepthM: r.DepthM, DurationMin: r.DurationMin, AscentRateMMin: r.AscentRateMMin, GasMix: r.GasMix, SegmentType: r.SegmentType, Notes: r.Notes, SequenceNo: 1}
	return create.ValidateBusiness()
}

func NewExposureSegmentResponse(item model.ExposureSegment) (ExposureSegmentResponse, error) {
	mix, err := decompression.DecodeGasMix(item.GasMixJSON)
	if err != nil {
		return ExposureSegmentResponse{}, fmt.Errorf("decode segment %d gas mix: %w", item.ID, err)
	}
	return ExposureSegmentResponse{ID: item.ID, PlanID: item.PlanID, SequenceNo: item.SequenceNo, DepthM: item.DepthM, DurationMin: item.DurationMin, AscentRateMMin: item.AscentRateMMin, GasMix: mix, SegmentType: item.SegmentType, Notes: item.Notes, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}, nil
}
