import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, RouterLinkStub, type VueWrapper } from '@vue/test-utils'
import RecipeConnectView from '@/custom/views/RecipeConnectView.vue'
import type { ApiKey } from '@/types'

const harness = vi.hoisted(() => ({
  app: { siteName: '魔法家族', siteLogo: '', apiBaseUrl: 'https://api.test/v1', fetchPublicSettings: vi.fn() },
  auth: { isAuthenticated: true, token: 'login-token-must-stay' },
  route: { query: {} as Record<string, string> }, list: vi.fn(), models: vi.fn()
}))
vi.mock('@/stores/app', () => ({ useAppStore: () => harness.app }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => harness.auth }))
vi.mock('vue-router', () => ({ useRoute: () => harness.route }))
vi.mock('@/api/keys', () => ({ keysAPI: { list: harness.list } }))
vi.mock('@/api/url', () => ({ buildGatewayUrl: () => 'https://fallback.test/v1' }))
vi.mock('@/custom/recipes/connect', async importOriginal => ({ ...await importOriginal<object>(), loadModels: harness.models }))
const global = { stubs: { RouterLink: RouterLinkStub, BrandThemeToggle: true } }
let wrapper: VueWrapper | undefined
let opener: { postMessage: ReturnType<typeof vi.fn> }
const nonce = 'a'.repeat(64)
function key(overrides: Partial<ApiKey> = {}): ApiKey {
  return { id: 1, key: 'fake-selected-key', name: '我的文字密钥', status: 'active', quota: 0, quota_used: 0, group_id: 1,
    group: { id: 1, name: '我的分组', status: 'active', platform: 'openai', allow_image_generation: true }, ...overrides } as ApiKey
}
async function page() {
  wrapper = mount(RecipeConnectView, { global })
  await flushPromises()
  return wrapper
}
async function choose(view: VueWrapper) {
  await view.get('#recipe-key').setValue('1')
  await flushPromises()
  await view.get('#recipe-model').setValue('model-a')
}
function ack(source = opener, origin = window.location.origin, token = nonce) {
  window.dispatchEvent(new MessageEvent('message', { source: source as unknown as Window, origin,
    data: { type: 'mofa-recipes-applied', nonce: token } }))
}
beforeEach(() => {
  vi.useRealTimers()
  opener = { postMessage: vi.fn() }
  Object.defineProperty(window, 'opener', { configurable: true, value: opener })
  harness.route.query = { origin: window.location.origin, nonce, kind: 'text' }
  harness.auth.isAuthenticated = true
  harness.app.apiBaseUrl = 'https://api.test/v1'
  harness.app.fetchPublicSettings.mockReset().mockResolvedValue({})
  harness.list.mockReset().mockResolvedValue({ items: [key()], pages: 1 })
  harness.models.mockReset().mockResolvedValue(['model-a', 'model-b'])
})
afterEach(() => { wrapper?.unmount(); wrapper = undefined; vi.useRealTimers() })

