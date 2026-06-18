// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/option"
	"github.com/gin-gonic/gin"
)

func registerOptionRoutes(apiGroup *gin.RouterGroup) {
	option.RegisterRoutes(apiGroup)
}
