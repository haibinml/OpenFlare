// Copyright 2025 linux.do
// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package router

import (
	"log"
	"strconv"

	router_root "Wavelet/OpenFlare/plugins/server/router/root"

	"Wavelet/OpenFlare/plugins/server/infra/config"
	"Wavelet/OpenFlare/plugins/server/oauth"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// BuildEngine 构造已挂全部中间件与路由的 gin engine，但不监听端口。
// 独立出来是为了让 Cordis 装配根可以用 driver_http.WithEngine 复用同一台引擎
// （内核目前没有 engine 级中间件贡献点，见 docs/superpowers/plans 附录 A）。
func BuildEngine() *gin.Engine {
	// 运行模式
	if config.Config.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化路由
	r := gin.New()
	// Legacy OpenFlare list endpoints register both /resource and /resource/; disable auto slash redirects.
	r.RedirectTrailingSlash = false
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	cfg := config.Config.Redis
	addrs := cfg.Addrs
	sessionAddr := "localhost:6379"
	if len(addrs) > 0 {
		sessionAddr = addrs[0]
	}

	sessionStore, err := redis.NewStoreWithDB(
		cfg.MinIdleConn,
		"tcp",
		sessionAddr,
		cfg.Username,
		cfg.Password,
		strconv.Itoa(cfg.DB),
		[]byte(config.Config.App.SessionSecret),
	)
	if err != nil {
		log.Fatalf("[API] init session store failed: %v\n", err)
	}

	// 设置 Session Redis Key 前缀
	if cfg.KeyPrefix != "" {
		if err := redis.SetKeyPrefix(sessionStore, cfg.KeyPrefix+"session:"); err != nil {
			log.Printf("[API] set session key prefix failed: %v\n", err)
		}
	}

	sessionStore.Options(oauth.GetSessionOptions(config.Config.App.SessionAge))

	r.Use(sessions.Sessions(config.Config.App.SessionCookieName, sessionStore))

	// 补充中间件
	r.Use(otelgin.Middleware(config.Config.App.AppName), errorHandlerMiddleware(), loggerMiddleware())

	// 前端 SPA 兜底（NoRoute + 静态资源）只有 gin engine 才能表达，
	// 内核暂无 NoRoute 贡献点；其余路由全部由 server 插件经 ctx.Router() 声明。
	router_root.RegisterFrontend(r)
	return r
}
