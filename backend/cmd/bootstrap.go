// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"

	"Wavelet/OpenFlare/plugins/server/platform/bootstrap"
	"Wavelet/pkg/trace"
)

func runBootstrap(opts bootstrap.Options) {
	ctx, span := trace.Start(context.Background(), "bootstrap.Init")
	defer span.End()
	bootstrap.Init(ctx, opts)
}
