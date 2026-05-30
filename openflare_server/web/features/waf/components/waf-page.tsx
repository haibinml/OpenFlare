'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Check, Globe2, Plus, Save, Search, ShieldCheck, Trash2 } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';

import { EmptyState } from '@/components/feedback/empty-state';
import { ErrorState } from '@/components/feedback/error-state';
import { InlineMessage } from '@/components/feedback/inline-message';
import { LoadingState } from '@/components/feedback/loading-state';
import { PageHeader } from '@/components/layout/page-header';
import { AppCard } from '@/components/ui/app-card';
import { Drawer } from '@/components/ui/drawer';
import { getProxyRoutes } from '@/features/proxy-routes/api/proxy-routes';
import type { ProxyRouteItem } from '@/features/proxy-routes/types';
import {
  DangerButton,
  PrimaryButton,
  ResourceField,
  ResourceInput,
  ResourceTextarea,
  SecondaryButton,
  ToggleField,
} from '@/features/shared/components/resource-primitives';
import {
  createWAFRuleGroup,
  deleteWAFRuleGroup,
  getWAFRuleGroups,
  replaceWAFRuleGroupSites,
  updateWAFRuleGroup,
} from '@/features/waf/api/waf';
import type { WAFRuleGroup, WAFRuleGroupPayload } from '@/features/waf/types';
import { cn } from '@/lib/utils/cn';

type FeedbackState = {
  tone: 'success' | 'danger' | 'info';
  message: string;
};

const emptyDraft: WAFRuleGroupPayload = {
  name: '',
  enabled: true,
  block_status_code: 418,
  block_response_body: '',
  ip_whitelist: [],
  ip_blacklist: [],
  country_whitelist: [],
  country_blacklist: [],
  region_whitelist: [],
  region_blacklist: [],
  remark: '',
};

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : '操作失败';
}

function listToText(items: string[]) {
  return items.join('\n');
}

