# System Architecture

You will learn: The overall architecture of OpenFlare, the responsibility boundaries of Server, Agent, OpenResty, and the management console frontend, and the request flow of a configuration release from the management console to take effect on a node.

OpenFlare consists of the Server, the Agent, local OpenResty on each node, and the management console frontend. The Server is the control plane, the Agent is the only controlled landing entry point on the node side, and OpenResty is the actual data plane.

```text
Browser
  |
  | Management UI / API
  v
OpenFlare Server (Gin + GORM + SQLite/PostgreSQL)
  |
  | Agent API / heartbeat / config pull
  v
OpenFlare Agent
  |
  | write config / openresty -t / reload / rollback
  v
OpenResty binary
  |
  | reverse proxy
  v
Origin
```

## Component Responsibilities

| Component | Responsibility |
| --- | --- |
| Server | Management UI, admin APIs, Agent APIs, configuration rendering, version publishing, data storage, and aggregate queries |
| Agent | Registration, heartbeats, synchronization, writing files, configuration validation, reloads, fallback/rollbacks, self-updating, and lightweight data collection |
| OpenResty | Receives real traffic, executes WAF, PoW, authentication, and reverse proxying according to configurations rendered by OpenFlare |
| Frontend | Manages website configurations, WAF, origins, certificates, nodes, versions, users, settings, and observability pages |

## Server

`openflare_server` is a monolithic control plane:

* Gin provides HTTP services.
* GORM accesses SQLite or PostgreSQL.
* The existing login system provides management console Sessions.
* Authentication source and external account binding support GitHub OAuth and standard OIDC.
* The Go Server hosts the static build output of `openflare_server/web`.

The Server does not directly SSH into nodes, nor does it modify node files online. It only saves the control plane state, generates complete configuration versions, and lets nodes actively pull them via the Agent API.

## Agent

`openflare_agent` is a Go monolithic application:

* Runs on nodes as a single binary.
* Reads or generates local node information upon startup.
* Performs periodic heartbeats to report status and fetch the active version summary.
* Pulls configurations, backs up old files, writes new files, validates, and reloads upon discovering a new version.
* Attempts to restore execution and roll back when the application fails.
* Maintains the WAF GeoIP mmdb; writes the built-in initial database on startup and updates it regularly based on configuration.

The Agent uniformly executes validations, reloads, starts, and restarts via the OpenResty binary pointed to by `openresty_path`; it falls back to calling `openresty` by default when not configured. In Docker deployments, the Agent image includes the OpenResty binary and follows the same binary control logic.

## Frontend

`openflare_server/web` is the official management console frontend:

* Next.js App Router.
* React 19.
* TypeScript.
* Tailwind CSS.
* TanStack Query manages server state.

The frontend is hosted by the Go Server after static export. All API requests must go through `lib/api/` uniformly and handle the `success/message/data` response structure.

## Data and Request Flow

### Management Console Request Flow

```text
Browser -> Frontend -> /api/* -> controller -> service -> model -> database
```

Mutation APIs on the management console use `POST`, while read-only APIs use `GET`. Both success and failure return a clear `message`.

### Agent Sync Flow

```text
Agent heartbeat -> Server returns active version summary
Agent discovers new version -> Pulls configuration details
Agent writes main configuration / route configuration / certificates / Lua resources / WAF runtime configuration
Agent executes OpenResty validation and reload
Agent reports application result
```

When WebSocket (WS) connection upgrade is enabled by default, the Agent first obtains settings through the HTTP heartbeat, and then attempts to connect to the Agent WebSocket. Once the WS connection is successful, periodic status reporting is carried by WS; when the Server publishes or activates a version, it broadcasts the active version summary to connected Agents, allowing them to enter the synchronization flow immediately. When the WS connection is disconnected or fails to establish, the Agent automatically falls back to the HTTP heartbeat.

### Reverse Proxy Flow

```text
Client -> OpenResty server block -> WAF Lua -> named upstream -> Origin
```

Website configuration is the aggregation boundary of reverse proxies. A website configuration can bind multiple domains and share site-level traffic limits, reverse proxies, and caching configurations.

WAF is executed in the OpenResty `access_by_lua_file` phase. Rules come from `waf_config.json` carried in the current active version; the global rule group takes effect by default, and websites can overlay custom rule groups.

## Core Objects

Currently active entities include:

* `proxy_routes`
* `origins`
* `config_versions`
* `nodes`
* `auth_sources`
* `external_accounts`
* `node_system_profiles`
* `apply_logs`
* `tls_certificates`
* `managed_domains`
* `node_request_reports`
* `node_access_logs`
* `node_metric_snapshots`
* `traffic_analytics_rollups`
* `node_health_events`
* `waf_rule_groups`
* `waf_rule_group_bindings`

## Key Design Decisions

| Decision | Reason |
| --- | --- |
| Complete configuration versions, instead of online patches | Gives previews, activations, history, and rollbacks stable boundaries |
| Active pull by Agents | Server does not need SSH permissions, nor does it expose remote command execution entry points |
| Global single active version | Reduces MVP complexity and ensures all nodes are consistent by default |
| Website configurations aggregate multiple domains | Supports sharing site-level policies for a business site while allowing certificate binding per domain |
| Server-side aggregation of observability data | Avoids inconsistent results caused by temporary frontend calculations |

## Contributor Reading Suggestions

If you want to modify architecture-related code, read these first:

1. [Product Boundary](./index.md)
2. [Release Model](./release-model.md)
3. [Development Constraints](./development.md)
4. [Repository Structure](../reference/repository.md)
