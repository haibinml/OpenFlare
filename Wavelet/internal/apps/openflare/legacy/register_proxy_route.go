// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/proxy_route"
	"github.com/gin-gonic/gin"
)

func registerProxyRouteRoutes(apiGroup *gin.RouterGroup) {
	proxyRouteGroup := apiGroup.Group("/proxy-routes")
	proxyRouteGroup.Use(compat.AdminAuth())
	{
		compat.RegisterCollection(proxyRouteGroup, "GET", proxy_route.GetProxyRoutes)
		proxyRouteGroup.GET("/:id", proxy_route.GetProxyRouteHandler)
		compat.RegisterCollection(proxyRouteGroup, "POST", proxy_route.CreateProxyRouteHandler)
		proxyRouteGroup.POST("/:id/update", proxy_route.UpdateProxyRouteHandler)
		proxyRouteGroup.POST("/:id/delete", proxy_route.DeleteProxyRouteHandler)
	}
}
