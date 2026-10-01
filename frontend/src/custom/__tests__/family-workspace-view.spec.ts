import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { reactive } from 'vue'
import FamilyWorkspaceView from '@/custom/views/FamilyWorkspaceView.vue'

const harness = vi.hoisted(() => ({ auth: { isAuthenticated: true }, route: { path: '/tools/recipes' }, list: vi.fn(), push: vi.fn() }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => harness.auth }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ apiBaseUrl: '' }) }))
vi.mock('@/api/keys', () => ({ keysAPI: { list: harness.list } }))
vi.mock('vue-router', () => ({ useRoute: () => harness.route, useRouter: () => ({ push: harness.push }) }))
let wrapper: VueWrapper | undefined
beforeEach(() => { harness.auth = reactive({ isAuthenticated: true }); harness.route = reactive({ path: '/tools/recipes' }); harness.list.mockReset().mockResolvedValue({ items: [], pages: 1 }); harness.push.mockReset(); document.documentElement.classList.remove('dark') })
afterEach(() => { wrapper?.unmount(); document.documentElement.classList.remove('dark'); vi.restoreAllMocks() })
async function page() {
  wrapper = mount(FamilyWorkspaceView, { global: { stubs: { AppLayout: { template: '<main><slot /></main>' } } }, attachTo: document.body })
  const iframe = wrapper.get('iframe').element as HTMLIFrameElement, child = iframe.contentWindow!
  const send = vi.spyOn(child, 'postMessage'); await wrapper.get('iframe').trigger('load')
  const nonce = send.mock.calls.at(-1)![0].nonce
  const message = (action: string, payload: Record<string, unknown> = {}) => window.dispatchEvent(new MessageEvent('message', { origin: location.origin, source: child, data: { type: 'mofa-host-request', nonce, id: '1', action, payload } }))
  return { send, message, nonce }
}
describe('真实工作区组件的宿主交互（模拟账号）', () => {
  it('当前正文加载静态模块、初始化同步浅色，切深色实时同步', async () => {
    const { send } = await page()
    expect(wrapper!.get('iframe').attributes('src')).toBe('/recipes/?embedded=1')
    expect(send.mock.calls[0][0]).toMatchObject({ type: 'mofa-host-init', theme: 'light' })
    document.documentElement.classList.add('dark'); await flushPromises()
    expect(send.mock.calls.at(-1)![0]).toMatchObject({ type: 'mofa-host-theme', theme: 'dark' })
    expect(wrapper!.findAll('iframe')).toHaveLength(1)
  })
  it('读取密钥仍由宿主认证API处理，只将安全目录回传子页', async () => {
    const { send, message } = await page(); message('keys', { kind: 'text' }); await flushPromises()
    expect(harness.list).toHaveBeenCalledWith(1, 100, { status: 'active' }, { signal: expect.any(AbortSignal) })
    expect(send.mock.calls.at(-1)![0]).toMatchObject({ type: 'mofa-host-response', result: [] })
  })
  it('退出通知子页清除连接，随后不再读取认证API', async () => {
    const { send, message } = await page(); harness.auth.isAuthenticated = false; await flushPromises()
    expect(send.mock.calls.at(-1)![0].type).toBe('mofa-host-session-ended')
    message('keys', { kind: 'text' }); await flushPromises(); expect(harness.list).not.toHaveBeenCalled()
  })
  it('返回家族首页走当前路由，工坊不开放配方配置通道', async () => {
    harness.route.path = '/tools/studio'; const { message } = await page()
    expect(wrapper!.get('iframe').attributes('src')).toBe('/studio/?embedded=1')
    message('keys', { kind: 'text' }); await flushPromises(); expect(harness.list).not.toHaveBeenCalled()
    message('family'); expect(harness.push).toHaveBeenCalledWith('/family')
    message('studio-submit'); expect(harness.push).toHaveBeenCalledWith('/tools/studio/submit')
  })
})
