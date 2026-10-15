import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h, ref } from 'vue'
import { createI18n } from 'vue-i18n'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import GroupRateMultiplier from '../GroupRateMultiplier.vue'
import { provideGroupRateDisplay } from '../rateDisplay'
import type { GroupRateSchedules } from '../api'

const mocks = vi.hoisted(() => ({ get: vi.fn() }))
vi.mock('../api', () => ({ getGroupRateSchedules: mocks.get }))

const serverStart = Date.parse('2026-10-15T13:06:10Z')
const schedules: GroupRateSchedules['schedules'] = {
  1: { rate_schedule_enabled: true, rate_schedule: [
    { starts_at: '2026-10-15T13:04:05Z', ends_at: '2026-10-15T13:06:11Z', rate_multiplier: 0.1 },
    { starts_at: '2026-10-15T13:06:19Z', ends_at: '2026-10-15T13:17:27Z', rate_multiplier: 0.8 },
  ] },
}

function render() {
  const enabled = ref(true)
  let reload!: () => Promise<void>
  const wrapper = mount(defineComponent({ setup() {
    reload = provideGroupRateDisplay(enabled).reload
    return () => h('div', [
      h(GroupRateMultiplier, { group: { id: 1, rate_multiplier: 1 } }),
      h(GroupRateMultiplier, { group: { id: 2, rate_multiplier: 0.25 } }),
    ])
  } }), { global: { plugins: [createI18n({ legacy: false, locale: 'zh', messages: { zh: {} } })] } })
  return { wrapper, enabled, reload }
}

beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date', 'performance', 'setTimeout', 'clearTimeout', 'setInterval', 'clearInterval'] })
  vi.setSystemTime(new Date('2027-03-01T00:00:00Z'))
  mocks.get.mockReset()
  mocks.get.mockImplementation(async () => ({ server_time: new Date(serverStart + performance.now()).toISOString(), schedules }))
})

afterEach(() => { vi.useRealTimers() })

describe('管理列表当前分组倍率', () => {
  it('使用服务器时间并自动经历0.1、8秒空档1、0.8及活动结束，不写回基础倍率', async () => {
    const { wrapper, reload } = render()
    await reload()
    await flushPromises()
    const rates = () => wrapper.findAll('[data-testid="current-rate"]').map(node => node.text())
    expect(rates()).toEqual(['0.1x', '0.25x'])
    expect(wrapper.get('[data-testid="rate-detail"]').text()).toContain('基础 1x')
    expect(mocks.get).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(1000)
    expect(rates()[0]).toBe('1x')
    await vi.advanceTimersByTimeAsync(7000)
    expect(rates()[0]).toBe('1x')
    await vi.advanceTimersByTimeAsync(1000)
    expect(rates()[0]).toBe('0.8x')
    expect(mocks.get).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(668_000)
    expect(rates()[0]).toBe('1x')
    expect(wrapper.find('[data-testid="rate-detail"]').exists()).toBe(false)
    wrapper.unmount()
    expect(vi.getTimerCount()).toBe(0)
  })

  it('读取失败时明确未知并保留基础值说明，重试后恢复', async () => {
    const { wrapper, reload } = render()
    expect(wrapper.findAll('[data-testid="current-rate"]')[0]!.text()).toBe('—')
    mocks.get.mockRejectedValueOnce(new Error('unavailable'))
    await reload()
    await flushPromises()
    expect(wrapper.text()).toContain('活动倍率未读取 · 基础 1x')
    await reload()
    await flushPromises()
    expect(wrapper.findAll('[data-testid="current-rate"]')[0]!.text()).toBe('0.1x')
    wrapper.unmount()
  })

  it('关闭活动后恢复基础倍率，0倍率也正确显示', async () => {
    const { wrapper, reload, enabled } = render()
    mocks.get.mockResolvedValueOnce({ server_time: new Date(serverStart).toISOString(), schedules: { 1: {
      rate_schedule_enabled: true, rate_schedule: [{ starts_at: '2026-10-15T13:00:00Z', ends_at: '2026-10-15T14:00:00Z', rate_multiplier: 0 }],
    } } })
    await reload()
    await flushPromises()
    expect(wrapper.findAll('[data-testid="current-rate"]')[0]!.text()).toBe('0x')
    mocks.get.mockResolvedValueOnce({ server_time: new Date(serverStart).toISOString(), schedules: {} })
    await reload()
    await flushPromises()
    expect(wrapper.findAll('[data-testid="current-rate"]')[0]!.text()).toBe('1x')
    const count = mocks.get.mock.calls.length
    enabled.value = false
    await flushPromises()
    await vi.advanceTimersByTimeAsync(30_000)
    expect(mocks.get).toHaveBeenCalledTimes(count)
    wrapper.unmount()
  })

  it('忽略过期的并发响应与卸载后返回的请求', async () => {
    const { wrapper, reload } = render()
    let finish!: (value: GroupRateSchedules) => void
    mocks.get.mockImplementationOnce(() => new Promise<GroupRateSchedules>(resolve => { finish = resolve }))
    const stale = reload()
    mocks.get.mockResolvedValueOnce({ server_time: new Date(serverStart).toISOString(), schedules: {} })
    await reload()
    finish({ server_time: new Date(serverStart).toISOString(), schedules })
    await stale
    await flushPromises()
    expect(wrapper.findAll('[data-testid="current-rate"]')[0]!.text()).toBe('1x')
    mocks.get.mockImplementationOnce(() => new Promise<GroupRateSchedules>(resolve => { finish = resolve }))
    const pending = reload()
    wrapper.unmount()
    finish({ server_time: new Date(serverStart).toISOString(), schedules })
    await pending
    expect(vi.getTimerCount()).toBe(0)
  })
})
