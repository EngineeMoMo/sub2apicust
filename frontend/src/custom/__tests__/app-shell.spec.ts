import { afterEach, describe, expect, it, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { earlyScript, localeChunks, renderAppShell } from '@/custom/shell/app-shell.mjs'

describe('首屏分包', () => {
  it('Airwallex 与 Stripe 一样独立成块，不进入每页都加载的 vendor-misc', () => {
    const viteConfig = readFileSync(resolve(process.cwd(), 'vite.config.ts'), 'utf8')
    const rule = viteConfig.indexOf("id.includes('/@airwallex/')")
    const miscFallback = viteConfig.indexOf("return 'vendor-misc'")
    expect(rule).toBeGreaterThan(-1)
    expect(viteConfig.slice(rule, miscFallback)).toContain("return 'vendor-airwallex'")
    expect(rule).toBeLessThan(miscFallback)
    expect(viteConfig).toContain('mofaAppShell()')
  })
})

const html = '<!doctype html><html><head>\n<meta charset="UTF-8" />\n<link rel="stylesheet" href="/assets/index.css"></head><body><div id="app"></div></body></html>'

afterEach(() => {
  document.documentElement.classList.remove('dark')
  document.documentElement.removeAttribute('data-shell')
  document.head.querySelectorAll('link[rel="modulepreload"]').forEach(link => link.remove())
  localStorage.removeItem('theme')
  localStorage.removeItem('sub2api_locale')
  window.history.replaceState({}, '', '/')
  vi.restoreAllMocks()
})

describe('首屏骨架注入', () => {
  it('把内联样式与脚本放在样式表之前，带 CSP nonce 占位，骨架在 #app 内', () => {
    const output = renderAppShell(html, { zh: 'assets/zh-abc.js' })
    const style = output.indexOf('<style nonce="__CSP_NONCE_VALUE__">')
    const script = output.indexOf('<script nonce="__CSP_NONCE_VALUE__">')
    expect(style).toBeGreaterThan(output.indexOf('<meta charset'))
    expect(script).toBeGreaterThan(style)
    expect(script).toBeLessThan(output.indexOf('<link rel="stylesheet"'))
    expect(output).toContain('<div id="app"><div class="mofa-shell" aria-hidden="true">')
    expect(output).toContain('"zh":"/assets/zh-abc.js"')
    expect(output).not.toMatch(/<\/script>[^]*<\/script>[^]*<div id="app">/)
  })

  it('重复执行不叠加，没有 #app 的页面原样返回', () => {
    const once = renderAppShell(html)
    expect(renderAppShell(once)).toBe(once)
    const other = '<html><head></head><body><main></main></body></html>'
    expect(renderAppShell(other)).toBe(other)
  })

  it('只识别语言包入口块，兼容 Windows 路径', () => {
    expect(localeChunks({
      a: { type: 'chunk', fileName: 'assets/index-1.js', facadeModuleId: '/repo/src/i18n/locales/zh/index.ts' },
      b: { type: 'chunk', fileName: 'assets/index-2.js', facadeModuleId: 'D:\\repo\\src\\i18n\\locales\\en\\index.ts' },
      c: { type: 'chunk', fileName: 'assets/common.js', facadeModuleId: '/repo/src/i18n/locales/zh/common.ts' },
      d: { type: 'asset', fileName: 'assets/index.css' }
    })).toEqual({ zh: 'assets/index-1.js', en: 'assets/index-2.js' })
    expect(localeChunks(undefined)).toEqual({})
  })
})

describe('早期脚本', () => {
  const run = (locales: Record<string, string>) => new Function(earlyScript(locales))()

  it('与 main.ts 同口径恢复深色，按路径选骨架外观', () => {
    localStorage.setItem('theme', 'dark')
    for (const [path, shell] of [['/home', 'public'], ['/plans', 'public'], ['/', 'public'], ['/login', 'auth'], ['/register', 'auth'], ['/dashboard', 'app'], ['/preview/keys', 'app'], ['/homework', 'app']]) {
      window.history.replaceState({}, '', path)
      run({})
      expect(document.documentElement.getAttribute('data-shell'), path).toBe(shell)
    }
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('没有保存主题时不加深色；按已保存语言或浏览器语言预加载对应语言包', () => {
    vi.spyOn(navigator, 'language', 'get').mockReturnValue('zh-CN')
    run({ zh: 'assets/zh.js', en: 'assets/en.js' })
    expect(document.documentElement.classList.contains('dark')).toBe(false)
    const links = () => [...document.head.querySelectorAll<HTMLLinkElement>('link[rel="modulepreload"]')].map(link => link.getAttribute('href'))
    expect(links()).toEqual(['/assets/zh.js'])
    localStorage.setItem('sub2api_locale', 'en')
    run({ zh: 'assets/zh.js', en: 'assets/en.js' })
    expect(links()).toEqual(['/assets/zh.js', '/assets/en.js'])
  })

  it('开发模式没有语言包清单时不插入预加载', () => {
    run({})
    expect(document.head.querySelectorAll('link[rel="modulepreload"]')).toHaveLength(0)
  })
})
