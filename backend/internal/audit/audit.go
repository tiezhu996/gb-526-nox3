package audit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"commercial-diving-decompression-control/backend/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Event struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	RequestID     string    `gorm:"size:64;not null;index" json:"request_id"`
	ActorID       uint      `gorm:"not null;index" json:"actor_id"`
	ActorUsername string    `gorm:"size:64;not null;index" json:"actor_username"`
	Action        string    `gorm:"size:64;not null;index" json:"action"`
	EntityType    string    `gorm:"size:64;not null;index" json:"entity_type"`
	EntityID      uint      `gorm:"not null;index" json:"entity_id"`
	BeforeSummary string    `gorm:"type:text;not null;default:''" json:"before_summary"`
	AfterSummary  string    `gorm:"type:text;not null;default:''" json:"after_summary"`
	CreatedAt     time.Time `gorm:"not null;index" json:"created_at"`
}

func (Event) TableName() string { return "audit_events" }

type Entry struct {
	RequestID     string
	ActorID       uint
	ActorUsername string
	Action        string
	EntityType    string
	EntityID      uint
	BeforeSummary string
	AfterSummary  string
}

type Filter struct {
	EntityType string
	Actor      string
	RequestID  string
	Page       int
	Size       int
}

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Record(ctx context.Context, entry Entry) error {
	return r.RecordWithDB(ctx, r.db, entry)
}

func (r *Repository) RecordWithDB(ctx context.Context, db *gorm.DB, entry Entry) error {
	event := Event{
		RequestID: entry.RequestID, ActorID: entry.ActorID, ActorUsername: entry.ActorUsername,
		Action: entry.Action, EntityType: entry.EntityType, EntityID: entry.EntityID,
		BeforeSummary: entry.BeforeSummary, AfterSummary: entry.AfterSummary,
	}
	if event.RequestID == "" {
		event.RequestID = "system"
	}
	if event.ActorUsername == "" {
		event.ActorUsername = "system"
	}
	if err := db.WithContext(ctx).Create(&event).Error; err != nil {
		return fmt.Errorf("record audit event: %w", err)
	}
	return nil
}

func (r *Repository) List(ctx context.Context, filter Filter) ([]Event, int64, error) {
	query := r.db.WithContext(ctx).Model(&Event{})
	if filter.EntityType != "" {
		query = query.Where("entity_type = ?", filter.EntityType)
	}
	if filter.Actor != "" {
		query = query.Where("LOWER(actor_username) LIKE ?", "%"+strings.ToLower(filter.Actor)+"%")
	}
	if filter.RequestID != "" {
		query = query.Where("request_id = ?", filter.RequestID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count audit events: %w", err)
	}
	var events []Event
	offset := (filter.Page - 1) * filter.Size
	if err := query.Order("created_at DESC, id DESC").Offset(offset).Limit(filter.Size).Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("list audit events: %w", err)
	}
	return events, total, nil
}

type Handler struct{ repo *Repository }

func NewHandler(repo *Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) List(c *gin.Context) {
	page, size := util.QueryPage(c)
	filter := Filter{
		EntityType: strings.TrimSpace(c.Query("entity_type")),
		Actor:      strings.TrimSpace(c.Query("actor")),
		RequestID:  strings.TrimSpace(c.Query("request_id")),
		Page:       page, Size: size,
	}
	events, total, err := h.repo.List(c.Request.Context(), filter)
	if err != nil {
		util.Fail(c, util.Internal(err))
		return
	}
	util.OK(c, util.Page{Items: events, Total: total, Page: page, Size: size})
}
