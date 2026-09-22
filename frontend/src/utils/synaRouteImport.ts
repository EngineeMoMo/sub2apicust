// [CUSTOM] SynaRoute 深链接一键导入（synaroute://），对标上游已有的 ccswitch:// 导入
// （见 utils/ccswitchImport.ts），形态刻意对齐。规格见
// docs「深链接一键导入」。本文件是纯函数：由 (platform, baseUrl, apiKey) 推导导入参数并构造
// synaroute:// URL，不触碰 DOM，便于单测。登记见项目根 CUSTOMIZATIONS.md。

import type { GroupPlatform } from '@/types'

export const GROK_SYNAROUTE_MODEL = 'grok-4.5'

// SynaRoute 支持的协议（确认框里的协议下拉）
export type SynaRouteProtocol = 'anthropic' | 'openai_chat' | 'openai_responses'
// SynaRoute 支持的分类
export type SynaRouteCategory = 'claude-cli' | 'claude-desktop' | 'codex'

export interface SynaRouteModels {
  opusModel?: string
  sonnetModel?: string
  haikuModel?: string
  fableModel?: string
}

export interface SynaRouteImportParams extends SynaRouteModels {
  endpoint: string
  apiKey?: string
  category?: SynaRouteCategory
  protocol?: SynaRouteProtocol
  name?: string
}

/**
 * UTF-8 安全的 base64 编码。API Key 通常是 ASCII（sk-...），但用 TextEncoder 兜住非 ASCII，
 * 避免 btoa 在 Latin1 范围外抛异常。
 */
function toBase64(input: string): string {
  const bytes = new TextEncoder().encode(input)
  let binary = ''
  for (const byte of bytes) binary += String.fromCharCode(byte)
  return btoa(binary)
}

/**
 * URL-safe base64（+/= -> -_ 去 padding）。规格推荐 URL-safe base64，
 * SynaRoute 解码端对标准/URL-safe、缺 padding 都兼容——用 URL-safe 最稳，URL 里不出现 +/=。
 */
export function toUrlSafeBase64(input: string): string {
  return toBase64(input).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

/**
 * 构造 synaroute://v1/import URL。所有值经 URLSearchParams 自动 URL 转义（端点含 ://、
 * base64 密钥可能含特殊字符）。host 固定 v1、path 固定 /import。
 */
export function buildSynaRouteUrl(params: SynaRouteImportParams): string {
  const search = new URLSearchParams()
  search.set('resource', 'provider')
  search.set('endpoint', params.endpoint)
  if (params.apiKey) search.set('apiKey', toUrlSafeBase64(params.apiKey))
  if (params.category) search.set('category', params.category)
  if (params.protocol) search.set('protocol', params.protocol)
  if (params.name) search.set('name', params.name)
  if (params.opusModel) search.set('opusModel', params.opusModel)
  if (params.sonnetModel) search.set('sonnetModel', params.sonnetModel)
  if (params.haikuModel) search.set('haikuModel', params.haikuModel)
  if (params.fableModel) search.set('fableModel', params.fableModel)
  return `synaroute://v1/import?${search.toString()}`
}

/** 去掉末尾的 /v1 与多余斜杠，得到 Anthropic/Claude 客户端要的 base（客户端自己接 /v1/messages）。 */
function toBaseRoot(baseUrl: string): string {
  return baseUrl.replace(/\/v1\/?$/, '').replace(/\/+$/, '')
}

/** 确保末尾是 /v1，得到 Codex/OpenAI Responses 客户端要的 base。 */
function ensureV1(baseUrl: string): string {
  const trimmed = toBaseRoot(baseUrl)
  return trimmed.endsWith('/v1') ? trimmed : `${trimmed}/v1`
}

export interface SynaRouteImportInput {
  platform?: GroupPlatform | null
  baseUrl: string
  apiKey: string
  name?: string
}

/**
 * 由密钥所属分组的 platform 推导 SynaRoute 导入参数（对齐 ccswitchImport 的 resolve*）。
 * SynaRoute 分类只有 claude-cli / codex，协议只有 anthropic / openai_responses：
 *  - openai              → codex / openai_responses，端点带 /v1
 *  - antigravity         → claude-cli / anthropic，端点为 {root}/antigravity
 *  - grok                → claude-cli / anthropic，附带 grok-4.5 三档 + Fable 映射
 *  - 其余（含默认 anthropic）→ claude-cli / anthropic，端点为去 /v1 的 root
 *  - gemini              → 无对应协议：只带端点/密钥/名字，分类与协议留空，由 SynaRoute 确认框让用户选
 */
export function buildSynaRouteImportDeeplink(input: SynaRouteImportInput): string {
  const platform: GroupPlatform = input.platform || 'anthropic'
  const baseRoot = toBaseRoot(input.baseUrl)
  const base = { apiKey: input.apiKey, name: input.name }

  let params: SynaRouteImportParams
  switch (platform) {
    case 'openai':
      params = { ...base, endpoint: ensureV1(input.baseUrl), category: 'codex', protocol: 'openai_responses' }
      break
    case 'antigravity':
      params = { ...base, endpoint: `${baseRoot}/antigravity`, category: 'claude-cli', protocol: 'anthropic' }
      break
    case 'grok':
      params = {
        ...base,
        endpoint: baseRoot,
        category: 'claude-cli',
        protocol: 'anthropic',
        opusModel: GROK_SYNAROUTE_MODEL,
        sonnetModel: GROK_SYNAROUTE_MODEL,
        haikuModel: GROK_SYNAROUTE_MODEL,
        fableModel: GROK_SYNAROUTE_MODEL
      }
      break
    case 'gemini':
      // SynaRoute 没有 gemini 协议/分类：留空让确认框里的下拉由用户选，端点仍带过去。
      params = { ...base, endpoint: baseRoot }
      break
    default:
      params = { ...base, endpoint: baseRoot, category: 'claude-cli', protocol: 'anthropic' }
  }

  return buildSynaRouteUrl(params)
}
