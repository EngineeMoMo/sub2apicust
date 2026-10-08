import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import PaymentCollectionSettings from '@/custom/components/PaymentCollectionSettings.vue'
import { collectionAllows, type CollectionPolicy } from '@/custom/paymentCollection'

const api = vi.hoisted(() => ({ getConfig: vi.fn(), updateConfig: vi.fn(), get: vi.fn(), post: vi.fn() }))
vi.mock('@/api/admin/payment', () => ({ adminPaymentAPI: api }))
vi.mock('@/api/client', () => ({ apiClient: api }))
const policy = (): CollectionPolicy => ({ enabled: true, single_max: 50, daily_max: 1000, timezone: 'Asia/Shanghai', quick_amounts: [10,20,40], allow_custom_amount: false, contact_text: '联系管理员' })

beforeEach(() => {
  vi.clearAllMocks()
  api.getConfig.mockResolvedValue({ data: { collection: policy() } })
  api.updateConfig.mockResolvedValue({ data: {} })
  api.get.mockResolvedValue({ data: { paid: 900, held: 40, remaining: 60, pending: [] } })
  api.post.mockResolvedValue({ data: {} })
})

describe('共享收款配置', () => {
  it('按实付分值处理50元与49.99元边界', () => {
    expect(collectionAllows(policy(), 50, 'CNY')).toBe(true)
    expect(collectionAllows(policy(), 50.01, 'CNY')).toBe(false)
    expect(collectionAllows({ ...policy(), single_max: 49.99 }, 50, 'CNY')).toBe(false)
    expect(collectionAllows(policy(), 10, 'USD')).toBe(false)
    expect(collectionAllows(policy(), Number.NaN, 'CNY')).toBe(false)
    expect(collectionAllows(undefined, 100, 'CNY')).toBe(true)
  })
  it('读取后台配置、显示占用并独立保存', async () => {
    const wrapper = mount(PaymentCollectionSettings)
    await flushPromises()
    expect(wrapper.text()).toContain('今日已收 ¥900.00')
    await wrapper.find('input[type="number"]').setValue(49.99)
    await wrapper.findAll('button').find(b => b.text() === '保存收款配置')!.trigger('click')
    await flushPromises()
    expect(api.updateConfig).toHaveBeenCalledWith({ collection: { ...policy(), single_max: 49.99 } })
    expect(wrapper.text()).toContain('收款配置已保存')
    wrapper.unmount()
  })
  it('读取失败不允许用默认值覆盖已有配置', async () => {
    api.getConfig.mockRejectedValue(new Error('配置读取失败'))
    const wrapper = mount(PaymentCollectionSettings)
    await flushPromises()
    expect(wrapper.findAll('button').map(b => b.text())).toEqual(['重新加载'])
    expect(api.updateConfig).not.toHaveBeenCalled()
    wrapper.unmount()
  })
  it('关单未确认仍保留错误提示并重新读取占用', async () => {
    api.get.mockResolvedValue({ data: { paid: 0, held: 50, remaining: 950, pending: [{order_id: 9, amount: 50, status: 'EXPIRED'}] } })
    api.post.mockRejectedValue(new Error('未能确认渠道关单'))
    const wrapper = mount(PaymentCollectionSettings)
    await flushPromises()
    await wrapper.findAll('button').find(b => b.text() === '查询并关闭渠道订单')!.trigger('click')
    await flushPromises()
    expect(api.post).toHaveBeenCalledWith('/admin/payment/collection/9/close')
    expect(wrapper.find('[role="alert"]').text()).toContain('未能确认渠道关单')
    expect(api.get).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })
})
