import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import type { UserSubscription } from '@/types'
import SubscriptionCard from '@/custom/components/SubscriptionCard.vue'
import SubscriptionsView from '@/views/user/SubscriptionsView.vue'
import { effectiveSubscriptionStatus, fetchSubscriptionResetTimes, remainingTimeLabel,
  subscriptionResetHint, type SubscriptionResetTimes } from '@/custom/subscriptions/timing'

const { getList, getProgress, push, showError } = vi.hoisted(() => ({ getList: vi.fn(), getProgress: vi.fn(), push: vi.fn(), showError: vi.fn() }))
const locale = ref('zh')
vi.mock('@/api/subscriptions', () => ({ default: { getMySubscriptions: getList } }))
vi.mock('@/api/client', () => ({ apiClient: { get: getProgress } }))
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<main><slot /></main>' } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError, cachedPublicSettings: { server_utc_offset: '+08:00' } }) }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))
vi.mock('@/utils/format', () => ({ formatDateTimeToMinute: (at: string) => new Date(at).toISOString().slice(0, 16) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ locale, t: (key: string, values: Record<string, string> = {}) => {
  const text: Record<string, string> = {
    'userSubscriptions.status.active': '有效', 'userSubscriptions.status.expired': '已过期',
    'userSubscriptions.status.revoked': '已撤销', 'userSubscriptions.daily': '每日',
    'userSubscriptions.weekly': '每周', 'userSubscriptions.monthly': '每月',
    'userSubscriptions.expires': '到期时间', 'userSubscriptions.windowNotActive': '等待首次使用',
    'userSubscriptions.unlimited': '无限制', 'userSubscriptions.unlimitedDesc': '该订阅无用量限制',
    'userSubscriptions.failedToLoad': '加载订阅失败', 'payment.renewNow': '续费',
    'userSubscriptions.resetIn': locale.value === 'zh' ? '{time} 后重置' : 'Resets in {time}',
    'userSubscriptions.quotaEndsIn': locale.value === 'zh' ? '额度将在 {time} 后结束' : 'Quota ends in {time}',
    'userSubscriptions.usageOf': '已用 {used} / {limit}'
  }
  return Object.entries(values).reduce((result, [name, value]) => result.replace(`{${name}}`, value), text[key] ?? key)
} }) }))

const now = Date.parse('2026-10-06T01:35:00Z')
const at = (hours: number) => new Date(now + hours * 3600_000).toISOString()
function subscription(overrides: Partial<UserSubscription> = {}): UserSubscription {
  return { id: 7, user_id: 3, group_id: 11, status: 'active', starts_at: at(-24), expires_at: at(24 * 40),
    daily_usage_usd: 4, weekly_usage_usd: 26.87, monthly_usage_usd: 3272.36,
    daily_window_start: at(-12), weekly_window_start: at(-12), monthly_window_start: at(-12),
    created_at: at(-24), updated_at: at(-24), group: { id: 11, name: '960元月订阅-1000$/周-4000$/月',
      platform: 'openai', rate_multiplier: 1, daily_limit_usd: 50, weekly_limit_usd: 1000, monthly_limit_usd: 4000 } as UserSubscription['group'], ...overrides }
}
const resets: SubscriptionResetTimes = { daily: at(14.5), weekly: at(156), monthly: at(36) }
const wrappers: Array<ReturnType<typeof mount>> = []
const card = (sub = subscription(), timings: SubscriptionResetTimes | undefined = resets) => {
  const wrapper = mount(SubscriptionCard, { props: { subscription: sub, now, resets: timings, utcOffset: '+08:00' } })
  wrappers.push(wrapper)
  return wrapper
}
const view = () => {
  const wrapper = mount(SubscriptionsView, { global: { stubs: { AppLayout: { template: '<main><slot /></main>' } } } })
  wrappers.push(wrapper)
  return wrapper
}
beforeEach(() => {
  vi.clearAllMocks()
  locale.value = 'zh'
  vi.useFakeTimers({ toFake: ['Date', 'setInterval', 'clearInterval'] })
  vi.setSystemTime(now)
  getList.mockResolvedValue([subscription()])
  getProgress.mockResolvedValue({ data: [{ subscription: { id: 7 }, progress: {
    daily: { resets_at: resets.daily }, weekly: { resets_at: resets.weekly }, monthly: { resets_at: resets.monthly }
  } }] })
})
afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  vi.useRealTimers()
})

