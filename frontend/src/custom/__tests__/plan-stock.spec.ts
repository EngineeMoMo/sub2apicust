import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { defineComponent } from 'vue'
import SubscriptionPlanCard from '@/components/payment/SubscriptionPlanCard.vue'
import PlanEditDialog from '@/views/admin/orders/PlanEditDialog.vue'
import PublicPlansView from '@/custom/views/PublicPlansView.vue'
import SubscriptionCard from '@/custom/components/SubscriptionCard.vue'
import SubscriptionStockPicker from '@/custom/components/SubscriptionStockPicker.vue'
import type { UserSubscription } from '@/types'
import type { SubscriptionPlan } from '@/types/payment'

const api = vi.hoisted(() => ({ createPlan: vi.fn(), updatePlan: vi.fn(), getPlans: vi.fn(), push: vi.fn(), fetchPublicPlans: vi.fn(), showError: vi.fn() }))
vi.mock('vue-router', async importOriginal => ({ ...await importOriginal<typeof import('vue-router')>(), useRouter: () => ({ push: api.push }) }))
vi.mock('@/api/admin/payment', () => ({ adminPaymentAPI: api }))
vi.mock('@/custom/guest/api', () => ({ fetchPublicPlans: api.fetchPublicPlans }))
vi.mock('@/custom/components/PublicLayout.vue', () => ({ default: { template: '<div><slot /></div>' } }))
vi.mock('@/custom/components/GuestAction.vue', () => ({ default: { template: '<button data-guest><slot /></button>' } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError: api.showError, showSuccess: vi.fn() }) }))
vi.mock('vue-i18n', async importOriginal => ({ ...await importOriginal<typeof import('vue-i18n')>(), useI18n: () => ({ locale: { value: 'zh' }, t: (key: string, args?: object) => key + (args ? JSON.stringify(args) : '') }) }))
const plan = (remaining?: number): SubscriptionPlan => ({ id: 1, group_id: 1, name: '月套餐', description: '', price: 60, features: [], validity_days: 30, validity_unit: 'days', for_sale: true, sort_order: 0, stock_limit: 10, stock_used: 10 - (remaining ?? 0), stock_remaining: remaining })
const dialog = defineComponent({ template: '<div><slot /><slot name="footer" /></div>' })
beforeEach(() => { vi.clearAllMocks(); api.updatePlan.mockResolvedValue({}); api.createPlan.mockResolvedValue({}) })

