// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package compat

import (
	"github.com/Rain-kl/Wavelet/internal/apps/oauth"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/gin-gonic/gin"
)

const openFlareTokenHeader = "OpenFlare-Token"

// Role constants mirror legacy OpenFlare role values.
const (
	RoleCommonUser = 1
	RoleAdminUser  = 10
	RoleRootUser   = 100
)

// RequireRole ensures the caller is authenticated with at least minRole.
// Phase 1: supports Wavelet session/access-token via oauth.GetUserFromRequest.
func RequireRole(minRole int) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := oauth.GetUserFromRequest(c)
		if err != nil || user == nil {
			Unauthorized(c, "无权进行此操作，未登录或 token 无效")
			c.Abort()
			return
		}
		role := resolveRole(user)
		if role < minRole {
			Fail(c, "无权进行此操作，权限不足")
			c.Abort()
			return
		}
		c.Set("of_user_id", user.ID)
		c.Set("of_role", role)
		c.Set("of_is_admin", user.IsAdmin)
		c.Next()
	}
}

// UserAuth requires role >= CommonUser.
func UserAuth() gin.HandlerFunc { return RequireRole(RoleCommonUser) }

// AdminAuth requires role >= AdminUser.
func AdminAuth() gin.HandlerFunc { return RequireRole(RoleAdminUser) }

// RootAuth requires role >= RootUser.
func RootAuth() gin.HandlerFunc { return RequireRole(RoleRootUser) }

func resolveRole(user *model.User) int {
	if user == nil {
		return 0
	}
	if user.IsAdmin {
		return RoleRootUser
	}
	return RoleCommonUser
}

// OpenFlareTokenHeader returns the legacy auth header name.
func OpenFlareTokenHeader() string {
	return openFlareTokenHeader
}

// BridgeOpenFlareToken maps OpenFlare-Token to X-Access-Token for legacy clients.
func BridgeOpenFlareToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Access-Token") == "" {
			if token := c.GetHeader(openFlareTokenHeader); token != "" {
				c.Request.Header.Set("X-Access-Token", token)
			}
		}
		c.Next()
	}
}
