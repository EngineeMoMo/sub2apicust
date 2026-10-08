export interface CollectionPolicy {
  enabled: boolean
  single_max: number
  daily_max: number
  timezone: string
  quick_amounts: number[]
  allow_custom_amount: boolean
  contact_text: string
}

export function collectionAllows(policy: CollectionPolicy | undefined, total: number, currency: string): boolean {
  return !policy?.enabled || (currency === 'CNY' && Number.isFinite(total) && total > 0 && Math.round(total * 100) <= Math.round(policy.single_max * 100))
}
