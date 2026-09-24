package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"commercial-diving-decompression-control/backend/internal/audit"
	"commercial-diving-decompression-control/backend/internal/model"
	"commercial-diving-decompression-control/backend/internal/util"
	"gorm.io/gorm"
)

type DiverProfileRepository struct {
	db    *gorm.DB
	audit *audit.Repository
}

func NewDiverProfileRepository(db *gorm.DB, auditRepo *audit.Repository) *DiverProfileRepository {
	return &DiverProfileRepository{db: db, audit: auditRepo}
}

func (r *DiverProfileRepository) List(ctx context.Context, search, status string, page, size int) ([]model.DiverProfile, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.DiverProfile{})
	if search != "" {
		like := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		query = query.Where("LOWER(profile_code) LIKE ? OR LOWER(display_name) LIKE ?", like, like)
	}
	if status != "" {
		query = query.Where("profile_status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count diver profiles: %w", err)
	}
	var items []model.DiverProfile
	if err := query.Order("profile_code ASC").Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list diver profiles: %w", err)
	}
	return items, total, nil
}

func (r *DiverProfileRepository) Get(ctx context.Context, id uint) (model.DiverProfile, error) {
	var item model.DiverProfile
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return model.DiverProfile{}, fmt.Errorf("get diver profile %d: %w", id, err)
	}
	return item, nil
}

func (r *DiverProfileRepository) Create(ctx context.Context, item *model.DiverProfile, entry audit.Entry) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return util.Conflict("PROFILE_CODE_EXISTS", "profile_code already exists", err)
			}
			return fmt.Errorf("create diver profile: %w", err)
		}
		entry.EntityID = item.ID
		if err := r.audit.RecordWithDB(ctx, tx, entry); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("create diver profile transaction: %w", err)
	}
	return nil
}

func (r *DiverProfileRepository) Update(ctx context.Context, current model.DiverProfile, changes map[string]any, entry audit.Entry) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		changes["version"] = gorm.Expr("version + 1")
		result := tx.Model(&model.DiverProfile{}).Where("id = ? AND version = ?", current.ID, current.Version).Updates(changes)
		if result.Error != nil {
			return fmt.Errorf("update diver profile: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return util.Conflict("PROFILE_VERSION_CONFLICT", "diver profile was changed by another user", nil)
		}
		entry.EntityID = current.ID
		return r.audit.RecordWithDB(ctx, tx, entry)
	})
	if err != nil {
		return fmt.Errorf("update diver profile transaction: %w", err)
	}
	return nil
}
