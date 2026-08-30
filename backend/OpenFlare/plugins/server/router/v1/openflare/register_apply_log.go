// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package openflare

import (
	"Wavelet/OpenFlare/plugins/server/openflare/apiutil"
	"Wavelet/OpenFlare/plugins/server/openflare/apply_log"
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
