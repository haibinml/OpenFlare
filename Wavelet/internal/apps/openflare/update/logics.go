// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package update

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Rain-kl/Wavelet/internal/apps/admin/updater"
	"github.com/Rain-kl/Wavelet/pkg/logger"
)

const (
	channelStable  = "stable"
	channelPreview = "preview"
)

// UpgradeLogRecord is a single upgrade log entry for the legacy update API.
type UpgradeLogRecord struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// LatestReleaseView mirrors the legacy OpenFlare latest-release payload.
type LatestReleaseView struct {
	TagName          string             `json:"tag_name"`
	Body             string             `json:"body"`
	HTMLURL          string             `json:"html_url"`
	PublishedAt      string             `json:"published_at"`
	Channel          string             `json:"channel"`
	Prerelease       bool               `json:"prerelease"`
	CurrentVersion   string             `json:"current_version"`
	HasUpdate        bool               `json:"has_update"`
	UpgradeSupported bool               `json:"upgrade_supported"`
	InProgress       bool               `json:"in_progress"`
	UpgradeStatus    string             `json:"upgrade_status"`
	UpgradeLogs      []UpgradeLogRecord `json:"upgrade_logs"`
}

// StreamSnapshot is pushed over the upgrade logs websocket.
type StreamSnapshot struct {
	InProgress    bool               `json:"in_progress"`
	UpgradeStatus string             `json:"upgrade_status"`
	UpgradeLogs   []UpgradeLogRecord `json:"upgrade_logs"`
}

type upgradeRequest struct {
	Channel string `json:"channel"`
}

var upgradeState struct {
	sync.Mutex
	inProgress bool
	status     string
	logs       []UpgradeLogRecord
}

var upgradeSubscribers struct {
	sync.Mutex
	nextID    int
	listeners map[int]chan StreamSnapshot
}

func init() {
	upgradeSubscribers.listeners = make(map[int]chan StreamSnapshot)
	upgradeState.status = "idle"
}

func normalizeChannel(channel string) string {
	switch strings.ToLower(strings.TrimSpace(channel)) {
	case channelPreview:
		return channelPreview
	default:
		return channelStable
	}
}

func isDevBuild(version string) bool {
	version = strings.TrimSpace(version)
	return version == "" || strings.EqualFold(version, "dev")
}

func mapStatusToLatestRelease(status updater.Status, channel string) *LatestReleaseView {
	inProgress, upgradeStatus, logs := snapshotUpgradeState()
	view := &LatestReleaseView{
		TagName:          status.LatestVersion,
		Body:             status.ReleaseNotes,
		HTMLURL:          status.ReleaseURL,
		PublishedAt:      status.PublishedAt,
		Channel:          channel,
		Prerelease:       status.Prerelease,
		CurrentVersion:   status.CurrentVersion,
		HasUpdate:        status.UpdateAvailable,
		UpgradeSupported: !isDevBuild(status.CurrentVersion) && runtime.GOOS != "windows",
		InProgress:       inProgress || updater.IsUpgrading(),
		UpgradeStatus:    upgradeStatus,
		UpgradeLogs:      logs,
	}
	if channel == channelPreview && status.Prerelease {
		view.HasUpdate = !isDevBuild(status.CurrentVersion)
	}
	return view
}

// GetLatestRelease returns the newest upstream release for the requested channel.
func GetLatestRelease(ctx context.Context, channel string) (*LatestReleaseView, error) {
	normalizedChannel := normalizeChannel(channel)
	status, err := updater.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	return mapStatusToLatestRelease(status, normalizedChannel), nil
}

