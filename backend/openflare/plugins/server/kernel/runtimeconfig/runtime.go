// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package runtimeconfig holds process-level config values bound from core.Context.
package runtimeconfig

import (
	"sync"

	"Wavelet/plugins/infra/database"
)

// Snapshot is the subset of host config remaining OF packages still need.
type Snapshot struct {
	SessionSecret   string
	DatabaseEnabled bool
	ClickHouse      database.ClickHouseConfig
}

var (
	mu      sync.RWMutex
	current Snapshot
)

// Set replaces the process snapshot.
func Set(s Snapshot) {
	mu.Lock()
	defer mu.Unlock()
	current = s
}

// Get returns the process snapshot.
func Get() Snapshot {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// SessionSecret returns the bound app.session_secret.
func SessionSecret() string {
	return Get().SessionSecret
}

// SetSessionSecret updates the bound session secret.
func SetSessionSecret(secret string) {
	s := Get()
	s.SessionSecret = secret
	Set(s)
}

// DatabaseEnabled reports whether PostgreSQL is the primary store.
func DatabaseEnabled() bool {
	return Get().DatabaseEnabled
}

// ClickHouseEnabled reports whether ClickHouse is enabled.
func ClickHouseEnabled() bool {
	return Get().ClickHouse.Enabled
}

// SetDatabaseEnabled updates the PostgreSQL enabled flag.
func SetDatabaseEnabled(enabled bool) {
	s := Get()
	s.DatabaseEnabled = enabled
	Set(s)
}

// SetClickHouseEnabled updates the ClickHouse enabled flag.
func SetClickHouseEnabled(enabled bool) {
	s := Get()
	s.ClickHouse.Enabled = enabled
	Set(s)
}

// Override replaces DB/CH enablement and returns a restore function.
func Override(databaseEnabled, clickHouseEnabled bool) func() {
	previous := Get()
	SetDatabaseEnabled(databaseEnabled)
	SetClickHouseEnabled(clickHouseEnabled)
	return func() { Set(previous) }
}
