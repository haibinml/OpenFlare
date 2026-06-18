// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package pages

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

func setupPagesTestDB(t *testing.T) func() {
	t.Helper()

	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)
	require.NoError(t, sqliteDB.AutoMigrate(
		&model.PagesProject{},
		&model.PagesDeployment{},
		&model.PagesDeploymentFile{},
	))

	db.SetDB(sqliteDB)
	return func() {
		db.SetDB(nil)
	}
}

func TestCreateProject(t *testing.T) {
	cleanup := setupPagesTestDB(t)
	defer cleanup()
	ctx := context.Background()

	project, err := CreateProject(ctx, Input{
		Name:               "Marketing Site",
		Slug:               "marketing-site",
		Description:        "public site",
		Enabled:            true,
		SPAFallbackEnabled: true,
		SPAFallbackPath:    "/index.html",
		EntryFile:          "index.html",
	})
	require.NoError(t, err)
	assert.NotZero(t, project.ID)
	assert.Equal(t, "Marketing Site", project.Name)
	assert.Equal(t, "marketing-site", project.Slug)
	assert.Equal(t, "public site", project.Description)
	assert.True(t, project.Enabled)
	assert.True(t, project.SPAFallbackEnabled)
	assert.Equal(t, "/index.html", project.SPAFallbackPath)
	assert.Equal(t, "index.html", project.EntryFile)
	assert.Equal(t, int64(0), project.DeploymentCount)

	_, err = CreateProject(ctx, Input{
		Name: "Duplicate Slug",
		Slug: "marketing-site",
	})
	require.Error(t, err)
	assert.Equal(t, errPagesSlugExists, err.Error())
}

func TestCreateProjectRejectsUnsafeFallbackPath(t *testing.T) {
	cleanup := setupPagesTestDB(t)
	defer cleanup()
	ctx := context.Background()

	_, err := CreateProject(ctx, Input{
		Name:               "Unsafe Fallback",
		Slug:               "unsafe-fallback",
		Enabled:            true,
		SPAFallbackEnabled: true,
		SPAFallbackPath:    "/index.html; proxy_pass http://evil",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "回退路径")
}
