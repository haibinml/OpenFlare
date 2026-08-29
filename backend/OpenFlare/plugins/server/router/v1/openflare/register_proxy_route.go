// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package openflare

import (
	"Wavelet/OpenFlare/plugins/server/openflare/apiutil"
	"Wavelet/OpenFlare/plugins/server/openflare/proxy_route"
	"Wavelet/core"
)

func registerProxyRouteRoutes(apiGroup core.RouterExtension) {
	proxyRouteGroup := apiGroup.Group("/proxy-routes")
	proxyRouteGroup.Use(apiutil.AdminMiddlewares()...)
	{
		apiutil.RegisterCollection(proxyRouteGroup, "GET", proxy_route.GetProxyRoutes)
		proxyRouteGroup.GET("/:id", proxy_route.GetProxyRouteHandler)
		apiutil.RegisterCollection(proxyRouteGroup, "POST", proxy_route.CreateProxyRouteHandler)
		proxyRouteGroup.POST("/:id/update", proxy_route.UpdateProxyRouteHandler)
		proxyRouteGroup.POST("/:id/delete", proxy_route.DeleteProxyRouteHandler)
	}
}
