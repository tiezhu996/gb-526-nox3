package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"commercial-diving-decompression-control/backend/internal/audit"
	"commercial-diving-decompression-control/backend/internal/constants"
	"commercial-diving-decompression-control/backend/internal/model"
	"commercial-diving-decompression-control/backend/internal/util"
	"gorm.io/gorm"
)

type DivePlanRepository struct {
	db    *gorm.DB
	audit *audit.Repository
}

func NewDivePlanRepository(db *gorm.DB, auditRepo *audit.Repository) *DivePlanRepository {
	return &DivePlanRepository{db: db, audit: auditRepo}
}

func (r *DivePlanRepository) List(ctx context.Context, search, status string, profileID uint, page, size int) ([]model.DivePlan, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.DivePlan{})
	if search != "" {
		query = query.Where("LOWER(plan_code) LIKE ?", "%"+strings.ToLower(strings.TrimSpace(search))+"%")
	}
	if status != "" {
		query = query.Where("plan_status = ?", status)
	}
	if profileID > 0 {
		query = query.Where("diver_profile_id = ?", profileID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count dive plans: %w", err)
	}
	var items []model.DivePlan
	if err := query.Order("planned_at DESC, id DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list dive plans: %w", err)
	}
	return items, total, nil
}

func (r *DivePlanRepository) Get(ctx context.Context, id uint) (model.DivePlan, error) {
	var item model.DivePlan
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return model.DivePlan{}, fmt.Errorf("get dive plan %d: %w", id, err)
	}
	return item, nil
}

func (r *DivePlanRepository) Create(ctx context.Context, item *model.DivePlan, entry audit.Entry) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return util.Conflict("PLAN_CODE_EXISTS", "plan_code already exists", err)
			}
			return fmt.Errorf("create dive plan: %w", err)
		}
		entry.EntityID = item.ID
		return r.audit.RecordWithDB(ctx, tx, entry)
	})
	if err != nil {
		return fmt.Errorf("create dive plan transaction: %w", err)
	}
	return nil
}

func (r *DivePlanRepository) Transition(ctx context.Context, current model.DivePlan, target constants.PlanStatus, reviewerID *uint, entry audit.Entry) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		changes := map[string]any{"plan_status": target, "version": gorm.Expr("version + 1")}
		if reviewerID != nil {
			changes["reviewed_by"] = *reviewerID
		}
		result := tx.Model(&model.DivePlan{}).Where("id = ? AND version = ? AND plan_status = ?", current.ID, current.Version, current.PlanStatus).Updates(changes)
		if result.Error != nil {
			return fmt.Errorf("transition dive plan: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return util.Conflict("PLAN_VERSION_CONFLICT", "dive plan state or version changed concurrently", nil)
		}
		entry.EntityID = current.ID
		return r.audit.RecordWithDB(ctx, tx, entry)
	})
	if err != nil {
		return fmt.Errorf("transition dive plan transaction: %w", err)
	}
	return nil
}

func (r *DivePlanRepository) ResetDraftAfterFailure(ctx context.Context, current model.DivePlan, reason string, entry audit.Entry) error {
	if current.PlanStatus == constants.PlanDraft {
		entry.EntityID = current.ID
		entry.Action = "dive_plan.model.rejected"
		entry.BeforeSummary = string(current.PlanStatus)
		entry.AfterSummary = reason
		return r.audit.Record(ctx, entry)
	}
	entry.Action = "dive_plan.model.reset"
	entry.BeforeSummary = string(current.PlanStatus)
	entry.AfterSummary = "draft: " + reason
	return r.Transition(ctx, current, constants.PlanDraft, nil, entry)
}
