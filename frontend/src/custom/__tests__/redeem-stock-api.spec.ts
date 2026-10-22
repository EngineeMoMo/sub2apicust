import { beforeEach, describe, expect, it, vi } from 'vitest'
import { generate, batchUpdate } from '@/api/admin/redeem'
const api = vi.hoisted(() => ({ post: vi.fn() }))
vi.mock('@/api/client', () => ({ apiClient: api }))
describe('订阅兑换库存API参数', () => {
 beforeEach(() => { vi.clearAllMocks(); api.post.mockResolvedValue({ data: [] }) })
 it('透传订阅套餐ID和有效天数', async () => {
  await generate(2, 'subscription', 0, 10, 7, undefined, 22)
  expect(api.post).toHaveBeenCalledWith('/admin/redeem-codes/generate', { count: 2, type: 'subscription', value: 0, group_id: 10, validity_days: 7, plan_id: 22 })
 })
 it('余额码不携带订阅库存参数', async () => {
  await generate(1, 'balance', 10, 10, 7, undefined, 22)
  expect(api.post).toHaveBeenCalledWith('/admin/redeem-codes/generate', { count: 1, type: 'balance', value: 10 })
 })
 it('批量修复旧订阅码时同时传分组和库存来源', async () => {
  await batchUpdate([1], { group_id: 10, plan_id: 22 })
  expect(api.post).toHaveBeenCalledWith('/admin/redeem-codes/batch-update', { ids: [1], fields: { group_id: 10, plan_id: 22 } })
 })
})
