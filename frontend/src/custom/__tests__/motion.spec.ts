import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { mount } from '@vue/test-utils'
import { defineComponent, h, withDirectives } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { motionAllowed, resetRevealForTest, vReveal } from '@/custom/motion/reveal'
import { finePointer, vPointer } from '@/custom/motion/pointer'
import { prefetchRoute, resetPrefetchForTest } from '@/custom/motion/prefetch'

const originalMatchMedia = window.matchMedia
function media(matches: Record<string, boolean>) {
  window.matchMedia = ((query: string) => ({ matches: Boolean(matches[query]), media: query, addEventListener() {}, removeEventListener() {} })) as unknown as typeof window.matchMedia
}

type Callback = (entries: { isIntersecting: boolean; target: Element }[]) => void
const io = vi.hoisted(() => ({ callback: undefined as Callback | undefined, observed: [] as Element[], unobserved: [] as Element[] }))
class FakeObserver {
  constructor(callback: Callback) { io.callback = callback }
  observe(element: Element) { io.observed.push(element) }
  unobserve(element: Element) { io.unobserved.push(element) }
  disconnect() {}
}

function revealHost(order?: number) {
  return defineComponent({ render: () => withDirectives(h('section', { class: 'target' }, '内容'), [[vReveal, order]]) })
}

beforeEach(() => {
  io.callback = undefined
  io.observed = []
  io.unobserved = []
  resetRevealForTest()
})
afterEach(() => {
  window.matchMedia = originalMatchMedia
  vi.unstubAllGlobals()
  resetRevealForTest()
})

describe('滚动出现 v-reveal', () => {
  it('减少动效时不隐藏内容、不注册观察器', () => {
    vi.stubGlobal('IntersectionObserver', FakeObserver)
    media({ '(prefers-reduced-motion: reduce)': true })
    const wrapper = mount(revealHost())
    expect(motionAllowed()).toBe(false)
    expect(document.documentElement.classList.contains('mofa-motion')).toBe(false)
    expect(io.observed).toHaveLength(0)
    expect(wrapper.get('.target').classes()).toContain('mofa-reveal')
    wrapper.unmount()
  })

  it('浏览器不支持 IntersectionObserver 时内容保持可见', () => {
    vi.stubGlobal('IntersectionObserver', undefined)
    media({})
    const wrapper = mount(revealHost())
    expect(motionAllowed()).toBe(false)
    expect(document.documentElement.classList.contains('mofa-motion')).toBe(false)
    wrapper.unmount()
  })

  it('进入视口后加 is-revealed 并停止观察，序号写入延迟且有上限', () => {
    vi.stubGlobal('IntersectionObserver', FakeObserver)
    media({})
    const wrapper = mount(revealHost(20))
    const element = wrapper.get('.target').element as HTMLElement
    expect(document.documentElement.classList.contains('mofa-motion')).toBe(true)
    expect(io.observed).toEqual([element])
    expect(element.style.getPropertyValue('--reveal-delay')).toBe('560ms')
    io.callback?.([{ isIntersecting: false, target: element }])
    expect(element.classList.contains('is-revealed')).toBe(false)
    io.callback?.([{ isIntersecting: true, target: element }])
    expect(element.classList.contains('is-revealed')).toBe(true)
    expect(io.unobserved).toContain(element)
    wrapper.unmount()
  })
})

describe('指针光斑 v-pointer', () => {
  const host = defineComponent({ render: () => withDirectives(h('div', { class: 'card' }), [[vPointer]]) })
  // 与浏览器一致：回调在下一帧执行，而不是在 requestAnimationFrame 调用内同步执行。
  let frames: FrameRequestCallback[] = []
  const flush = () => { const queued = frames; frames = []; queued.forEach(task => task(0)) }
  beforeEach(() => {
    frames = []
    vi.stubGlobal('requestAnimationFrame', (task: FrameRequestCallback) => frames.push(task))
    vi.stubGlobal('cancelAnimationFrame', vi.fn())
  })

  it('触屏或减少动效时指针移动不产生任何效果', () => {
    media({ '(hover: hover) and (pointer: fine)': false })
    expect(finePointer()).toBe(false)
    const wrapper = mount(host)
    const element = wrapper.get('.card').element as HTMLElement
    element.dispatchEvent(new MouseEvent('pointermove', { clientX: 10, clientY: 10 }))
    flush()
    expect(element.classList.contains('is-pointing')).toBe(false)
    expect(element.style.getPropertyValue('--mx')).toBe('')
    wrapper.unmount()
    media({ '(hover: hover) and (pointer: fine)': true, '(prefers-reduced-motion: reduce)': true })
    expect(finePointer()).toBe(false)
  })

  it('精细指针下按帧写入位置变量，同一帧只计算一次，离开与卸载时清理', () => {
    media({ '(hover: hover) and (pointer: fine)': true })
    const wrapper = mount(host, { attachTo: document.body })
    const element = wrapper.get('.card').element as HTMLElement
    const rect = vi.spyOn(element, 'getBoundingClientRect').mockReturnValue({ left: 100, top: 50, width: 200, height: 100 } as DOMRect)
    element.dispatchEvent(new MouseEvent('pointermove', { clientX: 120, clientY: 60 }))
    element.dispatchEvent(new MouseEvent('pointermove', { clientX: 150, clientY: 100 }))
    expect(frames).toHaveLength(1)
    flush()
    expect(rect).toHaveBeenCalledTimes(1)
    expect(element.style.getPropertyValue('--mx')).toBe('25.0%')
    expect(element.style.getPropertyValue('--my')).toBe('50.0%')
    expect(element.style.getPropertyValue('--tilt-x')).toBe('-0.500')
    expect(element.classList.contains('is-pointing')).toBe(true)
    element.dispatchEvent(new MouseEvent('pointerleave'))
    expect(element.classList.contains('is-pointing')).toBe(false)
    expect(element.style.getPropertyValue('--tilt-x')).toBe('0')
    const remove = vi.spyOn(element, 'removeEventListener')
    wrapper.unmount()
    expect(remove.mock.calls.map(([type]) => type)).toEqual(expect.arrayContaining(['pointermove', 'pointerleave']))
  })
})

