// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package legacy registers OpenFlare /api/* compatibility routes for the old frontend.
package legacy

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts all OpenFlare legacy API routes under the /api group.
func RegisterRoutes(apiGroup *gin.RouterGroup) {
	registerAuthRoutes(apiGroup)
	registerOptionRoutes(apiGroup)
	registerOriginRoutes(apiGroup)
	registerApplyLogRoutes(apiGroup)
	registerProxyRouteRoutes(apiGroup)
	registerNodeRoutes(apiGroup)
	registerWAFRoutes(apiGroup)
	registerTLSRoutes(apiGroup)
	registerConfigVersionRoutes(apiGroup)
	registerAgentRoutes(apiGroup)
	registerPagesRoutes(apiGroup)
	registerRelayFlaredRoutes(apiGroup)
	registerDashboardObsRoutes(apiGroup)
	registerMiscRoutes(apiGroup)
}
