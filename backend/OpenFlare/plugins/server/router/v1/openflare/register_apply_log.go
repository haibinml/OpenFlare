// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package openflare

import (
	"Wavelet/OpenFlare/plugins/server/openflare/apiutil"
	"Wavelet/OpenFlare/plugins/server/openflare/apply_log"
	"Wavelet/core"
)

func registerApplyLogRoutes(apiGroup core.RouterExtension) {
	applyLogRoute := apiGroup.Group("/apply-logs")
	applyLogRoute.Use(apiutil.AdminMiddlewares()...)
	{
		apiutil.RegisterCollection(applyLogRoute, "GET", apply_log.GetApplyLogs)
		applyLogRoute.POST("/cleanup", apply_log.CleanupApplyLogs)
	}
}
