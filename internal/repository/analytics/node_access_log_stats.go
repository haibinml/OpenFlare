// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package analytics

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// NodeAccessLogBucketAggregate is a folded bucket aggregate row.
type NodeAccessLogBucketAggregate struct {
	BucketEpoch      int64
	RequestCount     int64
	SuccessCount     int64
	ClientErrorCount int64
	ServerErrorCount int64
}

// NodeAccessLogBucketDimension is a bucket dimension value.
type NodeAccessLogBucketDimension struct {
	BucketEpoch int64
	Value       string
}

// NodeAccessLogIPAggregate is an IP aggregate row.
type NodeAccessLogIPAggregate struct {
	RemoteAddr       string
	RequestCount     int64
	SuccessCount     int64
	ClientErrorCount int64
	ServerErrorCount int64
	LastSeenEpoch    int64
}

// NodeAccessLogIPSummary is an IP summary row.
type NodeAccessLogIPSummary struct {
	RemoteAddr     string
	TotalRequests  int64
	RecentRequests int64
	LastSeenEpoch  int64
}

// NodeAccessLogIPTrend is an IP trend bucket row.
type NodeAccessLogIPTrend struct {
	BucketEpoch  int64
	RequestCount int64
}

// BucketAggregatesNodeAccessLogs returns folded bucket aggregates.
func BucketAggregatesNodeAccessLogs(ctx context.Context, filter NodeAccessLogFilter, bucketSeconds int64) ([]NodeAccessLogBucketAggregate, error) {
	conn, err := nodeAccessLogConn()
	if err != nil {
		return nil, err
	}
	clause, args := buildNodeAccessLogFilterClause(filter)
	bucketExpr := nodeAccessLogBucketEpochExpr(bucketSeconds)
	tableName := nodeAccessLogTableName()
	sql := fmt.Sprintf(`
SELECT
	%s AS bucket_epoch,
	count() AS request_count,
	countIf(status_code < 400) AS success_count,
	countIf(status_code >= 400 AND status_code < 500) AS client_error_count,
	countIf(status_code >= 500) AS server_error_count
FROM %s
WHERE %s
GROUP BY bucket_epoch`, bucketExpr, tableName, clause)
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("bucket aggregates node access logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []NodeAccessLogBucketAggregate
	for rows.Next() {
		var item NodeAccessLogBucketAggregate
		if err := rows.Scan(&item.BucketEpoch, &item.RequestCount, &item.SuccessCount, &item.ClientErrorCount, &item.ServerErrorCount); err != nil {
			return nil, fmt.Errorf("scan bucket aggregate row: %w", err)
		}
		result = append(result, item)
	}
	return result, nil
}

// BucketDimensionsNodeAccessLogs returns bucket dimension values.
func BucketDimensionsNodeAccessLogs(ctx context.Context, filter NodeAccessLogFilter, column string, bucketSeconds int64) ([]NodeAccessLogBucketDimension, error) {
	conn, err := nodeAccessLogConn()
	if err != nil {
		return nil, err
	}
	clause, args := buildNodeAccessLogFilterClause(filter)
	bucketExpr := nodeAccessLogBucketEpochExpr(bucketSeconds)
	tableName := nodeAccessLogTableName()
	sql := fmt.Sprintf(`
SELECT
	%s AS bucket_epoch,
	trim(%s) AS value
FROM %s
WHERE %s AND trim(%s) != ''
GROUP BY bucket_epoch, trim(%s)`, bucketExpr, column, tableName, clause, column, column)
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("bucket dimensions node access logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []NodeAccessLogBucketDimension
	for rows.Next() {
		var item NodeAccessLogBucketDimension
		if err := rows.Scan(&item.BucketEpoch, &item.Value); err != nil {
			return nil, fmt.Errorf("scan bucket dimension row: %w", err)
		}
		result = append(result, item)
	}
	return result, nil
}

// IPAggregatesNodeAccessLogs returns IP aggregate rows.
func IPAggregatesNodeAccessLogs(ctx context.Context, filter NodeAccessLogFilter, exactRemoteAddr bool) ([]NodeAccessLogIPAggregate, error) {
	conn, err := nodeAccessLogConn()
	if err != nil {
		return nil, err
	}
	clause, args := buildNodeAccessLogFilterClause(filter)
	queryClause := clause
	queryArgs := append([]any{}, args...)
	if exactRemoteAddr {
		trimmed := strings.TrimSpace(filter.RemoteAddr)
		if trimmed == "" {
			return []NodeAccessLogIPAggregate{}, nil
		}
		queryClause = combineNodeAccessLogSQLClauses(queryClause, "trim(remote_addr) = ?")
		queryArgs = append(queryArgs, trimmed)
	}
	lastSeenExpr := nodeAccessLogEpochExpr()
	tableName := nodeAccessLogTableName()
	sql := fmt.Sprintf(`
SELECT
	trim(remote_addr) AS remote_addr,
	count() AS request_count,
	countIf(status_code < 400) AS success_count,
	countIf(status_code >= 400 AND status_code < 500) AS client_error_count,
	countIf(status_code >= 500) AS server_error_count,
	max(%s) AS last_seen_epoch
FROM %s
WHERE %s AND trim(remote_addr) != ''
GROUP BY trim(remote_addr)`, lastSeenExpr, tableName, queryClause)
	rows, err := conn.Query(ctx, sql, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("ip aggregates node access logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []NodeAccessLogIPAggregate
	for rows.Next() {
		var item NodeAccessLogIPAggregate
		if err := rows.Scan(&item.RemoteAddr, &item.RequestCount, &item.SuccessCount, &item.ClientErrorCount, &item.ServerErrorCount, &item.LastSeenEpoch); err != nil {
			return nil, fmt.Errorf("scan ip aggregate row: %w", err)
		}
		result = append(result, item)
	}
	return result, nil
}

// IPSummariesNodeAccessLogs returns IP summary rows.
func IPSummariesNodeAccessLogs(ctx context.Context, filter NodeAccessLogFilter, recentSince time.Time) ([]NodeAccessLogIPSummary, error) {
	conn, err := nodeAccessLogConn()
	if err != nil {
		return nil, err
	}
	clause, args := buildNodeAccessLogFilterClause(filter)
	lastSeenExpr := nodeAccessLogEpochExpr()
	recentClause := "0"
	queryArgs := make([]any, 0, len(args)+1)
	if !recentSince.IsZero() {
		recentClause = "if(logged_at >= ?, 1, 0)"
		queryArgs = append(queryArgs, recentSince)
	}
	queryArgs = append(queryArgs, args...)
	tableName := nodeAccessLogTableName()
	sql := fmt.Sprintf(`
SELECT
	trim(remote_addr) AS remote_addr,
	count() AS total_requests,
	sum(%s) AS recent_requests,
	max(%s) AS last_seen_epoch
FROM %s
WHERE %s AND trim(remote_addr) != ''
GROUP BY trim(remote_addr)`, recentClause, lastSeenExpr, tableName, clause)
	rows, err := conn.Query(ctx, sql, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("ip summaries node access logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []NodeAccessLogIPSummary
	for rows.Next() {
		var item NodeAccessLogIPSummary
		if err := rows.Scan(&item.RemoteAddr, &item.TotalRequests, &item.RecentRequests, &item.LastSeenEpoch); err != nil {
			return nil, fmt.Errorf("scan ip summary row: %w", err)
		}
		result = append(result, item)
	}
	return result, nil
}

// IPTrendNodeAccessLogs returns IP trend bucket rows.
func IPTrendNodeAccessLogs(ctx context.Context, filter NodeAccessLogFilter, bucketSeconds int64) ([]NodeAccessLogIPTrend, error) {
	conn, err := nodeAccessLogConn()
	if err != nil {
		return nil, err
	}
	clause, args := buildNodeAccessLogFilterClause(filter)
	bucketExpr := nodeAccessLogBucketEpochExpr(bucketSeconds)
	tableName := nodeAccessLogTableName()
	sql := fmt.Sprintf(`
SELECT
	%s AS bucket_epoch,
	count() AS request_count
FROM %s
WHERE %s
GROUP BY bucket_epoch
ORDER BY bucket_epoch ASC`, bucketExpr, tableName, clause)
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ip trend node access logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []NodeAccessLogIPTrend
	for rows.Next() {
		var item NodeAccessLogIPTrend
		if err := rows.Scan(&item.BucketEpoch, &item.RequestCount); err != nil {
			return nil, fmt.Errorf("scan ip trend row: %w", err)
		}
		result = append(result, item)
	}
	return result, nil
}