export interface CollectionPolicy {
  enabled: boolean
  single_max: number
  daily_max: number
  timezone: string
  quick_amounts: number[]
  allow_custom_amount: boolean
  contact_text: string
  // 同一用户未付占用上限；旧配置缺省时后端按单笔上限的3倍推导。
  user_hold_max?: number
  // 无法证明付款期限的订单在期限后再等多少小时复查释放，默认24。
  unproven_release_hours?: number
  missing_provider_auto_release?: boolean
}

export interface CollectionHold {
  order_id: number
  amount: number
  status: string
  expires_at?: string
  manual_release_after?: string
}

// 后台时间按收款统计时区展示；无效时间显示占位，不抛错。
export function formatCollectionTime(value: string | undefined, timezone = 'Asia/Shanghai'): string {
  const time = value ? new Date(value) : undefined
  if (!time || Number.isNaN(time.getTime())) return '—'
  try { return time.toLocaleString('zh-CN', { timeZone: timezone, hour12: false }) }
  catch { return time.toLocaleString('zh-CN', { hour12: false }) }
}

export function canManuallyRelease(hold: CollectionHold, now = Date.now()): boolean {
  const after = hold.manual_release_after ? Date.parse(hold.manual_release_after) : Number.NaN
  return Number.isFinite(after) && now >= after
}

export function collectionAllows(policy: CollectionPolicy | undefined, total: number, currency: string): boolean {
  return !policy?.enabled || (currency === 'CNY' && Number.isFinite(total) && total > 0 && Math.round(total * 100) <= Math.round(policy.single_max * 100))
}
