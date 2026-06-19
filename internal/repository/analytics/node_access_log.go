// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/Rain-kl/Wavelet/internal/db"
	analyticsmodel "github.com/Rain-kl/Wavelet/internal/model/analytics"
)

// NodeAccessLogRegionCount aggregates access log regions.
type NodeAccessLogRegionCount struct {
	Region string
	Count  int64
}

func nodeAccessLogConn() (driver.Conn, error) {
	if db.ChConn == nil {
		return nil, fmt.Errorf("clickhouse connection is not initialized")
	}
	return db.ChConn, nil
}

// ListNodeAccessLogs returns access logs matching filter.
func ListNodeAccessLogs(ctx context.Context, filter NodeAccessLogFilter) ([]analyticsmodel.NodeAccessLog, error) {
	conn, err := nodeAccessLogConn()
	if err != nil {
		return nil, err
	}
	clause, args := buildNodeAccessLogFilterClause(filter)
	tableName := nodeAccessLogTableName()
	sql := fmt.Sprintf(`
SELECT id, node_id, logged_at, remote_addr, region, host, path, status_code, created_at
FROM %s
WHERE %s
ORDER BY %s`, tableName, clause, nodeAccessLogOrderClause(filter.SortBy, filter.SortOrder))
	if filter.PageSize > 0 {
		if filter.Page < 0 {
			filter.Page = 0
		}
		sql += " LIMIT ? OFFSET ?"
		args = append(args, filter.PageSize, filter.Page*filter.PageSize)
	}
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list node access logs: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanNodeAccessLogRows(rows)
}

func scanNodeAccessLogRows(rows driver.Rows) ([]analyticsmodel.NodeAccessLog, error) {
	var result []analyticsmodel.NodeAccessLog
	for rows.Next() {
		var item analyticsmodel.NodeAccessLog
		if err := rows.Scan(
			&item.ID,
			&item.NodeID,
			&item.LoggedAt,
			&item.RemoteAddr,
			&item.Region,
			&item.Host,
			&item.Path,
			&item.StatusCode,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan node access log row: %w", err)
		}
		item.LoggedAt = item.LoggedAt.UTC()
		item.CreatedAt = item.CreatedAt.UTC()
		result = append(result, item)
	}
	return result, nil
}

// CountNodeAccessLogs returns total records and distinct IPs matching filter.
func CountNodeAccessLogs(ctx context.Context, filter NodeAccessLogFilter) (int64, int64, error) {
	conn, err := nodeAccessLogConn()
	if err != nil {
		return 0, 0, err
	}
	clause, args := buildNodeAccessLogFilterClause(filter)
	tableName := nodeAccessLogTableName()

	var totalRecords int64
	countSQL := fmt.Sprintf("SELECT count() FROM %s WHERE %s", tableName, clause)
	if err := conn.QueryRow(ctx, countSQL, args...).Scan(&totalRecords); err != nil {
		return 0, 0, fmt.Errorf("count node access logs: %w", err)
	}

	ipSQL := fmt.Sprintf(`
SELECT count() FROM (
	SELECT trim(remote_addr) AS remote_addr
	FROM %s
	WHERE %s AND remote_addr != ''
	GROUP BY trim(remote_addr)
)`, tableName, clause)
	var totalIPs int64
	if err := conn.QueryRow(ctx, ipSQL, args...).Scan(&totalIPs); err != nil {
		return 0, 0, fmt.Errorf("count node access log ips: %w", err)
	}
	return totalRecords, totalIPs, nil
}

// RegionCountsNodeAccessLogs returns region counts for a node since a time.
func RegionCountsNodeAccessLogs(ctx context.Context, nodeID string, since time.Time, limit int) ([]NodeAccessLogRegionCount, error) {
	conn, err := nodeAccessLogConn()
	if err != nil {
		return nil, err
	}
	filter := NodeAccessLogFilter{NodeID: nodeID, Since: since}
	clause, args := buildNodeAccessLogFilterClause(filter)
	tableName := nodeAccessLogTableName()
	sql := fmt.Sprintf(`
SELECT trim(region) AS region, count() AS count
FROM %s
WHERE %s AND trim(region) != ''
GROUP BY trim(region)
ORDER BY count DESC, region ASC`, tableName, clause)
	if limit > 0 {
		sql += " LIMIT ?"
		args = append(args, limit)
	}
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("region counts node access logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []NodeAccessLogRegionCount
	for rows.Next() {
		var item NodeAccessLogRegionCount
		if err := rows.Scan(&item.Region, &item.Count); err != nil {
			return nil, fmt.Errorf("scan region count row: %w", err)
		}
		result = append(result, item)
	}
	return result, nil
}