// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"path/filepath"
	"testing"

	"Wavelet/core"
)

func testSource(t *testing.T) core.ConfigSource {
	t.Helper()
	return core.NewMapSource(map[string]any{
		"app": map[string]any{
			"addr": "127.0.0.1:0",
			"env":  "testing",
		},
		"redis": map[string]any{
			"enabled": false,
		},
		"database": map[string]any{
			"enabled":     false,
			"sqlite_path": filepath.Join(t.TempDir(), "openflare-cmd.db"),
		},
	})
}

func TestNewOpenFlareAppRegistersServerAndWaveletUser(t *testing.T) {
	app := newOpenFlareApp(core.ProfileAPI, core.WithConfigSource(testSource(t)))
	if err := app.Prepare(); err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, p := range app.Plugins() {
		names[p.Name()] = true
	}
	for _, n := range []string{"user", "auth", "cap", "admin", "server"} {
		if !names[n] {
			t.Errorf("missing plugin %s", n)
		}
	}

	if err := app.Reconcile(); err != nil {
		t.Fatal(err)
	}

	got := map[string]bool{}
	for _, rd := range app.Context().Router().Routes() {
		got[rd.Method+" "+rd.Path] = true
	}
	for _, want := range []string{
		"GET /api/health",
		"GET /api/v1/user/self",
		"GET /api/v1/d/nodes",
		"POST /api/cap/challenge",
	} {
		if !got[want] {
			t.Errorf("missing route %s", want)
		}
	}
}
