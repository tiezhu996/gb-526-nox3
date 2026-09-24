package dto

import (
	"fmt"
	"regexp"
	"strings"

	"commercial-diving-decompression-control/backend/internal/decompression"
	"commercial-diving-decompression-control/backend/internal/model"
)

var profileCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9-]{2,31}$`)

type CreateDiverProfileRequest struct {
	ProfileCode        string  `json:"profile_code" binding:"required,min=3,max=32"`
	DisplayName        string  `json:"display_name" binding:"required,min=2,max=100"`
	QualificationLevel string  `json:"qualification_level" binding:"required,oneof=trainee commercial advanced supervisor"`
	DefaultO2Fraction  float64 `json:"default_o2_fraction" binding:"required,gte=0,lte=1"`
	DefaultHeFraction  float64 `json:"default_he_fraction" binding:"gte=0,lte=1"`
	ProfileStatus      string  `json:"profile_status" binding:"required,oneof=active inactive training_hold"`
	LimitsNote         string  `json:"limits_note" binding:"max=500"`
}

type UpdateDiverProfileRequest struct {
	DisplayName        string  `json:"display_name" binding:"required,min=2,max=100"`
	QualificationLevel string  `json:"qualification_level" binding:"required,oneof=trainee commercial advanced supervisor"`
	DefaultO2Fraction  float64 `json:"default_o2_fraction" binding:"required,gte=0,lte=1"`
	DefaultHeFraction  float64 `json:"default_he_fraction" binding:"gte=0,lte=1"`
	ProfileStatus      string  `json:"profile_status" binding:"required,oneof=active inactive training_hold"`
	LimitsNote         string  `json:"limits_note" binding:"max=500"`
	Version            uint    `json:"version" binding:"required,min=1"`
}

type DiverProfileResponse struct {
	model.DiverProfile
	DefaultN2Fraction float64 `json:"default_n2_fraction"`
}

func (r CreateDiverProfileRequest) ValidateBusiness() error {
	code := strings.ToUpper(strings.TrimSpace(r.ProfileCode))
	if !profileCodePattern.MatchString(code) {
		return fmt.Errorf("profile_code must contain 3-32 uppercase letters, digits, or hyphens")
	}
	return validateDefaultMix(r.DefaultO2Fraction, r.DefaultHeFraction)
}

func (r UpdateDiverProfileRequest) ValidateBusiness() error {
	return validateDefaultMix(r.DefaultO2Fraction, r.DefaultHeFraction)
}

func validateDefaultMix(o2, he float64) error {
	n2 := 1 - o2 - he
	if n2 < 0 {
		return fmt.Errorf("default o2 and he fractions cannot exceed 1")
	}
	return decompression.GasMix{O2: o2, He: he, N2: n2}.Validate()
}

func NewDiverProfileResponse(item model.DiverProfile) DiverProfileResponse {
	return DiverProfileResponse{DiverProfile: item, DefaultN2Fraction: item.DefaultN2Fraction()}
}
