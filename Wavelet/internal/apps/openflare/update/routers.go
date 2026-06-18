// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package update

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgradeLogsUpgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool { return true },
}

// GetLatestReleaseHandler returns the newest GitHub release for the legacy update UI.
func GetLatestReleaseHandler(c *gin.Context) {
	release, err := GetLatestRelease(c.Request.Context(), c.Query("channel"))
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}
	compat.OK(c, release)
}

// UpgradeServerHandler schedules an automatic upgrade from the latest release.
func UpgradeServerHandler(c *gin.Context) {
	var request upgradeRequest
	if err := bindOptionalJSON(c.Request.Body, &request); err != nil {
		compat.Fail(c, "无效的参数")
		return
	}

	release, err := ScheduleUpgrade(c.Request.Context(), request.Channel)
	if err != nil {
		compat.Fail(c, err.Error())
		return
	}

	okWithMessage(c, release, "服务升级任务已启动，下载完成后将自动重启。")
}

// UploadManualServerBinaryHandler rejects manual uploads (feature disabled upstream).
func UploadManualServerBinaryHandler(c *gin.Context) {
	if err := UploadManualBinary(); err != nil {
		compat.Fail(c, err.Error())
		return
	}
}

// ConfirmManualServerUpgradeHandler rejects manual upgrades (feature disabled upstream).
func ConfirmManualServerUpgradeHandler(c *gin.Context) {
	if err := ConfirmManualUpgrade(); err != nil {
		compat.Fail(c, err.Error())
		return
	}
}

// StreamServerUpgradeLogsHandler streams upgrade progress snapshots over WebSocket.
func StreamServerUpgradeLogsHandler(c *gin.Context) {
	conn, err := upgradeLogsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	updates, unsubscribe := SubscribeUpgradeStream()
	defer unsubscribe()

	heartbeatTicker := time.NewTicker(15 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case snapshot, ok := <-updates:
			if !ok {
				return
			}
			if err := conn.WriteJSON(snapshot); err != nil {
				return
			}
		case <-heartbeatTicker.C:
			if err := conn.WriteJSON(StreamSnapshot{}); err != nil {
				return
			}
		case <-c.Request.Context().Done():
			return
		}
	}
}

func okWithMessage(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func bindOptionalJSON(body io.Reader, target any) error {
	if err := json.NewDecoder(body).Decode(target); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
