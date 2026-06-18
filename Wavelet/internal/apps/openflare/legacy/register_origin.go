// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/origin"
	"github.com/gin-gonic/gin"
)

func registerOriginRoutes(apiGroup *gin.RouterGroup) {
	originRoute := apiGroup.Group("/origins")
	originRoute.Use(compat.AdminAuth())
	{
		compat.RegisterCollection(originRoute, "GET", origin.GetOrigins)
		originRoute.GET("/:id", origin.GetOrigin)
		compat.RegisterCollection(originRoute, "POST", origin.CreateOriginHandler)
		originRoute.POST("/:id/update", origin.UpdateOriginHandler)
		originRoute.POST("/:id/delete", origin.DeleteOriginHandler)
	}
}
