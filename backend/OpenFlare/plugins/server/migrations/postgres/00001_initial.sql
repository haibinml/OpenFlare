-- +goose Up
CREATE TABLE IF NOT EXISTS of_acme_accounts (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL DEFAULT '',
    url VARCHAR(255) NOT NULL DEFAULT '',
    private_key TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS of_apply_logs (
    id BIGSERIAL PRIMARY KEY,
    node_id VARCHAR(64) NOT NULL,
    version VARCHAR(32) NOT NULL,
    result VARCHAR(32) NOT NULL,
    message TEXT,
    checksum VARCHAR(64) NOT NULL DEFAULT '',
    main_config_checksum VARCHAR(64) NOT NULL DEFAULT '',
    route_config_checksum VARCHAR(64) NOT NULL DEFAULT '',
    support_file_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_of_apply_logs_created_at ON of_apply_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_of_apply_logs_node_id ON of_apply_logs(node_id);

CREATE TABLE IF NOT EXISTS of_cf_connections (
    id BIGSERIAL PRIMARY KEY,
    source VARCHAR(32) NOT NULL DEFAULT '',
    dns_account_id BIGINT,
    "authorization" TEXT NOT NULL DEFAULT '',
    status VARCHAR(16) NOT NULL DEFAULT '',
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_cf_connections_dns_account_id ON of_cf_connections (dns_account_id);

CREATE TABLE IF NOT EXISTS of_cf_pointing_groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    primary_node_id BIGINT NOT NULL,
    backup_node_id BIGINT,
    active_node_id BIGINT NOT NULL,
    default_proxied BOOLEAN NOT NULL DEFAULT FALSE,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_cf_pointing_groups_active_node_id ON of_cf_pointing_groups (active_node_id);
CREATE INDEX IF NOT EXISTS idx_of_cf_pointing_groups_backup_node_id ON of_cf_pointing_groups (backup_node_id);
CREATE INDEX IF NOT EXISTS idx_of_cf_pointing_groups_primary_node_id ON of_cf_pointing_groups (primary_node_id);

CREATE TABLE IF NOT EXISTS of_cf_pointing_members (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL,
    zone_domain_id BIGINT NOT NULL,
    proxied BOOLEAN NOT NULL DEFAULT FALSE,
    cf_zone_id VARCHAR(64) NOT NULL DEFAULT '',
    cf_record_id VARCHAR(64) NOT NULL DEFAULT '',
    desired_ip VARCHAR(64) NOT NULL DEFAULT '',
    sync_status VARCHAR(16) NOT NULL DEFAULT 'pending',
    last_error TEXT NOT NULL DEFAULT '',
    synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_cf_pointing_members_group_id ON of_cf_pointing_members (group_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_cf_pointing_members_zone_domain_id ON of_cf_pointing_members (zone_domain_id);

CREATE TABLE IF NOT EXISTS of_config_versions (
    version VARCHAR(32) PRIMARY KEY,
    snapshot_json TEXT NOT NULL,
    main_config TEXT NOT NULL DEFAULT '',
    rendered_config TEXT NOT NULL,
    support_files_json TEXT NOT NULL DEFAULT '[]',
    checksum VARCHAR(64) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_by VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_config_versions_is_active ON of_config_versions (is_active);

CREATE TABLE IF NOT EXISTS of_dns_accounts (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(64) NOT NULL,
    "authorization" TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS of_node_access_logs (
    id              BIGINT NOT NULL,
    node_id         VARCHAR(64) NOT NULL DEFAULT '',
    logged_at       TIMESTAMPTZ NOT NULL,
    remote_addr     VARCHAR(128) NOT NULL DEFAULT '',
    region          VARCHAR(128) NOT NULL DEFAULT '',
    host            VARCHAR(255) NOT NULL DEFAULT '',
    path            VARCHAR(2048) NOT NULL DEFAULT '',
    user_agent      TEXT NOT NULL DEFAULT '',
    cache_status    VARCHAR(64) NOT NULL DEFAULT '',
    status_code     INTEGER NOT NULL DEFAULT 0,
    bytes_sent      BIGINT NOT NULL DEFAULT 0,
    request_length  BIGINT NOT NULL DEFAULT 0,
    request_time_ms INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, logged_at)
) PARTITION BY RANGE (logged_at);
CREATE INDEX IF NOT EXISTS idx_of_node_access_logs_host ON of_node_access_logs (host, logged_at DESC);
CREATE INDEX IF NOT EXISTS idx_of_node_access_logs_host_lower ON of_node_access_logs (lower(trim(host)));
CREATE INDEX IF NOT EXISTS idx_of_node_access_logs_logged_at ON of_node_access_logs (logged_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_of_node_access_logs_node_id ON of_node_access_logs (node_id, logged_at DESC);
CREATE INDEX IF NOT EXISTS idx_of_node_access_logs_remote_addr ON of_node_access_logs (remote_addr, logged_at DESC);
CREATE INDEX IF NOT EXISTS idx_of_node_access_logs_status_code ON of_node_access_logs (status_code, logged_at DESC);

-- +goose StatementBegin
DO $$
DECLARE
    d date;
BEGIN
    FOR d IN SELECT generate_series(date_trunc('month', now())::date, (date_trunc('month', now()) + interval '2 months')::date, interval '1 month')::date
    LOOP
        EXECUTE format('CREATE TABLE IF NOT EXISTS of_node_access_logs_%s PARTITION OF of_node_access_logs FOR VALUES FROM (%L) TO (%L)',
            to_char(d, 'YYYYMM'), d, d + interval '1 month');
    END LOOP;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS of_node_edge_health (
    id          BIGINT NOT NULL PRIMARY KEY,
    node_id     VARCHAR(64) NOT NULL DEFAULT '',
    captured_at TIMESTAMPTZ NOT NULL,
    status      VARCHAR(64) NOT NULL DEFAULT '',
    connections BIGINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_node_edge_health_node ON of_node_edge_health (node_id, captured_at DESC);

CREATE TABLE IF NOT EXISTS of_node_health_events (
    id BIGSERIAL PRIMARY KEY,
    node_id VARCHAR(64) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    severity VARCHAR(16) NOT NULL,
    status VARCHAR(16) NOT NULL,
    message TEXT,
    first_triggered_at TIMESTAMPTZ NOT NULL,
    last_triggered_at TIMESTAMPTZ NOT NULL,
    reported_at TIMESTAMPTZ NOT NULL,
    resolved_at TIMESTAMPTZ,
    metadata_json TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_node_health_events_event_type ON of_node_health_events (event_type);
CREATE INDEX IF NOT EXISTS idx_of_node_health_events_first_triggered_at ON of_node_health_events (first_triggered_at);
CREATE INDEX IF NOT EXISTS idx_of_node_health_events_last_triggered_at ON of_node_health_events (last_triggered_at);
CREATE INDEX IF NOT EXISTS idx_of_node_health_events_node_id ON of_node_health_events (node_id);
CREATE INDEX IF NOT EXISTS idx_of_node_health_events_reported_at ON of_node_health_events (reported_at);
CREATE INDEX IF NOT EXISTS idx_of_node_health_events_resolved_at ON of_node_health_events (resolved_at);
CREATE INDEX IF NOT EXISTS idx_of_node_health_events_status ON of_node_health_events (status);

CREATE TABLE IF NOT EXISTS of_node_metric_snapshots (
    id                 BIGINT NOT NULL PRIMARY KEY,
    node_id            VARCHAR(64) NOT NULL DEFAULT '',
    captured_at        TIMESTAMPTZ NOT NULL,
    cpu_usage_percent  DOUBLE PRECISION NOT NULL DEFAULT 0,
    memory_used_bytes  BIGINT NOT NULL DEFAULT 0,
    memory_total_bytes BIGINT NOT NULL DEFAULT 0,
    storage_used_bytes BIGINT NOT NULL DEFAULT 0,
    storage_total_bytes BIGINT NOT NULL DEFAULT 0,
    disk_read_bytes    BIGINT NOT NULL DEFAULT 0,
    disk_write_bytes   BIGINT NOT NULL DEFAULT 0,
    network_rx_bytes   BIGINT NOT NULL DEFAULT 0,
    network_tx_bytes   BIGINT NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_node_metric_snapshots_node ON of_node_metric_snapshots (node_id, captured_at DESC);

CREATE TABLE IF NOT EXISTS of_node_obs_frpc (
    id                     BIGINT NOT NULL PRIMARY KEY,
    node_id                VARCHAR(64) NOT NULL DEFAULT '',
    captured_at            TIMESTAMPTZ NOT NULL,
    tunnel_status          VARCHAR(16) NOT NULL DEFAULT '',
    connected_relays_count INTEGER NOT NULL DEFAULT 0,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_node_obs_frpc_node ON of_node_obs_frpc (node_id, captured_at DESC);

CREATE TABLE IF NOT EXISTS of_node_obs_frps (
    id                BIGINT NOT NULL PRIMARY KEY,
    node_id           VARCHAR(64) NOT NULL DEFAULT '',
    captured_at       TIMESTAMPTZ NOT NULL,
    frps_connections  INTEGER NOT NULL DEFAULT 0,
    frps_proxy_count  INTEGER NOT NULL DEFAULT 0,
    frps_client_count INTEGER NOT NULL DEFAULT 0,
    frps_proxies      TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_node_obs_frps_node ON of_node_obs_frps (node_id, captured_at DESC);

CREATE TABLE IF NOT EXISTS of_node_system_profiles (
    id BIGSERIAL PRIMARY KEY,
    node_id VARCHAR(64) NOT NULL,
    hostname VARCHAR(255) NOT NULL DEFAULT '',
    os_name VARCHAR(128) NOT NULL DEFAULT '',
    os_version VARCHAR(128) NOT NULL DEFAULT '',
    kernel_version VARCHAR(128) NOT NULL DEFAULT '',
    architecture VARCHAR(64) NOT NULL DEFAULT '',
    cpu_model VARCHAR(255) NOT NULL DEFAULT '',
    cpu_cores INTEGER NOT NULL DEFAULT 0,
    total_memory_bytes BIGINT NOT NULL DEFAULT 0,
    total_disk_bytes BIGINT NOT NULL DEFAULT 0,
    uptime_seconds BIGINT NOT NULL DEFAULT 0,
    reported_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_node_system_profiles_node_id ON of_node_system_profiles (node_id);
CREATE INDEX IF NOT EXISTS idx_of_node_system_profiles_reported_at ON of_node_system_profiles (reported_at);

CREATE TABLE IF NOT EXISTS of_nodes (
    id BIGSERIAL PRIMARY KEY,
    node_id VARCHAR(64) NOT NULL,
    name VARCHAR(128) NOT NULL,
    ip VARCHAR(64) NOT NULL DEFAULT '',
    ip_manual_override BOOLEAN NOT NULL DEFAULT FALSE,
    geo_name VARCHAR(128) NOT NULL DEFAULT '',
    geo_latitude DOUBLE PRECISION,
    geo_longitude DOUBLE PRECISION,
    geo_manual_override BOOLEAN NOT NULL DEFAULT FALSE,
    access_token VARCHAR(128) NOT NULL DEFAULT '',
    auto_update_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    update_requested BOOLEAN NOT NULL DEFAULT FALSE,
    update_channel VARCHAR(16) NOT NULL DEFAULT 'stable',
    update_tag VARCHAR(64) NOT NULL DEFAULT '',
    restart_openresty_requested BOOLEAN NOT NULL DEFAULT FALSE,
    version VARCHAR(64) NOT NULL DEFAULT '',
    ext_version VARCHAR(64) NOT NULL DEFAULT '',
    openresty_status VARCHAR(16) NOT NULL DEFAULT 'unknown',
    openresty_message TEXT,
    status VARCHAR(16) NOT NULL DEFAULT 'offline',
    current_version VARCHAR(32) NOT NULL DEFAULT '',
    last_seen_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    node_type VARCHAR(32) NOT NULL DEFAULT 'edge_node',
    relay_bind_port INTEGER NOT NULL DEFAULT 0,
    relay_vhost_http_port INTEGER NOT NULL DEFAULT 0,
    relay_auth_token VARCHAR(128) NOT NULL DEFAULT '',
    relay_agent_access_addr VARCHAR(255) NOT NULL DEFAULT '',
    relay_client_access_addr VARCHAR(255) NOT NULL DEFAULT '',
    relay_client_proxy_url VARCHAR(512) NOT NULL DEFAULT '',
    capabilities_json TEXT NOT NULL DEFAULT '[]',
    relay_status VARCHAR(16) NOT NULL DEFAULT 'unknown',
    relay_web_server_enabled BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_of_nodes_access_token ON of_nodes (access_token);
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_nodes_node_id ON of_nodes (node_id);

CREATE TABLE IF NOT EXISTS of_origins (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    remark VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_origins_address ON of_origins (address);

CREATE TABLE IF NOT EXISTS of_pages_deployment_files (
    id BIGSERIAL PRIMARY KEY,
    deployment_id BIGINT NOT NULL,
    path VARCHAR(2048) NOT NULL,
    size BIGINT NOT NULL DEFAULT 0,
    checksum VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_pages_deployment_files_deployment_id ON of_pages_deployment_files (deployment_id);

CREATE TABLE IF NOT EXISTS of_pages_deployments (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL,
    deployment_number INTEGER NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'uploaded',
    artifact_path VARCHAR(2048) NOT NULL DEFAULT '',
    file_count INTEGER NOT NULL DEFAULT 0,
    total_size BIGINT NOT NULL DEFAULT 0,
    created_by VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    activated_at TIMESTAMPTZ,
    upload_id BIGINT NOT NULL DEFAULT 0,
    source_type VARCHAR(32) NOT NULL DEFAULT '',
    source_identity CHAR(64),
    source_revision CHAR(64),
    source_label VARCHAR(255) NOT NULL DEFAULT '',
    source_meta TEXT NOT NULL DEFAULT '',
    trigger_type VARCHAR(32) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_of_pages_deployments_checksum ON of_pages_deployments (checksum);
CREATE INDEX IF NOT EXISTS idx_of_pages_deployments_project_id ON of_pages_deployments (project_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_pages_deployments_project_number
    ON of_pages_deployments (project_id, deployment_number);
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_pages_deployments_source_revision
    ON of_pages_deployments (project_id, source_identity, source_revision)
    WHERE source_identity IS NOT NULL AND source_revision IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_of_pages_deployments_status ON of_pages_deployments (status);
CREATE INDEX IF NOT EXISTS idx_of_pages_deployments_upload_id ON of_pages_deployments (upload_id);

CREATE TABLE IF NOT EXISTS of_pages_project_source_runtime (
    source_id BIGINT PRIMARY KEY,
    etag VARCHAR(512) NOT NULL DEFAULT '',
    last_seen_revision CHAR(64) NOT NULL DEFAULT '',
    last_seen_detail TEXT NOT NULL DEFAULT '',
    last_applied_revision CHAR(64) NOT NULL DEFAULT '',
    last_applied_detail TEXT NOT NULL DEFAULT '',
    sync_status VARCHAR(32) NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    last_checked_at TIMESTAMPTZ,
    last_synced_at TIMESTAMPTZ,
    next_check_at TIMESTAMPTZ,
    lease_expires_at TIMESTAMPTZ,
    lease_token VARCHAR(64) NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_pages_project_source_runtime_next_check_at
    ON of_pages_project_source_runtime (next_check_at);

CREATE TABLE IF NOT EXISTS of_pages_project_sources (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL,
    source_type VARCHAR(32) NOT NULL DEFAULT '',
    remote_url TEXT NOT NULL DEFAULT '',
    allow_insecure BOOLEAN NOT NULL DEFAULT FALSE,
    github_repository VARCHAR(255) NOT NULL DEFAULT '',
    release_selector VARCHAR(16) NOT NULL DEFAULT '',
    release_tag VARCHAR(255) NOT NULL DEFAULT '',
    asset_name VARCHAR(255) NOT NULL DEFAULT '',
    auto_update_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    check_interval_minutes INTEGER NOT NULL DEFAULT 0,
    config_version INTEGER NOT NULL DEFAULT 0,
    source_identity CHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_pages_project_sources_project_id
    ON of_pages_project_sources (project_id);

CREATE TABLE IF NOT EXISTS of_pages_projects (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(128) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    spa_fallback_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    spa_fallback_path VARCHAR(512) NOT NULL DEFAULT '/index.html',
    api_proxy_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    api_proxy_path VARCHAR(255) NOT NULL DEFAULT '',
    api_proxy_pass VARCHAR(2048) NOT NULL DEFAULT '',
    api_proxy_rewrite VARCHAR(255) NOT NULL DEFAULT '',
    active_deployment_id BIGINT,
    root_dir VARCHAR(512) NOT NULL DEFAULT '',
    entry_file VARCHAR(512) NOT NULL DEFAULT 'index.html',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    content_config_version INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_of_pages_projects_active_deployment_id ON of_pages_projects (active_deployment_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_pages_projects_slug ON of_pages_projects (slug);

CREATE TABLE IF NOT EXISTS of_proxy_routes (
    id BIGSERIAL PRIMARY KEY,
    site_name VARCHAR(255) NOT NULL DEFAULT '',
    origin_id BIGINT,
    origin_url VARCHAR(2048) NOT NULL,
    origin_host VARCHAR(255) NOT NULL DEFAULT '',
    upstreams TEXT NOT NULL DEFAULT '[]',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    enable_https BOOLEAN NOT NULL DEFAULT FALSE,
    redirect_http BOOLEAN NOT NULL DEFAULT FALSE,
    limit_conn_per_server INTEGER NOT NULL DEFAULT 0,
    limit_conn_per_ip INTEGER NOT NULL DEFAULT 0,
    limit_rate VARCHAR(32) NOT NULL DEFAULT '',
    cache_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    cache_policy VARCHAR(32) NOT NULL DEFAULT '',
    cache_rules TEXT NOT NULL DEFAULT '[]',
    custom_headers TEXT NOT NULL DEFAULT '[]',
    basic_auth_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    basic_auth_username VARCHAR(255) NOT NULL DEFAULT '',
    basic_auth_password VARCHAR(255) NOT NULL DEFAULT '',
    upstream_type VARCHAR(32) NOT NULL DEFAULT 'direct',
    tunnel_node_id BIGINT,
    tunnel_target_addr VARCHAR(512) NOT NULL DEFAULT '',
    tunnel_target_protocol VARCHAR(16) NOT NULL DEFAULT '',
    pages_project_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    limit_req_per_ip VARCHAR(32) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_of_proxy_routes_origin_id ON of_proxy_routes (origin_id);
CREATE INDEX IF NOT EXISTS idx_of_proxy_routes_pages_project_id ON of_proxy_routes (pages_project_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_proxy_routes_site_name ON of_proxy_routes (site_name);
CREATE INDEX IF NOT EXISTS idx_of_proxy_routes_tunnel_node_id ON of_proxy_routes (tunnel_node_id);

CREATE TABLE IF NOT EXISTS of_tls_certificates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    cert_pem TEXT NOT NULL,
    key_pem TEXT NOT NULL,
    not_before TIMESTAMPTZ,
    not_after TIMESTAMPTZ,
    remark VARCHAR(255) NOT NULL DEFAULT '',
    provider VARCHAR(64) NOT NULL DEFAULT 'upload',
    acme_account_id BIGINT NOT NULL DEFAULT 0,
    dns_account_id BIGINT NOT NULL DEFAULT 0,
    key_algorithm VARCHAR(32) NOT NULL DEFAULT '',
    auto_renew BOOLEAN NOT NULL DEFAULT FALSE,
    primary_domain VARCHAR(255) NOT NULL DEFAULT '',
    other_domains TEXT NOT NULL DEFAULT '',
    disable_cname BOOLEAN NOT NULL DEFAULT FALSE,
    skip_dns BOOLEAN NOT NULL DEFAULT FALSE,
    dns1 VARCHAR(128) NOT NULL DEFAULT '',
    dns2 VARCHAR(128) NOT NULL DEFAULT '',
    apply_status VARCHAR(64) NOT NULL DEFAULT 'ready',
    apply_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_tls_certificates_name ON of_tls_certificates (name);

CREATE TABLE IF NOT EXISTS of_waf_ip_groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(32) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ip_list TEXT NOT NULL DEFAULT '[]',
    auto_config TEXT NOT NULL DEFAULT '{}',
    ext_ips TEXT NOT NULL DEFAULT '[]',
    subscription_url VARCHAR(2048) NOT NULL DEFAULT '',
    subscription_format VARCHAR(32) NOT NULL DEFAULT 'text',
    subscription_mapping_rule VARCHAR(255) NOT NULL DEFAULT '',
    sync_interval_minutes INTEGER NOT NULL DEFAULT 1440,
    last_synced_at TIMESTAMPTZ,
    next_sync_at TIMESTAMPTZ,
    last_sync_status VARCHAR(32) NOT NULL DEFAULT '',
    last_sync_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_waf_ip_groups_next_sync_at ON of_waf_ip_groups (next_sync_at);
CREATE INDEX IF NOT EXISTS idx_of_waf_ip_groups_type ON of_waf_ip_groups (type);

CREATE TABLE IF NOT EXISTS of_waf_rule_group_bindings (
    id BIGSERIAL PRIMARY KEY,
    rule_group_id BIGINT NOT NULL,
    proxy_route_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sequence INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_waf_group_route ON of_waf_rule_group_bindings (rule_group_id, proxy_route_id);
CREATE INDEX IF NOT EXISTS idx_of_waf_rule_group_bindings_proxy_route_id ON of_waf_rule_group_bindings (proxy_route_id);

CREATE TABLE IF NOT EXISTS of_waf_rule_groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    is_global BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    graph TEXT NOT NULL DEFAULT '',
    revision BIGINT NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_of_waf_rule_groups_is_global ON of_waf_rule_groups (is_global);

CREATE TABLE IF NOT EXISTS of_zone_domains (
    id BIGSERIAL PRIMARY KEY,
    zone_id BIGINT NOT NULL,
    proxy_route_id BIGINT,
    domain VARCHAR(255) NOT NULL,
    cert_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_of_zone_domains_cert_id ON of_zone_domains (cert_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_zone_domains_domain ON of_zone_domains (domain);
CREATE INDEX IF NOT EXISTS idx_of_zone_domains_proxy_route_id ON of_zone_domains (proxy_route_id);
CREATE INDEX IF NOT EXISTS idx_of_zone_domains_zone_id ON of_zone_domains (zone_id);

CREATE TABLE IF NOT EXISTS of_zones (
    id BIGSERIAL PRIMARY KEY,
    domain VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_of_zones_domain ON of_zones (domain);

-- +goose Down
SELECT 1;
