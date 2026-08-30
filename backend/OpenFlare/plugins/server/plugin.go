// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package server 装载 OpenFlare 控制面业务路由：站点/区域、Cloudflare 接入、Pages 部署、
// WAF、节点与回源管理以及边缘协议接口。平台路径（user/admin/cap/health）由 Wavelet
// 域插件注册，本插件只挂 OpenFlare 自己的业务。
package server

import (
	ofrouter "Wavelet/OpenFlare/plugins/server/router/v1/openflare"
	"Wavelet/core"
	"Wavelet/core/contracts"
	"reflect"
)

// Plugin 实现 core.Plugin，是 OpenFlare 控制面的装载入口。
type Plugin struct{}

// New 创建 server 插件。
func New() *Plugin { return &Plugin{} }

// Name 返回插件标识。
func (p *Plugin) Name() string { return "server" }

// Inject waits for the same services as the Wavelet admin plugin so Apply runs
// after platform admin routes exist and OpenFlare can replace /admin/update*.
func (p *Plugin) Inject() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[contracts.DBService](),
		reflect.TypeFor[contracts.CacheService](),
		reflect.TypeFor[contracts.UserService](),
		reflect.TypeFor[contracts.AuthService](),
	}
}

// Apply 声明 OpenFlare 业务 HTTP 路由树。
func (p *Plugin) Apply(ctx *core.Context) error {
	var auth contracts.AuthService
	if err := core.Using[contracts.AuthService](ctx, func(s contracts.AuthService) { auth = s }); err != nil {
		return err
	}
	ofrouter.RegisterV1Routes(ctx.Router().Group("/api/v1"), auth)
	ofrouter.RegisterRoutes(ctx.Router().Group("/api/v1"), auth)
	return nil
}
