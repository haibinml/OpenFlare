// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package custom_events

import (
	"Wavelet/OpenFlare/plugins/server/admin/push"
	"Wavelet/OpenFlare/plugins/server/listener"
)

// Register wires push notification handlers for domain events and registers
// built-in event metadata. Must be called once during application bootstrap
// before push.SyncEvents.
func Register() {
	push.RegisterBuiltInEvent(AdminLogin)
	listener.OnAdminLoggedIn(handleAdminLogin)
}
