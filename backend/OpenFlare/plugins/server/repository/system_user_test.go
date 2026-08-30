// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"testing"

	"Wavelet/core/contracts"
	"Wavelet/pkg/idgen"
	adminmodel "Wavelet/plugins/domain/admin/model"
	"Wavelet/plugins/infra/database"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type stubUserService struct {
	contracts.UserService
	user *contracts.UserDTO
}

func (s stubUserService) GetUserByUsername(context.Context, string) (*contracts.UserDTO, error) {
	return s.user, nil
}

type stubAuthService struct {
	contracts.AuthService
	sources []contracts.AuthSourceViewDTO
}

func (s stubAuthService) ListAuthSources(context.Context) ([]contracts.AuthSourceViewDTO, error) {
	return s.sources, nil
}

func setupRepoTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := sqliteDB.AutoMigrate(&adminmodel.TaskExecution{}); err != nil {
		t.Fatalf("AutoMigrate(TaskExecution) error = %v", err)
	}
	if err := idgen.Init(1); err != nil {
		t.Fatalf("idgen.Init() error = %v", err)
	}
	database.SetDB(sqliteDB)
	return sqliteDB, func() { database.SetDB(nil) }
}

func TestGetActiveAuthSourcesUsesAuthService(t *testing.T) {
	SetAuthService(stubAuthService{})
	t.Cleanup(func() { SetAuthService(nil) })

	got, err := GetActiveAuthSources(context.Background())
	if err != nil {
		t.Fatalf("GetActiveAuthSources() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("GetActiveAuthSources() len = %d, want 0", len(got))
	}

	SetAuthService(stubAuthService{sources: []contracts.AuthSourceViewDTO{
		{ID: 1, Name: "inactive", Type: "oidc", DisplayName: "Off", IsActive: false},
		{ID: 2, Name: "github", Type: "oidc", DisplayName: "GitHub", IconURL: "/i.png", IsActive: true},
	}})

	got, err = GetActiveAuthSources(context.Background())
	if err != nil {
		t.Fatalf("GetActiveAuthSources() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("GetActiveAuthSources() len = %d, want 1", len(got))
	}
	if got[0].ID != 2 || got[0].Name != "github" || !got[0].IsActive {
		t.Fatalf("GetActiveAuthSources()[0] = %+v, want active github id=2", got[0])
	}
}

func TestGetSystemUserUsesUserService(t *testing.T) {
	SetUserService(stubUserService{user: &contracts.UserDTO{
		ID:       42,
		Username: "system",
		Nickname: "System User",
		IsActive: true,
	}})
	t.Cleanup(func() { SetUserService(nil) })

	got := GetSystemUser(context.Background())
	if got.ID != 42 || got.Username != "system" || got.Nickname != "System User" {
		t.Fatalf("GetSystemUser() = %+v, want id=42 username=system", got)
	}
}

func TestGetTaskExecutionByTaskIDUsesAdminStore(t *testing.T) {
	_, cleanup := setupRepoTestDB(t)
	t.Cleanup(cleanup)

	ctx := context.Background()
	row := &adminmodel.TaskExecution{
		ID:       7,
		TaskID:   "task-public-id",
		TaskType: "pages_source_action",
		Status:   adminmodel.TaskExecutionStatusPending,
	}
	if err := database.DB(ctx).Create(row).Error; err != nil {
		t.Fatalf("Create(TaskExecution) error = %v", err)
	}

	got, err := GetTaskExecutionByTaskID(ctx, "task-public-id")
	if err != nil {
		t.Fatalf("GetTaskExecutionByTaskID(%q) error = %v", "task-public-id", err)
	}
	if got.ID != 7 || got.TaskType != "pages_source_action" {
		t.Fatalf("GetTaskExecutionByTaskID() = %+v, want id=7 type=pages_source_action", got)
	}
}
