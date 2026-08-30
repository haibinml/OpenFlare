// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"Wavelet/OpenFlare/plugins/server/domain/option"
	"Wavelet/OpenFlare/plugins/server/kernel/apiutil"
	"Wavelet/core"
	"Wavelet/core/contracts"
)

func registerOptionRoutes(apiGroup core.RouterExtension, auth contracts.AuthService) {
	apiGroup.GET("/status", option.GetStatusHandler)

	optionRoute := apiGroup.Group("/option")
	optionRoute.Use(apiutil.AdminMiddlewares(auth)...)
	{
		apiutil.RegisterCollection(optionRoute, "GET", option.ListOptionsHandler)
		optionRoute.POST("/update", option.UpdateOptionHandler)
		optionRoute.POST("/update-batch", option.UpdateOptionsBatchHandler)
		optionRoute.POST("/geoip/lookup", option.LookupGeoIPHandler)
	}

	uptimeKumaRoute := apiGroup.Group("/uptimekuma")
	uptimeKumaRoute.Use(apiutil.AdminMiddlewares(auth)...)
	{
		uptimeKumaRoute.POST("/sync", option.SyncUptimeKumaHandler)
	}
}
