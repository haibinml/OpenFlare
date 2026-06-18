// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/node"
	"github.com/gin-gonic/gin"
)

func registerNodeRoutes(apiGroup *gin.RouterGroup) {
	nodeRoute := apiGroup.Group("/nodes")
	nodeRoute.Use(compat.AdminAuth())
	{
		nodeRoute.GET("/bootstrap-token", node.GetBootstrapTokenHandler)
		nodeRoute.POST("/bootstrap-token/rotate", node.RotateBootstrapTokenHandler)
		nodeRoute.GET("/", node.ListNodesHandler)
		nodeRoute.POST("/", node.CreateNodeHandler)
		nodeRoute.GET("/:id/agent-release", node.GetAgentReleaseHandler)
		nodeRoute.POST("/:id/update", node.UpdateNodeHandler)
		nodeRoute.POST("/:id/delete", node.DeleteNodeHandler)
		nodeRoute.POST("/:id/agent-update", node.RequestAgentUpdateHandler)
		nodeRoute.POST("/:id/openresty-restart", node.RequestOpenrestyRestartHandler)
		nodeRoute.POST("/:id/force-sync", node.RequestForceSyncHandler)
		nodeRoute.GET("/:id/observability", node.GetObservabilityHandler)
		nodeRoute.POST("/:id/observability/cleanup", node.CleanupHealthEventsHandler)
	}
}
