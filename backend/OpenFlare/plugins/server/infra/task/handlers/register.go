// Copyright 2025 linux.do
// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package handlers 注册异步任务处理器
package handlers

import (
	"Wavelet/OpenFlare/plugins/server/admin/push"
	"Wavelet/OpenFlare/plugins/server/infra/task"
	"Wavelet/OpenFlare/plugins/server/openflare"
	cf "Wavelet/OpenFlare/plugins/server/openflare/cloudflare"
	"Wavelet/OpenFlare/plugins/server/openflare/pages"
	"Wavelet/OpenFlare/plugins/server/openflare/tasks"
	"Wavelet/OpenFlare/plugins/server/openflare/tls"
	"Wavelet/OpenFlare/plugins/server/upload"
	"Wavelet/OpenFlare/plugins/server/user"
)

// Register registers all built-in task handlers and their metadata.
func Register() {
	task.RegisterHandler(upload.StorageMigrationTask, &upload.MigrationHandler{})
	task.RegisterTaskMeta(upload.StorageMigrationMeta)

	// system cleanup
	task.RegisterHandler(upload.SystemCleanupTask, &upload.SystemCleanupHandler{})
	task.RegisterTaskMeta(upload.SystemCleanupMeta)

	// upload
	task.RegisterHandler(upload.WarmImageCacheTask, &upload.WarmImageCacheHandler{})
	task.RegisterTaskMeta(upload.WarmImageCacheMeta)

	task.RegisterHandler(upload.RebuildUploadStatsTask, &upload.RebuildUploadStatsHandler{})
	task.RegisterTaskMeta(upload.RebuildUploadStatsMeta)

	// user
	task.RegisterHandler(user.SendEmailTask, &user.SendEmailHandler{})
	task.RegisterTaskMeta(user.SendEmailMeta)

	// push
	task.RegisterHandler(push.SendNotificationTask, &push.PushHandler{})
	task.RegisterTaskMeta(push.SendNotificationMeta)

	// openflare
	task.RegisterHandler(openflare.SSLRenewTask, &openflare.SSLRenewHandler{})
	task.RegisterTaskMeta(openflare.SSLRenewMeta)

	task.RegisterHandler(openflare.WAFIPGroupSyncTask, &openflare.WAFIPGroupSyncHandler{})
	task.RegisterTaskMeta(openflare.WAFIPGroupSyncMeta)

	task.RegisterHandler(openflare.UptimeKumaSyncTask, &openflare.UptimeKumaSyncHandler{})
	task.RegisterTaskMeta(openflare.UptimeKumaSyncMeta)

	task.RegisterHandler(openflare.LogDBSwitchTask, &tasks.LogDBSwitchHandler{})
	task.RegisterTaskMeta(openflare.LogDBSwitchMeta)

	task.RegisterHandler(cf.SyncMemberTask, &cf.SyncMemberTaskHandler{})
	task.RegisterTaskMeta(cf.SyncMemberMeta)
	task.RegisterHandler(cf.SyncGroupTask, &cf.SyncGroupTaskHandler{})
	task.RegisterTaskMeta(cf.SyncGroupMeta)
	task.RegisterHandler(cf.SyncByNodeTask, &cf.SyncByNodeTaskHandler{})
	task.RegisterTaskMeta(cf.SyncByNodeMeta)

	// pages source actions are only dispatched by the Pages domain API/scanner.
	task.RegisterHandler(pages.PagesSourceScanTask, &pages.SourceScanHandler{})
	task.RegisterTaskMeta(pages.PagesSourceScanMeta)

	task.RegisterHandler(pages.PagesSourceActionTask, &pages.SourceActionHandler{})
	task.RegisterTaskMeta(pages.PagesSourceActionMeta)

	// tls single renew
	task.RegisterHandler(tls.SSLSingleRenewTask, &tls.SSLSingleRenewHandler{})
	task.RegisterTaskMeta(tls.SSLSingleRenewMeta)
}
