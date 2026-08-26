// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

// wsClientCore holds the state and lifecycle shared by all WebSocket client
// variants (agent/relay/flared). Embed it; call close exactly-once semantics
// are guaranteed via once.
type wsClientCore struct {
	nodeID string
	conn   *websocket.Conn
	send   chan Message
	done   chan struct{}
	once   sync.Once
}

// close tears down the connection at most once.
func (c *wsClientCore) close() {
	if c == nil {
		return
	}
	c.once.Do(func() {
		close(c.done)
		if c.conn != nil {
			_ = c.conn.Close()
		}
	})
}

// enqueue best-effort delivers message; it never blocks and fails fast when
// the client is closed or its send buffer is full.
func (c *wsClientCore) enqueue(message Message) bool {
	select {
	case <-c.done:
		return false
	case c.send <- message:
		return true
	default:
		return false
	}
}
