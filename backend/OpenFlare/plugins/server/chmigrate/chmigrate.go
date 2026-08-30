// Copyright 2025 linux.do
// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package chmigrate applies ClickHouse of_node_* goose migrations.
package chmigrate

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"log"
	"time"

	"Wavelet/OpenFlare/plugins/server/runtimeconfig"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/pressly/goose/v3"
)

const (
	clickhouseMigrationDir      = "goose/clickhouse"
	clickhouseGooseVersionTable = "goose_clickhouse_version"
	clickhouseMaxExecTime       = 60
	clickhouseReadTimeoutFactor = 2
)

//go:embed goose/clickhouse/*.sql
var clickhouseMigrationFS embed.FS

// Up runs of_node_* goose migrations against ClickHouse when enabled.
func Up() error {
	if !runtimeconfig.ClickHouseEnabled() {
		return nil
	}

	cfg := runtimeconfig.Get().ClickHouse
	sqlDB := clickhouse.OpenDB(&clickhouse.Options{
		Addr: cfg.Hosts,
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": clickhouseMaxExecTime,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:     time.Duration(cfg.DialTimeout) * time.Second,
		MaxOpenConns:    cfg.MaxOpenConn,
		MaxIdleConns:    cfg.MaxIdleConn,
		ConnMaxLifetime: time.Duration(cfg.ConnMaxLifetime) * time.Second,
		ReadTimeout:     time.Duration(cfg.DialTimeout*clickhouseReadTimeoutFactor) * time.Second,
		BlockBufferSize: cfg.BlockBufferSize,
	})

	subFS, err := fs.Sub(clickhouseMigrationFS, clickhouseMigrationDir)
	if err != nil {
		closeClickHouseDB(sqlDB)
		log.Printf("[ClickHouse] get sub fs failed: %v\n", err)
		return err
	}

	provider, err := goose.NewProvider(
		"clickhouse",
		sqlDB,
		subFS,
		goose.WithTableName(clickhouseGooseVersionTable),
		goose.WithDisableGlobalRegistry(true),
	)
	if err != nil {
		closeClickHouseDB(sqlDB)
		log.Printf("[ClickHouse] create goose provider failed: %v\n", err)
		return err
	}
	if _, err := provider.GetDBVersion(context.Background()); err != nil {
		closeClickHouseDB(sqlDB)
		log.Printf("[ClickHouse] get goose version failed: %v\n", err)
		return err
	}

	if _, err := provider.Up(context.Background()); err != nil {
		closeClickHouseDB(sqlDB)
		log.Printf("[ClickHouse] goose migrate failed: %v\n", err)
		return err
	}
	if _, err := provider.GetDBVersion(context.Background()); err != nil {
		closeClickHouseDB(sqlDB)
		log.Printf("[ClickHouse] get migrated goose version failed: %v\n", err)
		return err
	}
	closeClickHouseDB(sqlDB)

	log.Println("[ClickHouse] goose migrate success")
	return nil
}

func closeClickHouseDB(sqlDB *sql.DB) {
	if err := sqlDB.Close(); err != nil {
		log.Printf("[ClickHouse] close sql db failed: %v\n", err)
	}
}
