package frps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"openflare/service"
)

type Manager struct {
	frpsPath   string
	dataDir    string
	configPath string
	agentToken string

	mu           sync.RWMutex
	activeConfig *service.RelayConfig
	cmd          *exec.Cmd
	status       string
	lastError    string
	generation   uint64
	stopping     bool
}

type RuntimeStatus struct {
	Status       string
	LastError    string
	Connections  int
	ProxyCount   int
	ClientCount  int
	Proxies      []service.RelayProxyStat
	ProcessAlive bool
}

func NewManager(frpsPath string, dataDir string, agentToken string) *Manager {
	return &Manager{
		frpsPath:   frpsPath,
		dataDir:    dataDir,
		configPath: filepath.Join(dataDir, "frps.toml"),
		status:     "unhealthy",
		agentToken: agentToken,
	}
}

func (m *Manager) GetVersion() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, m.frpsPath, "-v")
	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("failed to get frps version", "error", err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (m *Manager) GetStatus() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

type frpsServerInfo struct {
	Version        string         `json:"version"`
	BindPort       int            `json:"bind_port"`
	CurConns       int            `json:"cur_conns"`
	ClientCounts   int            `json:"client_counts"`
	ProxyTypeCount map[string]int `json:"proxy_type_count"`
}

type frpsProxyInfo struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	ClientAddr    string `json:"client_addr"`
	LastStartTime string `json:"last_start_time"`
	LastCloseTime string `json:"last_close_time"`
}

type frpsProxiesResponse struct {
	Proxies []frpsProxyInfo `json:"proxies"`
}

func (m *Manager) fetchFrpsTelemetry() (connections int, proxyCount int, clientCount int, proxies []service.RelayProxyStat, err error) {
	m.mu.RLock()
	activeConfig := m.activeConfig
	m.mu.RUnlock()

	if activeConfig == nil {
		return 0, 0, 0, nil, fmt.Errorf("activeConfig is nil")
	}

	dashboardPort := activeConfig.BindPort + 500
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// 1. Fetch serverinfo
	infoURL := fmt.Sprintf("http://127.0.0.1:%d/api/serverinfo", dashboardPort)
	req, err := http.NewRequest(http.MethodGet, infoURL, nil)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	password := m.agentToken
	if password == "" {
		password = "admin"
	}
	req.SetBasicAuth("admin", password)
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, 0, nil, fmt.Errorf("serverinfo returned status %d", resp.StatusCode)
	}

	var info frpsServerInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return 0, 0, 0, nil, err
	}

	connections = info.CurConns
	clientCount = info.ClientCounts

	totalProxies := 0
	for _, count := range info.ProxyTypeCount {
		totalProxies += count
	}
	proxyCount = totalProxies

	// 2. Fetch HTTP proxies list
	proxiesURL := fmt.Sprintf("http://127.0.0.1:%d/api/proxy/http", dashboardPort)
	req2, err := http.NewRequest(http.MethodGet, proxiesURL, nil)
	if err != nil {
		return connections, proxyCount, clientCount, nil, nil
	}
	req2.SetBasicAuth("admin", password)
	resp2, err := client.Do(req2)
	if err != nil {
		return connections, proxyCount, clientCount, nil, nil
	}
	defer resp2.Body.Close()

	if resp2.StatusCode == http.StatusOK {
		var listResp frpsProxiesResponse
		if err := json.NewDecoder(resp2.Body).Decode(&listResp); err == nil {
			for _, p := range listResp.Proxies {
				proxies = append(proxies, service.RelayProxyStat{
					Name:          p.Name,
					Type:          p.Type,
					Status:        p.Status,
					ClientAddr:    p.ClientAddr,
					LastStartTime: p.LastStartTime,
					LastCloseTime: p.LastCloseTime,
				})
			}
		}
	}

	return connections, proxyCount, clientCount, proxies, nil
}

func (m *Manager) GetRuntimeStatus() RuntimeStatus {
	m.mu.RLock()
	status := m.status
	lastError := m.lastError
	cmd := m.cmd
	m.mu.RUnlock()

	connections := 0
	proxyCount := 0
	clientCount := 0
	var proxies []service.RelayProxyStat

	if cmd != nil && cmd.Process != nil && status == "healthy" {
		if conns, pCount, cCount, pxs, err := m.fetchFrpsTelemetry(); err == nil {
			connections = conns
			proxyCount = pCount
			clientCount = cCount
			proxies = pxs
		} else {
			slog.Debug("failed to fetch frps telemetry", "error", err)
		}
	}

	return RuntimeStatus{
		Status:       status,
		LastError:    lastError,
		Connections:  connections,
		ProxyCount:   proxyCount,
		ClientCount:  clientCount,
		Proxies:      proxies,
		ProcessAlive: cmd != nil && cmd.Process != nil,
	}
}

