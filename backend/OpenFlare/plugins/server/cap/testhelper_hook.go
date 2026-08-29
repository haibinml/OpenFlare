// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package cap

import "Wavelet/OpenFlare/plugins/server/testhelper"

func init() {
	testhelper.RegisterCleanup(ResetRuntimeSettingsForTest)
}
