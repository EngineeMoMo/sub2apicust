import type { ApiKey } from '@/types'
import { gatewayBase, selectedConnection, usableKey, type RecipeConnectionKind } from '@/custom/recipes/connect'

interface Dependencies {
  list: (page: number, signal: AbortSignal) => Promise<{ items: ApiKey[]; pages: number }>
  models: (base: string, key: string, signal: AbortSignal) => Promise<string[]>
  base: () => string
}

export class WorkspaceConnections {
  private keys: ApiKey[] = []
  private catalogs = new Map<string, { models: string[]; time: number; key: string }>()
  private operation?: AbortController
  constructor(private dependencies: Dependencies) {}
  clear() { this.operation?.abort(); this.keys = []; this.catalogs.clear() }
  async execute(action: string, payload: Record<string, unknown>) {
    if (!['keys', 'models', 'apply'].includes(action)) throw new Error('不支持此配置操作。')
    const kind = payload.kind
    if (kind !== 'text' && kind !== 'image') throw new Error('请选择文字或生图连接。')
    this.operation?.abort()
    const operation = new AbortController()
    this.operation = operation
    const timer = setTimeout(() => operation.abort(), 30000)
    try {
      if (action === 'keys' || action === 'apply') {
        const fresh: ApiKey[] = []
        for (let page = 1; page <= 100; page++) {
          const result = await this.dependencies.list(page, operation.signal)
          if (operation.signal.aborted) throw new Error('配置读取已取消。')
          fresh.push(...result.items)
          if (page >= result.pages || !result.items.length) break
          if (page === 100) throw new Error('密钥过多，请先在密钥管理中整理。')
        }
        this.keys = fresh.filter(key => usableKey(key))
      }
      const allowed = this.keys.filter(key => !(kind === 'image' && key.group?.allow_image_generation === false))
      if (action === 'keys') {
        this.catalogs.clear()
        return allowed.map(key => ({ id: key.id, name: key.name, group: key.group?.name, protocol: key.group?.platform === 'openai' ? 'responses' : 'chat' }))
      }
      const key = allowed.find(entry => entry.id === payload.keyID)
      if (!key || !usableKey(key)) throw new Error('密钥不可用，请刷新或创建密钥。')
      const base = gatewayBase(this.dependencies.base(), '')
      const catalogID = kind + ':' + key.id + ':' + base
      if (action === 'models') {
        const models = await this.dependencies.models(base, key.key, operation.signal)
        if (operation.signal.aborted) throw new Error('模型目录读取已取消。')
        this.catalogs.set(catalogID, { models, time: Date.now(), key: key.key })
        return { models, protocol: key.group?.platform === 'openai' ? 'responses' : 'chat' }
      }
      const catalog = this.catalogs.get(catalogID)
      if (!catalog || catalog.key !== key.key || Date.now() - catalog.time > 120000) throw new Error('模型目录已过期，请重新选择密钥。')
      if (typeof payload.model !== 'string' || (payload.protocol !== 'chat' && payload.protocol !== 'responses')) throw new Error('请选择模型及接口格式。')
      return selectedConnection({ origin: location.origin, nonce: '', kind: kind as RecipeConnectionKind }, key, catalog.models, payload.model, base, payload.protocol)
    } finally { clearTimeout(timer) }
  }
}

export function acceptsWorkspaceMessage(event: MessageEvent, frame: Window | null | undefined, origin: string, nonce: string): boolean {
  const data = event.data
  return Boolean(frame && nonce && event.source === frame && event.origin === origin && data?.type === 'mofa-host-request'
    && data.nonce === nonce && typeof data.id === 'string' && /^\d{1,12}$/.test(data.id)
    && typeof data.action === 'string' && data.payload && typeof data.payload === 'object' && !Array.isArray(data.payload))
}
