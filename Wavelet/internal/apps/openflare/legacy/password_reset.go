// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	ofauth "github.com/Rain-kl/Wavelet/internal/apps/openflare/auth"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/gin-gonic/gin"
)

// SendPasswordResetEmail sends a password reset email for the legacy frontend.
func SendPasswordResetEmail(c *gin.Context) {
	email := c.Query("email")
	if err := ofauth.SendPasswordResetEmail(c.Request.Context(), email); err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OKMessage(c, "")
}
