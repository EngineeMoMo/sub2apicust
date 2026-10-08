import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, RouterLinkStub } from '@vue/test-utils'
import PublicGuideView from '@/custom/views/PublicGuideView.vue'
import { customRoutes } from '@/custom/routes'

const app = vi.hoisted(() => ({ apiBaseUrl: 'https://ai.mofamilys.com/v1', fetchPublicSettings: vi.fn(), siteName: '魔法家族', siteLogo: '' }))
vi.mock('@/stores', () => ({ useAppStore: () => app, useAuthStore: () => ({ isAuthenticated: false }) }))
const global = { stubs: { RouterLink: RouterLinkStub, GuestAction: { template: '<button><slot /></button>' } } }

beforeEach(() => {
  app.apiBaseUrl = 'https://ai.mofamilys.com/v1'
  app.fetchPublicSettings.mockReset().mockResolvedValue({})
})

describe('接入教程', () => {
  it('目录按场景分类且所有锚点都有对应正文', () => {
    const wrapper = mount(PublicGuideView, { props: { embedded: true }, global })
    const links = wrapper.findAll('nav[aria-label="教程目录"] a')
    expect(links).toHaveLength(6)
    for (const link of links) expect(wrapper.find(link.attributes('href')!).exists()).toBe(true)
    expect(wrapper.get('#guide-cli').text()).toContain('Codex CLI')
    expect(wrapper.get('#guide-cli').text()).toContain('Aider')
    expect(wrapper.get('#guide-desktop').text()).toContain('Cherry Studio')
    expect(wrapper.get('#guide-editors').text()).toContain('Continue')
    expect(wrapper.get('#guide-verify').text()).toContain('核对本站记录')
    wrapper.unmount()
  })
  it('公开和控制台路由分别保留匿名及认证边界', () => {
    expect(customRoutes.find(route => route.path === '/guide')?.meta?.requiresAuth).toBe(false)
    expect(customRoutes.find(route => route.path === '/help/guide')?.meta?.requiresAuth).toBe(true)
  })

  it('按协议显示站点地址与占位符，不读取密钥', () => {
    const wrapper = mount(PublicGuideView, { props: { embedded: true }, global })
    const content = wrapper.text()
    expect(content).toContain('https://ai.mofamilys.com/v1')
    expect(content).toContain('ANTHROPIC_BASE_URL="https://ai.mofamilys.com"')
    expect(content).toContain('wire_api = "responses"')
    expect(content).toContain('填写实际可用模型ID')
    expect(content).toContain('Claude Code 桌面端')
    expect(content).toContain('Cline')
    expect(content).not.toContain('sk-')
    expect(app.fetchPublicSettings).toHaveBeenCalledOnce()
    wrapper.unmount()
  })

  it('使用本机公开 API 地址渲染示例，保留嵌入正文', () => {
    app.apiBaseUrl = 'http://localhost:8080'
    const wrapper = mount(PublicGuideView, { props: { embedded: true }, global })
    expect(wrapper.findAll('main')).toHaveLength(0)
    expect(wrapper.text()).toContain('http://localhost:8080/v1')
    wrapper.unmount()
  })

  it.each([
    ['console', '/help/faq', '/purchase?tab=subscription'],
    ['preview', '/preview/faq', '/preview/plans'],
    ['public', '/faq', '/plans']
  ] as const)('%s 中的帮助与套餐链接保留对应浏览入口', (scope, faqTarget, plansTarget) => {
    const wrapper = mount(PublicGuideView, { props: { embedded: true, scope }, global })
    const targets = wrapper.findAllComponents(RouterLinkStub).map(link => link.props('to'))
    expect(targets).toContain(faqTarget)
    expect(targets).toContain(plansTarget)
    wrapper.unmount()
  })
})
