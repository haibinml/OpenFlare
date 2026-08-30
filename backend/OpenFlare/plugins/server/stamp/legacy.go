// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package stamp records pre-Cordis schema versions into w_schema_versions.
package stamp

import "Wavelet/core"

// Legacy is the pre-Cordis schema stamp hook. Task 14 fills the body; until then
// it is a no-op so plugin migrations can proceed on a fresh database.
func Legacy(*core.Context) error {
	return nil
}
