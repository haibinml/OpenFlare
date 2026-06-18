// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/agent"
	"github.com/gin-gonic/gin"
)

func registerAgentRoutes(apiGroup *gin.RouterGroup) {
	agent.RegisterRoutes(apiGroup)
}
