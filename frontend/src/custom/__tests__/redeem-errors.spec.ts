import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { apiClient } from '@/api/client'
import RedeemView from '@/views/user/RedeemView.vue'

const { refreshUser, fetchActiveSubscriptions, showError, showSuccess } = vi.hoisted(() => ({
  refreshUser: vi.fn(),
  fetchActiveSubscriptions: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/i18n', async () => ({
  ...await vi.importActual<typeof import('@/i18n')>('@/i18n'),
  getLocale: () => 'zh-CN',
}))
vi.mock('@/api', async () => {
  const { redeem } = await import('@/api/redeem')
  return {
    redeemAPI: { redeem, getHistory: vi.fn().mockResolvedValue({ items: [], total: 0 }) },
    authAPI: { getPublicSettings: vi.fn().mockResolvedValue({}) },
  }
})
vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ user: { balance: 10, concurrency: 2 }, refreshUser }),
}))
vi.mock('@/stores/subscriptions', () => ({
  useSubscriptionStore: () => ({ fetchActiveSubscriptions }),
}))
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess, showWarning: vi.fn() }),
}))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key }),
}))

describe('订阅兑换错误展示与真实客户端响应', () => {
  const originalAdapter = apiClient.defaults.adapter

  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  afterEach(() => {
    apiClient.defaults.adapter = originalAdapter
  })

  it.each([
    [400, 'REDEEM_CODE_INVALID', 'invalid subscription redeem code: missing group_id'],
    [400, 'GROUP_NOT_SUBSCRIPTION_TYPE', 'group is not a subscription type'],
    [400, 'SUBSCRIPTION_STOCK_PLAN_REQUIRED', '请先创建或选择该分组要扣库存的套餐；无法唯一确定套餐'],
    [400, 'SUBSCRIPTION_STOCK_PLAN_INVALID', '库存套餐不存在或不属于此订阅分组'],
    [409, 'PLAN_OUT_OF_STOCK', '套餐已售罄，请选择其他套餐或联系管理员'],
  ])('HTTP %s / %s 在页面和提示中显示服务器原始原因', async (status, reason, message) => {
    const adapter = vi.fn(async (config: InternalAxiosRequestConfig) => {
      throw new AxiosError('Request failed', 'ERR_BAD_REQUEST', config, undefined, {
        status: Number(status), statusText: 'Error', headers: {}, config,
        data: { code: status, reason, message },
      })
    })
    apiClient.defaults.adapter = adapter
    const wrapper = mount(RedeemView, {
      global: { stubs: { AppLayout: { template: '<div><slot /></div>' }, Icon: true } },
    })

    try {
      await flushPromises()
      await wrapper.get('input#code').setValue(' SUBSCRIPTION-CODE ')
      await wrapper.get('form').trigger('submit')
      await flushPromises()

      expect(adapter).toHaveBeenCalledOnce()
      const request = adapter.mock.calls[0][0]
      expect(request.url).toBe('/redeem')
      expect(JSON.parse(request.data)).toEqual({ code: 'SUBSCRIPTION-CODE' })
      expect(wrapper.text()).toContain(String(message))
      expect(wrapper.text()).not.toContain('redeem.failedToRedeem')
      expect(showError).toHaveBeenCalledWith(message)
      expect(showSuccess).not.toHaveBeenCalled()
      expect(refreshUser).not.toHaveBeenCalled()
      expect(fetchActiveSubscriptions).not.toHaveBeenCalled()
      expect((wrapper.get('input#code').element as HTMLInputElement).value).toBe(' SUBSCRIPTION-CODE ')
      expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
    } finally {
      wrapper.unmount()
    }
  })
})
