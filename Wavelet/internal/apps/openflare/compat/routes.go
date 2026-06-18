// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package compat

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterCollection registers a collection endpoint on both "" and "/" so requests
// work with or without a trailing slash. This avoids 301/308 redirect loops between
// Next.js dev proxy and Gin when paths disagree on trailing slashes.
//
// When the group base already ends with "/" (e.g. userGroup.Group("/")), only "/" is
// registered to avoid duplicate route panic.
func RegisterCollection(route *gin.RouterGroup, method string, handlers ...gin.HandlerFunc) {
	route.Handle(method, "/", handlers...)
	if !strings.HasSuffix(route.BasePath(), "/") {
		route.Handle(method, "", handlers...)
	}
}