// ScheduleUpgrade downloads the latest release and restarts with the staged binary.
func ScheduleUpgrade(ctx context.Context, channel string) (*LatestReleaseView, error) {
	normalizedChannel := normalizeChannel(channel)

	upgradeState.Lock()
	if upgradeState.inProgress || updater.IsUpgrading() {
		upgradeState.Unlock()
		return nil, errors.New("服务升级正在执行中，请稍后再试")
	}
	resetUpgradeLogsLocked()
	upgradeState.inProgress = true
	upgradeState.status = "running"
	appendUpgradeLogLocked("info", fmt.Sprintf("Automatic upgrade scheduled for channel: %s.", normalizedChannel))
	upgradeState.Unlock()
	broadcastUpgradeSnapshot()

	executable, stagedBinary, status, err := updater.PrepareUpgrade(ctx)
	if err != nil {
		recordUpgradeFailure(err)
		return nil, err
	}

	view := mapStatusToLatestRelease(status, normalizedChannel)
	view.InProgress = true
	view.UpgradeStatus = "running"
	view.UpgradeLogs = snapshotUpgradeLogs()

	appendUpgradeLogLocked("info", fmt.Sprintf("Upgrade package prepared: %s.", status.LatestVersion))
	broadcastUpgradeSnapshot()

	go func() {
		time.Sleep(time.Second)
		if err := updater.ApplyPreparedUpgrade(executable, stagedBinary); err != nil {
			updater.FinishUpgrade()
			recordUpgradeFailure(err)
			logger.ErrorF(context.Background(), "[Update] replace and restart failed: %v", err)
		}
	}()

	return view, nil
}

// UploadManualBinary is disabled in the legacy OpenFlare server.
func UploadManualBinary() error {
	return errors.New("手动升级功能已禁用")
}

// ConfirmManualUpgrade is disabled in the legacy OpenFlare server.
func ConfirmManualUpgrade() error {
	return errors.New("手动升级功能已禁用")
}

// SubscribeUpgradeStream registers a listener for upgrade websocket snapshots.
func SubscribeUpgradeStream() (<-chan StreamSnapshot, func()) {
	upgradeSubscribers.Lock()
	defer upgradeSubscribers.Unlock()

	id := upgradeSubscribers.nextID
	upgradeSubscribers.nextID++
	ch := make(chan StreamSnapshot, 1)
	upgradeSubscribers.listeners[id] = ch

	unsubscribe := func() {
		upgradeSubscribers.Lock()
		defer upgradeSubscribers.Unlock()
		if listener, ok := upgradeSubscribers.listeners[id]; ok {
			delete(upgradeSubscribers.listeners, id)
			close(listener)
		}
	}

	select {
	case ch <- currentUpgradeSnapshot():
	default:
	}

	return ch, unsubscribe
}

func snapshotUpgradeState() (bool, string, []UpgradeLogRecord) {
	upgradeState.Lock()
	defer upgradeState.Unlock()
	return upgradeState.inProgress || updater.IsUpgrading(), upgradeState.status, cloneUpgradeLogsLocked()
}

func snapshotUpgradeLogs() []UpgradeLogRecord {
	upgradeState.Lock()
	defer upgradeState.Unlock()
	return cloneUpgradeLogsLocked()
}

func currentUpgradeSnapshot() StreamSnapshot {
	inProgress, status, logs := snapshotUpgradeState()
	return StreamSnapshot{
		InProgress:    inProgress,
		UpgradeStatus: status,
		UpgradeLogs:   logs,
	}
}

func resetUpgradeLogsLocked() {
	upgradeState.logs = nil
}

func appendUpgradeLogLocked(level, message string) {
	upgradeState.logs = append(upgradeState.logs, UpgradeLogRecord{
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
}

func cloneUpgradeLogsLocked() []UpgradeLogRecord {
	if len(upgradeState.logs) == 0 {
		return []UpgradeLogRecord{}
	}
	cloned := make([]UpgradeLogRecord, len(upgradeState.logs))
	copy(cloned, upgradeState.logs)
	return cloned
}

func recordUpgradeFailure(err error) {
	upgradeState.Lock()
	upgradeState.inProgress = false
	upgradeState.status = "failed"
	appendUpgradeLogLocked("error", err.Error())
	upgradeState.Unlock()
	broadcastUpgradeSnapshot()
}

func broadcastUpgradeSnapshot() {
	snapshot := currentUpgradeSnapshot()
	upgradeSubscribers.Lock()
	defer upgradeSubscribers.Unlock()
	for _, listener := range upgradeSubscribers.listeners {
		select {
		case listener <- snapshot:
		default:
		}
	}
}
