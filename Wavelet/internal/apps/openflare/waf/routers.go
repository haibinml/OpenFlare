// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package waf

import (
	"errors"
	"strconv"

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

func routeIDParam(c *gin.Context) (uint, bool) {
	raw := c.Param("route_id")
	if raw == "" {
		compat.Fail(c, "invalid id")
		return 0, false
	}
	id64, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id64 == 0 {
		compat.Fail(c, "invalid id")
		return 0, false
	}
	return uint(id64), true
}

// ListRuleGroupsHandler lists all WAF rule groups.
func ListRuleGroupsHandler(c *gin.Context) {
	groups, err := ListRuleGroups(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, groups)
}

// GetRuleGroupHandler returns a WAF rule group by id.
func GetRuleGroupHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	group, err := GetRuleGroup(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, group)
}

// CreateRuleGroupHandler creates a WAF rule group.
func CreateRuleGroupHandler(c *gin.Context) {
	var input RuleGroupInput
	if !compat.BindJSON(c, &input) {
		return
	}
	group, err := CreateRuleGroup(c.Request.Context(), input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, group)
}

// UpdateRuleGroupHandler updates a WAF rule group.
func UpdateRuleGroupHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var input RuleGroupInput
	if !compat.BindJSON(c, &input) {
		return
	}
	group, err := UpdateRuleGroup(c.Request.Context(), id, input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, group)
}

// DeleteRuleGroupHandler deletes a WAF rule group.
func DeleteRuleGroupHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	if err := DeleteRuleGroup(c.Request.Context(), id); handleLogicError(c, err) {
		return
	}
	compat.OKMessage(c, "")
}

// ReplaceRuleGroupSitesHandler replaces site bindings for a rule group.
func ReplaceRuleGroupSitesHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var request IDsRequest
	if !compat.BindJSON(c, &request) {
		return
	}
	group, err := ReplaceRuleGroupSites(c.Request.Context(), id, request.IDs)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, group)
}

// GetSiteRuleGroupsHandler returns WAF rule groups for a proxy route.
func GetSiteRuleGroupsHandler(c *gin.Context) {
	routeID, ok := routeIDParam(c)
	if !ok {
		return
	}
	view, err := GetSiteRuleGroups(c.Request.Context(), routeID)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, view)
}

// ReplaceSiteRuleGroupsHandler replaces rule group bindings for a proxy route.
func ReplaceSiteRuleGroupsHandler(c *gin.Context) {
	routeID, ok := routeIDParam(c)
	if !ok {
		return
	}
	var request IDsRequest
	if !compat.BindJSON(c, &request) {
		return
	}
	view, err := ReplaceSiteRuleGroups(c.Request.Context(), routeID, request.IDs)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, view)
}

// ListIPGroupsHandler lists all WAF IP groups.
func ListIPGroupsHandler(c *gin.Context) {
	groups, err := ListIPGroups(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, groups)
}

// GetIPGroupHandler returns a WAF IP group by id.
func GetIPGroupHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	group, err := GetIPGroup(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, group)
}

// CreateIPGroupHandler creates a WAF IP group.
func CreateIPGroupHandler(c *gin.Context) {
	var input IPGroupInput
	if !compat.BindJSON(c, &input) {
		return
	}
	group, err := CreateIPGroup(c.Request.Context(), input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, group)
}

// UpdateIPGroupHandler updates a WAF IP group.
func UpdateIPGroupHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var input IPGroupInput
	if !compat.BindJSON(c, &input) {
		return
	}
	group, err := UpdateIPGroup(c.Request.Context(), id, input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, group)
}

// DeleteIPGroupHandler deletes a WAF IP group.
func DeleteIPGroupHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	if err := DeleteIPGroup(c.Request.Context(), id); handleLogicError(c, err) {
		return
	}
	compat.OKMessage(c, "")
}

// SyncIPGroupHandler triggers a stub sync for a WAF IP group.
func SyncIPGroupHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	result, err := SyncIPGroup(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, result)
}

// TestIPGroupAutoConfigHandler tests automatic IP group configuration (stub).
func TestIPGroupAutoConfigHandler(c *gin.Context) {
	var input IPGroupAutoTestInput
	if !compat.BindJSON(c, &input) {
		return
	}
	result, err := TestIPGroupAutoConfig(c.Request.Context(), input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, result)
}
