// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package fleet

import (
	"context"
	"testing"

	"Wavelet/OpenFlare/plugins/server/kernel/model"
	"Wavelet/OpenFlare/plugins/server/kernel/repository"
	db "Wavelet/plugins/infra/database"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUptimeKumaSyncHandlerSkipsWhenDisabled(t *testing.T) {
	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)
	require.NoError(t, sqliteDB.AutoMigrate(&model.SystemConfig{}))
	db.SetDB(sqliteDB)
	t.Cleanup(func() { db.SetDB(nil) })

	ctx := context.Background()
	require.NoError(t, repository.SaveOrUpdateSystemConfig(ctx, model.ConfigKeyUptimeKumaEnabled, "false"))

	result, err := (&UptimeKumaSyncHandler{}).Execute(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, result.Message, "未启用")
}
