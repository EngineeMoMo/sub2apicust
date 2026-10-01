import { afterEach, describe, expect, it, vi } from 'vitest'
import type { ApiKey } from '@/types'
import { acceptsWorkspaceMessage, WorkspaceConnections } from '@/custom/family/workspace'
import { customRoutes } from '@/custom/routes'

const key = (overrides: Partial<ApiKey> = {}) => ({ id: 1, name: '模拟密钥', key: 'fake-test-key', status: 'active', group_id: 1, quota: 0, quota_used: 0,
  group: { id: 1, name: '模拟分组', status: 'active', platform: 'openai', allow_image_generation: true }, ...overrides } as ApiKey)
function broker(entries = [key()]) {
  const list = vi.fn(async () => ({ items: entries, pages: 1 }))
  const models = vi.fn(async () => ['example-model'])
  return { list, models, connection: new WorkspaceConnections({ list, models, base: () => 'https://gateway.test/v1' }) }
}
afterEach(() => vi.restoreAllMocks())
describe('控制台子产品与配置授权', () => {
  it('两条工作区路由受登录保护，静态模块路径不暴露凭据', () => {
    for (const path of ['/tools/recipes', '/tools/studio']) expect(customRoutes.find(route => route.path === path)?.meta?.requiresAuth).toBe(true)
  })
  it('目录只返回名称与分组，不返回全部密钥原文，不自动读模型或生成', async () => {
    const { connection, models } = broker([key(), key({ id: 2, key: 'other-secret', status: 'inactive' })])
    const result = await connection.execute('keys', { kind: 'text' })
    expect(result).toEqual([{ id: 1, name: '模拟密钥', group: '模拟分组', protocol: 'responses' }])
    expect(JSON.stringify(result)).not.toMatch(/fake-test-key|other-secret/)
    expect(models).not.toHaveBeenCalled()
  })
  it('明确选择密钥及目录模型后才返回单个连接，再次校验密钥状态', async () => {
    const { connection, list, models } = broker()
    await connection.execute('keys', { kind: 'text' })
    expect(await connection.execute('models', { kind: 'text', keyID: 1 })).toEqual({ models: ['example-model'], protocol: 'responses' })
    expect(models).toHaveBeenCalledWith('https://gateway.test/v1', 'fake-test-key', expect.any(AbortSignal))
    const result = await connection.execute('apply', { kind: 'text', keyID: 1, model: 'example-model', protocol: 'responses' })
    expect(result).toMatchObject({ kind: 'text', key: 'fake-test-key', model: 'example-model', protocol: 'responses' })
    expect(list).toHaveBeenCalledTimes(2)
  })
  it('目录外模型、未读目录、用途不符与未知操作都拒绝', async () => {
    const { connection } = broker()
    await connection.execute('keys', { kind: 'text' })
    await expect(connection.execute('apply', { kind: 'text', keyID: 1, model: 'guess', protocol: 'chat' })).rejects.toThrow('目录已过期')
    await connection.execute('models', { kind: 'text', keyID: 1 })
    await expect(connection.execute('apply', { kind: 'text', keyID: 1, model: 'guess', protocol: 'chat' })).rejects.toThrow('目录中的模型')
    await expect(connection.execute('models', { kind: 'video', keyID: 1 })).rejects.toThrow('文字或生图')
    await expect(connection.execute('fetch', { kind: 'text' })).rejects.toThrow('不支持')
  })
  it('应用时重新验证停用及换钥，销毁后不能使用先前目录', async () => {
    const entries = [key()], { connection } = broker(entries)
    await connection.execute('keys', { kind: 'text' }); await connection.execute('models', { kind: 'text', keyID: 1 })
    entries[0].key = 'rotated-fake-key'
    await expect(connection.execute('apply', { kind: 'text', keyID: 1, model: 'example-model', protocol: 'chat' })).rejects.toThrow('目录已过期')
    entries[0].status = 'inactive'
    await expect(connection.execute('apply', { kind: 'text', keyID: 1, model: 'example-model', protocol: 'chat' })).rejects.toThrow('密钥不可用')
    connection.clear()
    await expect(connection.execute('models', { kind: 'text', keyID: 1 })).rejects.toThrow('密钥不可用')
  })
  it('生图筛选权限，文字目录不能用于生图授权', async () => {
    const entries = [key()], { connection } = broker(entries)
    await connection.execute('keys', { kind: 'text' }); await connection.execute('models', { kind: 'text', keyID: 1 })
    await expect(connection.execute('apply', { kind: 'image', keyID: 1, model: 'example-model', protocol: 'chat' })).rejects.toThrow('目录已过期')
    if (entries[0].group) entries[0].group.allow_image_generation = false
    await expect(connection.execute('keys', { kind: 'image' })).rejects.toThrow('所属分组均未开启生图权限')
  })

  it('生图目录不掩盖分组权限原因，也不误报空账号或混合权限', async () => {
    const entries = [key(), key({ id: 2, group: { ...key().group!, allow_image_generation: false } })]
    const { connection } = broker(entries)
    expect(await connection.execute('keys', { kind: 'image' })).toEqual([{ id: 1, name: '模拟密钥', group: '模拟分组', protocol: 'responses' }])
    expect(await broker([]).connection.execute('keys', { kind: 'image' })).toEqual([])
  })
  it('通道必须同时匹配来源、iframe、nonce及请求格式', () => {
    const frame = {} as Window, origin = location.origin, nonce = 'a'.repeat(64)
    const valid = { source: frame, origin, data: { type: 'mofa-host-request', nonce, id: '1', action: 'keys', payload: { kind: 'text' } } } as unknown as MessageEvent
    expect(acceptsWorkspaceMessage(valid, frame, origin, nonce)).toBe(true)
    for (const bad of [{ origin: 'https://evil.test' }, { source: {} }, { data: { ...valid.data, nonce: 'wrong' } }, { data: { ...valid.data, payload: [] } }, { data: { ...valid.data, id: '<script>' } }]) {
      expect(acceptsWorkspaceMessage({ ...valid, ...bad } as MessageEvent, frame, origin, nonce)).toBe(false)
    }
  })
})