describe('动效层样式契约', () => {
  const theme = readFileSync(resolve(process.cwd(), 'src/custom/theme.css'), 'utf8')
  const layer = theme.slice(theme.indexOf('2026-10-12 官网与登录页动效层'))

  // 删除指定 @media 块（按花括号配对），返回剩余文本。
  function stripBlocks(css: string, opener: RegExp) {
    let output = ''
    let index = 0
    for (let match = opener.exec(css.slice(index)); match; match = opener.exec(css.slice(index))) {
      const start = index + match.index
      let depth = 0
      let end = start
      for (let cursor = css.indexOf('{', start); cursor < css.length; cursor++) {
        if (css[cursor] === '{') depth++
        if (css[cursor] === '}' && --depth === 0) { end = cursor + 1; break }
      }
      if (end === start) throw new Error('theme.css 花括号不配对')
      output += css.slice(index, start)
      index = end
    }
    return output + css.slice(index)
  }

  it('新增动画全部受“减少动态效果”保护，主题扩散只由脚本在允许时触发', () => {
    expect(layer.length).toBeGreaterThan(1000)
    const ungated = stripBlocks(stripBlocks(layer, /@media[^{]*prefers-reduced-motion: no-preference[^{]*\{/), /@keyframes[^{]*\{/)
    const offenders = ungated.split('\n').filter(line => /animation(-name)?\s*:/.test(line) && !line.includes('mofa-theme-transition'))
    expect(offenders).toEqual([])
    expect(layer).toMatch(/@media \(prefers-reduced-motion: reduce\) \{[^}]*transition: none;/)
  })

  it('不使用强制覆盖，指针效果另限精细指针', () => {
    expect(layer).not.toContain('!important')
    expect(layer).toMatch(/@media \(hover: hover\) and \(pointer: fine\) and \(prefers-reduced-motion: no-preference\)/)
  })
})

describe('公开页预取', () => {
  beforeEach(() => resetPrefetchForTest())

  function router(loader: () => Promise<unknown>) {
    return createRouter({ history: createMemoryHistory(), routes: [
      { path: '/plans', component: loader },
      { path: '/:pathMatch(.*)*', component: { render: () => null } }
    ] })
  }

  it('只下载一次路由块，不发数据请求；未安装路由时直接跳过', async () => {
    const loader = vi.fn(() => Promise.resolve({ default: { render: () => null } }))
    const instance = router(loader)
    const fetch = vi.spyOn(window, 'fetch')
    prefetchRoute(instance, '/plans')
    prefetchRoute(instance, '/plans')
    await Promise.resolve()
    expect(loader).toHaveBeenCalledTimes(1)
    expect(fetch).not.toHaveBeenCalled()
    expect(() => prefetchRoute(undefined, '/plans')).not.toThrow()
    fetch.mockRestore()
  })

  it('省流模式或 2G 网络不预取，加载失败后允许重试', async () => {
    const loader = vi.fn(() => Promise.reject(new Error('offline')))
    const instance = router(loader)
    vi.stubGlobal('navigator', { ...navigator, connection: { saveData: true } })
    prefetchRoute(instance, '/plans')
    expect(loader).not.toHaveBeenCalled()
    vi.stubGlobal('navigator', { ...navigator, connection: { effectiveType: 'slow-2g' } })
    prefetchRoute(instance, '/plans')
    expect(loader).not.toHaveBeenCalled()
    vi.unstubAllGlobals()
    prefetchRoute(instance, '/plans')
    expect(loader).toHaveBeenCalledTimes(1)
    await new Promise(resolve => setTimeout(resolve))
    prefetchRoute(instance, '/plans')
    expect(loader).toHaveBeenCalledTimes(2)
  })
})