describe('登录后配置选择页', () => {
  it('无原配方窗口或来源未获准时不读取任何密钥', async () => {
    Object.defineProperty(window, 'opener', { configurable: true, value: null })
    let view = await page()
    expect(view.text()).toContain('未找到原配方页面')
    expect(harness.list).not.toHaveBeenCalled()
    view.unmount()
    Object.defineProperty(window, 'opener', { configurable: true, value: opener })
    harness.route.query.origin = 'https://evil.test'
    view = await page()
    expect(view.text()).toContain('尚未获准')
    expect(harness.list).not.toHaveBeenCalled()
  })
  it('未登录不读取密钥，登录保护仍交给既有路由守卫', async () => {
    harness.auth.isAuthenticated = false
    await page()
    expect(harness.list).not.toHaveBeenCalled()
    expect(harness.models).not.toHaveBeenCalled()
  })
  it('分页读取本人密钥，只显示有效名称分组，不显示原文或登录令牌', async () => {
    harness.list.mockResolvedValueOnce({ items: [key(), key({ id: 2, status: 'inactive' })], pages: 2 })
      .mockResolvedValueOnce({ items: [key({ id: 3, expires_at: '2000-01-01' }), key({ id: 4, name: '另一密钥' })], pages: 2 })
    const view = await page()
    expect(harness.list).toHaveBeenCalledTimes(2)
    expect(harness.list.mock.calls[1].slice(0, 3)).toEqual([2, 100, { status: 'active' }])
    expect(view.findAll('#recipe-key option')).toHaveLength(3)
    expect(view.text()).toContain('我的文字密钥 · 我的分组')
    expect(view.html()).not.toContain('fake-selected-key')
    expect(view.html()).not.toContain(harness.auth.token)
    expect(opener.postMessage).not.toHaveBeenCalled()
    expect(harness.models).not.toHaveBeenCalled()
  })
  it('文字选择目录与默认Responses，显式应用后严格等待原窗口确认', async () => {
    const view = await page()
    await choose(view)
    expect(harness.models.mock.calls[0].slice(0, 2)).toEqual(['https://api.test/v1', 'fake-selected-key'])
    expect((view.get('#recipe-protocol').element as HTMLSelectElement).value).toBe('responses')
    expect(opener.postMessage).not.toHaveBeenCalled()
    await view.get('form').trigger('submit')
    expect(opener.postMessage).toHaveBeenCalledWith({ type: 'mofa-recipes-connection', nonce,
      config: { kind: 'text', protocol: 'responses', base: 'https://api.test/v1', key: 'fake-selected-key', model: 'model-a', size: '' } }, window.location.origin)
    expect(JSON.stringify(opener.postMessage.mock.calls)).not.toContain(harness.auth.token)
    ack({ postMessage: vi.fn() }); ack(opener, 'https://evil.test'); ack(opener, window.location.origin, 'wrong')
    await flushPromises()
    expect(view.text()).not.toContain('配置已带回')
    ack()
    await flushPromises()
    expect(view.text()).toContain('配置已带回')
    expect(view.find('#recipe-key').exists()).toBe(false)
  })
  it('生图只列开放权限的密钥，不臆造目录中每个模型都能生图', async () => {
    harness.route.query.kind = 'image'
    harness.list.mockResolvedValue({ items: [key(), key({ id: 2, group: { ...key().group!, allow_image_generation: false } })], pages: 1 })
    const view = await page()
    expect(view.findAll('#recipe-key option')).toHaveLength(2)
    await choose(view)
    expect(view.find('#recipe-protocol').exists()).toBe(false)
    expect(view.text()).toContain('目录不代表每个模型都有生图能力')
    await view.get('form').trigger('submit')
    expect(opener.postMessage.mock.calls[0][0].config.protocol).toBe('images')
  })
  it('切换密钥时隔离迟到目录，刷新可重试空列表', async () => {
    harness.list.mockResolvedValue({ items: [key(), key({ id: 2, key: 'fake-second-key' })], pages: 1 })
    let resolveFirst: (models: string[]) => void = () => {}
    harness.models.mockImplementationOnce(() => new Promise<string[]>(resolve => { resolveFirst = resolve }))
    const view = await page()
    await view.get('#recipe-key').setValue('1')
    const firstSignal = harness.models.mock.calls[0][2] as AbortSignal
    await view.get('#recipe-key').setValue('2')
    await flushPromises()
    resolveFirst(['stale-model'])
    await flushPromises()
    expect(firstSignal.aborted).toBe(true)
    expect(view.findAll('#recipe-model option').map(option => option.text())).toEqual(['请选择模型', 'model-a', 'model-b'])
    harness.list.mockResolvedValue({ items: [], pages: 1 })
    await view.findAll('button').find(button => button.text() === '刷新密钥列表')!.trigger('click')
    await flushPromises()
    expect(view.text()).toContain('没有可用的密钥')
  })
  it('密钥读取失败可重试；确认超时不宣称导入成功', async () => {
    harness.list.mockRejectedValueOnce(new Error('raw-secret'))
    const view = await page()
    expect(view.text()).toContain('密钥列表读取失败')
    expect(view.text()).not.toContain('raw-secret')
    await view.findAll('button').find(button => button.text() === '刷新密钥列表')!.trigger('click')
    await flushPromises()
    await choose(view)
    vi.useFakeTimers()
    await view.get('form').trigger('submit')
    await vi.advanceTimersByTimeAsync(10001)
    expect(view.text()).toContain('未确认接收')
    ack()
    await flushPromises()
    expect(view.text()).not.toContain('配置已带回')
  })
})
