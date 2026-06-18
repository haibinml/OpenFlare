// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package waf

import (
	"context"
	"testing"

	"github.com/Rain-kl/Wavelet/internal/db"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupWAFTestDB(t *testing.T) func() {
	t.Helper()

	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)
	require.NoError(t, sqliteDB.AutoMigrate(
		&model.OpenFlareWAFRuleGroup{},
		&model.OpenFlareWAFIPGroup{},
		&model.OpenFlareWAFRuleGroupBinding{},
	))

	db.SetDB(sqliteDB)
	return func() {
		db.SetDB(nil)
	}
}

func TestCreateRuleGroup(t *testing.T) {
	cleanup := setupWAFTestDB(t)
	defer cleanup()
	ctx := context.Background()

	group, err := CreateRuleGroup(ctx, RuleGroupInput{
		Name:             "edge guard",
		Enabled:          true,
		BlockStatusCode:  451,
		IPWhitelist:      []string{" 192.0.2.1 ", "192.0.2.1", "198.51.100.0/24"},
		IPBlacklist:      []string{"203.0.113.10"},
		CountryBlacklist: []string{" cn ", "CN", "us"},
	})
	require.NoError(t, err)
	assert.NotZero(t, group.ID)
	assert.False(t, group.IsGlobal)
	assert.Equal(t, "edge guard", group.Name)
	require.Len(t, group.IPWhitelist, 2)
	assert.Equal(t, "192.0.2.1", group.IPWhitelist[0])
	assert.Equal(t, "198.51.100.0/24", group.IPWhitelist[1])
	require.Len(t, group.CountryBlacklist, 2)
	assert.Equal(t, "CN", group.CountryBlacklist[0])
	assert.Equal(t, "US", group.CountryBlacklist[1])

	_, err = CreateRuleGroup(ctx, RuleGroupInput{
		Name:        "bad ip",
		Enabled:     true,
		IPBlacklist: []string{"not-an-ip"},
	})
	require.Error(t, err)
}
