# AGENTS.md

本文件是 OpenFlare 的 AI 接手入口，不承载详细设计、规范和计划。接手项目时，请根据以下分层文档指引进行阅读与开发：

## 1. 核心必读文档（Level 3 & Level 4）- 必须阅读 ⚠️

为了理解 OpenFlare 的设计理念、产品边界、核心机制以及代码编写的工程约束，**AI 在接手项目时必须首先且完整阅读以下文档**：

### Level 3: 面向贡献者的参阅文档 (Contributor References)
* **[docs/design/index.md](./docs/design/index.md)**
  *作用：理解当前 MVP 的产品范围、系统边界、核心对象和长期约束。*
* **[docs/design/architecture.md](./docs/design/architecture.md)**
  *作用：理解 Server、Agent、OpenResty 与前端的职责边界与网络拓扑。*
* **[docs/design/release-model.md](./docs/design/release-model.md)**
  *作用：理解配置发布、激活、回滚与 Agent 节点配置应用的模型。*
* **[docs/design/development.md](./docs/design/development.md)**
  *作用：了解如何搭建本地开发环境，运行后端 Server、Agent 和前端开发服务器，以及运行测试与构建的命令。*
* **[docs/design/repository.md](./docs/design/repository.md)**
  *作用：熟悉仓库的整体物理结构和各子目录的职责。*

### Level 4: 面向 AI 的开发指导规范 (AI Guidelines)
* **[docs/guildline/development-constraints.md](./docs/guildline/development-constraints.md)**
  *作用：掌握核心后端/Agent/前端分层约束、数据模型规范、数据库迁移升级协议、API 与鉴权设计准则。*
* **[docs/guildline/Guidelines.md](./docs/guildline/Guidelines.md)**
  *作用：通用的 Go 后端开发与高质量编码准则，包括架构、并发、错误处理、安全及工作流程。*
* **[docs/guildline/Project.md](./docs/guildline/Project.md)**
  *作用：针对 OpenFlare 后端特定的控制器参数解析、响应处理、纯净工具类与数据库逻辑完全隔离、Go 泛型切片去重及 JSON 序列化避坑细则。*

---

## 2. 按需查阅文档（Level 2）- 根据需求阅读 💡

当开发任务涉及具体的系统部署、升级、接口联调或配置字段查阅时，**AI 应当根据需求阅读相应的参考手册**：

### Level 2: 面对高级用户/开发者的参阅文档 (Reference Manuals)
* **[docs/reference/configuration.md](./docs/reference/configuration.md)**
  *作用：系统启动时支持的所有环境变量、命令行参数、运行时 Option 选项和 Agent 配置文件字段。*
* **[docs/reference/cli.md](./docs/reference/cli.md)**
  *作用：Server 与 Agent 可用的命令行参数、安装/卸载脚本参数等参考。*
* **[docs/reference/api.md](./docs/reference/api.md)**
  *作用：管理端 API 与 Agent API 的响应结构、路径和详细鉴权约定。*
* **[docs/reference/deployment.md](./docs/reference/deployment.md)**
  *作用：理解 Server 和 Agent 的单机、Docker 部署配置，以及 Agent 接入、升级、卸载和联调步骤。*
* **[docs/reference/server.md](./docs/reference/server.md)**
  *作用：如何配置系统配置、服务环境变量并正确启动 Server 服务。*
* **[docs/reference/agent.md](./docs/reference/agent.md)**
  *作用：理解 Agent 接入的 discovery/agent 令牌鉴权机制、本地配置文件及 Docker 部署参数。*
* **[docs/reference/upgrade.md](./docs/reference/upgrade.md)**
  *作用：Server 及各代理节点 Agent 的升级步骤与维护策略。*

---

## 3. 新手与业务教程（Level 1）- 体验与排障参考 📘

如果任务涉及优化最终用户体验、丰富业务能力或排查常见故障，可参阅面向普通用户的指南：

### Level 1: 面向新手用户的教程文档 (Novice Tutorials)
* **[docs/guide/quick-start.md](./docs/guide/quick-start.md)**：五分钟内基于 Docker Compose 快速跑起 Server 和首个 Agent 节点的完整闭环。
* **[docs/guide/usage.md](./docs/guide/usage.md)**：反向代理网站、源站、证书托管、配置发布与回滚的常规界面操作与观测功能使用指南。
* **[docs/guide/sso.md](./docs/guide/sso.md)**：系统如何配置 GitHub OAuth 及标准 OIDC 第三方登录，以及绑定本地账户的流程。
* **[docs/guide/first-site.md](./docs/guide/first-site.md)**：从零开始配置、发布并验证第一个代理网站的完整步骤。
* **[docs/guide/troubleshooting.md](./docs/guide/troubleshooting.md)**：常见数据库迁移、节点离线、OpenResty 校验失败、SSL 证书失效等故障的表现症状及标准排障路径。

---

## 执行要求

* 如果实现内容超出 [产品边界](./docs/design/index.md)，先修改设计文档，再继续编码。
* 如果实现方式违反 [开发约束](./docs/guildline/development-constraints.md)，应优先调整方案，而不是绕过规范。
* 如果实现方式涉及后端代码逻辑，必须严格遵循 [docs/guildline/](./docs/guildline/) 下的所有开发准则。
* 如果需求与当前阶段原则冲突，优先遵守 [开发约束](./docs/guildline/development-constraints.md) 中的变更准入与验收标准。
* 如果任务涉及前端改造或管理端 UI，必须同时遵守 [开发约束](./docs/guildline/development-constraints.md) 中的前端规范。

## 文档维护要求

当以下内容发生变化时，应同步更新对应中文文档，不要同步英文文档：

* 产品范围或系统边界变化：更新 `docs/design/index.md`
* 系统结构、模块职责变化：更新 `docs/design/architecture.md`
* 发布、同步、回滚模型变化：更新 `docs/design/release-model.md`
* 业务分层、数据模型边界、接口约定、阶段原则、测试基线变化：更新 `docs/guildline/development-constraints.md`
* 后端开发规范、代码质量要求、重构模式、去重逻辑与避坑指南变化：更新 `docs/guildline/` 下的对应开发准则文件
* 产品启动、部署、升级、联调方式变化：更新 `docs/guide/quick-start.md`、`docs/reference/deployment.md` 和 `README.md`
* 用户操作路径、常见场景变化：更新 `docs/guide/usage.md`
* 本地开发、测试、构建方式变化：更新 `docs/design/development.md`
* 常见故障、排查路径变化：更新 `docs/guide/troubleshooting.md`
* 环境变量、命令行参数、运行时配置、Agent 配置变化：更新 `docs/reference/configuration.md`
