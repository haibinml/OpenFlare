// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package origin

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
		compat.Fail(c, errOriginNotFound)
		return true
	}
	compat.Fail(c, err.Error())
	return true
}

// GetOrigins 列出全部源站。
func GetOrigins(c *gin.Context) {
	origins, err := ListOrigins(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, origins)
}

// GetOrigin 获取源站详情。
func GetOrigin(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	detail, err := GetOriginDetail(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, detail)
}

// CreateOriginHandler 创建源站。
func CreateOriginHandler(c *gin.Context) {
	var input Input
	if !compat.BindJSON(c, &input) {
		return
	}
	origin, err := CreateOrigin(c.Request.Context(), input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, origin)
}

// UpdateOriginHandler 更新源站。
func UpdateOriginHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var input Input
	if !compat.BindJSON(c, &input) {
		return
	}
	origin, err := UpdateOrigin(c.Request.Context(), id, input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, origin)
}

// DeleteOriginHandler 删除源站。
func DeleteOriginHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	if err := DeleteOrigin(c.Request.Context(), id); handleLogicError(c, err) {
		return
	}
	compat.OK(c, nil)
}
