import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter, RouterView } from 'vue-router'
import { customRoutes } from '@/custom/routes'
import { previewSections } from '@/custom/guest/preview'

const { fetchPlans, fetchSettings } = vi.hoisted(() => ({ fetchPlans: vi.fn(), fetchSettings: vi.fn() }))
vi.mock('@/custom/guest/api', () => ({ fetchPublicPlans: fetchPlans }))
vi.mock('@/stores', () => ({
  useAuthStore: () => ({ isAuthenticated: false, get user(): never { throw new Error('Personal data accessed') } }),
  useAppStore: () => ({ siteName: '魔法家族', siteLogo: '', fetchPublicSettings: fetchSettings })
}))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

async function render(path = '/preview') {
  const router = createRouter({ history: createMemoryHistory(), routes: [
    ...customRoutes.filter(route => route.name === 'GuestPreview'),
    { path: '/:pathMatch(.*)*', component: { template: '<div />' } }
  ] })
  await router.push(path)
  await router.isReady()
  const wrapper = mount(RouterView, { global: { plugins: [router], stubs: {
    BrandThemeToggle: true,
    BaseDialog: { props: ['show'], template: '<div v-if="show" role="dialog"><slot /><slot name="footer" /></div>' }
  } } })
  await flushPromises()
  return { wrapper, router }
}

beforeEach(() => {
  fetchSettings.mockReset().mockResolvedValue({})
  fetchPlans.mockReset().mockResolvedValue({ purchase_enabled: true, plans: [{ id: 7, name: '真实目录测试套餐', price: 20, currency: 'USD', validity_days: 30, validity_unit: 'days', features: [], supported_model_scopes: [] }] })
})
afterEach(() => { document.body.innerHTML = '' })

describe('游客控制台预览', () => {
  it('概览只加载公开设置，不读取账户或套餐', async () => {
    const { wrapper } = await render()
    expect(wrapper.text()).toContain('先看看你的工作空间')
    expect(wrapper.findAll('main')).toHaveLength(1)
    expect(fetchSettings).toHaveBeenCalledOnce()
    expect(fetchPlans).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it.each(previewSections.filter(section => section.columns.length))('$label仅介绍字段，操作提示可取消并保留登录目标', async section => {
    const { wrapper, router } = await render(`/preview/${section.id}`)
    expect(wrapper.text()).toContain('这里是功能预览，不是你的账户数据')
    expect(fetchPlans).not.toHaveBeenCalled()
    await wrapper.get('.mofa-guest-locked button').trigger('click')
    const destination = new URL(wrapper.get('[role="dialog"] a').attributes('href')!, 'http://localhost')
    expect(destination.pathname).toBe('/login')
    expect(destination.searchParams.get('redirect')).toBe(section.target)
    await wrapper.get('[role="dialog"] button').trigger('click')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    expect(router.currentRoute.value.path).toBe(`/preview/${section.id}`)
    wrapper.unmount()
  })

  it('菜单导航保留外壳并收起手机菜单，套餐和FAQ嵌入正文', async () => {
    const { wrapper } = await render()
    await wrapper.get('.mofa-guest-menu-toggle').trigger('click')
    expect(wrapper.get('.mofa-guest-menu-toggle').attributes('aria-expanded')).toBe('true')
    await wrapper.get('nav a[href="/preview/plans"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('真实目录测试套餐')
    expect(wrapper.get('.mofa-guest-menu-toggle').attributes('aria-expanded')).toBe('false')
    expect(wrapper.get('nav a[href="/preview/plans"]').attributes('aria-current')).toBe('page')
    expect(wrapper.findAll('main')).toHaveLength(1)
    await wrapper.get('nav a[href="/preview/faq"]').trigger('click')
    await flushPromises()
    expect(wrapper.findAll('main')).toHaveLength(1)
    await wrapper.get('input').setValue('no-matching-faq-987')
    expect(wrapper.text()).toContain('没有找到')
    wrapper.unmount()
  })

  it('接入教程在预览右侧展示，不跳离预览或请求账户数据', async () => {
    const { wrapper, router } = await render('/preview/guide')
    expect(router.currentRoute.value.path).toBe('/preview/guide')
    expect(wrapper.findAll('main')).toHaveLength(1)
    expect(wrapper.text()).toContain('Claude Code 桌面端')
    expect(wrapper.text()).toContain('Codex CLI 与桌面端')
    expect(fetchPlans).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('无效预览地址返回概览', async () => {
    const { wrapper, router } = await render('/preview/unknown')
    expect(router.currentRoute.value.path).toBe('/preview')
    expect(wrapper.text()).toContain('先看看你的工作空间')
    wrapper.unmount()
  })
})
