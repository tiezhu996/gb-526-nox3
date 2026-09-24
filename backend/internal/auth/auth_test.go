package auth

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCurrentPrincipalUsesCurrentAccountState(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	user := User{Username: "planner", DisplayName: "Planner", Role: RolePlanner, PasswordHash: "unused", Active: true}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := NewService(NewRepository(db), "secret", 0)
	if err := db.Model(&user).Updates(map[string]any{"role": RoleSupervisor}).Error; err != nil {
		t.Fatalf("update role: %v", err)
	}
	principal, err := service.CurrentPrincipal(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("current principal: %v", err)
	}
	if principal.Role != RoleSupervisor {
		t.Fatalf("role = %q, want %q", principal.Role, RoleSupervisor)
	}

	if err := db.Model(&user).Update("active", false).Error; err != nil {
		t.Fatalf("deactivate user: %v", err)
	}
	if _, err := service.CurrentPrincipal(context.Background(), user.ID); err == nil {
		t.Fatal("inactive user unexpectedly authenticated")
	}
}
