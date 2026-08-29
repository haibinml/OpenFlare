// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package openflare

import (
	"Wavelet/OpenFlare/plugins/server/openflare/apiutil"
	"Wavelet/OpenFlare/plugins/server/openflare/zone"
	"Wavelet/core"
)

func registerZoneRoutes(apiGroup core.RouterExtension) {
	zoneGroup := apiGroup.Group("/zones")
	zoneGroup.Use(apiutil.AdminMiddlewares()...)
	apiutil.RegisterCollection(zoneGroup, "GET", zone.ListHandler)
	apiutil.RegisterCollection(zoneGroup, "POST", zone.CreateHandler)
	zoneGroup.GET("/:id/overview", zone.GetOverviewHandler)
	zoneGroup.GET("/:id/stats", zone.GetStatsHandler)
	zoneGroup.POST("/:id/update", zone.UpdateHandler)
	zoneGroup.POST("/:id/delete", zone.DeleteHandler)
	zoneGroup.POST("/:id/domains", zone.CreateDomainHandler)
	zoneGroup.POST("/:id/domains/:domainId/update", zone.UpdateDomainHandler)
	zoneGroup.POST("/:id/domains/:domainId/delete", zone.DeleteDomainHandler)
}
