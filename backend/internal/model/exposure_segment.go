package model

import "time"

type ExposureSegment struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PlanID         uint      `gorm:"not null;uniqueIndex:idx_plan_sequence;index" json:"plan_id"`
	SequenceNo     int       `gorm:"not null;uniqueIndex:idx_plan_sequence" json:"sequence_no"`
	DepthM         float64   `gorm:"not null" json:"depth_m"`
	DurationMin    float64   `gorm:"not null" json:"duration_min"`
	AscentRateMMin float64   `gorm:"not null;default:0" json:"ascent_rate_mmin"`
	GasMixJSON     string    `gorm:"type:text;not null" json:"gas_mix_json"`
	SegmentType    string    `gorm:"size:24;not null;index" json:"segment_type"`
	Notes          string    `gorm:"size:500;not null;default:''" json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (ExposureSegment) TableName() string { return "exposure_segments" }
