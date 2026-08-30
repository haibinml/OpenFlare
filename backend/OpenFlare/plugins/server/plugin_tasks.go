// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	"Wavelet/OpenFlare/plugins/server/openflare"
	cf "Wavelet/OpenFlare/plugins/server/openflare/cloudflare"
	"Wavelet/OpenFlare/plugins/server/openflare/pages"
	"Wavelet/OpenFlare/plugins/server/openflare/tasks"
	"Wavelet/OpenFlare/plugins/server/openflare/tls"
	oftask "Wavelet/OpenFlare/plugins/server/task"
	"Wavelet/core"
	"Wavelet/core/extpoints"
)

func registerOpenFlareTasks(ctx *core.Context) {
	registerOFTask(ctx, openflare.SSLRenewTask, &openflare.SSLRenewHandler{}, openflare.SSLRenewMeta)
	registerOFTask(ctx, openflare.WAFIPGroupSyncTask, &openflare.WAFIPGroupSyncHandler{}, openflare.WAFIPGroupSyncMeta)
	registerOFTask(ctx, openflare.UptimeKumaSyncTask, &openflare.UptimeKumaSyncHandler{}, openflare.UptimeKumaSyncMeta)
	registerOFTask(ctx, openflare.LogDBSwitchTask, &tasks.LogDBSwitchHandler{}, openflare.LogDBSwitchMeta)

	registerOFTask(ctx, cf.SyncMemberTask, &cf.SyncMemberTaskHandler{}, cf.SyncMemberMeta)
	registerOFTask(ctx, cf.SyncGroupTask, &cf.SyncGroupTaskHandler{}, cf.SyncGroupMeta)
	registerOFTask(ctx, cf.SyncByNodeTask, &cf.SyncByNodeTaskHandler{}, cf.SyncByNodeMeta)

	registerOFTask(ctx, pages.PagesSourceScanTask, &pages.SourceScanHandler{}, pages.PagesSourceScanMeta)
	registerOFTask(ctx, pages.PagesSourceActionTask, &pages.SourceActionHandler{}, pages.PagesSourceActionMeta)

	registerOFTask(ctx, tls.SSLSingleRenewTask, &tls.SSLSingleRenewHandler{}, tls.SSLSingleRenewMeta)
}

func registerOFTask(ctx *core.Context, pattern string, handler oftask.TaskHandler, meta oftask.TaskMeta) {
	opts := []extpoints.TaskOption{
		extpoints.WithTaskMeta(meta.ToDTO()),
	}
	if meta.MaxRetry > 0 {
		opts = append(opts, extpoints.WithTaskRetry(meta.MaxRetry))
	}
	ctx.Task().Register(pattern, func(c context.Context, payload []byte) error {
		_, err := handler.Execute(c, payload)
		return err
	}, opts...)
}
