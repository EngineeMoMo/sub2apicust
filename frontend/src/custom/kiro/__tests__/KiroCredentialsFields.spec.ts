import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import KiroCredentialsFields from '../KiroCredentialsFields.vue'

function mountFields(editing = false) {
  return mount(KiroCredentialsFields, {
    props: { modelValue: '', accountType: 'oauth', editing },
    global: { plugins: [createI18n({ legacy: false, locale: 'zh-CN', messages: {} })] }
  })
}

describe('Kiro credential fields', () => {
  it('requires an explicit estimated mode and preserves fields already entered', async () => {
    const wrapper = mountFields()
    expect(wrapper.get<HTMLSelectElement>('#kiro-usage-mode').element.value).toBe('verified')
    await wrapper.setProps({ modelValue: '{"refreshToken":"private","runtime_region":"us-east-1"}' })
    await wrapper.get('#kiro-usage-mode').setValue('estimated')
    const emitted = wrapper.emitted('update:modelValue')?.at(-1)?.[0] as string
    expect(JSON.parse(emitted)).toEqual({ refresh_token: 'private', runtime_region: 'us-east-1', usage_mode: 'estimated' })
    expect(wrapper.text()).toContain('按本站模型价格和倍率扣费')
  })
  it('allows editing only the mode without displaying saved secrets', async () => {
    const wrapper = mountFields(true)
    await wrapper.get('#kiro-usage-mode').setValue('estimated')
    expect(JSON.parse(wrapper.emitted('update:modelValue')?.at(-1)?.[0] as string)).toEqual({ usage_mode: 'estimated' })
    await wrapper.setProps({ modelValue: '{"usage_mode":"estimated"}' })
    await wrapper.get('#kiro-usage-mode').setValue('preserve')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([''])
  })
  it('does not overwrite incomplete credential JSON when changing mode', async () => {
    const wrapper = mountFields()
    await wrapper.setProps({ modelValue: '{"refreshToken":' })
    expect(wrapper.get<HTMLSelectElement>('#kiro-usage-mode').element.disabled).toBe(true)
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    expect(wrapper.get<HTMLTextAreaElement>('textarea').element.value).toBe('{"refreshToken":')
  })
  it('keeps saved credentials out of the edit form', () => {
    const wrapper = mountFields(true)
    expect(wrapper.find('#kiro-account-type').exists()).toBe(false)
    expect(wrapper.get('#kiro-usage-mode').element.value).toBe('preserve')
    expect(wrapper.get('textarea').element.value).toBe('')
    expect(wrapper.text()).toContain('留空保留现有凭据')
  })
  it('clears credentials when changing credential type', async () => {
    const wrapper = mountFields()
    await wrapper.get('#kiro-account-type').setValue('apikey')
    expect(wrapper.emitted('update:accountType')).toEqual([['apikey']])
    expect(wrapper.emitted('update:modelValue')).toEqual([['']])
  })
  it('reads a valid JSON file and clears the file input', async () => {
    const wrapper = mountFields()
    const input = wrapper.get('input[type="file"]')
    Object.defineProperty(input.element, 'files', { value: [{ size: 30, text: vi.fn().mockResolvedValue('{"refreshToken":"test"}') }] })
    await input.trigger('change')
    await vi.waitFor(() => expect(wrapper.emitted('update:modelValue')).toEqual([['{"refreshToken":"test"}']]))
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  })
  it('rejects files over the input limit', async () => {
    const wrapper = mountFields()
    const input = wrapper.get('input[type="file"]')
    Object.defineProperty(input.element, 'files', { value: [{ size: 2 * 1024 * 1024 }] })
    await input.trigger('change')
    expect(wrapper.find('[role="alert"]').exists()).toBe(true)
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })
})
