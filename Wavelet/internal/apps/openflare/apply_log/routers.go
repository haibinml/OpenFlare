// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package apply_log

import (
	"strconv"

	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/gin-gonic/gin"
)

// GetApplyLogs lists apply logs with pagination and optional node_id filter.
func GetApplyLogs(c *gin.Context) {
	result, err := ListPage(c.Request.Context(), ListQuery{
		NodeID:   c.Query("node_id"),
		PageNo:   readIntQuery(c, "pageNo", "page_no"),
		PageSize: readIntQuery(c, "pageSize", "page_size"),
	})
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, result)
}

// CleanupApplyLogs removes old apply logs or deletes all records.
func CleanupApplyLogs(c *gin.Context) {
	var input CleanupInput
	if !compat.BindJSON(c, &input) {
		return
	}

	result, err := Cleanup(c.Request.Context(), input)
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, result)
}

func readIntQuery(c *gin.Context, primary, secondary string) int {
	value := c.Query(primary)
	if value == "" {
		value = c.Query(secondary)
	}
	parsed, _ := strconv.Atoi(value)
	return parsed
}
