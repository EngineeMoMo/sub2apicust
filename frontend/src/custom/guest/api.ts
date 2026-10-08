import axios from 'axios'
import { getAPIBaseURL } from '@/api/url'

export interface PublicPlan {
  // [CUSTOM] -1 不限量；旧接口遗漏字段时兼容原购买行为。
  stock_limit?: number
  stock_used?: number
  stock_remaining?: number
  id: number
  name: string
  description: string
  price: number
  original_price?: number
  currency?: string
  validity_days: number
  validity_unit: string
  features: string[]
  daily_limit_usd?: number
  weekly_limit_usd?: number
  monthly_limit_usd?: number
  supported_model_scopes: string[]
}

export interface PublicCatalog {
  plans: PublicPlan[]
  purchase_enabled: boolean
}

const publicClient = axios.create({ baseURL: getAPIBaseURL(), timeout: 15000, withCredentials: false })

export async function fetchPublicPlans(): Promise<PublicCatalog> {
  const { data } = await publicClient.get('/payment/public/plans')
  if (data?.code !== 0 || !Array.isArray(data?.data?.plans)) throw new Error('套餐数据暂不可用')
  return data.data
}
