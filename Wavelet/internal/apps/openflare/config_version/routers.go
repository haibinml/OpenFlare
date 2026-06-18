// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package config_version

import (
	"errors"

	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleLogicError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		compat.Fail(c, "记录不存在")
		return true
	}
	compat.Fail(c, err.Error())
	return true
}

// ListConfigVersionsHandler lists config versions.
func ListConfigVersionsHandler(c *gin.Context) {
	versions, err := ListConfigVersions(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, versions)
}

// GetConfigVersionHandler returns a config version by id.
func GetConfigVersionHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	version, err := GetConfigVersionDetail(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, version)
}

// GetActiveConfigVersionHandler returns the active config version.
func GetActiveConfigVersionHandler(c *gin.Context) {
	version, err := GetActiveConfigVersion(c.Request.Context())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			compat.Fail(c, errNoActiveVersion)
			return
		}
		handleLogicError(c, err)
		return
	}
	compat.OK(c, version)
}

// PreviewConfigVersionHandler previews the current draft configuration.
func PreviewConfigVersionHandler(c *gin.Context) {
	preview, err := PreviewConfigVersion(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, preview)
}

// DiffConfigVersionHandler diffs the current draft against the active version.
func DiffConfigVersionHandler(c *gin.Context) {
	diff, err := DiffConfigVersion(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, diff)
}

// PublishConfigVersionHandler publishes a new config version.
func PublishConfigVersionHandler(c *gin.Context) {
	username := c.GetString("username")
	force := c.Query("force") == "true"
	version, err := PublishConfigVersion(c.Request.Context(), username, force)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, version)
}

// ActivateConfigVersionHandler activates an existing config version.
func ActivateConfigVersionHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	version, err := ActivateConfigVersion(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, version)
}

// CleanupConfigVersionsHandler removes old inactive config versions.
func CleanupConfigVersionsHandler(c *gin.Context) {
	var input CleanupInput
	if !compat.BindJSON(c, &input) {
		return
	}
	result, err := CleanupConfigVersions(c.Request.Context(), input.KeepCount)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, result)
}
