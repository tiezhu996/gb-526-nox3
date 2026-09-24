package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"commercial-diving-decompression-control/backend/internal/audit"
	"commercial-diving-decompression-control/backend/internal/constants"
	"commercial-diving-decompression-control/backend/internal/model"
	"commercial-diving-decompression-control/backend/internal/util"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newInsertTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dialector := sqlite.Open(fmt.Sprintf("file:seg-insert-%s?mode=memory&cache=shared", t.Name()))
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.DivePlan{}, &model.ExposureSegment{}, &audit.Event{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

func seedInsertPlan(t *testing.T, db *gorm.DB, status constants.PlanStatus, version uint, sequenceCount int) model.DivePlan {
	t.Helper()
	plan := model.DivePlan{PlanCode: fmt.Sprintf("PLAN-%s-%d", t.Name(), sequenceCount), DiverProfileID: 1, WorksitePressureBar: 1, BreathingMixJSON: `{"o2":0.21,"he":0,"n2":0.79}`, PlanStatus: status, CreatedBy: 1, Version: version, PlannedAt: time.Now().UTC()}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatalf("create plan: %v", err)
	}
	for index := 1; index <= sequenceCount; index++ {
		segment := model.ExposureSegment{PlanID: plan.ID, SequenceNo: index, DepthM: 10, DurationMin: 2, GasMixJSON: `{"o2":0.21,"he":0,"n2":0.79}`, SegmentType: "bottom", Notes: fmt.Sprintf("orig-%d", index)}
		if err := db.Create(&segment).Error; err != nil {
			t.Fatalf("create seed segment: %v", err)
		}
	}
	return plan
}

func newInsertItem(planID uint) *model.ExposureSegment {
	return &model.ExposureSegment{PlanID: planID, DepthM: 12, DurationMin: 1, GasMixJSON: `{"o2":0.21,"he":0,"n2":0.79}`, SegmentType: "bottom", Notes: "inserted"}
}

func appErrorCode(t *testing.T, err error) string {
	t.Helper()
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}

type insertPositionCase struct {
	name       string
	before     *int
	wantOrder  []string // notes in resulting sequence order
	wantStatus int
}

func TestExposureSegmentRepository_InsertPositions(t *testing.T) {
	ctx := context.Background()

	cases := []insertPositionCase{
		{name: "before first", before: intPtr(1), wantOrder: []string{"inserted", "orig-1", "orig-2", "orig-3"}, wantStatus: 4},
		{name: "before middle", before: intPtr(2), wantOrder: []string{"orig-1", "inserted", "orig-2", "orig-3"}, wantStatus: 4},
		{name: "before last", before: intPtr(3), wantOrder: []string{"orig-1", "orig-2", "inserted", "orig-3"}, wantStatus: 4},
		{name: "append after tail", before: intPtr(4), wantOrder: []string{"orig-1", "orig-2", "orig-3", "inserted"}, wantStatus: 4},
		{name: "append nil", before: nil, wantOrder: []string{"orig-1", "orig-2", "orig-3", "inserted"}, wantStatus: 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := newInsertTestDB(t)
			repo := NewExposureSegmentRepository(db, audit.NewRepository(db))
			plan := seedInsertPlan(t, db, constants.PlanDraft, 1, 3)

			err := repo.Insert(ctx, newInsertItem(plan.ID), 1, tc.before, audit.Entry{RequestID: "req-insert", ActorID: 1, ActorUsername: "planner", Action: "exposure_segment.insert"})
			if err != nil {
				t.Fatalf("insert returned error: %v", err)
			}
			var segments []model.ExposureSegment
			if err := db.Where("plan_id = ?", plan.ID).Order("sequence_no ASC").Find(&segments).Error; err != nil {
				t.Fatalf("reload segments: %v", err)
			}
			if len(segments) != len(tc.wantOrder) {
				t.Fatalf("segment count = %d, want %d", len(segments), len(tc.wantOrder))
			}
			for index, segment := range segments {
				if segment.SequenceNo != index+1 {
					t.Fatalf("segment at index %d has sequence_no %d, want continuous %d", index, segment.SequenceNo, index+1)
				}
				if segment.Notes != tc.wantOrder[index] {
					t.Fatalf("position %d note = %q, want %q", index+1, segment.Notes, tc.wantOrder[index])
				}
			}
			var refreshed model.DivePlan
			if err := db.First(&refreshed, plan.ID).Error; err != nil {
				t.Fatalf("reload plan: %v", err)
			}
			if refreshed.Version != 2 {
				t.Fatalf("plan version = %d, want 2", refreshed.Version)
			}
			var events []audit.Event
			if err := db.Where("action = ?", "exposure_segment.insert").Find(&events).Error; err != nil {
				t.Fatalf("load audit events: %v", err)
			}
			if len(events) != 1 {
				t.Fatalf("audit event count = %d, want 1", len(events))
			}
		})
	}
}

