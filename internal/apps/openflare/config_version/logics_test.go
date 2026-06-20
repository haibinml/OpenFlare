// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package config_version

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Rain-kl/Wavelet/internal/apps/openflare/waf"
	"github.com/Rain-kl/Wavelet/internal/db"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupConfigVersionTestDB(t *testing.T) func() {
	t.Helper()

	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)
	require.NoError(t, sqliteDB.AutoMigrate(
		&model.ProxyRoute{},
		&model.ConfigVersion{},
		&model.OpenFlareWAFRuleGroup{},
		&model.OpenFlareWAFRuleGroupBinding{},
		&model.OpenFlareWAFIPGroup{},
	))

	db.SetDB(sqliteDB)
	return func() {
		db.SetDB(nil)
	}
}

func TestPublishConfigVersionCreatesVersion(t *testing.T) {
	cleanup := setupConfigVersionTestDB(t)
	defer cleanup()
	ctx := context.Background()

	route := &model.ProxyRoute{
		SiteName:  "publish-site",
		Domain:    "publish.example.com",
		Domains:   `["publish.example.com"]`,
		OriginURL: "http://origin.publish.example.com:8080",
		Upstreams: `["http://origin.publish.example.com:8080"]`,
		Enabled:   true,
	}
	require.NoError(t, model.CreateProxyRouteRecord(ctx, route))

	version, err := PublishConfigVersion(ctx, "tester", false)
	require.NoError(t, err)
	require.NotNil(t, version)
	assert.NotZero(t, version.ID)
	assert.True(t, version.IsActive)
	assert.Equal(t, "tester", version.CreatedBy)
	assert.NotEmpty(t, version.Version)
	assert.NotEmpty(t, version.Checksum)
	assert.NotEmpty(t, version.SnapshotJSON)
	assert.NotEmpty(t, version.RenderedConfig)

	var snapshot snapshotDocument
	require.NoError(t, json.Unmarshal([]byte(version.SnapshotJSON), &snapshot))
	require.Len(t, snapshot.Routes, 1)
	assert.Equal(t, "publish-site", snapshot.Routes[0].SiteName)
	assert.Equal(t, "publish.example.com", snapshot.Routes[0].Domain)

	active, err := GetActiveConfigVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, version.ID, active.ID)

	_, err = PublishConfigVersion(ctx, "tester", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), errNoChangesToPublish)

	forced, err := PublishConfigVersion(ctx, "tester", true)
	require.NoError(t, err)
	assert.NotEqual(t, version.ID, forced.ID)
}

func TestBuildSnapshotWAFDocumentUsesNormalizedSiteNames(t *testing.T) {
	cleanup := setupConfigVersionTestDB(t)
	defer cleanup()
	ctx := context.Background()

	route := &model.ProxyRoute{
		Domain:    "Example.COM",
		Domains:   `["example.com","www.example.com"]`,
		OriginURL: "http://origin.example.com:8080",
		Upstreams: `["http://origin.example.com:8080"]`,
		Enabled:   true,
	}
	require.NoError(t, model.CreateProxyRouteRecord(ctx, route))

	require.NoError(t, waf.EnsureDefaultRuleGroup(ctx))
	globalGroup, err := model.GetGlobalOpenFlareWAFRuleGroup(ctx)
	require.NoError(t, err)

	customGroup := &model.OpenFlareWAFRuleGroup{
		Name:       "pow-group",
		Enabled:    true,
		PoWEnabled: true,
		PoWConfig:  `{"difficulty":4,"algorithm":"fast","session_ttl":600,"challenge_ttl":300}`,
	}
	require.NoError(t, model.CreateOpenFlareWAFRuleGroup(ctx, customGroup))
	require.NoError(t, model.ReplaceOpenFlareWAFRuleGroupBindings(ctx, customGroup.ID, []uint{route.ID}))

	bundle, err := buildCurrentConfigBundle(ctx, true)
	require.NoError(t, err)
	require.Len(t, bundle.SnapshotRoutes, 1)
	assert.Equal(t, "example.com", bundle.SnapshotRoutes[0].SiteName)

	require.NotEmpty(t, bundle.WAFSnapshot.Bindings)
	found := false
	for _, binding := range bundle.WAFSnapshot.Bindings {
		if binding.RouteID != route.ID {
			continue
		}
		found = true
		assert.Equal(t, "example.com", binding.SiteName)
		assert.Contains(t, binding.RuleGroupIDs, customGroup.ID)
	}
	assert.True(t, found, "expected WAF binding for enabled route")

	var wafRuntime struct {
		SiteRuleGroups map[string][]uint `json:"site_rule_groups"`
	}
	for _, file := range bundle.SupportFiles {
		if file.Path != "waf_config.json" {
			continue
		}
		require.NoError(t, json.Unmarshal([]byte(file.Content), &wafRuntime))
	}
	require.Contains(t, wafRuntime.SiteRuleGroups, "example.com")
	require.Contains(t, wafRuntime.SiteRuleGroups["example.com"], customGroup.ID)
	require.Contains(t, wafRuntime.SiteRuleGroups["example.com"], globalGroup.ID)
	assert.Contains(t, bundle.RouteConfig, `set $openflare_waf_site "example.com"`)
	assert.Contains(t, bundle.RouteConfig, `require("pow.runtime").check()`)
}
