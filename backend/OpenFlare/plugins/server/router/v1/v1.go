// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package v1 contains router registrations for API V1
package v1

import (
	ofrouter "Wavelet/OpenFlare/plugins/server/router/v1/openflare"
	"Wavelet/core"
	"Wavelet/core/contracts"
)

// RegisterV1Routes registers OpenFlare business routes under API V1.
// Platform user/admin/cap/health routes are owned by Wavelet domain plugins.
func RegisterV1Routes(apiV1Router core.RouterExtension, auth contracts.AuthService) {
	ofrouter.RegisterV1Routes(apiV1Router, auth)
	ofrouter.RegisterRoutes(apiV1Router, auth)
	RegisterCustomRoutes()
}
