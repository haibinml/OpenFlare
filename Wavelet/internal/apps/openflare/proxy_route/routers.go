// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package proxy_route

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
		compat.Fail(c, errProxyRouteNotFound)
		return true
	}
	compat.Fail(c, err.Error())
	return true
}

// GetProxyRoutes 列出全部代理规则。
func GetProxyRoutes(c *gin.Context) {
	routes, err := ListProxyRoutes(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, routes)
}

// GetProxyRouteHandler 获取代理规则详情。
func GetProxyRouteHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	route, err := GetProxyRoute(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, route)
}

// CreateProxyRouteHandler 创建代理规则。
func CreateProxyRouteHandler(c *gin.Context) {
	var input Input
	if !compat.BindJSON(c, &input) {
		return
	}
	route, err := CreateProxyRoute(c.Request.Context(), input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, route)
}

// UpdateProxyRouteHandler 更新代理规则。
func UpdateProxyRouteHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var input Input
	if !compat.BindJSON(c, &input) {
		return
	}
	route, err := UpdateProxyRoute(c.Request.Context(), id, input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, route)
}

// DeleteProxyRouteHandler 删除代理规则。
func DeleteProxyRouteHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	if err := DeleteProxyRoute(c.Request.Context(), id); handleLogicError(c, err) {
		return
	}
	compat.OK(c, nil)
}
