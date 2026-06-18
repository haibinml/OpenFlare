// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package compat provides OpenFlare legacy API compatibility helpers.
package compat

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope is the legacy OpenFlare frontend response format.
type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// OK sends a successful legacy response.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Success: true, Message: "", Data: data})
}

// OKMessage sends a successful legacy response with a message.
func OKMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Envelope{Success: true, Message: message, Data: nil})
}

// Fail sends a failed legacy response with HTTP 200 (OpenFlare convention).
func Fail(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Envelope{Success: false, Message: message, Data: nil})
}

// Unauthorized sends a 401 legacy response.
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Envelope{Success: false, Message: message, Data: nil})
}

// Forbidden sends a 403 legacy response.
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Envelope{Success: false, Message: message, Data: nil})
}