describe('订阅时间提示', () => {
  it('读取真实接口的嵌套结构和服务端日／周／月重置时间', async () => {
    expect(await fetchSubscriptionResetTimes()).toEqual({ 7: resets })
    expect(getProgress).toHaveBeenCalledWith('/subscriptions/progress')
  })
  it('使用服务端日窗口时间，而不在首次使用时刻加24小时', () => {
    expect(subscriptionResetHint(subscription(), 'daily', resets, now)).toEqual({ kind: 'countdown', at: at(14.5) })
  })
  it('重置时刻已过时不伪装成首次使用或已经发放额度', () => {
    expect(subscriptionResetHint(subscription(), 'weekly', { weekly: at(-1) }, now).kind).toBe('pending')
    expect(card(subscription(), { weekly: at(-1) }).get('[data-period="weekly"]').text()).toContain('已到重置时间，使用后更新')
  })
  it('未激活的窗口显示首次使用提示', () => {
    expect(subscriptionResetHint(subscription({ weekly_window_start: null }), 'weekly', resets, now).kind).toBe('unstarted')
    expect(card(subscription({ weekly_window_start: null })).get('[data-period="weekly"]').text()).toContain('等待首次使用')
  })
  it('缺失或无效重置时间明确显示未读取', () => {
    expect(subscriptionResetHint(subscription(), 'weekly', undefined, now).kind).toBe('unavailable')
    expect(subscriptionResetHint(subscription(), 'weekly', { weekly: 'invalid' }, now).kind).toBe('unavailable')
  })
  it('最后的不完整周期在到期结束，不承诺到期后重置', () => {
    const wrapper = card(subscription({ expires_at: at(72) }))
    expect(wrapper.get('[data-period="weekly"]').text()).toContain('额度将在 3 天 0 小时 后结束')
    expect(wrapper.get('[data-period="monthly"]').text()).toContain('1 天 12 小时 后重置')
  })
  it('日卡以订阅到期结束而不是再次发放每日额度', () => {
    const sub = subscription({ starts_at: at(-12), expires_at: at(12) })
    expect(subscriptionResetHint(sub, 'daily', { daily: at(12) }, now).kind).toBe('ends')
  })
  it.each(['expired', 'suspended', 'revoked'] as const)('%s 不显示未来额度重置倒计时', status => {
    const wrapper = card(subscription({ status }))
    expect(wrapper.text()).not.toContain('后重置')
    expect(wrapper.findAll('.mofa-subscription-reset time')).toHaveLength(0)
  })
  it('服务端状态仍是active但到期时间已过，页面显示已过期并保留日期', () => {
    const sub = subscription({ expires_at: at(-1) })
    expect(effectiveSubscriptionStatus(sub, now)).toBe('expired')
    const wrapper = card(sub)
    expect(wrapper.attributes('data-state')).toBe('expired')
    expect(wrapper.get('.mofa-subscription-expiry time').attributes('datetime')).toBe(at(-1))
    expect(wrapper.find('.mofa-subscription-remaining').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('后重置')
  })
  it('活跃卡片同时显示倒计时、精确日期和可访问用量条', () => {
    const wrapper = card()
    expect(wrapper.get('[data-period="weekly"]').text()).toContain('6 天 12 小时 后重置')
    expect(wrapper.get('[data-period="weekly"] time').attributes('datetime')).toBe(resets.weekly)
    expect(wrapper.get('[data-period="monthly"] progress').attributes('aria-valuetext')).toBe('已用 $3272.36 / $4000.00')
  })
  it('有效和过期订阅续费都保留原分组及购买路由', async () => {
    for (const status of ['active', 'expired'] as const) {
      await card(subscription({ status })).get('button').trigger('click')
    }
    expect(push).toHaveBeenCalledTimes(2)
    expect(push).toHaveBeenLastCalledWith({ path: '/purchase', query: { tab: 'subscription', group: '11' } })
  })
  it('暂停和撤销订阅不提供续费入口', () => {
    expect(card(subscription({ status: 'suspended' })).find('button').exists()).toBe(false)
    expect(card(subscription({ status: 'revoked' })).find('button').exists()).toBe(false)
  })
  it('页面停留自动更新分钟和到期状态，卸载后清理计时器', async () => {
    getList.mockResolvedValue([subscription({ expires_at: at(1 / 60) })])
    const wrapper = view()
    await flushPromises()
    expect(wrapper.get('article').attributes('data-state')).toBe('active')
    await vi.advanceTimersByTimeAsync(60_000)
    expect(wrapper.get('article').attributes('data-state')).toBe('expired')
    wrapper.unmount()
    wrappers.splice(wrappers.indexOf(wrapper), 1)
    expect(vi.getTimerCount()).toBe(0)
    expect(getProgress).toHaveBeenCalledTimes(1)
  })
  it('返回后台标签页时立即更新显示', async () => {
    const wrapper = view()
    await flushPromises()
    vi.setSystemTime(now + 3600_000)
    document.dispatchEvent(new Event('visibilitychange'))
    await flushPromises()
    expect(wrapper.get('[data-period="monthly"]').text()).toContain('1 天 11 小时 后重置')
  })
  it('时间接口失败保留订阅／用量，点击重试后恢复', async () => {
    getProgress.mockRejectedValueOnce(new Error('offline'))
    const wrapper = view()
    await flushPromises()
    expect(wrapper.find('article').exists()).toBe(true)
    expect(wrapper.text()).toContain('重置时间暂未读取')
    await wrapper.get('.mofa-subscription-notice button').trigger('click')
    await flushPromises()
    expect(wrapper.find('.mofa-subscription-notice').exists()).toBe(false)
    expect(wrapper.text()).toContain('6 天 12 小时 后重置')
  })
  it('时间接口慢响应不阻塞列表，返回后补齐倒计时', async () => {
    let resolve!: (value: unknown) => void
    getProgress.mockImplementationOnce(() => new Promise(done => { resolve = done }))
    const wrapper = view()
    await flushPromises()
    expect(wrapper.find('article').exists()).toBe(true)
    expect(wrapper.get('[data-period="weekly"]').text()).toContain('$26.87')
    expect(wrapper.find('.mofa-subscription-notice').exists()).toBe(false)
    resolve({ data: [{ subscription: { id: 7 }, progress: { weekly: { resets_at: resets.weekly } } }] })
    await flushPromises()
    expect(wrapper.get('[data-period="weekly"]').text()).toContain('6 天 12 小时 后重置')
  })
  it('时间接口延迟失败只提示重试，保持已展示列表', async () => {
    let reject!: (reason: Error) => void
    getProgress.mockImplementationOnce(() => new Promise((_, fail) => { reject = fail }))
    const wrapper = view()
    await flushPromises()
    expect(wrapper.find('article').exists()).toBe(true)
    reject(new Error('timeout'))
    await flushPromises()
    expect(wrapper.find('article').exists()).toBe(true)
    expect(wrapper.find('.mofa-subscription-notice').exists()).toBe(true)
    expect(showError).not.toHaveBeenCalled()
  })
  it.each(['resolve', 'reject'])('列表失败可立即重试，旧时间请求 %s 不覆盖新结果', async outcome => {
    let resolve!: (value: unknown) => void
    let reject!: (reason: Error) => void
    getProgress.mockImplementationOnce(() => new Promise((done, fail) => { resolve = done; reject = fail }))
    getList.mockRejectedValueOnce(new Error('offline'))
    const wrapper = view()
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('加载订阅失败')
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('6 天 12 小时 后重置')
    if (outcome === 'resolve') resolve({ data: [] })
    else reject(new Error('old timeout'))
    await flushPromises()
    expect(wrapper.text()).toContain('6 天 12 小时 后重置')
    expect(wrapper.find('.mofa-subscription-notice').exists()).toBe(false)
  })
  it('列表加载失败不误报无订阅，允许重试', async () => {
    getList.mockRejectedValueOnce(new Error('offline'))
    const wrapper = view()
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('加载订阅失败')
    expect(wrapper.find('article').exists()).toBe(false)
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.find('article').exists()).toBe(true)
  })
  it('英语显示本地化时间单位', () => {
    locale.value = 'en'
    expect(card().get('[data-period="weekly"]').text()).toContain('Resets in 6d 12h')
  })
  it('无效日期不产生NaN或负数倒计时', () => {
    expect(remainingTimeLabel('invalid', now, true)).toBe('')
    expect(card(subscription({ expires_at: 'invalid' })).text()).toContain('时间暂不可用')
  })
  it('不到一分钟时给出准确提示', () => {
    expect(remainingTimeLabel(at(1 / 120), now, true)).toBe('不到 1 分钟')
    expect(remainingTimeLabel(at(1 / 120), now, false)).toBe('less than 1 minute')
  })
})
