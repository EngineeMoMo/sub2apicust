import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { AdminGroup } from '@/types'
import type { GroupAccessPolicy } from '../api'
import GroupAccessPolicyButton from '../GroupAccessPolicyButton.vue'

const mocks = vi.hoisted(() => ({ get: vi.fn(), save: vi.fn(), groups: vi.fn(), success: vi.fn() }))
vi.mock('../api', async importOriginal => ({ ...await importOriginal<typeof import('../api')>(), getGroupAccessPolicy: mocks.get, saveGroupAccessPolicy: mocks.save }))
vi.mock('@/api/admin', () => ({ adminAPI: { groups: { getAll: mocks.groups } } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showSuccess: mocks.success }) }))

const group = { id: 1, name: '定制组', platform: 'anthropic', status: 'active' } as AdminGroup
const original: GroupAccessPolicy = { group_id: 1, monthly_recharge_enabled: false, monthly_recharge_min: 100, balance_enabled: false, balance_min: 20, expires_at: null, fallback_group_id: null, updated_at: null }

function render(currentGroup = group) {
  return mount(GroupAccessPolicyButton, {
    props: { group: currentGroup },
    global: {
      plugins: [createI18n({ legacy: false, locale: 'zh', messages: { zh: {} } })],
      stubs: { BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' }, Icon: true },
    },
  })
}

beforeEach(() => {
  vi.resetAllMocks()
  mocks.get.mockResolvedValue({ ...original })
  mocks.groups.mockResolvedValue([group, { ...group, id: 2, name: '常规组' }, { ...group, id: 3, platform: 'openai' }, { ...group, id: 4, subscription_type: 'subscription' }])
  mocks.save.mockResolvedValue({ ...original })
})

describe('分组开放规则配置', () => {
  it('预填三段福利倍率并保存23点到期兜底', async () => {
    mocks.get.mockResolvedValue({ ...original, rate_schedule_enabled: true, rate_schedule: [{ starts_at: '2026-11-01T00:00:00Z', ends_at: '2026-11-01T01:00:00Z', rate_multiplier: 0.1 }] })
    const wrapper = render()
    await wrapper.get('[data-testid="group-access-policy"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-testid="seed-periods"]').trigger('click')
    expect(wrapper.findAll('[data-testid="rate-period"]')).toHaveLength(3)
    await wrapper.get('[data-testid="fallback-group"]').setValue('2')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledWith(1, expect.objectContaining({
      rate_schedule_enabled: true, expires_at: '2026-11-01T15:00:00.000Z', fallback_group_id: 2,
      rate_schedule: [
        { starts_at: '2026-10-31T16:00:00.000Z', ends_at: '2026-11-01T10:00:00.000Z', rate_multiplier: 0.1 },
        { starts_at: '2026-11-01T10:00:00.000Z', ends_at: '2026-11-01T12:30:00.000Z', rate_multiplier: 0.15 },
        { starts_at: '2026-11-01T12:30:00.000Z', ends_at: '2026-11-01T15:00:00.000Z', rate_multiplier: 0.1 },
      ],
    }))
    expect(wrapper.emitted('saved')).toHaveLength(1)
    wrapper.unmount()
  })

  it('活动时段可以独立关闭并保留已配置的时段', async () => {
    mocks.get.mockResolvedValue({ ...original, rate_schedule_enabled: true, rate_schedule: [{ starts_at: '2026-11-01T00:00:00Z', ends_at: '2026-11-01T01:00:00Z', rate_multiplier: 0.1 }] })
    const wrapper = render()
    await wrapper.get('[data-testid="group-access-policy"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-testid="schedule-toggle"]').trigger('click')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledWith(1, expect.objectContaining({ rate_schedule_enabled: false, rate_schedule: [expect.objectContaining({ rate_multiplier: 0.1 })] }))
    wrapper.unmount()
  })

  it('两个门槛独立开关，保存时保留各自金额', async () => {
    const wrapper = render()
    await wrapper.get('[data-testid="group-access-policy"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-testid="monthly-toggle"]').trigger('click')
    await wrapper.get('[data-testid="balance-toggle"]').trigger('click')
    await wrapper.get('[data-testid="monthly-min"]').setValue('120')
    await wrapper.get('[data-testid="balance-min"]').setValue('30')
    await wrapper.get('[data-testid="monthly-toggle"]').trigger('click')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledWith(1, expect.objectContaining({ monthly_recharge_enabled: false, monthly_recharge_min: 120, balance_enabled: true, balance_min: 30 }))
    expect(mocks.success).toHaveBeenCalledOnce()
    wrapper.unmount()
  })

  it('到期时间与指定同平台分组一起保存', async () => {
    const wrapper = render()
    await wrapper.get('[data-testid="group-access-policy"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-testid="expiry-toggle"]').trigger('click')
    await wrapper.get('[data-testid="expires-at"]').setValue('2026-11-01T00:00')
    expect(wrapper.find('[data-testid="fallback-group"] option[value="3"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="fallback-group"] option[value="1"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="fallback-group"] option[value="4"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('到期删除并平移用户')
    expect(wrapper.text()).toContain('分时福利的时段结束只恢复基础倍率')
    await wrapper.get('[data-testid="fallback-group"]').setValue('2')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledWith(1, expect.objectContaining({ expires_at: '2026-10-31T16:00:00.000Z', fallback_group_id: 2 }))
    wrapper.unmount()
  })

  it('关闭到期设置后清除旧路由，保留金额门槛', async () => {
    mocks.get.mockResolvedValue({ ...original, balance_enabled: true, expires_at: '2026-10-31T16:00:00Z', fallback_group_id: 2 })
    const wrapper = render()
    await wrapper.get('[data-testid="group-access-policy"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-testid="expiry-toggle"]').trigger('click')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledWith(1, expect.objectContaining({ balance_enabled: true, expires_at: null, fallback_group_id: null }))
    wrapper.unmount()
  })

  it('订阅分组禁止开启到期删除，仍可配置福利倍率', async () => {
    const wrapper = render({ ...group, subscription_type: 'subscription' })
    await wrapper.get('[data-testid="group-access-policy"]').trigger('click')
    await flushPromises()
    expect(wrapper.get('[data-testid="expiry-toggle"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-testid="schedule-toggle"]').attributes('disabled')).toBeUndefined()
    await wrapper.get('[data-testid="schedule-toggle"]').trigger('click')
    await wrapper.get('[data-testid="seed-periods"]').trigger('click')
    expect(wrapper.find('[data-testid="expires-at"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('订阅分组可以关闭旧到期设置', async () => {
    mocks.get.mockResolvedValue({ ...original, expires_at: '2026-10-31T16:00:00Z', fallback_group_id: 2 })
    const wrapper = render({ ...group, subscription_type: 'subscription' })
    await wrapper.get('[data-testid="group-access-policy"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-testid="expiry-toggle"]').trigger('click')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(mocks.save).toHaveBeenCalledWith(1, expect.objectContaining({ expires_at: null, fallback_group_id: null }))
    wrapper.unmount()
  })

  it('读取失败时不能覆盖未读取的规则', async () => {
    mocks.get.mockRejectedValue(new Error('unavailable'))
    const wrapper = render()
    await wrapper.get('[data-testid="group-access-policy"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('form').exists()).toBe(false)
    expect(wrapper.get('[role="alert"]').text()).toContain('无法读取')
    expect(mocks.save).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('重复提交期间锁定保存并展示服务端冲突', async () => {
    let rejectSave: (reason: unknown) => void = () => {}
    mocks.save.mockReturnValue(new Promise((_, reject) => { rejectSave = reject }))
    const wrapper = render()
    await wrapper.get('[data-testid="group-access-policy"]').trigger('click')
    await flushPromises()
    await wrapper.get('form').trigger('submit')
    await wrapper.get('form').trigger('submit')
    expect(mocks.save).toHaveBeenCalledOnce()
    rejectSave({ response: { data: { message: '规则已被修改，请刷新' } } })
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('规则已被修改')
    expect(wrapper.find('form').exists()).toBe(true)
    wrapper.unmount()
  })
})