func (m *Manager) UpdateConfig(cfg *service.RelayConfig) {
	if cfg == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if config changed
	if m.activeConfig != nil &&
		m.activeConfig.BindPort == cfg.BindPort &&
		m.activeConfig.VhostHTTPPort == cfg.VhostHTTPPort &&
		m.activeConfig.AuthToken == cfg.AuthToken &&
		m.activeConfig.WebServerEnabled == cfg.WebServerEnabled {
		if m.cmd == nil && !m.stopping {
			slog.Warn("frps config unchanged but process is not running, restarting")
			if err := m.restartProcess(); err != nil {
				m.status = "unhealthy"
				m.lastError = err.Error()
				slog.Error("failed to restart frps with unchanged config", "error", err)
			}
		}
		return
	}

	m.activeConfig = cfg
	m.stopping = false
	slog.Info("relay config updated, reloading frps")

	if err := m.renderConfig(cfg); err != nil {
		slog.Error("failed to render frps config", "error", err)
		m.status = "unhealthy"
		m.lastError = err.Error()
		return
	}

	if err := m.restartProcess(); err != nil {
		slog.Error("failed to restart frps", "error", err)
		m.status = "unhealthy"
		m.lastError = err.Error()
	} else {
		m.status = "healthy"
		m.lastError = ""
	}
}

func (m *Manager) renderConfig(cfg *service.RelayConfig) error {
	if err := os.MkdirAll(m.dataDir, 0755); err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("bindPort = %d\n", cfg.BindPort))
	if cfg.VhostHTTPPort > 0 {
		buf.WriteString(fmt.Sprintf("vhostHTTPPort = %d\n", cfg.VhostHTTPPort))
	}
	if cfg.AuthToken != "" {
		buf.WriteString("[auth]\n")
		buf.WriteString("method = \"token\"\n")
		buf.WriteString(fmt.Sprintf("token = \"%s\"\n", cfg.AuthToken))
	}

	// WebServer configuration
	buf.WriteString("\n[webServer]\n")
	if cfg.WebServerEnabled {
		buf.WriteString("addr = \"0.0.0.0\"\n")
	} else {
		buf.WriteString("addr = \"127.0.0.1\"\n")
	}
	buf.WriteString(fmt.Sprintf("port = %d\n", cfg.BindPort+500))
	buf.WriteString("user = \"admin\"\n")

	password := m.agentToken
	if password == "" {
		password = "admin"
	}
	buf.WriteString(fmt.Sprintf("password = \"%s\"\n", password))

	return os.WriteFile(m.configPath, buf.Bytes(), 0644)
}

func (m *Manager) restartProcess() error {
	m.generation++
	generation := m.generation
	if m.cmd != nil && m.cmd.Process != nil {
		slog.Debug("stopping existing frps process")
		_ = m.cmd.Process.Kill()
		m.cmd = nil
	}
	return m.startProcessLocked(generation)
}

func (m *Manager) startProcessLocked(generation uint64) error {
	cmd := exec.Command(m.frpsPath, "-c", m.configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	m.cmd = cmd
	m.status = "healthy"
	m.lastError = ""

	go func(c *exec.Cmd) {
		err := c.Wait()
		slog.Warn("frps process exited", "error", err)
		m.mu.Lock()
		if m.cmd == c {
			m.cmd = nil
			m.status = "unhealthy"
			if err != nil {
				m.lastError = err.Error()
			} else {
				m.lastError = "frps process exited"
			}
		}
		shouldRestart := !m.stopping && m.generation == generation
		m.mu.Unlock()
		if !shouldRestart {
			return
		}
		time.Sleep(2 * time.Second)
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.stopping || m.generation != generation {
			return
		}
		slog.Warn("restarting frps after unexpected exit")
		if err := m.startProcessLocked(generation); err != nil {
			m.status = "unhealthy"
			m.lastError = err.Error()
			slog.Error("failed to auto restart frps", "error", err)
		}
	}(cmd)

	return nil
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopping = true
	m.generation++
	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
		m.cmd = nil
	}
	m.status = "unhealthy"
}
