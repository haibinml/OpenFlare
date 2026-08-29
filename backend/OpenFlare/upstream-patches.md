# 上游本地补丁清单

`scripts/sync-upstream.sh` 会用上游覆盖 `backend/{core,pkg,plugins}`。以下文件带有
OpenFlare 侧必需的内核补丁：同步后若丢失，`go build ./...` 或用例集会失败，
按本表重新应用，并应回流 Wavelet 上游后删除本条记录。

- `core/extpoints/router.go`
- `core/scoped_extpoints.go`

## core：RouterExtension 增加 HandleRaw / BasePath

`Handle` 经 `cleanPath` 归一化会剥掉尾部斜杠，无法表达 `/resource` 与 `/resource/`
两条不同路由；OpenFlare 的 20 个历史 list 端点两者都注册且部署关闭了
`RedirectTrailingSlash`，缺失即 404。

- `HandleRaw(method, path, handlers...)`：与组前缀拼接但**保留**尾部斜杠；
- `BasePath()`：返回组绝对前缀（根注册表为空串）；
- 作用域包装器为 `HandleRaw` 同样登记 `OnDispose` 反注册；
- 用例：`core/extpoints/router_raw_test.go`。

## 回流状态

上述内核补丁与 `pkg/util` 新增助手已提交到 Wavelet 分支
`feat/cordis-router-raw-routes`（worktree `/Users/ryan/Code/Go/Wavelet-router-raw`，
提交 `cb339ab`、`bad6fa7`、`8ff017b`）。**合并进 Wavelet main 后**：重跑
`scripts/sync-upstream.sh`，确认 `sync-upstream.sh --check` 零差异，然后清空本清单。

`pkg/util` 的版本比较 / 网络 / 格式化助手已回流，当前零漂移。
