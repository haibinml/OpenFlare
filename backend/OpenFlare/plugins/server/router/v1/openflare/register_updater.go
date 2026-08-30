// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package openflare

import (
	"Wavelet/OpenFlare/plugins/server/updater"
	"Wavelet/OpenFlare/plugins/server/openflare/apiutil"
	"Wavelet/core"
	"Wavelet/core/contracts"
)

func registerUpdaterRoutes(apiV1Router core.RouterExtension, auth contracts.AuthService) {
	// Wavelet admin already registered these paths. Replace them so gin does
	// not panic on duplicate method+path at driver_http Start, and so the
	// OpenFlare updater implementation is the one that runs.
	_ = apiV1Router.Unregister("GET", "/admin/update")
	_ = apiV1Router.Unregister("POST", "/admin/update/apply")

	adminRouter := apiV1Router.Group("/admin")
	adminRouter.Use(apiutil.AdminMiddlewares(auth)...)
	update := adminRouter.Group("/update")
	{
		update.GET("", updater.GetUpdateStatus)
		update.POST("/apply", updater.ApplyUpdate)
	}
}
