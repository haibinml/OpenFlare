// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package openflare

import (
	"Wavelet/OpenFlare/plugins/server/openflare/apiutil"
	"Wavelet/OpenFlare/plugins/server/openflare/dashboard"
	"Wavelet/core"
)

func registerDashboardRoutes(apiGroup core.RouterExtension) {
	dashboardRoute := apiGroup.Group("/dashboard")
	dashboardRoute.Use(apiutil.AdminMiddlewares()...)
	{
		dashboardRoute.GET("/overview", dashboard.GetOverviewHandler)
	}
}
