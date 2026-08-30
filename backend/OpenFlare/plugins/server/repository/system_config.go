// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"Wavelet/OpenFlare/plugins/server/model"
	db "Wavelet/plugins/infra/database"

	"gorm.io/gorm"
)

const configTypeSystem = "system"

// GetSystemConfigByKey loads a config row by key.
func GetSystemConfigByKey(ctx context.Context, key string) (model.SystemConfig, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return model.SystemConfig{}, errors.New(errDatabaseNotInitialized)
	}
	var sc model.SystemConfig
	if err := conn.Where("key = ?", key).First(&sc).Error; err != nil {
		return model.SystemConfig{}, err
	}
	return sc, nil
}

// ListSystemConfigsByKeys loads multiple config keys.
func ListSystemConfigsByKeys(ctx context.Context, keys []string) (map[string]model.SystemConfig, error) {
	result := make(map[string]model.SystemConfig, len(keys))
	if len(keys) == 0 {
		return result, nil
	}
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	var configs []model.SystemConfig
	if err := conn.Where("key IN ?", keys).Find(&configs).Error; err != nil {
		return nil, err
	}
	for i := range configs {
		result[configs[i].Key] = configs[i]
	}
	return result, nil
}

// GetIntByKey queries config and converts to int.
func GetIntByKey(ctx context.Context, key string) (int, error) {
	sc, err := GetSystemConfigByKey(ctx, key)
	if err != nil {
		return 0, err
	}
	value, err := strconv.Atoi(sc.Value)
	if err != nil {
		return 0, fmt.Errorf(errConfigIntParseFailed, key, sc.Value, err)
	}
	return value, nil
}

// GetBoolByKey queries config and converts to bool.
func GetBoolByKey(ctx context.Context, key string) (bool, error) {
	sc, err := GetSystemConfigByKey(ctx, key)
	if err != nil {
		return false, err
	}
	value, err := strconv.ParseBool(sc.Value)
	if err != nil {
		return false, fmt.Errorf(errConfigBoolParseFailed, key, sc.Value, err)
	}
	return value, nil
}

// CreateSystemConfig persists a new system config row.
func CreateSystemConfig(ctx context.Context, config *model.SystemConfig) error {
	conn := db.DB(ctx)
	if conn == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	return conn.Create(config).Error
}

// SaveOrUpdateSystemConfig creates or updates a config row.
func SaveOrUpdateSystemConfig(ctx context.Context, key, value string) error {
	conn := db.DB(ctx)
	if conn == nil {
		return errors.New(errDatabaseNotInitialized)
	}
	var sc model.SystemConfig
	err := conn.Where("key = ?", key).First(&sc).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		sc = model.SystemConfig{
			Key:        key,
			Value:      value,
			Type:       configTypeSystem,
			Visibility: model.ConfigVisibilityHidden,
		}
		return conn.Create(&sc).Error
	}
	sc.Value = value
	return conn.Save(&sc).Error
}

// InvalidateSystemConfigCache is a no-op: OF no longer owns the config cache.
func InvalidateSystemConfigCache(context.Context, string) error { return nil }

// InvalidateAllSystemConfigCaches is a no-op retained for migrator.
func InvalidateAllSystemConfigCaches(context.Context) error { return nil }

// StopSystemConfigCacheListener is a no-op retained for existing tests.
func StopSystemConfigCacheListener() {}

// ResetSystemConfigRAMCacheForTest is a no-op retained for existing tests.
func ResetSystemConfigRAMCacheForTest() {}

// ListAdminSystemConfigs returns configs, optionally filtered by type.
func ListAdminSystemConfigs(ctx context.Context, configType string) ([]model.SystemConfig, error) {
	conn := db.DB(ctx)
	if conn == nil {
		return nil, errors.New(errDatabaseNotInitialized)
	}
	query := conn.Order("created_at DESC")
	if configType != "" {
		query = query.Where("type = ?", configType)
	}
	var configs []model.SystemConfig
	if err := query.Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}
