import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { connectionRequest, gatewayBase, loadModels, selectedConnection, usableKey, catalogModels } from '@/custom/recipes/connect'
import { safeGuestRedirect } from '@/custom/guest/navigation'
import { customRoutes } from '@/custom/routes'
import type { ApiKey } from '@/types'

const nonce = 'a'.repeat(64)
const origin = 'https://ai.mofamilys.com'
const request = { origin, nonce, kind: 'text' as const }
function key(overrides: Partial<ApiKey> = {}): ApiKey {
  return { id: 1, name: '测试密钥', key: 'fake-key', status: 'active', group_id: 1, quota: 0, quota_used: 0,
    group: { id: 1, name: '测试分组', status: 'active', platform: 'openai', allow_image_generation: true }, ...overrides } as ApiKey
}

beforeEach(() => vi.restoreAllMocks())
afterEach(() => vi.unstubAllGlobals())

describe('魔法配方连接授权边界', () => {
  it('登录保护保留站内选择目标，URL不需要任何密钥', () => {
    expect(customRoutes.find(route => route.path === '/connect/recipes')?.meta?.requiresAuth).toBe(true)
    const path = '/connect/recipes?origin=' + encodeURIComponent(origin) + '&nonce=' + nonce + '&kind=text'
    expect(safeGuestRedirect(path)).toBe(path)
  })
  it('只接受同源或明确配置的独立配方来源，生产不隐式信任本机', () => {
    expect(connectionRequest(request, origin)).toEqual(request)
    expect(connectionRequest({ ...request, origin: 'https://recipes.test' }, origin, 'https://recipes.test').origin).toBe('https://recipes.test')
    expect(connectionRequest({ ...request, origin: 'http://127.0.0.1:4178' }, 'http://127.0.0.1:4175').origin).toBe('http://127.0.0.1:4178')
    expect(() => connectionRequest({ ...request, origin: 'http://127.0.0.1:4178' }, origin)).toThrow('尚未获准')
    expect(() => connectionRequest({ ...request, origin: 'https://evil.test' }, origin)).toThrow('尚未获准')
  })
  it.each([
    { nonce: 'wrong' }, { nonce: ['a'.repeat(64)] }, { kind: 'video' }, { kind: ['text'] },
    { origin: origin + '/path' }, { origin: origin + '/' }, { origin: origin + '?key=secret' },
    { origin: 'https://user:password@ai.mofamilys.com' }, { origin: 'http://remote.test' }
  ])('拒绝无效请求 %s', invalid => expect(() => connectionRequest({ ...request, ...invalid }, origin)).toThrow())
  it('独立来源配置不可包含路径、通配符或不安全 HTTP', () => {
    for (const configured of ['https://recipes.test/path', 'https://*.test', 'http://remote.test']) expect(() => connectionRequest(request, origin, configured)).toThrow()
  })
  it('密钥必须启用、有效分组、未过期且未耗尽配额', () => {
    expect(usableKey(key())).toBe(true)
    for (const invalid of [key({ status: 'inactive' }), key({ key: '' }), key({ group_id: null }), key({ group: undefined }), key({ group: { ...key().group!, status: 'inactive' } }), key({ expires_at: 'invalid' }), key({ expires_at: '2000-01-01' }), key({ quota: 2, quota_used: 2 })]) expect(usableKey(invalid)).toBe(false)
    expect(usableKey(key({ quota: 0, quota_used: 100 }))).toBe(true)
  })
  it('公开 API 地址优先，根地址补v1，已有路径保留', () => {
    expect(gatewayBase('', 'http://127.0.0.1:8080/v1')).toBe('http://127.0.0.1:8080/v1')
    expect(gatewayBase('https://api.test', '')).toBe('https://api.test/v1')
    expect(gatewayBase('https://api.test/gateway/v1/', '')).toBe('https://api.test/gateway/v1')
    expect(() => gatewayBase('https://api.test/?key=secret', '')).toThrow()
  })
  it('模型目录去重，不根据名称臆造支持能力', () => {
    expect(catalogModels({ data: [{ id: 'model-b' }, { id: 'model-a' }, { id: 'model-b' }, { id: '' }, { id: 7 }] })).toEqual(['model-a', 'model-b'])
    expect(() => catalogModels({ models: [] })).toThrow()
  })
  it('模型目录GET只带所选Key，不带登录Cookie、不跟随重定向', async () => {
    const fetcher = vi.fn().mockResolvedValue({ ok: true, json: async () => ({ data: [{ id: 'test-model' }] }) })
    vi.stubGlobal('fetch', fetcher)
    const signal = new AbortController().signal
    expect(await loadModels('https://api.test/v1', 'fake-key', signal)).toEqual(['test-model'])
    expect(fetcher).toHaveBeenCalledWith('https://api.test/v1/models', {
      headers: { Authorization: 'Bearer fake-key' }, signal, credentials: 'omit', redirect: 'error', referrerPolicy: 'no-referrer', cache: 'no-store'
    })
  })
  it('目录错误不回显接口可能包含的密钥或原始数据', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('secret-key')))
    await expect(loadModels('https://api.test/v1', 'fake', new AbortController().signal)).rejects.toThrow('模型目录连接失败')
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 403 }))
    await expect(loadModels('https://api.test/v1', 'fake', new AbortController().signal)).rejects.toThrow('HTTP 403')
  })
  it('只导出选择的连接，不导出账号凭据，生图禁用及目录外模型被拒绝', () => {
    expect(selectedConnection(request, key(), ['model'], 'model', 'https://api.test/v1', 'responses')).toEqual({ kind: 'text', protocol: 'responses', base: 'https://api.test/v1', key: 'fake-key', model: 'model', size: '' })
    expect(() => selectedConnection(request, key(), ['model'], 'outside', 'https://api.test/v1', 'chat')).toThrow('目录')
    expect(() => selectedConnection(request, key({ status: 'expired' }), ['model'], 'model', 'https://api.test/v1', 'chat')).toThrow('到期')
    const image = { ...request, kind: 'image' as const }
    expect(selectedConnection(image, key(), ['model'], 'model', 'https://api.test/v1', 'chat').protocol).toBe('images')
    expect(() => selectedConnection(image, key({ group: { ...key().group!, allow_image_generation: false } }), ['model'], 'model', 'https://api.test/v1', 'chat')).toThrow('生图权限')
  })
})
