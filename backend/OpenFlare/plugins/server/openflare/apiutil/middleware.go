// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package apiutil

import (
	"Wavelet/OpenFlare/plugins/server/admin"
	"Wavelet/OpenFlare/plugins/server/oauth"
)

// AdminMiddlewares returns Wavelet-standard middlewares for OpenFlare console routes.
// OpenFlare no longer distinguishes Admin vs Root tiers; all management endpoints share
// the same gate: user.IsAdmin for session users, token_admin for Access Token callers.
//
// 返回 []any 而非 []gin.HandlerFunc：内核 RouterExtension.Use 收 ...any，
// 而 Go 不允许把 []T 直接展开成 ...any。
func AdminMiddlewares() []any {
	return []any{oauth.LoginRequired(), admin.LoginAdminRequired()}
}
