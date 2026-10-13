import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, RouterLinkStub, type VueWrapper } from '@vue/test-utils'
import { reactive } from 'vue'
import PublicLayout from '@/custom/components/PublicLayout.vue'

const harness = vi.hoisted(() => ({ auth: {} as { isAuthenticated: boolean; isAdmin: boolean } }))
vi.mock('@/stores', () => ({
  useAuthStore: () => harness.auth,
  useAppStore: () => ({ siteName: '魔法家族', siteLogo: '', fetchPublicSettings: async () => ({}), cachedPublicSettings: { registration_enabled: true } })
}))

let wrapper: VueWrapper | undefined
function layout() {
  wrapper = mount(PublicLayout, { attachTo: document.body, slots: { default: '<p>正文</p>' }, global: { stubs: { RouterLink: RouterLinkStub, BrandThemeToggle: true } } })
  return wrapper
}
// 与浏览器一致：滚动回调在下一帧执行；每次设置滚动位置后手动推进一帧。
let frames: FrameRequestCallback[] = []
function setScroll(y: number) {
  Object.defineProperty(window, 'scrollY', { configurable: true, value: y })
  window.dispatchEvent(new Event('scroll'))
  const queued = frames
  frames = []
  queued.forEach(task => task(0))
}

beforeEach(() => {
  harness.auth = reactive({ isAuthenticated: false, isAdmin: false })
  frames = []
  vi.stubGlobal('requestAnimationFrame', (task: FrameRequestCallback) => frames.push(task))
  vi.stubGlobal('cancelAnimationFrame', vi.fn())
  setScroll(0)
})
afterEach(() => {
  wrapper?.unmount()
  wrapper = undefined
  vi.unstubAllGlobals()
  setScroll(0)
})

describe('官网页头菜单', () => {
  it('同一个导航只渲染一次，菜单按钮用 aria-expanded 控制展开', async () => {
    const view = layout()
    expect(view.findAll('nav[aria-label="官网导航"]')).toHaveLength(1)
    expect(view.findAllComponents(RouterLinkStub).filter(link => link.props('to') === '/preview')).toHaveLength(1)
    const toggle = view.get('.mofa-nav-toggle')
    expect(toggle.attributes('aria-controls')).toBe('public-nav')
    expect(view.get('#public-nav').exists()).toBe(true)
    expect(toggle.attributes('aria-expanded')).toBe('false')
    expect(toggle.text()).toContain('打开菜单')
    await toggle.trigger('click')
    expect(toggle.attributes('aria-expanded')).toBe('true')
    expect(toggle.text()).toContain('关闭菜单')
    expect(view.get('header').classes()).toContain('is-menu-open')
  })

  it('Esc 关闭并把焦点还给菜单按钮；点导航链接后关闭', async () => {
    const view = layout()
    const toggle = view.get('.mofa-nav-toggle')
    await toggle.trigger('click')
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await view.vm.$nextTick()
    expect(toggle.attributes('aria-expanded')).toBe('false')
    expect(document.activeElement).toBe(toggle.element)
    await toggle.trigger('click')
    await view.get('#public-nav a').trigger('click')
    expect(toggle.attributes('aria-expanded')).toBe('false')
  })

  it('点击页头之外或焦点离开页头时关闭', async () => {
    const view = layout()
    const toggle = view.get('.mofa-nav-toggle')
    await toggle.trigger('click')
    document.body.dispatchEvent(new Event('pointerdown', { bubbles: true }))
    await view.vm.$nextTick()
    expect(toggle.attributes('aria-expanded')).toBe('false')
    await toggle.trigger('click')
    const outside = document.createElement('button')
    document.body.appendChild(outside)
    view.get('header').element.dispatchEvent(new FocusEvent('focusout', { relatedTarget: outside }))
    await view.vm.$nextTick()
    expect(toggle.attributes('aria-expanded')).toBe('false')
    outside.remove()
  })
})

describe('官网页头滚动', () => {
  it('靠近顶部保持显示，向下滚动收起、向上滚动出现，菜单打开时不收起', async () => {
    const view = layout()
    const header = () => view.get('header').classes()
    setScroll(60)
    await view.vm.$nextTick()
    expect(header()).toContain('is-scrolled')
    expect(header()).not.toContain('is-hidden')
    setScroll(600)
    await view.vm.$nextTick()
    expect(header()).toContain('is-hidden')
    setScroll(560)
    await view.vm.$nextTick()
    expect(header()).not.toContain('is-hidden')
    await view.get('.mofa-nav-toggle').trigger('click')
    setScroll(1200)
    await view.vm.$nextTick()
    expect(header()).not.toContain('is-hidden')
  })

  it('焦点进入页头时立即出现，键盘用户不会操作看不见的链接', async () => {
    const view = layout()
    setScroll(200)
    setScroll(900)
    await view.vm.$nextTick()
    expect(view.get('header').classes()).toContain('is-hidden')
    view.get('header').element.dispatchEvent(new FocusEvent('focusin', { bubbles: true }))
    await view.vm.$nextTick()
    expect(view.get('header').classes()).not.toContain('is-hidden')
  })
})
