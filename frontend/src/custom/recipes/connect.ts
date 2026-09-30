import type { ApiKey } from '@/types'

export type RecipeConnectionKind = 'text' | 'image'
export interface RecipeConnectionRequest {
  origin: string
  nonce: string
  kind: RecipeConnectionKind
}
export interface RecipeConnectionConfig {
  kind: RecipeConnectionKind
  protocol: 'chat' | 'responses' | 'images'
  base: string
  key: string
  model: string
  size: string
}

function secureURL(value: string): URL {
  const url = new URL(value)
  const local = ['localhost', '127.0.0.1', '[::1]'].includes(url.hostname)
  if (url.protocol !== 'https:' && !(url.protocol === 'http:' && local)) throw new Error('连接地址需要 HTTPS，本机地址可以使用 HTTP。')
  if (url.username || url.password || url.search || url.hash || url.hostname.includes('*')) throw new Error('连接地址格式有误。')
  return url
}

export function connectionRequest(query: Record<string, unknown>, ownOrigin: string, configuredOrigin = ''): RecipeConnectionRequest {
  const origin = typeof query.origin === 'string' ? query.origin : ''
  const nonce = typeof query.nonce === 'string' ? query.nonce : ''
  if (!/^[a-f0-9]{64}$/.test(nonce) || typeof query.kind !== 'string' || !['text', 'image'].includes(query.kind)) throw new Error('连接请求无效，请从魔法配方的模型设置重新发起。')
  const target = secureURL(origin)
  if (target.origin !== origin) throw new Error('配方页面来源格式有误。')
  const allowed = new Set([ownOrigin])
  if (['localhost', '127.0.0.1', '[::1]'].includes(new URL(ownOrigin).hostname)) {
    for (const local of ['http://127.0.0.1:4178', 'http://localhost:4178', 'http://[::1]:4178']) allowed.add(local)
  }
  if (configuredOrigin) {
    const configured = secureURL(configuredOrigin)
    if (configured.origin !== configuredOrigin) throw new Error('配方来源配置需要完整 origin，不含路径。')
    allowed.add(configuredOrigin)
  }
  if (!allowed.has(origin)) throw new Error('此配方页面来源尚未获准接入，请联系站点管理员配置。')
  return { origin, nonce, kind: query.kind as RecipeConnectionKind }
}

export function usableKey(key: ApiKey, now = Date.now()): boolean {
  if (key.status !== 'active' || !key.key?.trim() || !key.group_id || !key.group || key.group.status !== 'active') return false
  if (key.expires_at && (!Number.isFinite(Date.parse(key.expires_at)) || Date.parse(key.expires_at) <= now)) return false
  return !(key.quota > 0 && key.quota_used >= key.quota)
}

export function gatewayBase(value: string, fallback: string): string {
  const url = secureURL(value.trim() || fallback)
  const path = url.pathname.replace(/\/+$/, '')
  url.pathname = path || '/v1'
  return url.href.replace(/\/+$/, '')
}

export function catalogModels(payload: unknown): string[] {
  const source = payload as { data?: unknown }
  if (!Array.isArray(source?.data)) throw new Error('模型目录格式不受支持，请刷新或检查接口。')
  const models = [...new Set(source.data.flatMap(item => {
    const id = item && typeof item === 'object' ? (item as { id?: unknown }).id : undefined
    return typeof id === 'string' && id.trim() && id.length <= 200 ? [id] : []
  }))]
  return models.sort((left, right) => left.localeCompare(right))
}

export async function loadModels(base: string, key: string, signal: AbortSignal): Promise<string[]> {
  let response: Response
  try {
    response = await fetch(base + '/models', {
      headers: { Authorization: 'Bearer ' + key }, signal,
      credentials: 'omit', redirect: 'error', referrerPolicy: 'no-referrer', cache: 'no-store'
    })
  } catch (error) {
    if (signal.aborted) throw error
    throw new Error('模型目录连接失败，请检查公开 API 地址、网络和跨域配置。')
  }
  if (!response.ok) throw new Error('无法读取此密钥的模型目录（HTTP ' + response.status + '），请检查密钥、分组和额度。')
  try { return catalogModels(await response.json()) }
  catch { throw new Error('模型目录格式不受支持，请刷新或检查接口。') }
}

export function selectedConnection(request: RecipeConnectionRequest, key: ApiKey, models: string[], model: string, base: string, protocol: 'chat' | 'responses'): RecipeConnectionConfig {
  if (!usableKey(key)) throw new Error('密钥已停用、到期或额度耗尽，请刷新密钥列表。')
  if (!models.includes(model)) throw new Error('请选择所选密钥目录中的模型。')
  if (!['chat', 'responses'].includes(protocol)) throw new Error('请选择正确的文字接口格式。')
  if (request.kind === 'image' && key.group?.allow_image_generation === false) throw new Error('此分组未开启生图权限，请选择其他密钥。')
  return { kind: request.kind, protocol: request.kind === 'image' ? 'images' : protocol, base: gatewayBase(base, ''), key: key.key, model, size: '' }
}
