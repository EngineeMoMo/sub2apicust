import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import RedeemView from '@/views/admin/RedeemView.vue'

const { listRedeemCodes, batchUpdateRedeemCodes, getAllGroups, showSuccess, showError, showInfo, generateRedeemCodes, getPlans } =
  vi.hoisted(() => ({
    listRedeemCodes: vi.fn(),
    generateRedeemCodes: vi.fn(),
    getPlans: vi.fn(),
    batchUpdateRedeemCodes: vi.fn(),
    getAllGroups: vi.fn(),
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showInfo: vi.fn()
  }))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    redeem: {
      list: listRedeemCodes,
      generate: generateRedeemCodes,
      delete: vi.fn(),
      batchDelete: vi.fn(),
      batchUpdate: batchUpdateRedeemCodes,
      exportCodes: vi.fn()
    },
    groups: {
      getAll: getAllGroups
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess,
    showError,
    showInfo
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const DataTableStub = {
  props: ['columns', 'data'],
  template: `
    <table>
      <thead>
        <tr>
          <th v-for="column in columns" :key="column.key">
            <slot :name="'header-' + column.key" :column="column">{{ column.label }}</slot>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in data" :key="row.id">
          <td v-for="column in columns" :key="column.key">
            <slot :name="'cell-' + column.key" :row="row" :value="row[column.key]">
              {{ row[column.key] }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
  `
}

const SelectStub = {
  props: ['modelValue', 'options'],
  emits: ['update:modelValue', 'change'],
  setup(props: { options: Array<{ value: unknown; label: string }> }, { emit }: { emit: (event: string, ...args: unknown[]) => void }) {
    const onChange = (event: Event) => {
      const raw = (event.target as HTMLSelectElement).value
      const option = props.options.find((item) => String(item.value ?? '') === raw)
      const value = option ? option.value : raw
      emit('update:modelValue', value)
      emit('change', value, option ?? null)
    }
    return { onChange }
  },
  template: `
    <select v-bind="$attrs" :value="modelValue ?? ''" @change="onChange">
      <option v-for="option in options" :key="String(option.value ?? '')" :value="option.value ?? ''">
        {{ option.label }}
      </option>
    </select>
  `
}


vi.mock('@/api/admin/payment', () => ({ adminPaymentAPI: { getPlans } }))

let cleanup = () => {}
afterEach(() => cleanup())
function mountAdminRedeem() {
 const wrapper = mount(RedeemView, {
  attachTo: document.body,
  global: { stubs: {
   AppLayout: { template: '<div><slot /></div>' },
   TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
   DataTable: DataTableStub, Pagination: true, ConfirmDialog: true, Select: SelectStub, GroupBadge: true, GroupOptionItem: true, Icon: true, Teleport: true
  } }
 })
 cleanup = () => wrapper.unmount()
 return wrapper
}

describe('订阅兑换码库存来源', () => {
 beforeEach(() => {
  vi.clearAllMocks()
  localStorage.clear()
  document.body.innerHTML = ''
  listRedeemCodes.mockResolvedValue({ items: [{ id: 1, code: 'SUB-OLD', type: 'subscription', value: 0, status: 'unused', group_id: null, validity_days: 7, used_by: null, used_at: null, created_at: '2026-01-01T00:00:00Z', expires_at: null }], total: 1, page: 1, page_size: 20, pages: 1 })
  getAllGroups.mockResolvedValue([{ id: 10, name: 'A', platform: 'openai', subscription_type: 'subscription', rate_multiplier: 1 }, { id: 20, name: 'B', platform: 'openai', subscription_type: 'subscription', rate_multiplier: 1 }])
  getPlans.mockResolvedValue({ data: [{ id: 21, group_id: 10, name: '售罄', stock_remaining: 0 }, { id: 22, group_id: 10, name: '可用', stock_remaining: 2 }, { id: 31, group_id: 20, name: '其他', stock_remaining: -1 }] })
  generateRedeemCodes.mockResolvedValue([{ id: 2, code: 'SUB-NEW' }])
  batchUpdateRedeemCodes.mockResolvedValue({ updated: 1 })
 })
 async function openGeneration() {
  const wrapper = mountAdminRedeem()
  await flushPromises()
  await wrapper.findAll('button').find(button => button.text() === 'admin.redeem.generateCodes')!.trigger('click')
  await wrapper.get('form').findAll('select')[0].setValue('subscription')
  await flushPromises()
  await wrapper.get('form').findAll('select')[1].setValue('10')
  await flushPromises()
  return wrapper
 }
 it('明确选择库存套餐后生成，并传递套餐ID', async () => {
  const wrapper = await openGeneration()
  expect(wrapper.get('#redeem-generate-stock-plan option[value="21"]').attributes('disabled')).toBeDefined()
  await wrapper.get('#redeem-generate-stock-plan').setValue('22')
  await wrapper.get('form').trigger('submit')
  await flushPromises()
  expect(generateRedeemCodes).toHaveBeenCalledWith(1, 'subscription', 10, 10, 30, undefined, 22)
 })
 it('未选择套餐时阻止生成', async () => {
  const wrapper = await openGeneration()
  await wrapper.get('form').trigger('submit')
  expect(generateRedeemCodes).not.toHaveBeenCalled()
  expect(showError).toHaveBeenCalledWith('payment.stock.selectAvailable')
 })
 it('切换分组清除旧来源，避免把A组套餐发给B组', async () => {
  const wrapper = await openGeneration()
  await wrapper.get('#redeem-generate-stock-plan').setValue('22')
  await wrapper.get('form').findAll('select')[1].setValue('20')
  await flushPromises()
  expect((wrapper.get('#redeem-generate-stock-plan').element as HTMLSelectElement).value).toBe('')
  await wrapper.get('form').trigger('submit')
  expect(generateRedeemCodes).not.toHaveBeenCalled()
 })
 it('库存读取失败时阻止生成', async () => {
  getPlans.mockRejectedValue(new Error('offline'))
  const wrapper = await openGeneration()
  expect(wrapper.get('#redeem-generate-stock-plan').attributes('disabled')).toBeDefined()
  await wrapper.get('form').trigger('submit')
  expect(generateRedeemCodes).not.toHaveBeenCalled()
 })
 it('展示生成接口返回的具体错误', async () => {
  generateRedeemCodes.mockRejectedValueOnce({ message: '库存套餐不存在或不属于此订阅分组' })
  const wrapper = await openGeneration()
  await wrapper.get('#redeem-generate-stock-plan').setValue('22')
  await wrapper.get('form').trigger('submit')
  await flushPromises()
  expect(showError).toHaveBeenCalledWith('库存套餐不存在或不属于此订阅分组')
 })
 it('跨页选中的订阅码继续禁止清空分组', async () => {
  const wrapper = mountAdminRedeem()
  await flushPromises()
  await wrapper.get('[data-test="select-code"]').setValue(true)
  listRedeemCodes.mockResolvedValueOnce({ items: [{ id: 2, code: 'BAL', type: 'balance', value: 2, status: 'unused', used_by: null, used_at: null, created_at: '2026-01-01T00:00:00Z', expires_at: null }], total: 1, page: 1, page_size: 20, pages: 1 })
  await wrapper.findAll('select')[0].setValue('balance')
  await flushPromises()
  await wrapper.get('[data-test="select-code"]').setValue(true)
  await wrapper.get('[data-test="batch-update-open"]').trigger('click')
  await wrapper.get('[data-test="batch-field-group"]').setValue(true)
  await flushPromises()
  const groupSelect = wrapper.get('[data-test="batch-update-form"]').get('select')
  expect(groupSelect.findAll('option').map(option => option.text())).not.toContain('admin.redeem.clearGroup')
  expect(wrapper.find('#redeem-batch-stock-plan').exists()).toBe(true)
 })
 it('批量编辑禁止清空订阅分组，并能为旧码绑定分组和套餐', async () => {
  const wrapper = mountAdminRedeem()
  await flushPromises()
  await wrapper.get('[data-test="select-code"]').setValue(true)
  await wrapper.get('[data-test="batch-update-open"]').trigger('click')
  await wrapper.get('[data-test="batch-field-group"]').setValue(true)
  await flushPromises()
  const groupSelect = wrapper.get('[data-test="batch-update-form"]').get('select')
  expect(groupSelect.findAll('option').map(option => option.text())).not.toContain('admin.redeem.clearGroup')
  await wrapper.get('[data-test="batch-update-form"]').trigger('submit')
  expect(batchUpdateRedeemCodes).not.toHaveBeenCalled()
  await groupSelect.setValue('10')
  await flushPromises()
  await wrapper.get('#redeem-batch-stock-plan').setValue('22')
  await wrapper.get('[data-test="batch-update-form"]').trigger('submit')
  await flushPromises()
  expect(batchUpdateRedeemCodes).toHaveBeenCalledWith([1], { group_id: 10, plan_id: 22 })
 })
})
