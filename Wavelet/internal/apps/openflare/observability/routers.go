// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package observability

import (
	"strconv"

	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts legacy OpenFlare access log routes.
func RegisterRoutes(apiGroup *gin.RouterGroup) {
	accessLogRoute := apiGroup.Group("/access-logs")
	accessLogRoute.Use(compat.AdminAuth())
	{
		compat.RegisterCollection(accessLogRoute, "GET", getAccessLogsHandler)
		accessLogRoute.GET("/folds", getFoldedAccessLogsHandler)
		accessLogRoute.GET("/folds/ip-summary", getFoldedAccessLogIPsHandler)
		accessLogRoute.GET("/ip-summary", getAccessLogIPSummariesHandler)
		accessLogRoute.GET("/ip-summary/trend", getAccessLogIPTrendHandler)
		accessLogRoute.POST("/cleanup", cleanupAccessLogsHandler)
	}
}

func getAccessLogsHandler(c *gin.Context) {
	logs, err := ListAccessLogs(c.Request.Context(), readAccessLogQuery(c))
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, logs)
}

func getFoldedAccessLogsHandler(c *gin.Context) {
	query := readAccessLogQuery(c)
	query.FoldMinutes = readQueryInt(c, "fold_minutes")
	logs, err := ListFoldedAccessLogs(c.Request.Context(), query)
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, logs)
}

func getFoldedAccessLogIPsHandler(c *gin.Context) {
	result, err := ListFoldedAccessLogIPs(c.Request.Context(), FoldedAccessLogIPQuery{
		NodeID:          c.Query("node_id"),
		RemoteAddr:      c.Query("remote_addr"),
		Host:            c.Query("host"),
		Path:            c.Query("path"),
		BucketStartedAt: c.Query("bucket_started_at"),
		FoldMinutes:     readQueryInt(c, "fold_minutes"),
		Page:            readQueryInt(c, "p"),
		PageSize:        readQueryInt(c, "page_size"),
		SortBy:          c.Query("sort_by"),
		SortOrder:       c.Query("sort_order"),
	})
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, result)
}

func getAccessLogIPSummariesHandler(c *gin.Context) {
	result, err := ListAccessLogIPSummaries(c.Request.Context(), AccessLogIPSummaryQuery{
		NodeID:     c.Query("node_id"),
		RemoteAddr: c.Query("remote_addr"),
		Host:       c.Query("host"),
		Page:       readQueryInt(c, "p"),
		PageSize:   readQueryInt(c, "page_size"),
		SortBy:     c.Query("sort_by"),
		SortOrder:  c.Query("sort_order"),
	})
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, result)
}

func getAccessLogIPTrendHandler(c *gin.Context) {
	result, err := GetAccessLogIPTrend(c.Request.Context(), AccessLogIPTrendQuery{
		NodeID:        c.Query("node_id"),
		RemoteAddr:    c.Query("remote_addr"),
		Host:          c.Query("host"),
		Hours:         readQueryInt(c, "hours"),
		BucketMinutes: readQueryInt(c, "bucket_minutes"),
	})
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, result)
}

func cleanupAccessLogsHandler(c *gin.Context) {
	var input AccessLogCleanupInput
	if !compat.BindJSON(c, &input) {
		return
	}
	result, err := CleanupAccessLogs(c.Request.Context(), input)
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, result)
}

func readAccessLogQuery(c *gin.Context) AccessLogQuery {
	return AccessLogQuery{
		NodeID:     c.Query("node_id"),
		RemoteAddr: c.Query("remote_addr"),
		Host:       c.Query("host"),
		Path:       c.Query("path"),
		Page:       readQueryInt(c, "p"),
		PageSize:   readQueryInt(c, "page_size"),
		SortBy:     c.Query("sort_by"),
		SortOrder:  c.Query("sort_order"),
	}
}

func readQueryInt(c *gin.Context, key string) int {
	value, _ := strconv.Atoi(c.DefaultQuery(key, "0"))
	return value
}
