# OpenFlare 后端迁移 — 任务拆分与委派

> **状态**：进行中  
> **基建**：Batch 0 已完成（`Wavelet/internal/apps/openflare/compat/` + `legacy/register*.go` + `router.go` 挂载）

## 任务隔离规则

| 规则 | 说明 |
|---|---|
| 文件所有权 | 每个任务 **仅修改** 自己的 `internal/apps/openflare/<module>/`、对应 `legacy/register_<module>.go`、`internal/model/openflare_<module>.go`、goose SQL |
| 禁止修改 | `v1/user.go`、`v1/admin.go`、`model/users.go`、其他任务的 `register_*.go` |
| 响应格式 | 旧前端兼容 API 使用 `compat.OK/Fail/Unauthorized`（`{success,message,data}`） |
| 鉴权 | 管理端 `compat.AdminAuth()` / `compat.RootAuth()`；用户 `compat.UserAuth()` |
| Logic 层 | `logics.go` 使用 `context.Context`，不依赖 `*gin.Context` |
| 数据源 | `db.DB(ctx)` 获取 GORM |
| 质量门禁 | 完成后 `go build ./...` 并通过本模块测试 |

## 任务队列

| ID | 板块 | 状态 | 负责文件 |
|---|---|---|---|
| T-AUTH | 认证/用户/OAuth/Cap | ✅ 完成 | `legacy/register_auth.go`, `openflare/auth/`, `legacy/auth_*.go` |
| T-OPTION | 状态/公告/Option | ✅ 完成 | `legacy/register_option.go`, `openflare/option/` |
| T-ORIGIN | 源站 | ✅ 完成 | `legacy/register_origin.go`, `openflare/origin/` |
| T-APPLYLOG | 应用日志 | ✅ 完成 | `legacy/register_apply_log.go`, `openflare/apply_log/` |
| T-PROXY | 代理规则 | ✅ 完成 | `legacy/register_proxy_route.go`, `openflare/proxy_route/` |
| T-NODE | 节点管理 | ✅ 完成 | `legacy/register_node.go`, `openflare/node/` |
| T-WAF | WAF | ✅ 完成 | `legacy/register_waf.go`, `openflare/waf/` |
| T-TLS | TLS/证书/域名/DNS | ✅ 完成 | `legacy/register_tls.go`, `openflare/tls/` |
| T-CFGVER | 配置版本 | ✅ 完成 | `legacy/register_config_version.go`, `openflare/config_version/` |
| T-AGENT | Agent API + WS | ✅ 完成 | `legacy/register_agent.go`, `openflare/agent/`, `openflare/websocket/` |
| T-PAGES | Pages 托管 | ✅ 完成 | `legacy/register_pages.go`, `openflare/pages/` |
| T-RELAY | Relay + Flared | ✅ 完成 | `legacy/register_relay_flared.go`, `openflare/relay/`, `openflare/flared/` |
| T-OBS | 仪表盘 + 可观测 | ✅ 完成 | `legacy/register_dashboard_obs.go`, `openflare/dashboard/`, `openflare/observability/` |
| T-MISC | 升级/GeoIP/UptimeKuma | ✅ 完成 | `legacy/register_misc.go`, `openflare/update/`, `openflare/geoip/` |

## 集成测试结果（2026-06-19）

| 测试包 | 场景 | 结果 |
|---|---|---|
| `integration/auth_option_test.go` | 登录/self/option 权限/热重载 | ✅ 5/5 |
| `integration/core_chain_test.go` | 源站→规则→发布→节点→apply-log | ✅ 6/6 |
| `integration/security_test.go` | WAF/TLS/域名/DNS | ✅ 7/7 |
| `integration/agent_protocol_test.go` | Agent/Relay/Flared 协议 | ✅ 5/5 |

```bash
go test ./internal/apps/openflare/... -count=1   # 全部通过
```

**已知问题**：`go build ./...` 因 `github.com/boj/redistore` 与 `go-redis/v9` 版本不兼容失败（Wavelet 既有依赖问题，与 OpenFlare 迁移无关）。OpenFlare 包与 integration 测试均可独立编译通过。

## 源代码参考

- 旧后端：`openflare-server/internal/controller/`、`service/`、`model/`
- 迁移计划：`docs/plan/20260618-openflare-wavelet-backend-migration.md`
- Wavelet 规范：`Wavelet/AGENTS.md`