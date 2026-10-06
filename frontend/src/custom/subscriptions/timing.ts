import { onMounted, onUnmounted, ref } from 'vue'
import { apiClient } from '@/api/client'
import type { UserSubscription } from '@/types'

export const quotaPeriods = ['daily', 'weekly', 'monthly'] as const
export type QuotaPeriod = typeof quotaPeriods[number]
export type SubscriptionResetTimes = Partial<Record<QuotaPeriod, string>>

// 原进度接口返回 { subscription, progress }，重置时间以服务端 resets_at 为准。
interface SubscriptionTimingResponse {
  subscription: { id: number }
  progress: Partial<Record<QuotaPeriod, { resets_at: string } | null>>
}

export async function fetchSubscriptionResetTimes(): Promise<Record<number, SubscriptionResetTimes>> {
  const { data } = await apiClient.get<SubscriptionTimingResponse[]>('/subscriptions/progress')
  return Object.fromEntries(data.map(({ subscription, progress }) => [subscription.id,
    Object.fromEntries(quotaPeriods.flatMap(period => {
      const at = progress[period]?.resets_at
      return at ? [[period, at]] : []
    }))
  ]))
}

export function useSubscriptionClock() {
  const now = ref(Date.now())
  let timer: ReturnType<typeof setInterval> | undefined
  const update = () => { now.value = Date.now() }
  onMounted(() => {
    update()
    timer = setInterval(update, 15_000)
    document.addEventListener('visibilitychange', update)
  })
  onUnmounted(() => {
    clearInterval(timer)
    document.removeEventListener('visibilitychange', update)
  })
  return now
}

export function effectiveSubscriptionStatus(subscription: UserSubscription, now: number): UserSubscription['status'] {
  if (subscription.status !== 'active') return subscription.status
  const expiresAt = subscription.expires_at ? Date.parse(subscription.expires_at) : NaN
  return Number.isFinite(expiresAt) && expiresAt <= now ? 'expired' : 'active'
}

export type ResetHint = { kind: 'countdown' | 'pending' | 'ends' | 'unstarted' | 'unavailable' | 'inactive'; at?: string }

export function subscriptionResetHint(subscription: UserSubscription, period: QuotaPeriod,
  resets: SubscriptionResetTimes | undefined, now: number): ResetHint {
  if (effectiveSubscriptionStatus(subscription, now) !== 'active') return { kind: 'inactive' }
  const windowStart = subscription[`${period}_window_start`]
  if (!windowStart) return { kind: 'unstarted' }
  const at = resets?.[period]
  const resetAt = at ? Date.parse(at) : NaN
  if (!Number.isFinite(resetAt)) return { kind: 'unavailable' }
  const expiresAt = subscription.expires_at ? Date.parse(subscription.expires_at) : NaN
  // 最后一个不完整周期不再发放新额度；日卡同样以订阅到期结束。
  if (Number.isFinite(expiresAt) && resetAt >= expiresAt) return { kind: 'ends', at: subscription.expires_at! }
  return { kind: resetAt <= now ? 'pending' : 'countdown', at }
}

export function remainingTimeLabel(target: string, now: number, zh: boolean): string {
  const difference = Date.parse(target) - now
  if (!Number.isFinite(difference) || difference <= 0) return ''
  if (difference < 60_000) return zh ? '不到 1 分钟' : 'less than 1 minute'
  const minutes = Math.ceil(difference / 60_000)
  const days = Math.floor(minutes / 1440)
  const hours = Math.floor((minutes % 1440) / 60)
  const remainder = minutes % 60
  if (days) return zh ? `${days} 天 ${hours} 小时` : `${days}d ${hours}h`
  if (hours) return zh ? `${hours} 小时 ${remainder} 分钟` : `${hours}h ${remainder}m`
  return zh ? `${remainder} 分钟` : `${remainder}m`
}
