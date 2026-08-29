// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	"Wavelet/OpenFlare/plugins/server/model"
	"Wavelet/OpenFlare/plugins/server/repository"
	"Wavelet/OpenFlare/plugins/server/testhelper"
)

func TestListVisibleSystemConfigsUsesRedisCache(t *testing.T) {
	dbConn, _, cleanup := testhelper.SetupTestEnvironment(t)
	defer cleanup()
	ctx := context.Background()

	if err := repository.InvalidateVisibleSystemConfigsCache(ctx); err != nil {
		t.Fatalf("InvalidateVisibleSystemConfigsCache() error = %v", err)
	}

	if _, err := repository.ListVisibleSystemConfigs(ctx); err != nil {
		t.Fatalf("ListVisibleSystemConfigs() warm cache error = %v", err)
	}

	if err := dbConn.Create(&model.SystemConfig{
		Key:         "cache_probe_public_key",
		Value:       "cache_probe_public_value",
		Type:        "system",
		Visibility:  model.ConfigVisibilityVisible,
		Description: "cache probe",
	}).Error; err != nil {
		t.Fatalf("Create(cache_probe_public_key) error = %v", err)
	}

	cached, err := repository.ListVisibleSystemConfigs(ctx)
	if err != nil {
		t.Fatalf("ListVisibleSystemConfigs() cached call error = %v", err)
	}
	for _, item := range cached {
		if item.Key == "cache_probe_public_key" {
			t.Fatal("cached visible config list should be stale before invalidation")
		}
	}

	// Since system configs are now purely cached in process-local RAM (L1) and not written to Redis (L2),
	// we do not verify the existence of the legacy Redis key here.

	if err := repository.InvalidateVisibleSystemConfigsCache(ctx); err != nil {
		t.Fatalf("InvalidateVisibleSystemConfigsCache() error = %v", err)
	}

	refreshed, err := repository.ListVisibleSystemConfigs(ctx)
	if err != nil {
		t.Fatalf("ListVisibleSystemConfigs() refreshed call error = %v", err)
	}

	var found bool
	for _, item := range refreshed {
		if item.Key == "cache_probe_public_key" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("refreshed visible config list should include newly created public config")
	}
}
