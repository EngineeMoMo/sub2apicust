import { apiClient } from '@/api/client'

export interface GroupRatePeriod {
  starts_at: string
  ends_at: string
  rate_multiplier: number
}

export interface GroupAccessPolicy {
  group_id: number
  monthly_recharge_enabled: boolean
  monthly_recharge_min: number
  balance_enabled: boolean
  balance_min: number
  expires_at: string | null
  fallback_group_id: number | null
  updated_at: string | null
  rate_schedule_enabled?: boolean
  rate_schedule?: GroupRatePeriod[]
}

export async function getGroupAccessPolicy(groupId: number): Promise<GroupAccessPolicy> {
  const { data } = await apiClient.get<GroupAccessPolicy>(`/admin/groups/${groupId}/access-policy`)
  return data
}

export async function saveGroupAccessPolicy(groupId: number, policy: GroupAccessPolicy): Promise<GroupAccessPolicy> {
  const { data } = await apiClient.put<GroupAccessPolicy>(`/admin/groups/${groupId}/access-policy`, policy)
  return data
}

export function toBeijingDateTime(value: string | null): string {
  if (!value) return ''
  const instant = Date.parse(value)
  if (!Number.isFinite(instant)) return ''
  return new Date(instant + 8 * 3600_000).toISOString().slice(0, 19)
}

export function fromBeijingDateTime(value: string): string | null {
  if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(:\d{2})?$/.test(value)) return null
  const normalized = value.length === 16 ? `${value}:00` : value
  const instant = Date.parse(`${normalized}+08:00`)
  if (!Number.isFinite(instant)) return null
  const iso = new Date(instant).toISOString()
  return toBeijingDateTime(iso) === normalized ? iso : null
}

export function validGroupThreshold(value: number, enabled: boolean): boolean {
  return Number.isFinite(value) && value >= 0 && value <= 1e12 &&
    Math.abs(value - Math.round(value * 100) / 100) < 1e-9 && (!enabled || value > 0)
}

export function normalizeGroupRateSchedule(periods: GroupRatePeriod[]): GroupRatePeriod[] | null {
  if (periods.length > 32) return null
  const result: GroupRatePeriod[] = []
  for (const period of periods) {
    const start = fromBeijingDateTime(period.starts_at)
    const end = fromBeijingDateTime(period.ends_at)
    if (!start || !end || Date.parse(start) >= Date.parse(end) || !Number.isFinite(period.rate_multiplier) || period.rate_multiplier < 0 || period.rate_multiplier > 1000) return null
    result.push({ starts_at: start, ends_at: end, rate_multiplier: period.rate_multiplier })
  }
  result.sort((first, second) => Date.parse(first.starts_at) - Date.parse(second.starts_at))
  if (result.some((period, index) => index > 0 && Date.parse(period.starts_at) < Date.parse(result[index - 1]!.ends_at))) return null
  return result
}
