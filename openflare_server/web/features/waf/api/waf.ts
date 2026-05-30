import { apiRequest } from '@/lib/api/client';

import type {
  WAFRuleGroup,
  WAFRuleGroupPayload,
  WAFSiteRuleGroups,
} from '@/features/waf/types';

export function getWAFRuleGroups() {
  return apiRequest<WAFRuleGroup[]>('/waf/rule-groups');
}

export function createWAFRuleGroup(payload: WAFRuleGroupPayload) {
  return apiRequest<WAFRuleGroup>('/waf/rule-groups', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateWAFRuleGroup(id: number, payload: WAFRuleGroupPayload) {
  return apiRequest<WAFRuleGroup>(`/waf/rule-groups/${id}/update`, {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function deleteWAFRuleGroup(id: number) {
  return apiRequest<void>(`/waf/rule-groups/${id}/delete`, {
    method: 'POST',
  });
}

export function replaceWAFRuleGroupSites(id: number, ids: number[]) {
  return apiRequest<WAFRuleGroup>(`/waf/rule-groups/${id}/sites`, {
    method: 'POST',
    body: JSON.stringify({ ids }),
  });
}

export function getWAFSiteRuleGroups(routeId: number) {
  return apiRequest<WAFSiteRuleGroups>(`/waf/sites/${routeId}/rule-groups`);
}

export function replaceWAFSiteRuleGroups(routeId: number, ids: number[]) {
  return apiRequest<WAFSiteRuleGroups>(`/waf/sites/${routeId}/rule-groups`, {
    method: 'POST',
    body: JSON.stringify({ ids }),
  });
}
