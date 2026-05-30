package controller

import (
	"encoding/json"
	"net/http"
	"openflare/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type wafIDsRequest struct {
	IDs []uint `json:"ids"`
}

func ListWAFRuleGroups(c *gin.Context) {
	groups, err := service.ListWAFRuleGroups()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": groups})
}

func GetWAFRuleGroup(c *gin.Context) {
	id, ok := parseUintPathParam(c, "id")
	if !ok {
		return
	}
	group, err := service.GetWAFRuleGroup(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": group})
}

func CreateWAFRuleGroup(c *gin.Context) {
	var input service.WAFRuleGroupInput
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid payload"})
		return
	}
	group, err := service.CreateWAFRuleGroup(input)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": group})
}

func UpdateWAFRuleGroup(c *gin.Context) {
	id, ok := parseUintPathParam(c, "id")
	if !ok {
		return
	}
	var input service.WAFRuleGroupInput
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid payload"})
		return
	}
	group, err := service.UpdateWAFRuleGroup(id, input)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": group})
}

func DeleteWAFRuleGroup(c *gin.Context) {
	id, ok := parseUintPathParam(c, "id")
	if !ok {
		return
	}
	if err := service.DeleteWAFRuleGroup(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

func ReplaceWAFRuleGroupSites(c *gin.Context) {
	id, ok := parseUintPathParam(c, "id")
	if !ok {
		return
	}
	var request wafIDsRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid payload"})
		return
	}
	group, err := service.ReplaceWAFRuleGroupSites(id, request.IDs)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": group})
}

func GetWAFSiteRuleGroups(c *gin.Context) {
	routeID, ok := parseUintPathParam(c, "route_id")
	if !ok {
		return
	}
	view, err := service.GetWAFSiteRuleGroups(routeID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": view})
}

func ReplaceWAFSiteRuleGroups(c *gin.Context) {
	routeID, ok := parseUintPathParam(c, "route_id")
	if !ok {
		return
	}
	var request wafIDsRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid payload"})
		return
	}
	view, err := service.ReplaceWAFSiteRuleGroups(routeID, request.IDs)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": view})
}

func parseUintPathParam(c *gin.Context, name string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return 0, false
	}
	return uint(id), true
}
