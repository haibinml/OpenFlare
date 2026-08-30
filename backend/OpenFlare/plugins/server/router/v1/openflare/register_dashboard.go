// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package openflare

import (
	"Wavelet/OpenFlare/plugins/server/openflare/apiutil"
	"Wavelet/OpenFlare/plugins/server/openflare/dashboard"
	"Wavelet/core"
	"Wavelet/core/contracts"
)

func registerDashboardRoutes(apiGroup core.RouterExtension, auth contracts.AuthService) {
	dashboardRoute := apiGroup.Group("/dashboard")
	dashboardRoute.Use(apiutil.AdminMiddlewares(auth)...)
	{
		dashboardRoute.GET("/overview", dashboard.GetOverviewHandler)
	}
}
