// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package compat

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// IDParam parses :id from the URL path.
func IDParam(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	if raw == "" {
		Fail(c, "无效的 ID")
		return 0, false
	}
	id64, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id64 == 0 {
		Fail(c, "无效的 ID")
		return 0, false
	}
	return uint(id64), true
}

// BindJSON binds JSON body; returns false after writing a failure response.
func BindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		Fail(c, "参数错误")
		return false
	}
	return true
}
