package database

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"commercial-diving-decompression-control/backend/internal/audit"
	"commercial-diving-decompression-control/backend/internal/auth"
	"commercial-diving-decompression-control/backend/internal/config"
	"commercial-diving-decompression-control/backend/internal/constants"
	"commercial-diving-decompression-control/backend/internal/decompression"
	"commercial-diving-decompression-control/backend/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.DBDriver {
	case "sqlite":
		dialector = sqlite.Open(cfg.DBDSN)
	case "postgres":
		dialector = postgres.Open(cfg.DBDSN)
	default:
		return nil, fmt.Errorf("open database: unsupported driver %q", cfg.DBDriver)
	}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Warn), TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("open %s database: %w", cfg.DBDriver, err)
	}
	if cfg.DBAutoMigrate {
		if err := migrate(db); err != nil {
			return nil, err
		}
		if err := backfillRiskBands(db); err != nil {
			return nil, err
		}
		if err := seed(db, cfg); err != nil {
			return nil, err
		}
	}
	return db, nil
}

func backfillRiskBands(db *gorm.DB) error {
	var assessments []model.DecompressionAssessment
	if err := db.Find(&assessments).Error; err != nil {
		return fmt.Errorf("load assessments for risk-band backfill: %w", err)
	}
	for _, assessment := range assessments {
		var flags []decompression.RiskFlag
		if err := json.Unmarshal([]byte(assessment.RiskFlagsJSON), &flags); err != nil {
			return fmt.Errorf("decode assessment %d risk flags for backfill: %w", assessment.ID, err)
		}
		band := decompression.HighestRiskBand(flags)
		if assessment.HighestRiskBand == band {
			continue
		}
		if err := db.Model(&model.DecompressionAssessment{}).Where("id = ?", assessment.ID).Update("highest_risk_band", band).Error; err != nil {
			return fmt.Errorf("backfill assessment %d risk band: %w", assessment.ID, err)
		}
	}
	return nil
}

func migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&auth.User{}, &model.DiverProfile{}, &model.DivePlan{}, &model.ExposureSegment{}, &model.DecompressionAssessment{}, &audit.Event{}); err != nil {
		return fmt.Errorf("auto migrate database: %w", err)
	}
	return nil
}

func seed(db *gorm.DB, cfg config.Config) error {
	if err := auth.SeedUsers(db); err != nil {
		return err
	}
	var profileCount int64
	if err := db.Model(&model.DiverProfile{}).Count(&profileCount).Error; err != nil {
		return fmt.Errorf("count diver profiles: %w", err)
	}
	if profileCount > 0 {
		return nil
	}
	profiles := []model.DiverProfile{
		{ProfileCode: "TRN-ALPHA", DisplayName: "Training Profile Alpha", QualificationLevel: "commercial", DefaultO2Fraction: 0.21, DefaultHeFraction: 0, ProfileStatus: "active", LimitsNote: "Synthetic minimal training record; no medical data.", Version: 1},
		{ProfileCode: "TRN-BRAVO", DisplayName: "Training Profile Bravo", QualificationLevel: "advanced", DefaultO2Fraction: 0.18, DefaultHeFraction: 0.35, ProfileStatus: "active", LimitsNote: "Offline comparison profile; supervisor interpretation required.", Version: 1},
	}
	if err := db.Create(&profiles).Error; err != nil {
		return fmt.Errorf("seed diver profiles: %w", err)
	}
	var planner auth.User
	if err := db.Where("username = ?", "planner").First(&planner).Error; err != nil {
		return fmt.Errorf("find seed planner: %w", err)
	}
	air, err := decompression.EncodeGasMix(decompression.GasMix{O2: 0.21, N2: 0.79})
	if err != nil {
		return err
	}
	trimix, err := decompression.EncodeGasMix(decompression.GasMix{O2: 0.18, He: 0.35, N2: 0.47})
	if err != nil {
		return err
	}
	plans := []model.DivePlan{
		{PlanCode: "TRAIN-30A", DiverProfileID: profiles[0].ID, WorksitePressureBar: 1, BreathingMixJSON: air, PlanStatus: constants.PlanDraft, CreatedBy: planner.ID, Version: 1, PlannedAt: time.Now().UTC().Add(24 * time.Hour)},
		{PlanCode: "COMPARE-42B", DiverProfileID: profiles[1].ID, WorksitePressureBar: 1, BreathingMixJSON: trimix, PlanStatus: constants.PlanPendingReview, CreatedBy: planner.ID, Version: 3, PlannedAt: time.Now().UTC().Add(48 * time.Hour)},
	}
	if err := db.Create(&plans).Error; err != nil {
		return fmt.Errorf("seed dive plans: %w", err)
	}
	segments := []model.ExposureSegment{
		{PlanID: plans[0].ID, SequenceNo: 1, DepthM: 30, DurationMin: 3, GasMixJSON: air, SegmentType: "descent", Notes: "Training descent assumption"},
		{PlanID: plans[0].ID, SequenceNo: 2, DepthM: 30, DurationMin: 24, GasMixJSON: air, SegmentType: "bottom", Notes: "Constant-depth work assumption"},
		{PlanID: plans[0].ID, SequenceNo: 3, DepthM: 15, DurationMin: 2, AscentRateMMin: 8, GasMixJSON: air, SegmentType: "ascent", Notes: "Transition assumption only"},
		{PlanID: plans[0].ID, SequenceNo: 4, DepthM: 0, DurationMin: 2, AscentRateMMin: 8, GasMixJSON: air, SegmentType: "ascent", Notes: "Surface transition assumption"},
		{PlanID: plans[1].ID, SequenceNo: 1, DepthM: 42, DurationMin: 4, GasMixJSON: trimix, SegmentType: "descent", Notes: "Comparison input"},
		{PlanID: plans[1].ID, SequenceNo: 2, DepthM: 42, DurationMin: 20, GasMixJSON: trimix, SegmentType: "bottom", Notes: "Comparison input"},
		{PlanID: plans[1].ID, SequenceNo: 3, DepthM: 18, DurationMin: 2, AscentRateMMin: 12, GasMixJSON: trimix, SegmentType: "ascent", Notes: "Comparison transition"},
		{PlanID: plans[1].ID, SequenceNo: 4, DepthM: 0, DurationMin: 2, AscentRateMMin: 9, GasMixJSON: trimix, SegmentType: "ascent", Notes: "Comparison transition"},
	}
	if err := db.Create(&segments).Error; err != nil {
		return fmt.Errorf("seed exposure segments: %w", err)
	}
	result, err := decompression.Run(plans[1], profiles[1], segments[4:], cfg.ModelVersion, cfg.MaxSegments)
	if err != nil {
		return fmt.Errorf("seed assessment model: %w", err)
	}
	snapshot, curves, flags, assumptions, err := decompression.MarshalResult(result)
	if err != nil {
		return err
	}
	assessment := model.DecompressionAssessment{PlanID: plans[1].ID, AssessmentStatus: string(constants.PlanPendingReview), AlgorithmVersion: cfg.ModelVersion, InputSnapshotJSON: snapshot, CompartmentLoadsJSON: curves, RiskFlagsJSON: flags, HighestRiskBand: decompression.HighestRiskBand(result.RiskFlags), ComparativeScore: result.ComparativeScore, AssumptionsJSON: assumptions}
	if err := db.Create(&assessment).Error; err != nil {
		return fmt.Errorf("seed assessment: %w", err)
	}
	slog.Info("database seed completed", "profiles", len(profiles), "plans", len(plans), "segments", len(segments), "assessments", 1)
	return nil
}

func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql database: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	return nil
}
