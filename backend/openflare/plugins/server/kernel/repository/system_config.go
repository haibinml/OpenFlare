// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"errors"

	"Wavelet/openflare/plugins/server/kernel/model"
	adminrepo "Wavelet/plugins/domain/admin/repository"
	db "Wavelet/plugins/infra/database"
)

const configTypeSystem = "system"

// ensureAdminStore points OF config access at Wavelet's admin repository so
// reads hit the same cache that SaveOrUpdateSystemConfig invalidates.
func ensureAdminStore(ctx context.Context) error {
	if conn := db.DB(ctx); conn != nil {
		adminrepo.SetDBService(db.NewService(conn))
	}
	if adminrepo.GetDB(ctx) == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	return nil
}

// GetSystemConfigByKey loads a config row by key through the admin store cache.
func GetSystemConfigByKey(ctx context.Context, key string) (model.SystemConfig, error) {
	if err := ensureAdminStore(ctx); err != nil {
		return model.SystemConfig{}, err
	}
	return adminrepo.GetSystemConfigByKey(ctx, key)
}

// ListSystemConfigsByKeys loads multiple config keys through the admin store cache.
func ListSystemConfigsByKeys(ctx context.Context, keys []string) (map[string]model.SystemConfig, error) {
	if err := ensureAdminStore(ctx); err != nil {
		return nil, err
	}
	return adminrepo.ListSystemConfigsByKeys(ctx, keys)
}

// ListVisibleSystemConfigs returns visibility=1 configs from the admin store cache.
func ListVisibleSystemConfigs(ctx context.Context) ([]model.SystemConfig, error) {
	if err := ensureAdminStore(ctx); err != nil {
		return nil, err
	}
	return adminrepo.ListVisibleSystemConfigs(ctx)
}

// GetIntByKey queries config and converts to int.
func GetIntByKey(ctx context.Context, key string) (int, error) {
	if err := ensureAdminStore(ctx); err != nil {
		return 0, err
	}
	return adminrepo.GetIntByKey(ctx, key)
}

// GetBoolByKey queries config and converts to bool.
func GetBoolByKey(ctx context.Context, key string) (bool, error) {
	if err := ensureAdminStore(ctx); err != nil {
		return false, err
	}
	return adminrepo.GetBoolByKey(ctx, key)
}

// CreateSystemConfig persists a new system config row.
func CreateSystemConfig(ctx context.Context, config *model.SystemConfig) error {
	if err := ensureAdminStore(ctx); err != nil {
		return err
	}
	return adminrepo.CreateSystemConfigRecord(ctx, config)
}

// SaveOrUpdateSystemConfig creates or updates a config row and invalidates the admin cache.
func SaveOrUpdateSystemConfig(ctx context.Context, key, value string) error {
	if err := ensureAdminStore(ctx); err != nil {
		return err
	}
	return adminrepo.SaveOrUpdateSystemConfig(ctx, key, value)
}

// InvalidateSystemConfigCache evicts one key from Wavelet's system-config cache.
func InvalidateSystemConfigCache(ctx context.Context, key string) error {
	if err := ensureAdminStore(ctx); err != nil {
		return err
	}
	return adminrepo.InvalidateSystemConfigCache(ctx, key)
}

// InvalidateAllSystemConfigCaches evicts the whole Wavelet system-config cache.
func InvalidateAllSystemConfigCaches(ctx context.Context) error {
	if err := ensureAdminStore(ctx); err != nil {
		return err
	}
	return adminrepo.InvalidateAllSystemConfigCaches(ctx)
}

// StopSystemConfigCacheListener is retained for existing tests.
func StopSystemConfigCacheListener() {
	adminrepo.StopSystemConfigCacheListener()
}

// ResetSystemConfigRAMCacheForTest clears the process-local admin config cache.
func ResetSystemConfigRAMCacheForTest() {
	adminrepo.ResetSystemConfigRAMCacheForTest()
}

// ListAdminSystemConfigs returns configs, optionally filtered by type.
func ListAdminSystemConfigs(ctx context.Context, configType string) ([]model.SystemConfig, error) {
	if err := ensureAdminStore(ctx); err != nil {
		return nil, err
	}
	return adminrepo.ListAdminSystemConfigs(ctx, configType)
}
