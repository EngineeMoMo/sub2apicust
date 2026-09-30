import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, RouterLinkStub, type VueWrapper } from '@vue/test-utils'
import { reactive } from 'vue'
import FamilyHomeView from '@/custom/views/FamilyHomeView.vue'
import ApiLandingView from '@/custom/views/ApiLandingView.vue'
import { customRoutes } from '@/custom/routes'

const harness = vi.hoisted(() => ({
  auth: {} as { isAuthenticated: boolean; isAdmin: boolean; user?: { username: string } },
  user: vi.fn(), token: vi.fn(), settings: vi.fn()
}))
vi.mock('@/stores', () => ({
  useAuthStore: () => harness.auth,
  useAppStore: () => ({ siteName: '魔法家族', siteLogo: '', fetchPublicSettings: harness.settings,
    cachedPublicSettings: { registration_enabled: true } })
}))
const global = { stubs: {
  RouterLink: RouterLinkStub, BrandThemeToggle: true,
  BrandPanel: { template: '<section data-testid="original-brand"><slot name="actions" /></section>' },
  GuestAction: { props: ['to'], template: '<button :data-target="to"><slot /></button>' }
} }
let wrapper: VueWrapper | undefined
function page() { wrapper = mount(FamilyHomeView, { global, attachTo: document.body }); return wrapper }
function linkTo(view: VueWrapper, text: string) {
  return view.findAllComponents(RouterLinkStub).find(link => link.text() === text)?.props('to')
}

beforeEach(() => {
  vi.stubEnv('VITE_MAGIC_RECIPES_URL', 'https://recipes.example.test/')
  vi.stubEnv('VITE_MAGIC_STUDIO_URL', 'https://studio.example.test/')
  vi.stubEnv('DEV', false)
  harness.settings.mockReset().mockResolvedValue({})
  harness.user.mockReset().mockReturnValue({ username: '测试用户' })
  harness.token.mockReset().mockReturnValue('private-login-token')
  harness.auth = reactive({ isAuthenticated: false, isAdmin: false,
    get user() { return harness.user() }, get token() { return harness.token() } })
})
afterEach(() => { wrapper?.unmount(); wrapper = undefined; window.history.replaceState({}, '', window.location.pathname); vi.unstubAllEnvs(); vi.restoreAllMocks() })

