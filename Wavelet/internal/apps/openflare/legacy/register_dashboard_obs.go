// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/dashboard"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/observability"
	"github.com/gin-gonic/gin"
)

func registerDashboardObsRoutes(apiGroup *gin.RouterGroup) {
	dashboard.RegisterRoutes(apiGroup)
	observability.RegisterRoutes(apiGroup)
}
