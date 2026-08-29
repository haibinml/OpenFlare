// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package openflare

import (
	"Wavelet/OpenFlare/plugins/server/openflare/apiutil"
	"Wavelet/OpenFlare/plugins/server/openflare/origin"
	"Wavelet/core"
)

func registerOriginRoutes(apiGroup core.RouterExtension) {
	originRoute := apiGroup.Group("/origins")
	originRoute.Use(apiutil.AdminMiddlewares()...)
	{
		apiutil.RegisterCollection(originRoute, "GET", origin.GetOrigins)
		originRoute.GET("/:id", origin.GetOrigin)
		apiutil.RegisterCollection(originRoute, "POST", origin.CreateOriginHandler)
		originRoute.POST("/:id/update", origin.UpdateOriginHandler)
		originRoute.POST("/:id/delete", origin.DeleteOriginHandler)
	}
}
