// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	"strconv"
	"strings"

	ofauth "github.com/Rain-kl/Wavelet/internal/apps/openflare/auth"
	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/gin-gonic/gin"
)

type linkExistingRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// OAuthAuthorize starts OAuth authorization for a legacy auth source.
func OAuthAuthorize(c *gin.Context) {
	url, err := ofauth.OAuthAuthorize(c.Request.Context(), c, c.Param("source"))
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, gin.H{"authorize_url": url})
}

// OAuthCallback handles the legacy GET OAuth callback.
func OAuthCallback(c *gin.Context) {
	result, err := ofauth.OAuthCallback(c.Request.Context(), c, c.Param("source"))
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, result)
}

// LinkExistingOAuthAccount binds a pending OAuth account to an existing user.
func LinkExistingOAuthAccount(c *gin.Context) {
	var req linkExistingRequest
	if !compat.BindJSON(c, &req) {
		return
	}
	result, err := ofauth.LinkExistingOAuthAccount(c.Request.Context(), c, ofauth.LinkExistingInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, result)
}

// ListExternalAccounts returns external account bindings for the current user.
func ListExternalAccounts(c *gin.Context) {
	userID := callerUserID(c)
	accounts, err := model.ListExternalAccountsByUserID(c.Request.Context(), userID)
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, accounts)
}

// DeleteExternalAccount removes an external account binding.
func DeleteExternalAccount(c *gin.Context) {
	userID := callerUserID(c)
	if userID == 0 {
		compat.Unauthorized(c, "无权进行此操作，未登录或 token 无效")
		return
	}
	rawID := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || id == 0 {
		compat.Fail(c, "绑定记录 ID 无效")
		return
	}
	if err := model.DeleteExternalAccountForUser(c.Request.Context(), id, userID); err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OKMessage(c, "")
}
