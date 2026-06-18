// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"time"

	"github.com/Rain-kl/Wavelet/internal/model"
)

const (
	nodeStatusOnline  = "online"
	nodeStatusOffline = "offline"
	nodeStatusPending = "pending"
)

func computeNodeStatus(node *model.OpenFlareNode) string {
	if node == nil {
		return nodeStatusOffline
	}
	if node.LastSeenAt == nil || node.LastSeenAt.IsZero() {
		return nodeStatusPending
	}
	if time.Since(*node.LastSeenAt) > model.NodeOfflineThreshold {
		return nodeStatusOffline
	}
	return nodeStatusOnline
}

func nodeViewLastSeenAt(node *model.OpenFlareNode) any {
	if node == nil || node.LastSeenAt == nil {
		return time.Time{}
	}
	return *node.LastSeenAt
}
