// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"log"
	"time"

	"Wavelet/OpenFlare/plugins/server"
	"Wavelet/OpenFlare/plugins/server/infra/config"
	"Wavelet/OpenFlare/plugins/server/platform/bootstrap"
	"Wavelet/OpenFlare/plugins/server/router"
	"Wavelet/core"
	"Wavelet/plugins/drivers/driver_http"
)

// defaultShutdownTimeout 在配置未给出优雅退出超时时兜底。
const defaultShutdownTimeout = 10 * time.Second

// runHTTPApp 以 Cordis 方式启动控制面：装配根持有 gin engine（引擎级中间件与前端
// SPA 兜底），server 插件经 ctx.Router() 声明路由，driver_http 负责监听与优雅退出。
// 本函数阻塞至收到退出信号。
func runHTTPApp(mode string) {
	engine := router.BuildEngine()
	timeout := shutdownTimeout()

	app := core.NewApp(
		core.WithProfile(core.ProfileAPI),
		core.WithShutdownTimeout(timeout),
	)
	httpDriver := driver_http.New(
		driver_http.WithEngine(engine),
		driver_http.WithAddr(config.Config.App.Addr),
	)
	app.Use(server.New())
	app.Use(httpDriver)

	if err := app.Prepare(); err != nil {
		log.Fatalf("[API] prepare failed: %v\n", err)
	}

	printStartupBanner(startupState{
		mode:           mode,
		relationalDB:   latestMigrationState.relationalDB,
		clickHouseDB:   latestMigrationState.clickHouseDB,
		listensForHTTP: true,
	})

	if err := app.Run(); err != nil {
		log.Printf("[API] server failed: %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	bootstrap.Stop(ctx)
	log.Println("[API] server exited")
}

func shutdownTimeout() time.Duration {
	if config.Config.App.GracefulShutdownTimeout <= 0 {
		return defaultShutdownTimeout
	}
	return time.Duration(config.Config.App.GracefulShutdownTimeout) * time.Second
}
