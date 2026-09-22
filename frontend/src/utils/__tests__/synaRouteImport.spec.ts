// [CUSTOM] SynaRoute 深链接导入 util 测试。登记见项目根 CUSTOMIZATIONS.md。
import { describe, expect, it } from 'vitest'

import {
  buildSynaRouteImportDeeplink,
  buildSynaRouteUrl,
  toUrlSafeBase64
} from '../synaRouteImport'

describe('toUrlSafeBase64', () => {
  it('URL-safe：不含 + / =，且可解回原文', () => {
    const encoded = toUrlSafeBase64('sk-abc+/=')
    expect(encoded).not.toMatch(/[+/=]/)
    const padded = encoded.replace(/-/g, '+').replace(/_/g, '/')
    expect(atob(padded)).toBe('sk-abc+/=')
  })

  it('UTF-8 安全（非 ASCII 不抛异常）', () => {
    expect(() => toUrlSafeBase64('密钥🔑')).not.toThrow()
    expect(toUrlSafeBase64('密钥🔑')).toMatch(/^[A-Za-z0-9_-]+$/)
  })
})

describe('buildSynaRouteUrl', () => {
  it('resource 固定 provider，host=v1、path=/import', () => {
    const url = buildSynaRouteUrl({ endpoint: 'https://relay.example.com/v1' })
    expect(url.startsWith('synaroute://v1/import?')).toBe(true)
    const q = new URLSearchParams(url.split('?')[1])
    expect(q.get('resource')).toBe('provider')
    expect(q.get('endpoint')).toBe('https://relay.example.com/v1')
  })

  it('apiKey 以 base64 写入；缺省时不写该参数', () => {
    const withKey = new URLSearchParams(
      buildSynaRouteUrl({ endpoint: 'https://a', apiKey: 'sk-x' }).split('?')[1]
    )
    expect(withKey.get('apiKey')).toBe(toUrlSafeBase64('sk-x'))
    const noKey = new URLSearchParams(
      buildSynaRouteUrl({ endpoint: 'https://a' }).split('?')[1]
    )
    expect(noKey.has('apiKey')).toBe(false)
  })

  it('端点与中文名经 URL 转义后可原样解回', () => {
    const q = new URLSearchParams(
      buildSynaRouteUrl({ endpoint: 'https://relay.example.com/v1', name: '我的中转站' }).split('?')[1]
    )
    expect(q.get('endpoint')).toBe('https://relay.example.com/v1')
    expect(q.get('name')).toBe('我的中转站')
  })
})

describe('buildSynaRouteImportDeeplink', () => {
  const parse = (url: string) => new URLSearchParams(url.split('?')[1])

  it('anthropic（默认）→ claude-cli / anthropic，端点去掉 /v1', () => {
    const q = parse(buildSynaRouteImportDeeplink({ platform: 'anthropic', baseUrl: 'https://h/v1', apiKey: 'k', name: '站点' }))
    expect(q.get('category')).toBe('claude-cli')
    expect(q.get('protocol')).toBe('anthropic')
    expect(q.get('endpoint')).toBe('https://h')
    expect(q.get('name')).toBe('站点')
    expect(q.get('apiKey')).toBe(toUrlSafeBase64('k'))
  })

  it('platform 缺省时回退 anthropic', () => {
    const q = parse(buildSynaRouteImportDeeplink({ platform: null, baseUrl: 'https://h', apiKey: 'k' }))
    expect(q.get('protocol')).toBe('anthropic')
    expect(q.get('category')).toBe('claude-cli')
  })

  it('openai → codex / openai_responses，端点带 /v1（不重复）', () => {
    expect(parse(buildSynaRouteImportDeeplink({ platform: 'openai', baseUrl: 'https://h', apiKey: 'k' })).get('endpoint')).toBe('https://h/v1')
    const q = parse(buildSynaRouteImportDeeplink({ platform: 'openai', baseUrl: 'https://h/v1', apiKey: 'k' }))
    expect(q.get('endpoint')).toBe('https://h/v1')
    expect(q.get('category')).toBe('codex')
    expect(q.get('protocol')).toBe('openai_responses')
  })

  it('antigravity → 端点加 /antigravity', () => {
    expect(parse(buildSynaRouteImportDeeplink({ platform: 'antigravity', baseUrl: 'https://h', apiKey: 'k' })).get('endpoint')).toBe('https://h/antigravity')
  })

  it('grok → 带 grok-4.5 三档 + Fable 映射', () => {
    const q = parse(buildSynaRouteImportDeeplink({ platform: 'grok', baseUrl: 'https://h', apiKey: 'k' }))
    expect(q.get('opusModel')).toBe('grok-4.5')
    expect(q.get('sonnetModel')).toBe('grok-4.5')
    expect(q.get('haikuModel')).toBe('grok-4.5')
    expect(q.get('fableModel')).toBe('grok-4.5')
  })

  it('gemini → 无协议/分类，仅带端点与密钥，交由确认框选择', () => {
    const q = parse(buildSynaRouteImportDeeplink({ platform: 'gemini', baseUrl: 'https://h/v1', apiKey: 'k' }))
    expect(q.has('protocol')).toBe(false)
    expect(q.has('category')).toBe(false)
    expect(q.get('endpoint')).toBe('https://h')
    expect(q.get('apiKey')).toBe(toUrlSafeBase64('k'))
  })
})
