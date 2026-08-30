// Copyright 2025 linux.do
// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package testhelper 提供测试辅助工具
package testhelper

import (
	"context"
	"testing"

	"Wavelet/OpenFlare/plugins/server/model"
	analyticsmodel "Wavelet/OpenFlare/plugins/server/model/analytics"
	"Wavelet/OpenFlare/plugins/server/repository"
	"Wavelet/OpenFlare/plugins/server/repository/logstore"
	"Wavelet/pkg/idgen"
	db "Wavelet/plugins/infra/database"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	configTypeSystem   = "system"
	configTypeBusiness = "business"
	configValueTrue    = "true"
	configValueFalse   = "false"
)

// SetupTestEnvironment initializes an in-memory SQLite DB and seeds default
// configurations. Redis is no longer owned by OpenFlare.
func SetupTestEnvironment(t *testing.T) (*gorm.DB, any, func()) {
	t.Helper()
	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to open in-memory SQLite db: %v", err)
	}

	if sqlDB, err := sqliteDB.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	err = sqliteDB.AutoMigrate(
		&model.User{},
		&model.AuthSource{},
		&model.ExternalAccount{},
		&model.SystemConfig{},
		&model.Upload{},
		&model.UploadStat{},
		&model.TaskExecution{},
		&model.Template{},
		&model.AccessToken{},
		&model.Schedule{},
	)
	if err != nil {
		t.Fatalf("failed to auto migrate tables: %v", err)
	}

	db.SetDB(sqliteDB)
	if err := idgen.Init(1); err != nil {
		t.Fatalf("idgen.Init: %v", err)
	}
	seedDefaultConfigs(t, sqliteDB)

	cleanup := func() {
		runExtraCleanups()
		repository.StopSystemConfigCacheListener()
		repository.ResetSystemConfigRAMCacheForTest()
		db.SetDB(nil)
	}

	return sqliteDB, nil, cleanup
}

func getSeedConfigsPart1() []model.SystemConfig {
	return []model.SystemConfig{
		{Key: model.ConfigKeyUploadAllowedExtensions, Value: "jpg,png,webp", Type: configTypeSystem, Description: "允许上传的图片扩展名（逗号分隔）"},
		{Key: model.ConfigKeySiteName, Value: "OpenFlare", Type: configTypeSystem, Description: "系统平台的展示名称"},
		{Key: model.ConfigKeyPasswordLoginEnabled, Value: configValueTrue, Type: configTypeSystem},
		{Key: model.ConfigKeyRegistrationEnabled, Value: configValueFalse, Type: configTypeSystem},
		{Key: model.ConfigKeyPasswordRegisterEnabled, Value: configValueFalse, Type: configTypeSystem},
		{Key: model.ConfigKeyOIDCLoginEnabled, Value: configValueTrue, Type: configTypeSystem},
		{Key: model.ConfigKeyMaxAPIKeysPerUser, Value: "5", Type: "business"},
		{Key: model.ConfigKeyCapLoginEnabled, Value: configValueFalse, Type: configTypeSystem},
		{Key: model.ConfigKeyCapAutoSolve, Value: configValueTrue, Type: configTypeSystem},
		{Key: model.ConfigKeyCapChallengeCount, Value: "1", Type: configTypeSystem},
		{Key: model.ConfigKeyCapChallengeSize, Value: "32", Type: configTypeSystem},
		{Key: model.ConfigKeyCapChallengeDifficulty, Value: "4", Type: configTypeSystem},
		{Key: model.ConfigKeyCapChallengeTTL, Value: "600", Type: configTypeSystem},
		{Key: model.ConfigKeyCapTokenTTL, Value: "1200", Type: configTypeSystem},
	}
}

func getSeedConfigsPart2() []model.SystemConfig {
	return []model.SystemConfig{
		{Key: model.ConfigKeyServerAddress, Value: "", Type: configTypeSystem},
		{Key: model.ConfigKeySMTPHost, Value: "", Type: configTypeSystem},
		{Key: model.ConfigKeySMTPPort, Value: "587", Type: configTypeSystem},
		{Key: model.ConfigKeySMTPUsername, Value: "", Type: configTypeSystem},
		{Key: model.ConfigKeySMTPPassword, Value: "", Type: configTypeSystem},
		{Key: model.ConfigKeyEmailLoginVerificationEnabled, Value: configValueFalse, Type: configTypeSystem},
		{Key: model.ConfigKeyEmailRegisterVerificationEnabled, Value: configValueFalse, Type: configTypeSystem},
		{Key: model.ConfigKeyMenuDisplayConfig, Value: "{}", Type: configTypeSystem},
		{Key: model.ConfigKeySearchEngineIndexingEnabled, Value: configValueFalse, Type: configTypeSystem},
		{Key: model.ConfigKeyFileAccessWhitelist, Value: `["avatar"]`, Type: configTypeSystem},
		{Key: model.ConfigKeyDiskCacheMaxSizeMB, Value: "100", Type: configTypeSystem},
		{Key: model.ConfigKeyDiskCacheTTLMinutes, Value: "60", Type: configTypeSystem},
		{Key: model.ConfigKeyDiskCacheLRUEnabled, Value: configValueTrue, Type: configTypeSystem},
		{Key: model.ConfigKeyLoginSessionTTLHours, Value: "0", Type: configTypeSystem},
		{Key: model.ConfigKeyUpdateUpstreamRepository, Value: "Rain-kl/OpenFlare", Type: configTypeSystem},
		{Key: model.ConfigKeyStorageConfig, Value: `{"driver":"local","local":{"root":"."},"s3":{"region":"us-east-1"},"r2":{"region":"auto"},"minio":{"region":"us-east-1","path_style":true},"oss":{},"webdav":{}}`, Type: configTypeSystem},
		{Key: model.ConfigKeyRelayFRPSWebUIEnabled, Value: configValueFalse, Type: configTypeBusiness},
		{Key: model.ConfigKeyRelayFRPSWebUIPort, Value: "17500", Type: configTypeBusiness},
		{Key: model.ConfigKeyPagesMaxPackageSizeMB, Value: "100", Type: configTypeBusiness},
		{Key: model.ConfigKeyPagesMaxHistoryCount, Value: "20", Type: configTypeBusiness},
		{Key: model.ConfigKeyLogRetentionDaysPostgres, Value: "90", Type: configTypeBusiness},
		{Key: model.ConfigKeyLogRetentionDaysSQLite, Value: "90", Type: configTypeBusiness},
		{Key: model.ConfigKeyLogRetentionDaysClickHouse, Value: "90", Type: configTypeBusiness},
	}
}

