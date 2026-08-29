// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package openflare

import (
	"Wavelet/OpenFlare/plugins/server/openflare/flared"
	"Wavelet/OpenFlare/plugins/server/openflare/relay"
	"Wavelet/core"
)

func registerRelayRoutes(apiV1Router core.RouterExtension) {
	relayRoute := apiV1Router.Group("/relay")
	relayRoute.Use(relay.Auth())
	{
		relayRoute.POST("/heartbeat", relay.PostHeartbeat)
		relayRoute.GET("/ws", relay.GetWebSocket)
	}
}

func registerTunnelRoutes(apiV1Router core.RouterExtension) {
	tunnelRoute := apiV1Router.Group("/tunnel")
	tunnelRoute.Use(flared.TunnelAuth())
	{
		tunnelRoute.POST("/heartbeat", flared.PostHeartbeat)
		tunnelRoute.GET("/config/active", flared.GetActiveConfig)
		tunnelRoute.POST("/apply-log", flared.PostApplyLog)
		tunnelRoute.GET("/ws", flared.GetWebSocket)
	}
}
