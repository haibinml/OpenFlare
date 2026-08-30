// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package server 装载 OpenFlare 控制面业务路由：站点/区域、Cloudflare 接入、Pages 部署、
// WAF、节点与回源管理以及边缘协议接口。平台路径（user/admin/cap/health）由 Wavelet
// 域插件注册，本插件只挂 OpenFlare 自己的业务。
package server

import (
	"Wavelet/OpenFlare/plugins/server/migrate"
	"Wavelet/OpenFlare/plugins/server/ofevents"
	"Wavelet/OpenFlare/plugins/server/openflare/chwriter"
	ofgeoip "Wavelet/OpenFlare/plugins/server/openflare/geoip"
	"Wavelet/OpenFlare/plugins/server/openflare/ofupload"
	"Wavelet/OpenFlare/plugins/server/publicconfig"
	"Wavelet/OpenFlare/plugins/server/repository"
	"Wavelet/OpenFlare/plugins/server/repository/logstore"
	ofrouter "Wavelet/OpenFlare/plugins/server/router/v1/openflare"
	"Wavelet/OpenFlare/plugins/server/runtimeconfig"
	oftask "Wavelet/OpenFlare/plugins/server/task"
	"Wavelet/core"
	"Wavelet/core/contracts"
	"Wavelet/pkg/logger"
	"Wavelet/plugins/infra/database"
	"context"
	"embed"
	"reflect"

	"Wavelet/OpenFlare/plugins/server/openflare/credential"
	_ "Wavelet/docs"
	adminservice "Wavelet/plugins/domain/admin/service"

	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//go:embed migrate/postgres/*.sql migrate/sqlite/*.sql
var serverMigrations embed.FS

// Plugin 实现 core.Plugin，是 OpenFlare 控制面的装载入口。
type Plugin struct{}

// New 创建 server 插件。
func New() *Plugin { return &Plugin{} }

// Name 返回插件标识。
func (p *Plugin) Name() string { return "server" }

// Inject waits for platform services so Apply runs after Wavelet domain plugins.
func (p *Plugin) Inject() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[contracts.DBService](),
		reflect.TypeFor[contracts.CacheService](),
		reflect.TypeFor[contracts.UserService](),
		reflect.TypeFor[contracts.AuthService](),
		reflect.TypeFor[contracts.TaskService](),
		reflect.TypeFor[contracts.StorageService](),
	}
}

// Apply 声明 OpenFlare 业务 HTTP 路由树、公共配置、推送事件与异步任务。
func (p *Plugin) Apply(ctx *core.Context) error {
	var chCfg database.ClickHouseConfig
	_ = ctx.Config().Bind("clickhouse", &chCfg)
	runtimeconfig.Set(runtimeconfig.Snapshot{
		SessionSecret:   ctx.Config().String("app.session_secret", ""),
		DatabaseEnabled: ctx.Config().Bool("database.enabled", false),
		ClickHouse:      chCfg,
	})
	credential.SetSessionSecret(runtimeconfig.SessionSecret())

	ctx.Migrations().Register("server", serverMigrations)
	if err := migrate.UpClickHouse(); err != nil {
		return err
	}

	if ts, err := core.Inject[contracts.TaskService](ctx); err == nil && ts != nil {
		oftask.SetService(ts)
	} else {
		core.When[contracts.TaskService](ctx, oftask.SetService)
	}
	if user, err := core.Inject[contracts.UserService](ctx); err == nil && user != nil {
		repository.SetUserService(user)
	} else {
		core.When[contracts.UserService](ctx, repository.SetUserService)
	}
	if storage, err := core.Inject[contracts.StorageService](ctx); err == nil && storage != nil {
		ofupload.SetStorage(storage)
	} else {
		core.When[contracts.StorageService](ctx, ofupload.SetStorage)
	}

	core.Provide[contracts.PublicConfigProvider](ctx, publicconfig.New(ctx))
	if pr, err := core.Inject[contracts.PushRegistry](ctx); err == nil {
		for _, meta := range ofevents.All() {
			pr.RegisterBuiltInEvent(meta)
		}
		_ = pr.SyncEvents(ctx.GoContext())
	}

	registerOpenFlareTasks(ctx)
	bindLogstore(ctx)

	if err := ofgeoip.EnsureRuntimeProvider(ctx.GoContext()); err != nil {
		logger.ErrorF(ctx.GoContext(), "[server] init GeoIP provider failed: %v", err)
	}
	chwriter.Init(ctx.GoContext())

	var auth contracts.AuthService
	if err := core.Using[contracts.AuthService](ctx, func(s contracts.AuthService) { auth = s }); err != nil {
		return err
	}
	repository.SetAuthService(auth)
	ofrouter.RegisterV1Routes(ctx.Router().Group("/api/v1"), auth)
	ofrouter.RegisterRoutes(ctx.Router().Group("/api/v1"), auth)

	ctx.Router().GET("/robots.txt", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(adminservice.RobotsTxtBody(c.Request.Context())))
	})
	env := ctx.Config().String("app.env", "production")
	if env != "production" && env != "prod" {
		ctx.Router().GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	return nil
}

func bindLogstore(ctx *core.Context) {
	logstore.SetConfigReader(func(goCtx context.Context, key string) (string, error) {
		cfg, err := repository.GetSystemConfigByKey(goCtx, key)
		if err != nil {
			return "", err
		}
		return cfg.Value, nil
	})
	logstore.Init(ctx.GoContext())
}
