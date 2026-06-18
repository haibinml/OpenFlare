// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/apply_log"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/gin-gonic/gin"
)

func registerApplyLogRoutes(apiGroup *gin.RouterGroup) {
	applyLogRoute := apiGroup.Group("/apply-logs")
	applyLogRoute.Use(compat.AdminAuth())
	{
		compat.RegisterCollection(applyLogRoute, "GET", apply_log.GetApplyLogs)
		applyLogRoute.POST("/cleanup", apply_log.CleanupApplyLogs)
	}
}