function textToList(text: string) {
  return text
    .split(/[\n,，\s]+/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function buildDraft(group: WAFRuleGroup | null): WAFRuleGroupPayload {
  if (!group) {
    return { ...emptyDraft };
  }
  return {
    name: group.name,
    enabled: group.enabled,
    block_status_code: group.block_status_code || 418,
    block_response_body: group.block_response_body ?? '',
    ip_whitelist: group.ip_whitelist ?? [],
    ip_blacklist: group.ip_blacklist ?? [],
    country_whitelist: group.country_whitelist ?? [],
    country_blacklist: group.country_blacklist ?? [],
    region_whitelist: group.region_whitelist ?? [],
    region_blacklist: group.region_blacklist ?? [],
    remark: group.remark ?? '',
  };
}

function ruleCount(group: WAFRuleGroup) {
  return (
    group.ip_whitelist.length +
    group.ip_blacklist.length +
    group.country_whitelist.length +
    group.country_blacklist.length
  );
}

function SiteApplyDrawer({
  group,
  routes,
  open,
  onOpenChange,
  onSave,
  pending,
}: {
  group: WAFRuleGroup | null;
  routes: ProxyRouteItem[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (ids: number[]) => void;
  pending: boolean;
}) {
  const [keyword, setKeyword] = useState('');
  const [selectedIDs, setSelectedIDs] = useState<number[]>([]);

  useEffect(() => {
    setSelectedIDs(group?.applied_site_ids ?? []);
    setKeyword('');
  }, [group, open]);

  const filteredRoutes = useMemo(() => {
    const normalized = keyword.trim().toLowerCase();
    if (!normalized) {
      return routes;
    }
    return routes.filter((route) =>
      [route.site_name, route.primary_domain, ...route.domains]
        .join(' ')
        .toLowerCase()
        .includes(normalized),
    );
  }, [keyword, routes]);

  const selectedSet = useMemo(() => new Set(selectedIDs), [selectedIDs]);
  const toggleID = (id: number) => {
    setSelectedIDs((current) =>
      current.includes(id)
        ? current.filter((item) => item !== id)
        : [...current, id].sort((left, right) => left - right),
    );
  };

  const selectFiltered = () => {
    const next = new Set(selectedIDs);
    filteredRoutes.forEach((route) => next.add(route.id));
    setSelectedIDs([...next].sort((left, right) => left - right));
  };

  return (
    <Drawer
      open={open}
      onOpenChange={onOpenChange}
      direction="right"
      title={group ? `应用 ${group.name}` : '应用规则组'}
      description="选择这个自定义规则组要叠加到哪些网站。"
      footer={
        <div className="flex justify-end gap-3">
          <SecondaryButton type="button" onClick={() => onOpenChange(false)}>
            取消
          </SecondaryButton>
          <PrimaryButton
            type="button"
            disabled={!group || pending}
            onClick={() => onSave(selectedIDs)}
          >
            {pending ? '保存中...' : '保存应用范围'}
          </PrimaryButton>
        </div>
      }
    >
      <div className="space-y-4">
        <div className="flex items-center gap-3 rounded-2xl border border-[var(--border-default)] bg-[var(--surface-elevated)] px-4 py-3">
          <Search className="h-4 w-4 text-[var(--foreground-secondary)]" />
          <input
            value={keyword}
            onChange={(event) => setKeyword(event.target.value)}
            placeholder="搜索网站或域名"
            className="min-w-0 flex-1 bg-transparent text-sm text-[var(--foreground-primary)] outline-none placeholder:text-[var(--foreground-muted)]"
          />
          <button
            type="button"
            onClick={selectFiltered}
            className="text-xs font-medium text-[var(--brand-primary)]"
          >
            全选当前
          </button>
        </div>
        <div className="space-y-2">
          {filteredRoutes.map((route) => (
            <button
              key={route.id}
              type="button"
              onClick={() => toggleID(route.id)}
              className={cn(
                'flex w-full items-center gap-3 rounded-2xl border px-4 py-3 text-left transition',
                selectedSet.has(route.id)
                  ? 'border-[var(--border-strong)] bg-[var(--accent-soft)]'
                  : 'border-[var(--border-default)] bg-[var(--surface-elevated)] hover:bg-[var(--surface-muted)]',
              )}
            >
              <span
                className={cn(
                  'flex h-5 w-5 items-center justify-center rounded-md border',
                  selectedSet.has(route.id)
                    ? 'border-[var(--brand-primary)] bg-[var(--brand-primary)] text-[var(--foreground-inverse)]'
                    : 'border-[var(--border-default)]',
                )}
              >
                {selectedSet.has(route.id) ? <Check className="h-3 w-3" /> : null}
              </span>
              <span className="min-w-0 flex-1">
                <span className="block truncate text-sm font-medium text-[var(--foreground-primary)]">
                  {route.site_name}
                </span>
                <span className="block truncate text-xs text-[var(--foreground-secondary)]">
                  {route.domains.join(', ')}
                </span>
              </span>
            </button>
          ))}
        </div>
      </div>
    </Drawer>
  );
}

export function WAFPage() {
  const queryClient = useQueryClient();
  const [selectedID, setSelectedID] = useState<number | null>(null);
  const [draft, setDraft] = useState<WAFRuleGroupPayload>(emptyDraft);
  const [feedback, setFeedback] = useState<FeedbackState | null>(null);
  const [applyGroup, setApplyGroup] = useState<WAFRuleGroup | null>(null);

  const groupsQuery = useQuery({
    queryKey: ['waf', 'rule-groups'],
    queryFn: getWAFRuleGroups,
  });
  const routesQuery = useQuery({
    queryKey: ['proxy-routes'],
    queryFn: getProxyRoutes,
  });

  const groups = useMemo(() => groupsQuery.data ?? [], [groupsQuery.data]);
  const routes = useMemo(() => routesQuery.data ?? [], [routesQuery.data]);
  const selectedGroup = useMemo(
    () =>
      selectedID === 0
        ? null
        : (groups.find((group) => group.id === selectedID) ?? groups[0] ?? null),
    [groups, selectedID],
  );

  useEffect(() => {
    if (selectedGroup) {
      setSelectedID(selectedGroup.id);
      setDraft(buildDraft(selectedGroup));
    }
  }, [selectedGroup]);

  const invalidate = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['waf', 'rule-groups'] }),
      queryClient.invalidateQueries({ queryKey: ['config-versions', 'diff'] }),
    ]);
  };

  const saveMutation = useMutation({
    mutationFn: (payload: WAFRuleGroupPayload) => {
      if (selectedGroup) {
        return updateWAFRuleGroup(selectedGroup.id, payload);
      }
      return createWAFRuleGroup(payload);
    },
    onSuccess: async (group) => {
      setSelectedID(group.id);
      setFeedback({ tone: 'success', message: 'WAF 规则组已保存。' });
      await invalidate();
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteWAFRuleGroup,
    onSuccess: async () => {
      setSelectedID(null);
      setFeedback({ tone: 'success', message: 'WAF 规则组已删除。' });
      await invalidate();
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  const applyMutation = useMutation({
    mutationFn: ({ id, ids }: { id: number; ids: number[] }) =>
      replaceWAFRuleGroupSites(id, ids),
    onSuccess: async () => {
      setApplyGroup(null);
      setFeedback({ tone: 'success', message: '规则组应用范围已更新。' });
      await invalidate();
    },
    onError: (error) => {
      setFeedback({ tone: 'danger', message: getErrorMessage(error) });
    },
  });

  if (groupsQuery.isLoading || routesQuery.isLoading) {
    return <LoadingState />;
  }
  if (groupsQuery.isError) {
    return <ErrorState title="WAF 加载失败" description={getErrorMessage(groupsQuery.error)} />;
  }
  if (routesQuery.isError) {
    return <ErrorState title="网站列表加载失败" description={getErrorMessage(routesQuery.error)} />;
  }
  if (!selectedGroup && groups.length === 0) {
    return <EmptyState title="WAF 尚未初始化" description="刷新页面后系统会自动创建全局规则组。" />;
  }

  const enabledCount = groups.filter((group) => group.enabled).length;
  const protectedSites = new Set(groups.flatMap((group) => group.applied_site_ids));
  const totalRules = groups.reduce((sum, group) => sum + ruleCount(group), 0);

  return (
    <>
      <div className="space-y-6">
        <PageHeader
          title="WAF"
          description="按规则组维护 IP 与地域黑白名单，全局规则始终应用到所有网站。"
          action={
            <PrimaryButton
              type="button"
              onClick={() => {
                setSelectedID(0);
                setDraft({ ...emptyDraft, name: '自定义规则组' });
              }}
            >
              <Plus className="mr-2 h-4 w-4" />
              新建规则组
            </PrimaryButton>
          }
        />

        {feedback ? <InlineMessage tone={feedback.tone} message={feedback.message} /> : null}

        <div className="grid gap-4 xl:grid-cols-3">
          <AppCard>
            <p className="text-sm text-[var(--foreground-secondary)]">启用规则组</p>
            <p className="mt-2 text-3xl font-semibold text-[var(--foreground-primary)]">{enabledCount}</p>
          </AppCard>
          <AppCard>
            <p className="text-sm text-[var(--foreground-secondary)]">自定义覆盖网站</p>
            <p className="mt-2 text-3xl font-semibold text-[var(--foreground-primary)]">{protectedSites.size}</p>
          </AppCard>
          <AppCard>
            <p className="text-sm text-[var(--foreground-secondary)]">黑白名单条目</p>
            <p className="mt-2 text-3xl font-semibold text-[var(--foreground-primary)]">{totalRules}</p>
          </AppCard>
        </div>

        <div className="grid gap-5 xl:grid-cols-[360px_minmax(0,1fr)]">
          <AppCard title="规则组">
            <div className="space-y-2">
              {groups.map((group) => (
                <button
                  key={group.id}
                  type="button"
                  onClick={() => setSelectedID(group.id)}
                  className={cn(
                    'w-full rounded-2xl border px-4 py-3 text-left transition',
                    selectedGroup?.id === group.id
                      ? 'border-[var(--border-strong)] bg-[var(--accent-soft)]'
                      : 'border-[var(--border-default)] bg-[var(--surface-elevated)] hover:bg-[var(--surface-muted)]',
                  )}
                >
                  <span className="flex items-center justify-between gap-3">
                    <span className="flex min-w-0 items-center gap-2">
                      {group.is_global ? <Globe2 className="h-4 w-4" /> : <ShieldCheck className="h-4 w-4" />}
                      <span className="truncate text-sm font-semibold text-[var(--foreground-primary)]">
                        {group.name}
                      </span>
                    </span>
                    <span className="text-xs text-[var(--foreground-secondary)]">
                      {group.enabled ? '启用' : '停用'}
                    </span>
                  </span>
                  <span className="mt-2 block text-xs text-[var(--foreground-secondary)]">
                    {group.is_global ? '应用全部网站' : `已应用 ${group.applied_site_count} 个网站`} · {ruleCount(group)} 条规则
                  </span>
                </button>
              ))}
            </div>
          </AppCard>

          <AppCard
            title={selectedGroup ? selectedGroup.name : '新建规则组'}
            description="白名单命中后直接放行；未命中白名单时继续判断黑名单。"
            action={
              selectedGroup && !selectedGroup.is_global ? (
                <SecondaryButton type="button" onClick={() => setApplyGroup(selectedGroup)}>
                  一键应用
                </SecondaryButton>
              ) : null
            }
          >
            <div className="grid gap-5 xl:grid-cols-2">
              <ResourceField label="规则组名称">
                <ResourceInput
                  value={draft.name}
                  disabled={selectedGroup?.is_global}
                  onChange={(event) => setDraft((current) => ({ ...current, name: event.target.value }))}
                />
              </ResourceField>
              <ResourceField label="拦截状态码">
                <ResourceInput
                  type="number"
                  min={400}
                  max={599}
                  value={draft.block_status_code}
                  onChange={(event) =>
                    setDraft((current) => ({ ...current, block_status_code: Number(event.target.value) }))
                  }
                />
              </ResourceField>
              <ToggleField
                label="启用规则组"
                checked={draft.enabled}
                onChange={(checked) => setDraft((current) => ({ ...current, enabled: checked }))}
              />
              <ResourceField label="备注">
                <ResourceInput
                  value={draft.remark}
                  onChange={(event) => setDraft((current) => ({ ...current, remark: event.target.value }))}
                />
              </ResourceField>
              <ResourceField label="IP / IP 段白名单" hint="每行一个 IP 或 CIDR。">
                <ResourceTextarea
                  value={listToText(draft.ip_whitelist)}
                  onChange={(event) =>
                    setDraft((current) => ({ ...current, ip_whitelist: textToList(event.target.value) }))
                  }
                />
              </ResourceField>
              <ResourceField label="IP / IP 段黑名单" hint="每行一个 IP 或 CIDR。">
                <ResourceTextarea
                  value={listToText(draft.ip_blacklist)}
                  onChange={(event) =>
                    setDraft((current) => ({ ...current, ip_blacklist: textToList(event.target.value) }))
                  }
                />
              </ResourceField>
              <ResourceField label="国家白名单" hint="ISO 两位国家代码，例如 CN、US。">
                <ResourceTextarea
                  value={listToText(draft.country_whitelist)}
                  onChange={(event) =>
                    setDraft((current) => ({ ...current, country_whitelist: textToList(event.target.value) }))
                  }
                />
              </ResourceField>
              <ResourceField label="国家黑名单" hint="ISO 两位国家代码，例如 CN、US。">
                <ResourceTextarea
                  value={listToText(draft.country_blacklist)}
                  onChange={(event) =>
                    setDraft((current) => ({ ...current, country_blacklist: textToList(event.target.value) }))
                  }
                />
              </ResourceField>
              <ResourceField label="拦截页面" className="xl:col-span-2" hint="留空时只返回状态码。">
                <ResourceTextarea
                  value={draft.block_response_body}
                  onChange={(event) =>
                    setDraft((current) => ({ ...current, block_response_body: event.target.value }))
                  }
                />
              </ResourceField>
            </div>
            <div className="mt-6 flex flex-wrap justify-between gap-3">
              <div>
                {selectedGroup && !selectedGroup.is_global ? (
                  <DangerButton
                    type="button"
                    disabled={deleteMutation.isPending}
                    onClick={() => {
                      if (window.confirm(`确认删除 WAF 规则组 ${selectedGroup.name} 吗？`)) {
                        deleteMutation.mutate(selectedGroup.id);
                      }
                    }}
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    删除
                  </DangerButton>
                ) : null}
              </div>
              <PrimaryButton
                type="button"
                disabled={saveMutation.isPending}
                onClick={() => saveMutation.mutate(draft)}
              >
                <Save className="mr-2 h-4 w-4" />
                {saveMutation.isPending ? '保存中...' : '保存规则组'}
              </PrimaryButton>
            </div>
          </AppCard>
        </div>
      </div>
      <SiteApplyDrawer
        group={applyGroup}
        routes={routes}
        open={Boolean(applyGroup)}
        pending={applyMutation.isPending}
        onOpenChange={(open) => {
          if (!open) {
            setApplyGroup(null);
          }
        }}
        onSave={(ids) => {
          if (applyGroup) {
            applyMutation.mutate({ id: applyGroup.id, ids });
          }
        }}
      />
    </>
  );
}
