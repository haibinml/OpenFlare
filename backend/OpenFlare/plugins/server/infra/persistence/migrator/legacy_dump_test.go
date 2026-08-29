// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"Wavelet/OpenFlare/plugins/server/openflare/zone"

	"github.com/glebarez/sqlite"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// 记账表随迁移路径不同而不同（历史链用 goose_db_version，插件化后用 w_schema_versions），
// 比对 schema 时必须排除，否则 A/B 两路必然假性不一致。
var bookkeepingTables = []string{"goose_db_version", "w_schema_versions"}

// TestDumpLegacySchema 把当前 embed 内的完整历史链应用到临时 sqlite 库并导出 schema，
// 作为 Cordis 改造前后 schema 一致性的基线（A 路）与新架构产出（B 路）的唯一事实来源。
//
// OF_DUMP_SCHEMA=<path> 导出表/索引定义；OF_DUMP_VERSIONS=<path> 导出已应用版本序列。
func TestDumpLegacySchema(t *testing.T) {
	dumpPath := os.Getenv("OF_DUMP_SCHEMA")
	if dumpPath == "" && os.Getenv("OF_DUMP_VERSIONS") == "" {
		t.Skip("OF_DUMP_SCHEMA / OF_DUMP_VERSIONS not set")
	}

	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := conn.DB()
	require.NoError(t, err)
	// :memory: 连接池仅一条连接可见，否则 goose 与 dump 会看到不同的空库。
	sqlDB.SetMaxOpenConns(1)

	goose.SetBaseFS(migrationFS)
	require.NoError(t, goose.SetDialect(dialectSqlite))

	// 与 Migrate() 保持同一顺序：SQL 到 zone 导入标记 → Go 侧导入 → 其余 SQL。
	require.NoError(t, goose.UpTo(sqlDB, "goose/sqlite", zoneImportSQLVersion))
	require.NoError(t, runZoneImport(sqlDB))
	require.NoError(t, goose.Up(sqlDB, "goose/sqlite"))

	if dumpPath != "" {
		schema := dumpSchema(t, sqlDB)
		require.NoError(t, os.WriteFile(dumpPath, []byte(schema), 0o600))
	}
	if vPath := os.Getenv("OF_DUMP_VERSIONS"); vPath != "" {
		versions := dumpVersions(t, sqlDB)
		require.NoError(t, os.WriteFile(vPath, []byte(versions), 0o600))
	}
}

// runZoneImport 在空库上执行 Go 侧 zone 导入，验证升级窗口钩子本身可运行（空库应为零导入）。
func runZoneImport(sqlDB *sql.DB) error {
	ctx := context.Background()
	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin zone import: %w", err)
	}
	report, err := zone.ImportLegacyTx(ctx, tx, false)
	if err != nil {
		_ = tx.Rollback()
		return report.LogAndReturn(err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit zone import: %w", err)
	}
	return nil
}

func dumpSchema(t *testing.T, sqlDB *sql.DB) string {
	t.Helper()
	query := `SELECT type, name, COALESCE(sql, '') FROM sqlite_master
		WHERE type IN ('table', 'index') AND name NOT LIKE 'sqlite_%'
		ORDER BY type, name`
	rows, err := sqlDB.Query(query)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	var b strings.Builder
	for rows.Next() {
		var objType, name, ddl string
		require.NoError(t, rows.Scan(&objType, &name, &ddl))
		if isBookkeeping(name) {
			continue
		}
		fmt.Fprintf(&b, "-- %s %s\n%s;\n", objType, name, strings.TrimSpace(ddl))
	}
	require.NoError(t, rows.Err())
	return b.String() + fmt.Sprintf("-- total objects (excl. bookkeeping): %d\n", countObjects(t, sqlDB))
}

func countObjects(t *testing.T, sqlDB *sql.DB) int {
	t.Helper()
	query := `SELECT COUNT(*) FROM sqlite_master WHERE type IN ('table','index')
		AND name NOT LIKE 'sqlite_%'`
	var n int
	require.NoError(t, sqlDB.QueryRow(query).Scan(&n))
	for _, tbl := range bookkeepingTables {
		var has int
		require.NoError(t, sqlDB.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE name = ?", tbl).Scan(&has))
		n -= has
	}
	return n
}

func dumpVersions(t *testing.T, sqlDB *sql.DB) string {
	t.Helper()
	rows, err := sqlDB.Query(
		"SELECT version_id FROM goose_db_version WHERE is_applied = 1 ORDER BY version_id")
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	var b strings.Builder
	for rows.Next() {
		var v int64
		require.NoError(t, rows.Scan(&v))
		fmt.Fprintf(&b, "%d\n", v)
	}
	require.NoError(t, rows.Err())
	return b.String()
}

func isBookkeeping(name string) bool {
	for _, tbl := range bookkeepingTables {
		if name == tbl {
			return true
		}
	}
	return false
}
