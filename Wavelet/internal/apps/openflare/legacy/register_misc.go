// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/update"
	"github.com/gin-gonic/gin"
)

func registerMiscRoutes(apiGroup *gin.RouterGroup) {
	updateRoute := apiGroup.Group("/update")
	updateRoute.Use(compat.RootAuth())
	{
		updateRoute.GET("/latest-release", update.GetLatestReleaseHandler)
		updateRoute.GET("/logs/ws", update.StreamServerUpgradeLogsHandler)
		updateRoute.POST("/manual-upload", update.UploadManualServerBinaryHandler)
		updateRoute.POST("/manual-upgrade", update.ConfirmManualServerUpgradeHandler)
		updateRoute.POST("/upgrade", update.UpgradeServerHandler)
	}
}
