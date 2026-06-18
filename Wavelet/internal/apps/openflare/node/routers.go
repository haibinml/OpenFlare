// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleLogicError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		compat.Fail(c, errNodeNotFound)
		return true
	}
	compat.Fail(c, err.Error())
	return true
}

// ListNodesHandler lists all nodes.
func ListNodesHandler(c *gin.Context) {
	nodes, err := ListNodes(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, nodes)
}

// CreateNodeHandler creates a node.
func CreateNodeHandler(c *gin.Context) {
	var input Input
	if !compat.BindJSON(c, &input) {
		return
	}
	view, err := CreateNode(c.Request.Context(), input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, view)
}

// UpdateNodeHandler updates a node.
func UpdateNodeHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var input Input
	if !compat.BindJSON(c, &input) {
		return
	}
	view, err := UpdateNode(c.Request.Context(), id, input)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, view)
}

// DeleteNodeHandler deletes a node.
func DeleteNodeHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	if err := DeleteNode(c.Request.Context(), id); handleLogicError(c, err) {
		return
	}
	compat.OKMessage(c, "")
}

// GetBootstrapTokenHandler returns the global discovery token.
func GetBootstrapTokenHandler(c *gin.Context) {
	view, err := GetBootstrapToken(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, view)
}

// RotateBootstrapTokenHandler rotates the global discovery token.
func RotateBootstrapTokenHandler(c *gin.Context) {
	view, err := RotateBootstrapToken(c.Request.Context())
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, view)
}

// GetAgentReleaseHandler returns the latest agent release for a node.
func GetAgentReleaseHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	release, err := GetAgentRelease(c.Request.Context(), id, c.Query("channel"))
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, release)
}

// RequestAgentUpdateHandler requests agent self-update on a node.
func RequestAgentUpdateHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var request AgentUpdateInput
	if c.Request.ContentLength > 0 {
		if err := bindOptionalJSON(c.Request.Body, &request); err != nil {
			compat.Fail(c, "参数错误")
			return
		}
	}
	view, err := RequestAgentUpdate(c.Request.Context(), id, request)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, view)
}

// RequestOpenrestyRestartHandler requests openresty restart on a node.
func RequestOpenrestyRestartHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	view, err := RequestOpenrestyRestart(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, view)
}

// RequestForceSyncHandler requests force sync on a node.
func RequestForceSyncHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	view, err := RequestForceSync(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, view)
}

// GetObservabilityHandler returns node observability details.
func GetObservabilityHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	var query ObservabilityQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		compat.Fail(c, "参数错误")
		return
	}
	view, err := GetObservability(c.Request.Context(), id, query)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, view)
}

// CleanupHealthEventsHandler cleans up node health events.
func CleanupHealthEventsHandler(c *gin.Context) {
	id, ok := compat.IDParam(c)
	if !ok {
		return
	}
	result, err := CleanupHealthEvents(c.Request.Context(), id)
	if handleLogicError(c, err) {
		return
	}
	compat.OK(c, result)
}

func bindOptionalJSON(body io.Reader, target any) error {
	if err := json.NewDecoder(body).Decode(target); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
