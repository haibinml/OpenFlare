// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package server 装载 OpenFlare 控制面插件：站点/区域、Cloudflare 接入、Pages 部署、
// WAF、节点与回源管理、边缘协议接口以及根级默认路由，全部经 ctx.Router() 声明，
// 由 driver_http 挂载到装配根提供的 gin engine 上。
package server

import (
	"Wavelet/OpenFlare/plugins/server/infra/config"
	router_root "Wavelet/OpenFlare/plugins/server/router/root"
	v1 "Wavelet/OpenFlare/plugins/server/router/v1"
	"Wavelet/core"
)

// Plugin 实现 core.Plugin，是 OpenFlare 控制面的装载入口。
type Plugin struct{}

// New 创建 server 插件。
func New() *Plugin { return &Plugin{} }

// Name 返回插件标识。
func (p *Plugin) Name() string { return "server" }

// Apply 声明控制面 HTTP 路由树。
//
// 中间件（Recovery/CORS/session/otelgin/错误与日志）留在引擎层，由装配根的
// router.BuildEngine 提供；内核暂无 engine 级中间件贡献点，见计划附录 A。
// 任务、设置与迁移仍由 platform/bootstrap 显式装配，待后续迁入 Apply。
func (p *Plugin) Apply(ctx *core.Context) error {
	root := ctx.Router()
	router_root.RegisterRootRoutes(root)

	api := root.Group(config.Config.App.APIPrefix)
	v1.RegisterV1Routes(api.Group("/v1"), api)
	return nil
}
