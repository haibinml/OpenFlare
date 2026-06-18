// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package legacy

import (
	"net/http"

	"github.com/Rain-kl/Wavelet/internal/apps/cap"
	"github.com/gin-gonic/gin"
)

type capRedeemRequest struct {
	Token     string `json:"token" binding:"required"`
	Solutions []int  `json:"solutions" binding:"required"`
}

// GetCapChallenge generates a CAP challenge for the legacy frontend.
func GetCapChallenge(c *gin.Context) {
	scope := c.Param("scope")
	if scope == "" {
		scope = c.Query("scope")
	}
	if scope == "" {
		scope = "login"
	}

	mgr := cap.GetDefaultManager()
	resp, err := mgr.Generate(c.Request.Context(), scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, cap.RedeemResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RedeemCapChallenge redeems a CAP challenge for the legacy frontend.
func RedeemCapChallenge(c *gin.Context) {
	scope := c.Param("scope")
	if scope == "" {
		scope = c.Query("scope")
	}
	if scope == "" {
		scope = "login"
	}

	var req capRedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, cap.RedeemResponse{
			Success: false,
			Error:   "无效的参数",
		})
		return
	}

	mgr := cap.GetDefaultManager()
	resp, err := mgr.Redeem(c.Request.Context(), req.Token, req.Solutions, scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, cap.RedeemResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	if !resp.Success {
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	c.JSON(http.StatusOK, resp)
}
