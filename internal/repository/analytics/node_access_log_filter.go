// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package analytics

import (
	"fmt"
	"strings"
	"time"
)

// NodeAccessLogFilter scopes ClickHouse node access log queries.
type NodeAccessLogFilter struct {
	NodeID     string
	RemoteAddr string
	Host       string
	Path       string
	Since      time.Time
	Until      time.Time
	Page       int
	PageSize   int
	SortBy     string
	SortOrder  string
}

func buildNodeAccessLogFilterClause(filter NodeAccessLogFilter) (string, []any) {
	parts := make([]string, 0, 6)
	args := make([]any, 0, 6)
	if trimmed := strings.TrimSpace(filter.NodeID); trimmed != "" {
		parts = append(parts, "node_id = ?")
		args = append(args, trimmed)
	}
	if trimmed := strings.TrimSpace(filter.RemoteAddr); trimmed != "" {
		parts = append(parts, "remote_addr LIKE ?")
		args = append(args, trimmed+"%")
	}
	if trimmed := strings.TrimSpace(filter.Host); trimmed != "" {
		parts = append(parts, "host LIKE ?")
		args = append(args, trimmed+"%")
	}
	if trimmed := strings.TrimSpace(filter.Path); trimmed != "" {
		parts = append(parts, "path LIKE ?")
		args = append(args, trimmed+"%")
	}
	if !filter.Since.IsZero() {
		parts = append(parts, "logged_at >= ?")
		args = append(args, filter.Since.UTC())
	}
	if !filter.Until.IsZero() {
		parts = append(parts, "logged_at < ?")
		args = append(args, filter.Until.UTC())
	}
	if len(parts) == 0 {
		return "1", nil
	}
	return strings.Join(parts, " AND "), args
}

func combineNodeAccessLogSQLClauses(left string, right string) string {
	if strings.TrimSpace(left) == "" || left == "TRUE" || left == "1" {
		return right
	}
	return left + " AND " + right
}

func nodeAccessLogOrderClause(sortBy string, sortOrder string) string {
	direction := "DESC"
	if normalizeNodeAccessLogSortOrder(sortOrder) == "asc" {
		direction = "ASC"
	}
	column := "logged_at"
	switch strings.TrimSpace(sortBy) {
	case "status_code":
		column = "status_code"
	case "remote_addr":
		column = "remote_addr"
	case "host":
		column = "host"
	case "path":
		column = "path"
	}
	if column == "logged_at" {
		return column + " " + direction + ", id " + direction
	}
	return column + " " + direction + ", logged_at " + direction + ", id " + direction
}

func normalizeNodeAccessLogSortOrder(sortOrder string) string {
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		return "asc"
	}
	return "desc"
}

func nodeAccessLogBucketEpochExpr(bucketSeconds int64) string {
	return fmt.Sprintf("toInt64(intDiv(toUnixTimestamp(logged_at), %d) * %d)", bucketSeconds, bucketSeconds)
}

func nodeAccessLogEpochExpr() string {
	return "toInt64(toUnixTimestamp(logged_at))"
}

func nodeAccessLogTableName() string {
	return "of_node_access_logs"
}