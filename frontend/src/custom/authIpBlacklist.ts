import { apiClient } from '@/api/client'

export interface AuthIPConfig {
  enabled: boolean
  threshold: number
  window_seconds: number
  version: number
  block_seconds: number
  block_mode: 'timed' | 'permanent'
}

export interface AuthIPEntry {
  id: number
  ip: string
  generation: number
  blocked: boolean
  blocked_at: string
  trigger_path: string
  trigger_count: number
  window_seconds: number
  first_event_at: string | null
  last_event_at: string | null
  rule_version: number
  blocked_until?: string | null
  source: 'auto' | 'manual'
  reason: string
}

export interface AuthIPManualBlock {
  ip: string
  block_mode: 'timed' | 'permanent'
  block_seconds: number
  reason: string
}

export interface AuthIPPage {
  items: AuthIPEntry[]
  total: number
  page: number
  page_size: number
  pages: number
}

const base = '/admin/ip-blacklist'
export const authIPBlacklistAPI = {
  async config(): Promise<AuthIPConfig> {
    return (await apiClient.get<AuthIPConfig>(`${base}/config`)).data
  },
  async save(config: AuthIPConfig): Promise<AuthIPConfig> {
    return (await apiClient.put<AuthIPConfig>(`${base}/config`, config)).data
  },
  async list(ip: string, page: number): Promise<AuthIPPage> {
    return (await apiClient.get<AuthIPPage>(`${base}/entries`, { params: { ip, page, page_size: 20 } })).data
  },
  async unblock(entry: AuthIPEntry): Promise<AuthIPEntry> {
    return (await apiClient.post<AuthIPEntry>(`${base}/entries/${entry.id}/unblock`, { generation: entry.generation })).data
  },
  async add(input: AuthIPManualBlock): Promise<AuthIPEntry> {
    return (await apiClient.post<AuthIPEntry>(`${base}/entries`, input)).data
  }
}

export function authIPError(error: unknown): { conflict: boolean; message: string } {
  const detail = error as { status?: number; message?: string } | null
  return { conflict: detail?.status === 409, message: detail?.message || '请求失败，请稍后重试' }
}

export function formatAuthIPTime(value: string | null, timezone: string): string {
  if (!value) return '—'
  const instant = new Date(value)
  return Number.isNaN(instant.getTime()) ? '—' : instant.toLocaleString('zh-CN', { timeZone: timezone, hour12: false })
}
