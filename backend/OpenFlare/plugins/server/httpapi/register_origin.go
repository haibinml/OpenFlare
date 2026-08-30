// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"Wavelet/OpenFlare/plugins/server/domain/site/origin"
	"Wavelet/OpenFlare/plugins/server/kernel/apiutil"
	"Wavelet/core"
	"Wavelet/core/contracts"
)

func registerOriginRoutes(apiGroup core.RouterExtension, auth contracts.AuthService) {
	originRoute := apiGroup.Group("/origins")
	originRoute.Use(apiutil.AdminMiddlewares(auth)...)
	{
		apiutil.RegisterCollection(originRoute, "GET", origin.GetOrigins)
		originRoute.GET("/:id", origin.GetOrigin)
		apiutil.RegisterCollection(originRoute, "POST", origin.CreateOriginHandler)
		originRoute.POST("/:id/update", origin.UpdateOriginHandler)
		originRoute.POST("/:id/delete", origin.DeleteOriginHandler)
	}
}
