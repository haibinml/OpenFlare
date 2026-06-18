// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// AgentWSConnectedLastSeenValue is the sentinel last_seen_at value when agent WS is connected.
	AgentWSConnectedLastSeenValue = "__OPENFLARE_AGENT_WS_CONNECTED__"

	agentMessageTypeForceSyncConfig = "force_sync_config"
)

type agentClient struct {
	nodeID string
	conn   *websocket.Conn
	send   chan Message
	done   chan struct{}
	once   sync.Once
}

func (c *agentClient) close() {
	if c == nil {
		return
	}
	c.once.Do(func() {
		close(c.done)
		_ = c.conn.Close()
	})
}

type agentHub struct {
	mu      sync.RWMutex
	clients map[string]*agentClient
}

var defaultAgentHub = &agentHub{clients: make(map[string]*agentClient)}

// ServeAgent handles an upgraded agent websocket connection.
func ServeAgent(c *gin.Context, nodeID string) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Debug("agent ws upgrade failed", "node_id", nodeID, "error", err)
		return
	}

	client := &agentClient{
		nodeID: nodeID,
		conn:   conn,
		send:   make(chan Message, 16),
		done:   make(chan struct{}),
	}
	defaultAgentHub.register(client)
	defer defaultAgentHub.unregister(client)

	slog.Debug("agent ws connected", "node_id", nodeID, "remote", c.Request.RemoteAddr)

	go client.writePump()
	client.readPump()
}

func (h *agentHub) register(client *agentClient) {
	h.mu.Lock()
	if existing := h.clients[client.nodeID]; existing != nil {
		existing.close()
	}
	h.clients[client.nodeID] = client
	h.mu.Unlock()
}

func (h *agentHub) unregister(client *agentClient) {
	h.mu.Lock()
	if current := h.clients[client.nodeID]; current == client {
		delete(h.clients, client.nodeID)
	}
	h.mu.Unlock()
	client.close()
}

// IsAgentConnected reports whether an agent websocket is active.
func IsAgentConnected(nodeID string) bool {
	defaultAgentHub.mu.RLock()
	client := defaultAgentHub.clients[nodeID]
	defaultAgentHub.mu.RUnlock()
	if client == nil {
		return false
	}
	select {
	case <-client.done:
		return false
	default:
		return true
	}
}

// SendForceSyncConfig notifies an agent to force sync configuration.
func SendForceSyncConfig(nodeID string, payload any) bool {
	defaultAgentHub.mu.RLock()
	client := defaultAgentHub.clients[nodeID]
	defaultAgentHub.mu.RUnlock()
	if client == nil {
		return false
	}
	select {
	case <-client.done:
		return false
	case client.send <- Message{Type: agentMessageTypeForceSyncConfig, Payload: payload}:
		return true
	default:
		return false
	}
}

func (c *agentClient) readPump() {
	defer c.close()
	_ = c.conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			slog.Debug("agent ws read closed", "node_id", c.nodeID, "error", err)
			return
		}

		var message Message
		if err = json.Unmarshal(data, &message); err != nil {
			slog.Debug("agent ws invalid message", "node_id", c.nodeID, "error", err)
			continue
		}

		switch message.Type {
		case messageTypePing:
			_ = c.enqueue(Message{Type: messageTypePong})
		case messageTypePong:
		default:
			_ = c.enqueue(Message{Type: messageTypeNotify, Payload: gin.H{
				"echo":    true,
				"type":    message.Type,
				"payload": message.Payload,
			}})
		}
	}
}

func (c *agentClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case message := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteJSON(message); err != nil {
				slog.Debug("agent ws write failed", "node_id", c.nodeID, "error", err)
				c.close()
				return
			}
		case <-ticker.C:
			select {
			case <-c.done:
				return
			case c.send <- Message{Type: messageTypePing}:
			default:
			}
		}
	}
}

func (c *agentClient) enqueue(message Message) bool {
	select {
	case <-c.done:
		return false
	case c.send <- message:
		return true
	default:
		return false
	}
}