func TestExposureSegmentRepository_InsertIntoEmptyPlan(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name   string
		before *int
	}{
		{name: "nil appends into empty", before: nil},
		{name: "position 1 into empty", before: intPtr(1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := newInsertTestDB(t)
			repo := NewExposureSegmentRepository(db, audit.NewRepository(db))
			plan := seedInsertPlan(t, db, constants.PlanDraft, 1, 0)

			item := newInsertItem(plan.ID)
			if err := repo.Insert(ctx, item, 1, tc.before, audit.Entry{}); err != nil {
				t.Fatalf("insert into empty plan: %v", err)
			}
			if item.SequenceNo != 1 {
				t.Fatalf("new segment sequence = %d, want 1", item.SequenceNo)
			}
		})
	}
}

func TestExposureSegmentRepository_InsertRejectionsKeepOrder(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name      string
		status    constants.PlanStatus
		version   uint
		clientVer uint
		before    *int
		wantCode  string
	}{
		{name: "stale version", status: constants.PlanDraft, version: 2, clientVer: 1, before: intPtr(1), wantCode: "PLAN_VERSION_CONFLICT"},
		{name: "modeled plan locked", status: constants.PlanModeled, version: 1, clientVer: 1, before: intPtr(1), wantCode: "PLAN_NOT_EDITABLE"},
		{name: "pending review locked", status: constants.PlanPendingReview, version: 1, clientVer: 1, before: intPtr(1), wantCode: "PLAN_NOT_EDITABLE"},
		{name: "position zero", status: constants.PlanDraft, version: 1, clientVer: 1, before: intPtr(0), wantCode: "INVALID_INSERT_POSITION"},
		{name: "position past tail plus one", status: constants.PlanDraft, version: 1, clientVer: 1, before: intPtr(5), wantCode: "INVALID_INSERT_POSITION"},
		{name: "negative position", status: constants.PlanDraft, version: 1, clientVer: 1, before: intPtr(-2), wantCode: "INVALID_INSERT_POSITION"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := newInsertTestDB(t)
			repo := NewExposureSegmentRepository(db, audit.NewRepository(db))
			plan := seedInsertPlan(t, db, tc.status, tc.version, 3)

			err := repo.Insert(ctx, newInsertItem(plan.ID), tc.clientVer, tc.before, audit.Entry{ActorUsername: "planner"})
			if code := appErrorCode(t, err); code != tc.wantCode {
				t.Fatalf("error code = %q, want %q (err=%v)", code, tc.wantCode, err)
			}
			var segments []model.ExposureSegment
			if err := db.Where("plan_id = ?", plan.ID).Order("sequence_no ASC").Find(&segments).Error; err != nil {
				t.Fatalf("reload segments: %v", err)
			}
			if len(segments) != 3 {
				t.Fatalf("segment count changed to %d, want 3", len(segments))
			}
			for index, segment := range segments {
				if segment.SequenceNo != index+1 || segment.Notes != fmt.Sprintf("orig-%d", index+1) {
					t.Fatalf("original order mutated at %d: seq=%d note=%q", index+1, segment.SequenceNo, segment.Notes)
				}
			}
			var refreshed model.DivePlan
			if err := db.First(&refreshed, plan.ID).Error; err != nil {
				t.Fatalf("reload plan: %v", err)
			}
			if refreshed.Version != tc.version {
				t.Fatalf("plan version changed to %d, want %d", refreshed.Version, tc.version)
			}
			var eventCount int64
			if err := db.Model(&audit.Event{}).Where("action = ?", "exposure_segment.insert").Count(&eventCount).Error; err != nil {
				t.Fatalf("count audit events: %v", err)
			}
			if eventCount != 0 {
				t.Fatalf("audit event written on failed insert: %d", eventCount)
			}
		})
	}
}

func intPtr(value int) *int { return &value }
