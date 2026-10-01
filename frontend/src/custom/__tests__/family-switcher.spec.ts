import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, RouterLinkStub, type VueWrapper } from '@vue/test-utils'
import { reactive } from 'vue'
import FamilyProductSwitcher from '@/custom/components/FamilyProductSwitcher.vue'

const harness = vi.hoisted(() => ({
  auth: { isAuthenticated: true, isAdmin: false },
  route: { path: '/dashboard', query: {} as Record<string, unknown>, hash: '' },
  replace: vi.fn()
}))
vi.mock('@/stores', () => ({ useAuthStore: () => harness.auth }))
vi.mock('vue-router', () => ({ useRoute: () => harness.route, useRouter: () => ({ replace: harness.replace }) }))
let wrapper: VueWrapper | undefined
function page() { wrapper = mount(FamilyProductSwitcher, { global: { stubs: { RouterLink: RouterLinkStub } } }); return wrapper }

beforeEach(() => {
  vi.stubEnv('DEV', false)
  vi.stubEnv('VITE_MAGIC_RECIPES_URL', 'https://recipes.example.test/')
  vi.stubEnv('VITE_MAGIC_STUDIO_URL', 'https://studio.example.test/')
  harness.auth = reactive({ isAuthenticated: true, isAdmin: false })
  harness.route = reactive({ path: '/dashboard', query: {}, hash: '' })
  harness.replace.mockReset()
})
afterEach(() => { wrapper?.unmount(); wrapper = undefined; vi.unstubAllEnvs(); vi.restoreAllMocks() })

describe('控制台顶部产品切换', () => {
  it('有四个真实产品入口，不自动打开窗口或请求接口', () => {
    const open = vi.spyOn(window, 'open')
    const fetch = vi.spyOn(window, 'fetch')
    const view = page()
    expect(view.get('nav').attributes('aria-label')).toBe('魔法家族产品切换')
    expect(view.findAll('nav > *')).toHaveLength(4)
    expect(view.get('.mofa-product-current').text()).toBe('API 控制台')
    expect(view.findAll('a[target="_blank"]')).toHaveLength(2)
    for (const link of view.findAll('a[target="_blank"]')) expect(link.attributes('rel')).toBe('noopener noreferrer')
    expect(open).not.toHaveBeenCalled()
    expect(fetch).not.toHaveBeenCalled()
    expect(view.find('aside').exists()).toBe(false)
  })

  it('登录携带配方意图时仍停留控制台，显式打开按钮仅含公开地址', async () => {
    harness.route.query = { product: 'recipes', keep: '1' }
    const open = vi.spyOn(window, 'open')
    const view = page()
    expect(view.get('aside').text()).toContain('已进入 API 控制台')
    expect(view.get('aside a').attributes('href')).toBe('https://recipes.example.test/')
    expect(open).not.toHaveBeenCalled()
    await view.get('[aria-label="留在 API 控制台"]').trigger('click')
    expect(harness.replace).toHaveBeenCalledWith({ path: '/dashboard', query: { keep: '1' }, hash: '' })
  })

  it.each(['https://evil.example/', ['recipes', 'studio'], 'unknown'])('不把未知或重复意图当外部跳转 %s', product => {
    harness.route.query = { product }
    expect(page().find('aside').exists()).toBe(false)
  })

  it('未配置独立地址时使用内置产品，继续打开仍由用户主动点击', () => {
    vi.stubEnv('VITE_MAGIC_RECIPES_URL', '')
    vi.stubEnv('VITE_MAGIC_STUDIO_URL', '')
    harness.route.query = { product: 'recipes' }
    const view = page()
    expect(view.findAll('button:disabled')).toHaveLength(0)
    expect(view.findAllComponents(RouterLinkStub).find(link => link.text().includes('打开魔法配方'))?.props('to')).toBe('/tools/recipes')
    expect(view.findAllComponents(RouterLinkStub).find(link => link.text().includes('魔法工坊'))?.props('to')).toBe('/tools/studio')
  })

  it('显式错误的外部地址仍禁用，不悄悄回退或展示打开提示', () => {
    vi.stubEnv('VITE_MAGIC_RECIPES_URL', 'http://unsafe.example.test/')
    vi.stubEnv('VITE_MAGIC_STUDIO_URL', 'https://studio.test/?key=secret')
    harness.route.query = { product: 'recipes' }
    const view = page()
    expect(view.findAll('button:disabled')).toHaveLength(2)
    expect(view.find('aside').exists()).toBe(false)
  })

  it('密钥页仍属于 API 控制台，桌面入口明确仅配置，管理员入口随角色更新', async () => {
    const view = page()
    harness.route.path = '/keys'
    harness.auth.isAdmin = true
    await view.vm.$nextTick()
    expect(view.get('.mofa-product-current').text()).toBe('API 控制台')
    expect(view.get('[aria-label="SynaRoute（配置 API 密钥）"]').attributes('aria-current')).toBeUndefined()
    expect(view.findAllComponents(RouterLinkStub)[0].props('to')).toBe('/admin/dashboard')
  })

  it('退出后没有继续打开提示，其他页面不消费控制台意图', async () => {
    harness.route.query = { product: 'studio' }
    const view = page()
    expect(view.get('aside').text()).toContain('魔法工坊')
    harness.auth.isAuthenticated = false
    await view.vm.$nextTick()
    expect(view.find('aside').exists()).toBe(false)
    harness.auth.isAuthenticated = true
    harness.route.path = '/usage'
    await view.vm.$nextTick()
    expect(view.find('aside').exists()).toBe(false)
  })
})
