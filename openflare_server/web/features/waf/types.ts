import type { ProxyRoutePoWConfig } from '@/features/proxy-routes/types';

export interface WAFRuleGroup {
  id: number;
  name: string;
  enabled: boolean;
  is_global: boolean;
  block_status_code: number;
  block_response_body: string;
  ip_whitelist: string[];
  ip_blacklist: string[];
  country_whitelist: string[];
  country_blacklist: string[];
  region_whitelist: string[];
  region_blacklist: string[];
  pow_enabled: boolean;
  pow_config: ProxyRoutePoWConfig;
  remark: string;
  applied_site_ids: number[];
  applied_site_count: number;
  created_at: string;
  updated_at: string;
}

export interface WAFRuleGroupPayload {
  name: string;
  enabled: boolean;
  block_status_code: number;
  block_response_body: string;
  ip_whitelist: string[];
  ip_blacklist: string[];
  country_whitelist: string[];
  country_blacklist: string[];
  region_whitelist: string[];
  region_blacklist: string[];
  pow_enabled: boolean;
  pow_config: ProxyRoutePoWConfig;
  remark: string;
}

export interface WAFSiteRuleGroups {
  route_id: number;
  global_rule_group: WAFRuleGroup | null;
  rule_groups: WAFRuleGroup[];
  applied_rule_groups: WAFRuleGroup[];
  applied_ids: number[];
}
