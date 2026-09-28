import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, create } = vi.hoisted(() => {
  const get = vi.fn()
  return { get, create: vi.fn(() => ({ get })) }
})
vi.mock('axios', () => ({ default: { create } }))
vi.mock('@/api/url', () => ({ getAPIBaseURL: () => '/api/v1' }))
import { fetchPublicPlans } from '@/custom/guest/api'

beforeEach(() => { get.mockReset() })

describe('游客目录请求隔离', () => {
  it('独立客户端不附带认证拦截器或凭证', async () => {
    localStorage.setItem('auth_token', 'private-test-token')
    get.mockResolvedValue({ data: { code: 0, data: { plans: [], purchase_enabled: false } } })
    expect(await fetchPublicPlans()).toEqual({ plans: [], purchase_enabled: false })
    expect(create).toHaveBeenCalledWith({ baseURL: '/api/v1', timeout: 15000, withCredentials: false })
    expect(get).toHaveBeenCalledTimes(1)
    expect(get).toHaveBeenCalledWith('/payment/public/plans')
    localStorage.removeItem('auth_token')
  })
  it.each([{ code: 500 }, '<html>not a catalog</html>', { code: 0, data: {} }])('拒绝无效响应', async data => {
    get.mockResolvedValue({ data })
    await expect(fetchPublicPlans()).rejects.toThrow('套餐数据暂不可用')
  })
  it('请求错误交给页面展示，不进行登录回跳', async () => {
    const error = new Error('network unavailable')
    get.mockRejectedValue(error)
    await expect(fetchPublicPlans()).rejects.toBe(error)
  })
})
