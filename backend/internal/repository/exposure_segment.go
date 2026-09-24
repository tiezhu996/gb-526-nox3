package repository

import (
	"context"
	"errors"
	"fmt"

	"commercial-diving-decompression-control/backend/internal/audit"
	"commercial-diving-decompression-control/backend/internal/constants"
	"commercial-diving-decompression-control/backend/internal/model"
	"commercial-diving-decompression-control/backend/internal/util"
	"gorm.io/gorm"
)

type ExposureSegmentRepository struct {
	db    *gorm.DB
	audit *audit.Repository
}

func NewExposureSegmentRepository(db *gorm.DB, auditRepo *audit.Repository) *ExposureSegmentRepository {
	return &ExposureSegmentRepository{db: db, audit: auditRepo}
}

func (r *ExposureSegmentRepository) ListByPlan(ctx context.Context, planID uint) ([]model.ExposureSegment, error) {
	var items []model.ExposureSegment
	if err := r.db.WithContext(ctx).Where("plan_id = ?", planID).Order("sequence_no ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list exposure segments for plan %d: %w", planID, err)
	}
	return items, nil
}

func (r *ExposureSegmentRepository) Get(ctx context.Context, id uint) (model.ExposureSegment, error) {
	var item model.ExposureSegment
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return model.ExposureSegment{}, fmt.Errorf("get exposure segment %d: %w", id, err)
	}
	return item, nil
}

func requireDraftPlan(tx *gorm.DB, planID, version uint) (model.DivePlan, error) {
	var plan model.DivePlan
	if err := tx.First(&plan, planID).Error; err != nil {
		return model.DivePlan{}, fmt.Errorf("get segment plan %d: %w", planID, err)
	}
	if plan.Version != version {
		return model.DivePlan{}, util.Conflict("PLAN_VERSION_CONFLICT", "dive plan was changed by another user", nil)
	}
	if plan.PlanStatus != constants.PlanDraft {
		return model.DivePlan{}, util.Conflict("PLAN_NOT_EDITABLE", "exposure segments can only change while the plan is draft", nil)
	}
	return plan, nil
}

func bumpPlanVersion(tx *gorm.DB, plan model.DivePlan) error {
	result := tx.Model(&model.DivePlan{}).Where("id = ? AND version = ? AND plan_status = ?", plan.ID, plan.Version, constants.PlanDraft).Update("version", gorm.Expr("version + 1"))
	if result.Error != nil {
		return fmt.Errorf("increment plan input version: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return util.Conflict("PLAN_VERSION_CONFLICT", "dive plan input version changed concurrently", nil)
	}
	return nil
}

func (r *ExposureSegmentRepository) Create(ctx context.Context, item *model.ExposureSegment, planVersion uint, entry audit.Entry) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		plan, err := requireDraftPlan(tx, item.PlanID, planVersion)
		if err != nil {
			return err
		}
		if err := tx.Create(item).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return util.Conflict("SEGMENT_SEQUENCE_CONFLICT", "sequence_no already exists for this plan", err)
			}
			return fmt.Errorf("create exposure segment: %w", err)
		}
		if err := bumpPlanVersion(tx, plan); err != nil {
			return err
		}
		entry.EntityID = item.ID
		return r.audit.RecordWithDB(ctx, tx, entry)
	})
	if err != nil {
		return fmt.Errorf("create exposure segment transaction: %w", err)
	}
	return nil
}

func (r *ExposureSegmentRepository) Update(ctx context.Context, current model.ExposureSegment, planVersion uint, changes map[string]any, entry audit.Entry) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		plan, err := requireDraftPlan(tx, current.PlanID, planVersion)
		if err != nil {
			return err
		}
		if err := tx.Model(&model.ExposureSegment{}).Where("id = ? AND plan_id = ?", current.ID, current.PlanID).Updates(changes).Error; err != nil {
			return fmt.Errorf("update exposure segment: %w", err)
		}
		if err := bumpPlanVersion(tx, plan); err != nil {
			return err
		}
		entry.EntityID = current.ID
		return r.audit.RecordWithDB(ctx, tx, entry)
	})
	if err != nil {
		return fmt.Errorf("update exposure segment transaction: %w", err)
	}
	return nil
}

func (r *ExposureSegmentRepository) Reorder(ctx context.Context, planID, planVersion uint, orderedIDs []uint, entry audit.Entry) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		plan, err := requireDraftPlan(tx, planID, planVersion)
		if err != nil {
			return err
		}
		var segments []model.ExposureSegment
		if err := tx.Where("plan_id = ?", planID).Find(&segments).Error; err != nil {
			return fmt.Errorf("load segments for reorder: %w", err)
		}
		if len(segments) != len(orderedIDs) {
			return util.Unprocessable("SEGMENT_ORDER_MISMATCH", "ordered_ids must include every plan segment exactly once", nil)
		}
		known := make(map[uint]bool, len(segments))
		for _, segment := range segments {
			known[segment.ID] = true
		}
		seen := make(map[uint]bool, len(orderedIDs))
		for _, id := range orderedIDs {
			if !known[id] || seen[id] {
				return util.Unprocessable("SEGMENT_ORDER_MISMATCH", "ordered_ids contains an unknown or repeated segment", nil)
			}
			seen[id] = true
		}
		for index, id := range orderedIDs {
			if err := tx.Model(&model.ExposureSegment{}).Where("id = ? AND plan_id = ?", id, planID).Update("sequence_no", -(index + 1)).Error; err != nil {
				return fmt.Errorf("stage segment reorder: %w", err)
			}
		}
		for index, id := range orderedIDs {
			if err := tx.Model(&model.ExposureSegment{}).Where("id = ? AND plan_id = ?", id, planID).Update("sequence_no", index+1).Error; err != nil {
				return fmt.Errorf("finish segment reorder: %w", err)
			}
		}
		if err := bumpPlanVersion(tx, plan); err != nil {
			return err
		}
		entry.EntityID = planID
		return r.audit.RecordWithDB(ctx, tx, entry)
	})
	if err != nil {
		return fmt.Errorf("reorder exposure segments transaction: %w", err)
	}
	return nil
}
