import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import UsageToolCosts from '../UsageToolCosts.vue'
import zh from '@/i18n/locales/zh'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => zh.usage[key.split('.').at(-1) as keyof typeof zh.usage] }) }))

const render = (tools?: Record<string, { calls: number; unit_price: number; total_cost: number; actual_cost: number }>) => mount(UsageToolCosts, {
 props: { tools },
})

describe('Claude 工具收费快照', () => {
 it('显示实际次数、基础单价与倍率后的附加费', () => {
  const wrapper = render({ web_search: { calls: 2, unit_price: 0.01, total_cost: 0.02, actual_cost: 0.003 } })
  expect(wrapper.text()).toContain('网页搜索 × 2')
  expect(wrapper.text()).toContain('$0.01000000')
  expect(wrapper.text()).toContain('$0.02000000')
  expect(wrapper.text()).toContain('$0.00300000')
  expect(wrapper.text()).toContain('已包含在本次总费用')
 })
 it('0 单价仍保留次数', () => {
  const wrapper = render({ web_search: { calls: 1, unit_price: 0, total_cost: 0, actual_cost: 0 } })
  expect(wrapper.text()).toContain('网页搜索 × 1')
  expect(wrapper.text()).toContain('$0.00000000')
 })
 it('旧日志和无工具日志不显示收费', () => {
  expect(render().find('[data-testid="tool-costs"]').exists()).toBe(false)
  expect(render({}).find('[data-testid="tool-costs"]').exists()).toBe(false)
 })
})
