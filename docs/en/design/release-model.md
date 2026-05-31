# Release Model

You will learn: Why OpenFlare uses a complete configuration version as the release unit, and how publishing, activation, Agent application, and rollback work.

OpenFlare's release model is centered on complete configuration versions rather than modifying node configurations online.

Standard link:

```text
Modify rules -> Preview / View diff -> Publish -> Generate complete configuration version -> Activate version -> Agent pulls -> Local application -> Report result
```

## Publishing Rules

When publishing, the Server must:

1. Read all enabled `proxy_routes`.
2. Read the Server side OpenResty main configuration template, performance parameters, cache parameters, and necessary Lua resources.
3. Read domain and certificate binding relationships.
4. Read the WAF global rule group, custom rule groups, and website binding relationships.
5. Render the complete OpenResty configuration and WAF runtime configuration.
6. Calculate the `checksum`.
7. Write to `config_versions`.
8. Switch the activated version.
9. Let the Agent discover and apply it in subsequent heartbeats.

The version number format is fixed as `YYYYMMDD-NNN`.

## Preview and Publishing

Preview and diff are read-only capabilities and do not generate release records.

Publishing generates a new complete configuration version. The version must contain sufficient information so that future rollbacks can be re-applied based on historical snapshots, without relying on current mutable configurations.

## Activating Version

There can only be one activated version globally at a time. Differentiated versions grouped by nodes are currently not supported.

The Agent obtains the activated version summary through the heartbeat; only when the remote version or checksum is inconsistent with the local state does the Agent enter the synchronization flow. When Agent WS connection upgrade is enabled and the connection is available, the Server will broadcast the latest active version summary after successfully publishing or activating a version. Upon receiving it, the Agent immediately pulls and applies the configuration using the ordinary synchronization flow. When WS is unavailable, changes are still discovered at HTTP heartbeat intervals.

## Immutable History

Historical versions are immutable. Rollback is not achieved by modifying older versions, but by reactivating older versions.

The result of doing this is:

* Every version can be traced back.
* The rollback link is consistent with the ordinary release application link.
* The Agent does not need to understand "reverse patch", but only needs to apply a target version.

## Agent Application Policy

When discovering a new version, the Agent will:

1. Pull the details of the target version.
2. Back up old files.
3. Write the main configuration, route configurations, certificates, necessary Lua resources, and WAF/PoW runtime configurations.
4. Execute OpenResty configuration verification.
5. reload; if it is not started during runtime, try to start OpenResty with the current configuration.
6. Report success, warning, or failure.

If the activation of the new configuration fails, the Agent must try to restore execution; report a warning when the rollback succeeds. If there is no historical main configuration to roll back to locally, the Agent will write the built-in safe fallback configuration and try to pull up OpenResty: this configuration only listens to port `80` externally, contains no user routes, uniformly returns `503 Service Unavailable` and `OpenFlare: No Valid Configuration`, and retains the local `stub_status` health check entry. If fallback startup is successful, it still blocks the failed target version and reports a warning; report a failure when there is a historical main configuration but it still cannot recover after rollback.

Once a target `version + checksum` application fails and rolls back, the Agent will block repeated applications of this target in its local state. Only when the remote activated version or checksum changes is it allowed to try again.

## Design Constraints

* Publishing must read all enabled site configurations, rather than only rendering the modified object this time.
* Rollback is achieved by reactivating older versions, without modifying historical versions.
* The Agent API is fixed to use the node-exclusive `agent_token`; the first access can use the `discovery_token`.
* The Server does not provide remote shell or arbitrary command execution entries.
* The configuration version must save complete snapshots, rendering results, and `checksum`.
* WAF rule groups and website binding relationships must enter the snapshot and checksum along with the complete configuration version, and must not rely on the current mutable WAF configuration when rolling back.
