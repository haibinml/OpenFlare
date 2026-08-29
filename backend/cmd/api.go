// Copyright 2025 linux.do
// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"Wavelet/OpenFlare/plugins/server/platform/bootstrap"

	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "wavelet API",
	Run: func(_ *cobra.Command, _ []string) {
		bootstrap.RegisterAPI()
		runBootstrap(bootstrap.Options{API: true})
		runHTTPApp("API")
	},
}