describe('统一产品首页', () => {
  it('产品展台默认预览配方，只显示一个面板，切换预览不登录或调用模型', async () => {
    const fetch = vi.spyOn(window, 'fetch')
    const open = vi.spyOn(window, 'open')
    const view = page()
    expect(view.findAll('[role="tab"]')).toHaveLength(4)
    expect(view.get('#family-tab-recipes').attributes('aria-selected')).toBe('true')
    expect(view.findAll('[role="tabpanel"]').filter(panel => panel.isVisible())).toHaveLength(1)
    for (const product of ['studio', 'synaroute', 'api', 'recipes']) {
      await view.get('#family-tab-' + product).trigger('click')
      expect(view.get('#' + product).isVisible()).toBe(true)
      expect(view.findAll('[role="tab"][aria-selected="true"]')).toHaveLength(1)
      expect(view.findAll('[role="tabpanel"]').filter(panel => panel.isVisible())).toHaveLength(1)
      expect(view.get('#' + product).attributes('aria-labelledby')).toBe('family-tab-' + product)
    }
    expect(fetch).not.toHaveBeenCalled()
    expect(open).not.toHaveBeenCalled()
  })

  it('预览支持左右方向键、首尾键和循环焦点，Tab仅停在当前标签', async () => {
    const view = page()
    await view.get('#family-tab-recipes').trigger('keydown', { key: 'ArrowRight' })
    expect(document.activeElement).toBe(view.get('#family-tab-studio').element)
    expect(view.get('#family-tab-studio').attributes('tabindex')).toBe('0')
    expect(view.get('#family-tab-recipes').attributes('tabindex')).toBe('-1')
    await view.get('#family-tab-studio').trigger('keydown', { key: 'End' })
    expect(document.activeElement).toBe(view.get('#family-tab-synaroute').element)
    await view.get('#family-tab-synaroute').trigger('keydown', { key: 'ArrowRight' })
    expect(document.activeElement).toBe(view.get('#family-tab-api').element)
    await view.get('#family-tab-api').trigger('keydown', { key: 'ArrowLeft' })
    expect(document.activeElement).toBe(view.get('#family-tab-synaroute').element)
    await view.get('#family-tab-synaroute').trigger('keydown', { key: 'Home' })
    expect(document.activeElement).toBe(view.get('#family-tab-api').element)
  })

  it('已知产品锚点选择对应预览，未知锚点不改变入口或触发跳转', () => {
    window.history.replaceState({}, '', '#studio')
    expect(page().get('#studio').isVisible()).toBe(true)
    wrapper?.unmount()
    window.history.replaceState({}, '', '#unknown')
    expect(page().get('#recipes').isVisible()).toBe(true)
  })

  it('结构示意与模型授权说明保留，详情可展开而非占满首屏', async () => {
    const view = page()
    expect(view.get('#recipes figcaption').text()).toContain('非模型输出')
    expect(view.get('.mofa-family-session details').attributes('open')).toBeUndefined()
    expect(view.get('.mofa-family-session details').text()).toContain('仍由你选择已有密钥并明确授权')
    expect(view.get('.mofa-family-session summary').text()).toContain('登录与模型授权')
  })

  it('匿名首页仅加载公开设置，不读用户凭据，不自动打开产品或生成请求', () => {
    const fetch = vi.spyOn(window, 'fetch')
    const open = vi.spyOn(window, 'open')
    const view = page()
    expect(view.findAll('.mofa-family-product')).toHaveLength(4)
    expect(harness.settings).toHaveBeenCalledOnce()
    expect(harness.user).not.toHaveBeenCalled()
    expect(harness.token).not.toHaveBeenCalled()
    expect(fetch).not.toHaveBeenCalled()
    expect(open).not.toHaveBeenCalled()
    expect(view.text()).toContain('SynaRoute')
    expect(view.get('[data-testid="original-brand"]').text()).toContain('先看看订阅套餐')
    expect(view.find('.mofa-family-account').exists()).toBe(false)
    expect(linkTo(view, '登录')).toBe('/login')
    expect(linkTo(view, '注册')).toBe('/register')
    expect(linkTo(view, '登录后进入控制台')).toEqual({ path: '/login', query: { redirect: '/dashboard' } })
    expect(linkTo(view, '登录后打开魔法配方')).toEqual({ path: '/login', query: { redirect: '/dashboard?product=recipes' } })
    expect(linkTo(view, '登录后打开魔法工坊')).toEqual({ path: '/login', query: { redirect: '/dashboard?product=studio' } })
  })

  it('有效登录态直接进入本站产品，独立配方只传公开地址并在新窗口打开', async () => {
    const view = page()
    harness.auth.isAuthenticated = true
    await view.vm.$nextTick()
    expect(harness.user).not.toHaveBeenCalled()
    expect(linkTo(view, '进入控制台')).toBe('/dashboard')
    expect(linkTo(view, '配置 SynaRoute')).toBe('/keys')
    expect(view.get('#recipes a').attributes()).toMatchObject({
      href: 'https://recipes.example.test/', target: '_blank', rel: 'noopener noreferrer'
    })
    expect(view.findAllComponents(RouterLinkStub).some(link => {
      const target = link.props('to')
      return typeof target === 'object' && target.path === '/login'
    })).toBe(false)
    expect(harness.token).not.toHaveBeenCalled()
    expect(view.html()).not.toContain('private-login-token')
    expect(view.get('#studio a').attributes()).toMatchObject({
      href: 'https://studio.example.test/', target: '_blank', rel: 'noopener noreferrer'
    })
  })

  it('管理员入口与退出状态响应式更新，不改变其他登录流程', async () => {
    const view = page()
    harness.auth.isAuthenticated = true
    harness.auth.isAdmin = true
    await view.vm.$nextTick()
    const entries = view.findAllComponents(RouterLinkStub).filter(link => link.text() === '进入控制台')
    expect(entries).toHaveLength(2)
    expect(entries.every(link => link.props('to') === '/admin/dashboard')).toBe(true)
    harness.auth.isAuthenticated = false
    await view.vm.$nextTick()
    expect(view.text()).not.toContain('测试用户')
    expect(view.get('#recipes').text()).toContain('登录后打开魔法配方')
    expect(linkTo(view, '游客预览控制台')).toBe('/preview')
  })

  it('生产地址未配置时独立产品不可点击，不虚构已上线', () => {
    vi.stubEnv('VITE_MAGIC_RECIPES_URL', '')
    vi.stubEnv('VITE_MAGIC_STUDIO_URL', '')
    const view = page()
    expect(view.get('#recipes button').attributes('disabled')).toBeDefined()
    expect(view.get('#recipes').text()).toContain('待配置发布地址')
    expect(view.get('#recipes').text()).toContain('发布后开放')
    expect(view.get('#studio button').attributes('disabled')).toBeDefined()
    expect(view.get('#studio').text()).toContain('待配置发布地址')
    expect(view.get('#studio').text()).toContain('内容首版已完成')
  })

  it('原创图例及非实测说明可读屏，API节点保留原M轮廓', () => {
    const view = page()
    const caption = view.get('#studio figcaption')
    expect(caption.text()).toBe('原创视觉示例 · 非模板实测')
    expect(caption.element.closest('[aria-hidden="true"]')).toBeNull()
    expect(view.get('#studio img[alt="纸艺狐狸原创图例"]').exists()).toBe(true)
    expect(view.get('#api .mofa-api-node-mark').attributes('style')).toContain('--mofa-node-mask: url(')
  })

  it('家族首页与原 API 介绍都保留公开路由，配置选择页仍要求登录', () => {
    for (const path of ['/family', '/api']) {
      expect(customRoutes.find(route => route.path === path)?.meta?.requiresAuth).toBe(false)
    }
    expect(customRoutes.find(route => route.path === '/connect/recipes')?.meta?.requiresAuth).toBe(true)
    wrapper = mount(ApiLandingView, { global })
    expect(wrapper.text()).toContain('先了解服务，再决定如何接入')
    expect(wrapper.text()).toContain('管理 API Key')
    expect(linkTo(wrapper, '先看看订阅套餐')).toBe('/plans')
    expect(linkTo(wrapper, '查看接入教程')).toBe('/guide')
    expect(linkTo(wrapper, '登录')).toBe('/login')
  })
})
