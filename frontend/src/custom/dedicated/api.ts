import { apiClient } from '@/api/client'
export interface DedicatedBinding {
  config_status?: string
  user_ids?: number[]
  users?: Array<{ id: number; name: string }>
  platform?: string
  user_name?: string
  account_name?: string
  group_name?: string
  id: number
  user_id: number
  account_id: number
  group_id: number
  label: string
  expires_at: string
  revoked_at: string | null
  updated_at: string
}
export interface DedicatedInput {
  expected_updated_at?: string
  reactivate?: boolean
  user_ids?: number[]
  user_id: number
  account_id: number
  group_id: number
  label: string
  expires_at: string
}
export interface DedicatedWindow {
  key: string
  remaining_percent: number | null
  resets_at: string | null
  stale: boolean
}
export interface DedicatedView {
  group_name?: string
  id: number
  label: string
  platform: string
  group_id: number
  expires_at: string
  status: string
  last_used_at: string | null
  sampled_at: string | null
  checked_at: string
  quota_state: string
  windows: DedicatedWindow[]
}
export interface DedicatedChoice { id: number; label: string }
export type ChoiceKind = 'users' | 'accounts' | 'groups'

export function dedicatedDisplayName(name: string | undefined, id: number): string {
  return name?.trim() ? `${name.trim()} #${id}` : `#${id}`
}

export const dedicatedAPI = {
  async mine(page = 1): Promise<DedicatedView[]> {
    return (await apiClient.get('/dedicated-accounts', { params: { page } })).data
  },
  async list(page = 1): Promise<DedicatedBinding[]> {
    return (await apiClient.get('/admin/dedicated-accounts', { params: { page } })).data
  },
  async save(id: number | null, input: DedicatedInput): Promise<DedicatedBinding> {
    return (id
      ? await apiClient.put('/admin/dedicated-accounts/' + id, input)
      : await apiClient.post('/admin/dedicated-accounts', input)).data
  },
  async revoke(id: number): Promise<void> {
    await apiClient.post('/admin/dedicated-accounts/' + id + '/revoke')
  },
  async remove(id: number): Promise<void> {
    await apiClient.delete('/admin/dedicated-accounts/' + id)
  },
  async choices(kind: ChoiceKind, search: string, platform: string): Promise<DedicatedChoice[]> {
    const params = { page: 1, page_size: 30, search, status: 'active', ...(kind !== 'users' ? { platform } : {}), ...(kind === 'groups' ? { is_exclusive: true } : {}) }
    const { data } = await apiClient.get('/admin/' + kind, { params })
    const items = data.items as Array<{ id: number; name?: string; email?: string; type?: string; subscription_type?: string; parent_account_id?: number | null }>
    return items.filter(item => kind !== 'accounts' || (!item.parent_account_id && (item.type === 'oauth' || (platform === 'anthropic' && item.type === 'setup-token'))))
      .filter(item => kind !== 'groups' || item.subscription_type === 'standard')
      .map(item => ({ id: item.id, label: '#' + item.id + ' · ' + (item.email || item.name || '') }))
  }
}

export function remainingQuota(window: DedicatedWindow, sampledAt: string | null, now: number): number | null {
  const sampled = sampledAt ? Date.parse(sampledAt) : NaN
  const reset = window.resets_at ? Date.parse(window.resets_at) : null
  if (window.stale || !Number.isFinite(sampled) || now - sampled > 15 * 60_000 || sampled > now + 60_000 || (reset !== null && (!Number.isFinite(reset) || now >= reset))) return null
  const value = window.remaining_percent
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= 100 ? value : null
}

export function effectiveStatus(account: DedicatedView, now: number): string {
  if (account.status === 'revoked') return 'revoked'
  if (!Number.isFinite(Date.parse(account.expires_at))) return 'unavailable'
  return now >= Date.parse(account.expires_at) ? 'expired' : account.status
}

export function dedicatedErrorCode(error: unknown): string {
  const candidate = error as { response?: { data?: { code?: unknown; reason?: unknown } }; code?: unknown; reason?: unknown }
  const code = candidate?.response?.data?.reason ?? candidate?.response?.data?.code ?? candidate?.reason ?? candidate?.code
  return typeof code === 'string' ? code : ''
}
