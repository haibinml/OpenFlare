// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package root registers custom business routes and frontend serving.
package root

import (
	"Wavelet/core"

	"github.com/gin-gonic/gin"
)

// RegisterFrontend is a package-level variable overridden by frontend.go when built with embed_frontend.
//
// 前端 SPA 兜底只能表达在 gin engine 上（NoRoute + 静态资源），内核暂无 NoRoute
// 贡献点，因此由装配根的 router.BuildEngine 调用，不走插件的声明式路由。
var RegisterFrontend = func(_ *gin.Engine) {
	// No-op by default
}

// RegisterRootRoutes registers the root-level API routes declared by the server plugin.
func RegisterRootRoutes(r core.RouterExtension) {
	// 1. Default root routes (/f/:id, /robots.txt, and /swagger/*any)
	RegisterDefaultRootRoutes(r)

	// 2. Register custom serving
	RegisterCustomRootRoutes(r)
}
