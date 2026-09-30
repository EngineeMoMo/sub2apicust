import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, RouterLinkStub } from '@vue/test-utils'
import { reactive } from 'vue'
import BrandHomeView from '@/custom/views/BrandHomeView.vue'
import ConsoleFaqView from '@/custom/views/ConsoleFaqView.vue'
import { customRoutes } from '@/custom/routes'

const mocks = vi.hoisted(() => ({ auth: { isAuthenticated: false, isAdmin: false } }))
vi.mock('@/stores', () => ({
  useAuthStore: () => mocks.auth,
  useAppStore: () => ({ siteName: '魔法家族', siteLogo: '', fetchPublicSettings: async () => ({}), cachedPublicSettings: {} })
}))
const global = { stubs: {
  RouterLink: RouterLinkStub,
  BrandThemeToggle: true,
  GuestAction: true,
  BrandPanel: { template: '<section><slot name="actions" /></section>' },
  AppLayout: { template: '<div><aside>控制台侧栏</aside><main><slot /></main></div>' }
} }
beforeEach(() => { mocks.auth = reactive({ isAuthenticated: false, isAdmin: false }) })

describe('登录态与控制台帮助', () => {
  it('登录后官网导航和主按钮同时隐藏游客入口，退出后恢复', async () => {
    const wrapper = mount(BrandHomeView, { global })
    const entries = () => wrapper.findAllComponents(RouterLinkStub).filter(link => link.props('to') === '/preview')
    expect(entries()).toHaveLength(2)
    mocks.auth.isAuthenticated = true
    await wrapper.vm.$nextTick()
    expect(entries()).toHaveLength(0)
    expect(wrapper.text()).toContain('进入控制台')
    mocks.auth.isAuthenticated = false
    await wrapper.vm.$nextTick()
    expect(entries()).toHaveLength(2)
    wrapper.unmount()
  })
  it('FAQ保留控制台侧栏，正文只有一层main并支持搜索', async () => {
    const wrapper = mount(ConsoleFaqView, { global })
    expect(wrapper.get('aside').text()).toBe('控制台侧栏')
    expect(wrapper.findAll('main')).toHaveLength(1)
    expect(wrapper.find('nav[aria-label="官网导航"]').exists()).toBe(false)
    await wrapper.get('input').setValue('no-faq-123')
    expect(wrapper.text()).toContain('没有找到相关问题')
    expect(customRoutes.find(route => route.name === 'ConsoleFaq')?.meta?.requiresAuth).toBe(true)
    expect(customRoutes.find(route => route.name === 'PublicFaq')?.meta?.requiresAuth).toBe(false)
    wrapper.unmount()
  })
})
