package geoip

import (
	"context"
	"testing"

	"github.com/Rain-kl/Wavelet/internal/db"
	"github.com/Rain-kl/Wavelet/internal/model"
	pkggeoip "github.com/Rain-kl/Wavelet/pkg/geoip"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestEnsureRuntimeProviderInitializesConfiguredProvider(t *testing.T) {
	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := sqliteDB.AutoMigrate(&model.OpenFlareOption{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db.SetDB(sqliteDB)
	t.Cleanup(func() {
		db.SetDB(nil)
		model.ResetOptionMapForTest()
		ResetRuntimeForTest()
	})

	ctx := context.Background()
	model.ResetOptionMapForTest()
	ResetRuntimeForTest()
	if err := model.UpdateOpenFlareOption(ctx, "GeoIPProvider", pkggeoip.ProviderIPInfo); err != nil {
		t.Fatalf("update option: %v", err)
	}

	if err := EnsureRuntimeProvider(ctx); err != nil {
		t.Fatalf("EnsureRuntimeProvider error = %v", err)
	}
	if pkggeoip.CurrentProvider == nil || pkggeoip.CurrentProvider.Name() == "EmptyProvider" {
		t.Fatalf("expected ipinfo provider, got %#v", pkggeoip.CurrentProvider)
	}
}
