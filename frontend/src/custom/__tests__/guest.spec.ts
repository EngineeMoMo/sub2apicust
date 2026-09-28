import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { safeGuestRedirect, selectedPublicPlanID } from '@/custom/guest/navigation'
import { faqItems, searchFaq } from '@/custom/guest/faq'
import PublicFaqView from '@/custom/views/PublicFaqView.vue'
import PublicPlansView from '@/custom/views/PublicPlansView.vue'
import GuestAction from '@/custom/components/GuestAction.vue'

const { fetchPlans, push, auth } = vi.hoisted(() => ({ fetchPlans: vi.fn(), push: vi.fn(), auth: { isAuthenticated: false } }))
vi.mock('@/custom/guest/api', () => ({ fetchPublicPlans: fetchPlans }))
vi.mock('@/stores', () => ({ useAuthStore: () => auth }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key === 'payment.perMonth' ? '月' : key }) }))

const global = {
  stubs: {
    PublicLayout: { template: '<main><slot /></main>' },
    RouterLink: RouterLinkStub,
    BaseDialog: { props: ['show'], template: '<div v-if="show" role="dialog"><slot /><slot name="footer" /></div>' }
  }
}
const plan = { id: 7, name: '测试套餐', description: '说明', price: 20, currency: 'USD', validity_days: 1, validity_unit: 'months', features: ['测试权益'], supported_model_scopes: [] }

beforeEach(() => { fetchPlans.mockReset(); push.mockReset(); auth.isAuthenticated = false })

describe('游客导航', () => {
  it.each(['https://evil.test', '//evil.test', '/\\evil.test', '/%2fexample.com', '/%5cexample.com', '/%0aevil', '/x/..//evil.test', '/login', '/register', undefined, ['/', '/keys']])('拒绝不安全回跳 %s', value => {
    expect(safeGuestRedirect(value)).toBe('/dashboard')
  })
  it('保留站内套餐目标', () => expect(safeGuestRedirect('/purchase?tab=subscription&plan=7')).toBe('/purchase?tab=subscription&plan=7'))
  it.each(['0', '-1', '1.5', '1e3', '9007199254740992', ['7'], undefined])('拒绝无效套餐编号 %s', value => expect(selectedPublicPlanID(value)).toBeNull())
  it('解析有效套餐编号', () => expect(selectedPublicPlanID('7')).toBe(7))
  it('游客先提示，可取消，不发起导航', async () => {
    const wrapper = mount(GuestAction, { props: { to: '/purchase?tab=subscription&plan=7' }, global })
    await wrapper.get('button').trigger('click')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
    expect(wrapper.getComponent(RouterLinkStub).props('to')).toEqual({ path: '/login', query: { redirect: '/purchase?tab=subscription&plan=7' } })
    await wrapper.get('[role="dialog"] button').trigger('click')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    expect(push).not.toHaveBeenCalled()
  })
  it('已登录进入目标，不自动下单', async () => {
    auth.isAuthenticated = true
    const wrapper = mount(GuestAction, { props: { to: '/purchase?tab=subscription&plan=7' }, global })
    await wrapper.get('button').trigger('click')
    expect(push).toHaveBeenCalledWith('/purchase?tab=subscription&plan=7')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
  })
})

describe('公开套餐', () => {
  it('显示真实价格、周期与权益，不把未知额度标成无限', async () => {
    fetchPlans.mockResolvedValue({ plans: [plan], purchase_enabled: true })
    const wrapper = mount(PublicPlansView, { global })
    expect(wrapper.text()).toContain('正在读取')
    await flushPromises()
    expect(wrapper.text()).toContain('测试套餐')
    expect(wrapper.text()).toContain('USD')
    expect(wrapper.text()).toContain('测试权益')
    expect(wrapper.text()).toContain('有效期：月')
    expect(wrapper.text()).toContain('未提供额度说明')
    expect(wrapper.text()).not.toContain('无限')
    expect(wrapper.getComponent(GuestAction).props('to')).toBe('/purchase?tab=subscription&plan=7')
  })
  it('支付关闭仍展示套餐且不允许购买', async () => {
    fetchPlans.mockResolvedValue({ plans: [plan], purchase_enabled: false })
    const wrapper = mount(PublicPlansView, { global })
    await flushPromises()
    expect(wrapper.text()).toContain('测试套餐')
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
    expect(wrapper.findComponent(GuestAction).exists()).toBe(false)
  })
  it('空目录与失败重试，不跳转登录', async () => {
    fetchPlans.mockRejectedValueOnce(new Error('offline')).mockResolvedValueOnce({ plans: [], purchase_enabled: true })
    const wrapper = mount(PublicPlansView, { global })
    await flushPromises()
    expect(wrapper.text()).toContain('暂时无法读取套餐')
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('暂无在售套餐')
    expect(push).not.toHaveBeenCalled()
  })
})

describe('常见问题', () => {
  it('提供四类十二题并可搜索', () => {
    expect(faqItems).toHaveLength(12)
    expect(new Set(faqItems.map(item => item.id)).size).toBe(12)
    expect(new Set(faqItems.map(item => item.group)).size).toBe(4)
    expect(searchFaq(' api key ').length).toBeGreaterThan(0)
  })
  it('使用原生可键盘展开问答，搜索无结果可恢复', async () => {
    const wrapper = mount(PublicFaqView, { global })
    expect(wrapper.findAll('details summary')).toHaveLength(12)
    await wrapper.get('input').setValue('api key')
    expect(wrapper.findAll('details').length).toBeLessThan(12)
    expect(wrapper.get('details').attributes('open')).toBeDefined()
    await wrapper.get('input').setValue('不可能匹配的关键词')
    expect(wrapper.text()).toContain('没有找到相关问题')
    await wrapper.get('button').trigger('click')
    expect(wrapper.findAll('details summary')).toHaveLength(12)
  })
})
