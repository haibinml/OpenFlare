// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	ofauth "github.com/Rain-kl/Wavelet/internal/apps/openflare/auth"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/gin-gonic/gin"
)

// GetStatus returns public server status for the legacy frontend.
func GetStatus(c *gin.Context) {
	data, err := ofauth.BuildPublicStatus(c.Request.Context())
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, data)
}

// GetNotice returns the legacy notice content.
func GetNotice(c *gin.Context) {
	compat.OK(c, ofauth.GetNotice(c.Request.Context()))
}

// GetAbout returns the legacy about content.
func GetAbout(c *gin.Context) {
	compat.OK(c, ofauth.GetAbout(c.Request.Context()))
}
