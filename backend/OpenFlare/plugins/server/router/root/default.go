// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package root

import (
	publicconfig "Wavelet/OpenFlare/plugins/server/config"
	"Wavelet/OpenFlare/plugins/server/health"
	"Wavelet/OpenFlare/plugins/server/infra/config"
	"Wavelet/OpenFlare/plugins/server/upload"
	"Wavelet/core"
	_ "Wavelet/docs" // Swagger documentation generation setup

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RegisterDefaultRootRoutes registers default routes that belong to the root path.
func RegisterDefaultRootRoutes(r core.RouterExtension) {
	// 1. Serve files by ID
	r.GET("/f/:id", upload.ServeFileByID)

	// 2. Dynamic robots.txt serving
	r.GET("/robots.txt", publicconfig.GetRobotsTXT)

	// 3. Swagger routes (Non-production only)
	if !config.Config.App.IsProduction() {
		r.GET(config.Config.App.APIPrefix+"/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 4. Health check
	r.GET(config.Config.App.APIPrefix+"/health", health.Health)
}
