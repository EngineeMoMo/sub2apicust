import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, RouterLinkStub, type VueWrapper } from '@vue/test-utils'
import { reactive } from 'vue'
import { createPinia } from 'pinia'
import { mkdirSync, readFileSync, readdirSync, writeFileSync } from 'node:fs'
import { resolve } from 'node:path'
import AppHeader from '@/components/layout/AppHeader.vue'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import SubscriptionProgressMini from '@/components/common/SubscriptionProgressMini.vue'
import zh from '@/i18n/locales/zh'

const harness = vi.hoisted(() => ({
  auth: {} as { isAuthenticated: boolean; isAdmin: boolean; user: unknown },
  toggle: vi.fn(),
  app: { siteName: '魔法家族', docUrl: 'https://docs.example.com', contactInfo: '', toggleMobileSidebar: vi.fn(), showError: vi.fn(), showSuccess: vi.fn() },
  subscription: { activeSubscriptions: [{ id: 1, group_id: 1, group: { name: '模拟订阅', daily_limit_usd: 10 }, daily_usage_usd: 2 }], hasActiveSubscriptions: true, fetchActiveSubscriptions: vi.fn().mockResolvedValue(undefined) }
}))
vi.mock('@/stores', () => ({
  useAuthStore: () => harness.auth,
  useAppStore: () => harness.app,
  useOnboardingStore: () => ({}),
  useSubscriptionStore: () => harness.subscription
}))
vi.mock('@/stores/app', () => ({ useAppStore: () => harness.app }))
vi.mock('vue-i18n', async original => {
  const actual = await original<typeof import('vue-i18n')>()
  const { ref } = await import('vue')
  const { default: messages } = await import('@/i18n/locales/zh')
  const locale = ref('zh')
  return { ...actual, useI18n: () => ({ locale, t: (key: string) => {
    const message = key.split('.').reduce<unknown>((current, part) => current && typeof current === 'object' ? (current as Record<string, unknown>)[part] : undefined, messages)
    return typeof message === 'string' ? message : key
  } }) }
})
vi.mock('@/stores/announcements', async () => {
  const { defineStore } = await import('pinia')
  return { useAnnouncementStore: defineStore('family-preview-announcements', {
    state: () => ({ announcements: [], loading: false, unreadCount: 0, currentPopup: null }),
    actions: { markAsRead: vi.fn(), markAllAsRead: vi.fn() }
  }) }
})
vi.mock('@/i18n', () => ({ setLocale: vi.fn(), availableLocales: [{ code: 'zh', name: '简体中文', flag: '🇨🇳' }, { code: 'en', name: 'English', flag: '🇺🇸' }] }))
vi.mock('vue-router', () => ({ useRoute: () => ({ path: '/dashboard', query: {}, hash: '' }), useRouter: () => ({}) }))
vi.mock('@/utils/featureFlags', () => ({ FeatureFlags: { subscription: 'subscription', modelPlaza: 'modelPlaza' }, isFeatureFlagEnabled: () => true }))
vi.mock('@/custom/brand/useWorkspaceHeading', () => ({ useWorkspaceHeading: () => ({ pageTitle: 'API 控制台' }) }))
function mountHeader() {
  return mount(AppHeader, { global: { plugins: [createPinia()], stubs: { RouterLink: RouterLinkStub } } })
}
let wrapper: VueWrapper | undefined

beforeEach(() => {
  vi.stubEnv('DEV', false)
  vi.stubEnv('VITE_MAGIC_RECIPES_URL', 'http://127.0.0.1:4178')
  vi.stubEnv('VITE_MAGIC_STUDIO_URL', 'http://127.0.0.1:4179')
  harness.toggle.mockReset()
  harness.app.toggleMobileSidebar = harness.toggle
  harness.subscription.fetchActiveSubscriptions.mockClear()
  harness.auth = reactive({ isAuthenticated: true, isAdmin: false, user: { username: '界面演示', role: 'user', balance: 25.36, frozen_balance: 3 } })
})
afterEach(() => { wrapper?.unmount(); wrapper = undefined; vi.unstubAllEnvs() })

describe('原页头装配产品切换', () => {
  it('保留账户与移动菜单，匿名页头不装配产品切换', async () => {
    wrapper = mountHeader()
    expect(wrapper.find('[aria-label="魔法家族产品切换"]').exists()).toBe(true)
    expect(wrapper.get('[aria-label="' + zh.common.userMenu + '"]').exists()).toBe(true)
    await wrapper.get('[aria-label="' + zh.common.toggleMenu + '"]').trigger('click')
    expect(harness.toggle).toHaveBeenCalledOnce()
    harness.auth.user = null
    harness.auth.isAuthenticated = false
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.mofa-product-switcher').exists()).toBe(false)
  })

  it('真实公告、语言、订阅、文档、模型广场与冻结余额同时装配，不请求业务接口', () => {
    wrapper = mountHeader()
    expect(wrapper.findComponent(AnnouncementBell).exists()).toBe(true)
    expect(wrapper.findComponent(LocaleSwitcher).get('button').text()).toContain('ZH')
    expect(wrapper.findComponent(SubscriptionProgressMini).get('button').text()).toContain('1')
    expect(wrapper.get('a[href="https://docs.example.com/"]').text()).toContain(zh.nav.docs)
    expect(wrapper.text()).toContain(zh.nav.modelPlaza)
    expect(wrapper.text()).toContain('$25.36')
    expect(wrapper.text()).toContain('$3.00')
    expect(harness.subscription.fetchActiveSubscriptions).toHaveBeenCalledOnce()
  })

  it('导出真实组件的无凭据视觉夹具，仅用于布局检查', () => {
    wrapper = mountHeader()
    expect(wrapper.get('.mofa-product-capsule').text()).toContain('魔法工坊')
    if (process.env.FAMILY_CAPTURE !== '1') return
    const directory = resolve(process.cwd(), process.env.FAMILY_CAPTURE_DIRECTORY || '../output/family-layout-20260930')
    mkdirSync(directory, { recursive: true })
    const assets = resolve(process.cwd(), '../backend/internal/web/dist/assets')
    const css = readdirSync(assets).filter(name => name.startsWith('index-') && name.endsWith('.css')).map(name => readFileSync(resolve(assets, name), 'utf8')).join('\n')
    const mark = 'data:image/webp;base64,' + readFileSync(resolve(process.cwd(), 'src/custom/assets/mofa-mark.webp')).toString('base64')
    const header = wrapper.html().replaceAll('/src/custom/assets/mofa-mark.webp', mark)
    const html = '<!doctype html><html lang="zh-CN" class="dark"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>产品切换布局夹具 · 非真实登录</title><style>' + css + '</style><body><div class="mofa-workspace"><aside class="fixed inset-y-0 left-0 hidden w-64 p-6 lg:block"><h2>魔法家族</h2><p class="mt-8">API 控制台</p></aside><div class="mofa-workspace-shell lg:ml-64">' + header + '<main class="mofa-workspace-main p-8"><h1>API 控制台</h1><p class="mt-6">产品切换布局夹具 · 模拟账户，非真实登录</p><p class="mt-4">顶部为实际 AppHeader、产品切换、公告、语言与订阅组件渲染；数据为模拟，无凭据及业务接口连接。静态夹具不用于点击验收。</p></main></div></div></body></html>'
    writeFileSync(resolve(directory, 'header-preview.html'), html)
  })
})
