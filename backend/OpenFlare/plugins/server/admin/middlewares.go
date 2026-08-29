// Copyright 2025 linux.do
// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"Wavelet/OpenFlare/plugins/server/model"
	"Wavelet/pkg/logger"
	"Wavelet/pkg/response"
	otel_trace "Wavelet/pkg/trace"

	"Wavelet/OpenFlare/plugins/server/oauth"

	"github.com/gin-gonic/gin"
)

// LoginAdminRequired 返回管理员权限校验中间件
func LoginAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// init trace
		ctx, span := otel_trace.Start(c.Request.Context(), "LoginAdminRequired")
		defer span.End()

		user, _ := oauth.GetFromContext[*model.User](c, oauth.UserObjKey)

		// 如果是通过 Access Token 鉴权，需要检查令牌本身是否具有管理员权限
		if tokenAuth, _ := oauth.GetFromContext[bool](c, oauth.TokenAuthKey); tokenAuth {
			tokenAdmin, _ := oauth.GetFromContext[bool](c, oauth.TokenAdminKey)
			if !tokenAdmin {
				response.AbortNotFound(c, TokenAdminRequired)
				return
			}
		}

		if !user.IsAdmin {
			response.AbortNotFound(c, AdminRequired)
			return
		}

		// log
		logger.InfoF(ctx, "[LoginAdminRequired] %d %s", user.ID, user.Username)

		// next
		c.Next()
	}
}
