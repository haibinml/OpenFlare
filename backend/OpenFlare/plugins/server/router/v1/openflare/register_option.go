// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package openflare

import (
	"Wavelet/OpenFlare/plugins/server/openflare/apiutil"
	"Wavelet/OpenFlare/plugins/server/openflare/option"
	"Wavelet/core"
)

func registerOptionRoutes(apiGroup core.RouterExtension) {
	apiGroup.GET("/status", option.GetStatusHandler)

	optionRoute := apiGroup.Group("/option")
	optionRoute.Use(apiutil.AdminMiddlewares()...)
	{
		apiutil.RegisterCollection(optionRoute, "GET", option.ListOptionsHandler)
		optionRoute.POST("/update", option.UpdateOptionHandler)
		optionRoute.POST("/update-batch", option.UpdateOptionsBatchHandler)
		optionRoute.POST("/geoip/lookup", option.LookupGeoIPHandler)
	}

	uptimeKumaRoute := apiGroup.Group("/uptimekuma")
	uptimeKumaRoute.Use(apiutil.AdminMiddlewares()...)
	{
		uptimeKumaRoute.POST("/sync", option.SyncUptimeKumaHandler)
	}
}
