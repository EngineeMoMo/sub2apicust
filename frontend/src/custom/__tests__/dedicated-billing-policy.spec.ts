import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import DedicatedBillingPolicy from '@/custom/components/DedicatedBillingPolicy.vue'
import { dedicatedAPI, type DedicatedBillingPolicy as Policy } from '@/custom/dedicated/api'

vi.mock('@/custom/dedicated/api', async importOriginal => {
  const original = await importOriginal<typeof import('@/custom/dedicated/api')>()
  return { ...original, dedicatedAPI: { ...original.dedicatedAPI, billingPolicy: vi.fn(), saveBillingPolicy: vi.fn() } }
})
const policy: Policy = { concurrency_limit: 2, rpm_limit: 30, daily_request_limit: 0, max_body_bytes: 2097152, allow_images: false, updated_at: null }
let wrapper: VueWrapper
beforeEach(() => { vi.mocked(dedicatedAPI.billingPolicy).mockReset().mockResolvedValue({ ...policy }); vi.mocked(dedicatedAPI.saveBillingPolicy).mockReset().mockResolvedValue({ ...policy, updated_at: '2026-10-01T01:00:00Z' }) })
afterEach(() => wrapper?.unmount())
async function page(locale = 'zh') { wrapper = mount(DedicatedBillingPolicy, { props: { bindingId: 1 }, attachTo: document.body, global: { plugins: [createI18n({ legacy: false, locale, messages: {} })] } }); await flushPromises(); return wrapper }
describe('包号共享使用限制', () => {
  it('显示共享范围、UTC日窗口、默认值与生图关闭', async () => {
    const view = await page()
    expect(view.text()).toContain('所有成员、专属组和密钥共用')
    expect(view.text()).toContain('UTC零点')
    expect(view.findAll('input[type=number]').map(field => (field.element as HTMLInputElement).value)).toEqual(['2', '30', '0', '2097152'])
    expect((view.get('input[type=checkbox]').element as HTMLInputElement).checked).toBe(false)
    expect(dedicatedAPI.saveBillingPolicy).not.toHaveBeenCalled()
  })
  it('保存具体限制并携带原始版本，不更改余额或绑定', async () => {
    const view = await page()
    await view.findAll('input[type=number]')[2]!.setValue('500')
    await view.get('form').trigger('submit')
    await flushPromises()
    expect(dedicatedAPI.saveBillingPolicy).toHaveBeenCalledWith(1, { ...policy, daily_request_limit: 500 })
    expect(view.text()).toContain('使用限制已保存')
    await view.get('form').trigger('submit'); await flushPromises()
    expect(vi.mocked(dedicatedAPI.saveBillingPolicy).mock.calls[1]?.[1].updated_at).toBe('2026-10-01T01:00:00Z')
  })
  it('版本冲突提示刷新且聚焦错误，不自动覆盖新值', async () => {
    vi.mocked(dedicatedAPI.saveBillingPolicy).mockRejectedValue({ response: { data: { reason: 'DEDICATED_ACCOUNT_STALE' } } })
    const view = await page(); await view.get('form').trigger('submit'); await flushPromises()
    expect(view.get('[role=alert]').text()).toContain('已被其他操作修改')
    expect(document.activeElement).toBe(view.get('[role=alert]').element)
    expect(dedicatedAPI.saveBillingPolicy).toHaveBeenCalledTimes(1)
  })
  it('迟到读取不能覆盖另一个包号的限制', async () => {
    let complete!: (value: Policy) => void
    vi.mocked(dedicatedAPI.billingPolicy).mockImplementationOnce(() => new Promise(resolve => { complete = resolve })).mockResolvedValueOnce({ ...policy, rpm_limit: 99 })
    wrapper = mount(DedicatedBillingPolicy, { props: { bindingId: 1 }, global: { plugins: [createI18n({ legacy: false, locale: 'zh', messages: {} })] } })
    await wrapper.setProps({ bindingId: 2 }); await flushPromises(); complete({ ...policy, rpm_limit: 1 }); await flushPromises()
    expect((wrapper.findAll('input[type=number]')[1]!.element as HTMLInputElement).value).toBe('99')
  })
  it('英文同样说明免扣和原计费边界', async () => {
    const view = await page('en'); expect(view.text()).toContain('billed at zero'); expect(view.text()).toContain('public groups retain normal billing')
  })
})
