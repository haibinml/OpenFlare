// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts legacy OpenFlare dashboard routes.
func RegisterRoutes(apiGroup *gin.RouterGroup) {
	dashboardRoute := apiGroup.Group("/dashboard")
	dashboardRoute.Use(compat.AdminAuth())
	{
		dashboardRoute.GET("/overview", getOverviewHandler)
	}
}

func getOverviewHandler(c *gin.Context) {
	overview, err := GetOverview(c.Request.Context())
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, overview)
}
