// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"Wavelet/OpenFlare/plugins/server/infra/config"
	"Wavelet/core"

	"github.com/gin-gonic/gin"
)

// baselineRoutesFile 是改造前遗留注册路径导出的 (方法 路径) 全集。
const baselineRoutesFile = "docs/superpowers/specs/baseline/routes-engine.txt"

// TestPluginRoutesMatchLegacyBaseline 保证 server 插件经 ctx.Router() 声明的路由
// 与迁移前的 gin 路由表逐条一致：少一条即接口消失，多一条即路由漂移。
func TestPluginRoutesMatchLegacyBaseline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if config.Config.App.APIPrefix == "" {
		config.Config.App.APIPrefix = "/api"
	}

	ctx := core.NewContext(context.Background())
	if err := New().Apply(ctx); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	got := routeSet(ctx)
	if out := os.Getenv("OF_DUMP_ROUTES"); out != "" {
		if err := os.WriteFile(out, []byte(strings.Join(sortedKeys(got), "\n")+"\n"), 0o600); err != nil {
			t.Fatalf("write route dump: %v", err)
		}
	}

	want := loadBaseline(t)
	missing := make([]string, 0)
	extra := make([]string, 0)
	for k := range want {
		if !got[k] {
			missing = append(missing, k)
		}
	}
	for k := range got {
		if !want[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)

	if len(missing) > 0 || len(extra) > 0 {
		t.Errorf("route table drifted: baseline=%d plugin=%d\nmissing (%d):\n  %s\nunexpected (%d):\n  %s",
			len(want), len(got), len(missing), strings.Join(missing, "\n  "),
			len(extra), strings.Join(extra, "\n  "))
	}
}

func routeSet(ctx *core.Context) map[string]bool {
	set := make(map[string]bool)
	for _, rd := range ctx.Router().Routes() {
		set[rd.Method+" "+rd.Path] = true
	}
	return set
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func loadBaseline(t *testing.T) map[string]bool {
	t.Helper()
	path := locateFile(t, baselineRoutesFile)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read baseline %s: %v", path, err)
	}
	set := make(map[string]bool)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			set[line] = true
		}
	}
	if len(set) == 0 {
		t.Fatalf("baseline %s is empty", path)
	}
	return set
}

// locateFile 从测试文件所在目录向上查找，避免依赖测试运行深度。
func locateFile(t *testing.T, rel string) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	for range 8 {
		candidate := filepath.Join(dir, rel)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		dir = filepath.Join(dir, "..")
	}
	t.Fatalf("%s not found above %s", rel, filepath.Dir(thisFile))
	return ""
}
