package model

import "time"

type DiverProfile struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	ProfileCode        string    `gorm:"size:32;not null;uniqueIndex" json:"profile_code"`
	DisplayName        string    `gorm:"size:100;not null" json:"display_name"`
	QualificationLevel string    `gorm:"size:40;not null;index" json:"qualification_level"`
	DefaultO2Fraction  float64   `gorm:"not null" json:"default_o2_fraction"`
	DefaultHeFraction  float64   `gorm:"not null" json:"default_he_fraction"`
	ProfileStatus      string    `gorm:"size:24;not null;index" json:"profile_status"`
	LimitsNote         string    `gorm:"size:500;not null;default:''" json:"limits_note"`
	Version            uint      `gorm:"not null;default:1" json:"version"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (DiverProfile) TableName() string { return "diver_profiles" }

func (p DiverProfile) DefaultN2Fraction() float64 {
	return 1 - p.DefaultO2Fraction - p.DefaultHeFraction
}