describe('套餐库存交互', () => {
  it.each([0, undefined])('我的订阅库存 %s 不可续订', async remaining => {
    const wrapper = mount(SubscriptionCard, { props: { subscription: { id: 1, group_id: 1, status: 'active', expires_at: '2099-01-01' } as UserSubscription, now: Date.now(), plans: [plan(remaining)] } })
    expect(wrapper.get('[data-test="subscription-stock"]').text()).toContain(remaining === 0 ? 'payment.stock.soldOut' : 'payment.stock.loadFailed')
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
    await wrapper.get('button').trigger('click')
    expect(api.push).not.toHaveBeenCalled()
  })
  it.each([-1, 3])('我的订阅有库存 %s 可以续订', async remaining => {
    const wrapper = mount(SubscriptionCard, { props: { subscription: { id: 1, group_id: 1, status: 'active', expires_at: '2099-01-01' } as UserSubscription, now: Date.now(), plans: [plan(0), { ...plan(remaining), id: 2 }] } })
    await wrapper.get('button').trigger('click')
    expect(api.push).toHaveBeenCalledWith({ path: '/purchase', query: { tab: 'subscription', group: '1' } })
  })
  it.each(['stockLoading', 'stockError'] as const)('我的订阅 %s 不伪装成可续订', key => {
    const wrapper = mount(SubscriptionCard, { props: { subscription: { id: 1, group_id: 1, status: 'active', expires_at: '2099-01-01' } as UserSubscription, now: Date.now(), plans: [plan(3)], [key]: true } })
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
  })
  it('管理员选择套餐显示库存，售罄不可选，切换分组清除旧选择', async () => {
    api.getPlans.mockResolvedValue({ data: [plan(0), { ...plan(2), id: 2 }] })
    const wrapper = mount(SubscriptionStockPicker, { props: { groupId: 1, modelValue: null } })
    await flushPromises()
    expect(wrapper.get('option[value="1"]').attributes('disabled')).toBeDefined()
    await wrapper.get('select').setValue('2')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([2])
    await wrapper.setProps({ modelValue: 2 })
    expect(wrapper.emitted('available')?.at(-1)).toEqual([true])
    await wrapper.setProps({ groupId: 99 })
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([null])
    expect(wrapper.emitted('available')?.at(-1)).toEqual([false])
  })
  it('管理员库存读取失败禁止分配并可重试', async () => {
    api.getPlans.mockRejectedValueOnce(new Error('offline')).mockResolvedValue({ data: [plan(2)] })
    const wrapper = mount(SubscriptionStockPicker, { props: { groupId: 1, modelValue: 1 } })
    await flushPromises()
    expect(wrapper.get('select').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[role="alert"]').text()).toContain('payment.stock.loadFailed')
    await wrapper.get('button').trigger('click'); await flushPromises()
    expect(wrapper.emitted('available')?.at(-1)).toEqual([true])
  })
  it.each([undefined, -1, 1])('库存 %s 仍可购买', async remaining => {
    const wrapper = mount(SubscriptionPlanCard, { props: { plan: plan(remaining) }, global: { plugins: [createPinia()] } })
    expect(wrapper.get('button').attributes('disabled')).toBeUndefined()
    await wrapper.get('button').trigger('click')
    expect(wrapper.emitted('select')).toHaveLength(1)
  })
  it('售罄显示原因并禁止选择和续费', async () => {
    const wrapper = mount(SubscriptionPlanCard, { props: { plan: plan(0) }, global: { plugins: [createPinia()] } })
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[role="status"]').text()).toContain('payment.stock.soldOut')
    await wrapper.get('button').trigger('click')
    expect(wrapper.emitted('select')).toBeUndefined()
  })
  it('编辑限额包含零且不向后端覆盖已占用计数', async () => {
    const wrapper = mount(PlanEditDialog, { props: { show: false, plan: plan(4), groups: [] }, global: { stubs: { BaseDialog: dialog, Select: true, Icon: true, GroupBadge: true } } })
    await wrapper.setProps({ show: true })
    expect((wrapper.get('#plan-stock-limit').element as HTMLInputElement).value).toBe('10')
    await wrapper.get('#plan-stock-limit').setValue('0')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(api.updatePlan).toHaveBeenCalledWith(1, expect.objectContaining({ stock_limit: 0 }))
    expect(api.updatePlan.mock.calls[0][1]).not.toHaveProperty('stock_used')
    expect(wrapper.emitted('saved')).toHaveLength(1)
  })
  it.each(['-2', '1.5', '', '2147483648'])('拒绝非法库存 %s', async value => {
    const wrapper = mount(PlanEditDialog, { props: { show: false, plan: plan(4), groups: [] }, global: { stubs: { BaseDialog: dialog, Select: true, Icon: true, GroupBadge: true } } })
    await wrapper.setProps({ show: true })
    await wrapper.get('#plan-stock-limit').setValue(value)
    await wrapper.get('form').trigger('submit')
    expect(api.updatePlan).not.toHaveBeenCalled()
    expect(api.showError).toHaveBeenCalledWith('payment.stock.invalid')
  })
  it('再次创建恢复不限量默认值', async () => {
    const wrapper = mount(PlanEditDialog, { props: { show: false, plan: plan(0), groups: [] }, global: { stubs: { BaseDialog: dialog, Select: true, Icon: true, GroupBadge: true } } })
    await wrapper.setProps({ show: true }); await wrapper.setProps({ show: false, plan: null }); await wrapper.setProps({ show: true })
    expect((wrapper.get('#plan-stock-limit').element as HTMLInputElement).value).toBe('-1')
  })
  it('游客售罄页不引导登录购买', async () => {
    api.fetchPublicPlans.mockResolvedValue({ purchase_enabled: true, plans: [{ ...plan(0), supported_model_scopes: [] }] })
    const wrapper = mount(PublicPlansView, { props: { embedded: true }, global: { stubs: { RouterLink: true, GuestAction: { template: '<button data-guest><slot /></button>' } } } })
    await flushPromises()
    expect(wrapper.text()).toContain('已售罄')
    expect(wrapper.find('[data-guest]').exists()).toBe(false)
    expect(wrapper.get('article button').attributes('disabled')).toBeDefined()
  })
})
