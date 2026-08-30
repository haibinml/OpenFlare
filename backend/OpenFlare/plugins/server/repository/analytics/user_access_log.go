// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package analytics

import (
	"context"
	"time"

	analyticsmodel "Wavelet/OpenFlare/plugins/server/model/analytics"
	risklogstore "Wavelet/plugins/domain/risk_control/logstore"
)

func toRiskFilter(filter analyticsmodel.AccessLogFilter) risklogstore.AccessLogFilter {
	return risklogstore.AccessLogFilter{
		UserIDs:   filter.UserIDs,
		Path:      filter.Path,
		StartTime: filter.StartTime,
		EndTime:   filter.EndTime,
	}
}

// BatchInsert writes user access logs via Wavelet risk_control.
func BatchInsert(ctx context.Context, logs []analyticsmodel.UserAccessLog) error {
	return risklogstore.BatchInsert(ctx, logs)
}

// DeleteAllUserAccessLogs truncates user access logs via Wavelet risk_control.
func DeleteAllUserAccessLogs(ctx context.Context) (int64, error) {
	return risklogstore.DeleteAllUserAccessLogs(ctx)
}

// CountAccessLogs counts user access logs via Wavelet risk_control.
func CountAccessLogs(ctx context.Context, filter analyticsmodel.AccessLogFilter) (uint64, error) {
	return risklogstore.CountAccessLogs(ctx, toRiskFilter(filter))
}

// ListAccessLogs lists user access logs via Wavelet risk_control.
func ListAccessLogs(ctx context.Context, filter analyticsmodel.AccessLogFilter, page, pageSize int) ([]analyticsmodel.UserAccessLog, uint64, error) {
	return risklogstore.ListAccessLogs(ctx, toRiskFilter(filter), page, pageSize)
}

// GetDailyTrend returns the daily trend via Wavelet risk_control.
func GetDailyTrend(ctx context.Context, days int) ([]analyticsmodel.DailyTrend, error) {
	src, err := risklogstore.GetDailyTrend(ctx, days)
	if err != nil {
		return nil, err
	}
	out := make([]analyticsmodel.DailyTrend, len(src))
	for i, v := range src {
		out[i] = analyticsmodel.DailyTrend{Date: v.Date, Count: v.Count}
	}
	return out, nil
}

// GetBrowserDistribution returns browser share via Wavelet risk_control.
func GetBrowserDistribution(ctx context.Context, startTime time.Time) ([]analyticsmodel.BrowserShare, error) {
	src, err := risklogstore.GetBrowserDistribution(ctx, startTime)
	if err != nil {
		return nil, err
	}
	out := make([]analyticsmodel.BrowserShare, len(src))
	for i, v := range src {
		out[i] = analyticsmodel.BrowserShare{Browser: v.Browser, Count: v.Count}
	}
	return out, nil
}

// GetTopActiveUsers returns top users via Wavelet risk_control.
func GetTopActiveUsers(ctx context.Context, startTime time.Time, limit int) ([]analyticsmodel.TopUser, error) {
	src, err := risklogstore.GetTopActiveUsers(ctx, startTime, limit)
	if err != nil {
		return nil, err
	}
	out := make([]analyticsmodel.TopUser, len(src))
	for i, v := range src {
		out[i] = analyticsmodel.TopUser{UserID: v.UserID, Count: v.Count}
	}
	return out, nil
}
