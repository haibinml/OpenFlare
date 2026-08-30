// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"Wavelet/OpenFlare/plugins/server/domain/site/apply_log"
	"Wavelet/OpenFlare/plugins/server/kernel/apiutil"
	"Wavelet/core"
	"Wavelet/core/contracts"
)

func registerApplyLogRoutes(apiGroup core.RouterExtension, auth contracts.AuthService) {
	applyLogRoute := apiGroup.Group("/apply-logs")
	applyLogRoute.Use(apiutil.AdminMiddlewares(auth)...)
	{
		apiutil.RegisterCollection(applyLogRoute, "GET", apply_log.GetApplyLogs)
		applyLogRoute.POST("/cleanup", apply_log.CleanupApplyLogs)
	}
}
