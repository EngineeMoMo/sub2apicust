import { describe, expect, it, vi } from 'vitest'
import { assign } from '@/api/admin/subscriptions'

const { post } = vi.hoisted(() => ({ post: vi.fn() }))
vi.mock('@/api/client', () => ({ apiClient: { post } }))

describe('single-user subscription assignment API', () => {
  it('sends the explicit operation key and the selected stock plan', async () => {
    const request = { user_id: 11, group_id: 7, plan_id: 71, validity_days: 3 }
    post.mockResolvedValue({ data: { id: 1 } })
    expect(await assign(request, 'assign-once')).toEqual({ id: 1 })
    expect(post).toHaveBeenCalledWith('/admin/subscriptions/assign', request, { headers: { 'Idempotency-Key': 'assign-once' } })
  })
})