func seedDefaultConfigs(t *testing.T, tx *gorm.DB) {
	t.Helper()
	defaultConfigs := append(getSeedConfigsPart1(), getSeedConfigsPart2()...)
	if err := tx.Create(&defaultConfigs).Error; err != nil {
		t.Fatalf("failed to seed default system configs: %v", err)
	}

	publicKeys := []string{
		model.ConfigKeyUploadAllowedExtensions,
		model.ConfigKeySiteName,
		model.ConfigKeyPasswordLoginEnabled,
		model.ConfigKeyRegistrationEnabled,
		model.ConfigKeyPasswordRegisterEnabled,
		model.ConfigKeyOIDCLoginEnabled,
		model.ConfigKeyMaxAPIKeysPerUser,
		model.ConfigKeyCapLoginEnabled,
		model.ConfigKeyCapAutoSolve,
		model.ConfigKeyEmailLoginVerificationEnabled,
		model.ConfigKeyEmailRegisterVerificationEnabled,
		model.ConfigKeyMenuDisplayConfig,
		model.ConfigKeySearchEngineIndexingEnabled,
		model.ConfigKeyFileAccessWhitelist,
	}
	if err := tx.Model(&model.SystemConfig{}).
		Where("key IN ?", publicKeys).
		Update("visibility", model.ConfigVisibilityVisible).Error; err != nil {
		t.Fatalf("failed to seed public system config visibility: %v", err)
	}
}

// SetupLogStoresForTest 将 logstore 指向测试已通过 db.SetDB 注入的 sqlite 库。
func SetupLogStoresForTest(t *testing.T) {
	t.Helper()

	gdb := db.DB(context.Background())
	require.NoError(t, idgen.Init(1))
	require.NoError(t, gdb.AutoMigrate(
		&analyticsmodel.NodeAccessLog{},
		&analyticsmodel.UserAccessLog{},
		&analyticsmodel.NodeMetricSnapshot{},
		&analyticsmodel.NodeEdgeHealth{},
		&analyticsmodel.NodeObsFrps{},
		&analyticsmodel.NodeObsFrpc{},
	))

	logstore.ResetForTest()
	logstore.SetConfigReader(func(_ context.Context, key string) (string, error) {
		if key == model.ConfigKeyLogDatabase {
			return "sqlite", nil
		}
		return "", nil
	})
	store, err := logstore.Active(context.Background())
	require.NoError(t, err)

	logstore.SetAccessLogHooks(logstore.AccessLogHooks{
		QueueNodeAccessLogs: func(logs []analyticsmodel.NodeAccessLog) {
			if err := store.AccessLogs.BatchInsertNodeAccessLogs(context.Background(), logs); err != nil {
				t.Errorf("batch insert node access logs failed in test hook: %v", err)
			}
		},
	})
	logstore.SetObservabilityHooks(logstore.ObservabilityHooks{
		QueueMetricSnapshot: func(record analyticsmodel.NodeMetricSnapshot) {
			if err := store.Observability.BatchInsertNodeMetricSnapshots(context.Background(), []analyticsmodel.NodeMetricSnapshot{record}); err != nil {
				t.Errorf("batch insert node metric snapshots failed in test hook: %v", err)
			}
		},
		QueueEdgeHealth: func(record analyticsmodel.NodeEdgeHealth) {
			if err := store.Observability.BatchInsertNodeEdgeHealth(context.Background(), []analyticsmodel.NodeEdgeHealth{record}); err != nil {
				t.Errorf("batch insert node edge health failed in test hook: %v", err)
			}
		},
		QueueNodeObsFrps: func(record analyticsmodel.NodeObsFrps) {
			if err := store.Observability.BatchInsertNodeObsFrps(context.Background(), []analyticsmodel.NodeObsFrps{record}); err != nil {
				t.Errorf("batch insert node obs frps failed in test hook: %v", err)
			}
		},
		QueueNodeObsFrpc: func(record analyticsmodel.NodeObsFrpc) {
			if err := store.Observability.BatchInsertNodeObsFrpc(context.Background(), []analyticsmodel.NodeObsFrpc{record}); err != nil {
				t.Errorf("batch insert node obs frpc failed in test hook: %v", err)
			}
		},
	})

	t.Cleanup(func() {
		logstore.SetAccessLogHooks(logstore.AccessLogHooks{})
		logstore.SetObservabilityHooks(logstore.ObservabilityHooks{})
		logstore.ResetForTest()
	})
}
