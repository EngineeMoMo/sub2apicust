export interface FamilyProduct {
  id: string
  name: string
  purpose: string
  description: string
  state: string
  account: string
  destination?: string
  external?: boolean
  action: string
}

export function recipeDestination(value: string, origin: string, development: boolean): { href?: string; preview: boolean; invalid: boolean } {
  const site = new URL(origin)
  const local = ['localhost', '127.0.0.1', '[::1]'].includes(site.hostname)
  const preview = !value.trim() && development && local
  const configured = value.trim() || (preview ? 'http://127.0.0.1:4178' : '')
  if (!configured) return { preview: false, invalid: false }
  try {
    const url = new URL(configured, origin)
    const localTarget = ['localhost', '127.0.0.1', '[::1]'].includes(url.hostname)
    const localPreview = local && localTarget && site.protocol === 'http:' && url.protocol === 'http:'
    if (url.protocol !== 'https:' && !localPreview) throw new Error('不安全地址')
    if (url.username || url.password || url.search || url.hash || configured.startsWith('//') || url.hostname.includes('*')) throw new Error('地址不能含凭据或参数')
    if (localPreview || (development && local && localTarget)) url.searchParams.set('api_site', origin)
    return { href: url.href, preview: preview || localPreview, invalid: false }
  } catch { return { preview: false, invalid: true } }
}

export function studioDestination(value: string, origin: string, development: boolean): ReturnType<typeof recipeDestination> {
  const configured = value.trim() || (development && ['localhost', '127.0.0.1', '[::1]'].includes(new URL(origin).hostname) ? 'http://127.0.0.1:4179' : '')
  const destination = recipeDestination(configured, origin, false)
  if (destination.href) {
    const url = new URL(destination.href)
    url.searchParams.delete('api_site')
    destination.href = url.href
  }
  return destination
}

export function familyProducts(recipe: ReturnType<typeof recipeDestination>, admin: boolean, studio: ReturnType<typeof recipeDestination> = { preview: false, invalid: false }): FamilyProduct[] {
  return [
    { id: 'api', name: '魔法 API', purpose: '模型接入与账户管理', description: '创建密钥、查看用量，管理余额和订阅。给你的 AI 工具提供模型服务。', state: '本站服务', account: '共用当前登录', destination: admin ? '/admin/dashboard' : '/dashboard', action: '进入控制台' },
    { id: 'recipes', name: '魔法配方', purpose: '可复用的 AI 工作方法', description: '工作、大学与创作配方。填材料整理提示词，接入自己的文字或生图模型后运行。', state: recipe.preview ? '本机预览' : recipe.href ? '独立产品' : recipe.invalid ? '地址配置有误' : '待配置发布地址', account: '模型连接复用魔法 API 登录，首次需授权密钥', destination: recipe.href, external: true, action: '打开魔法配方' },
    { id: 'studio', name: '魔法工坊', purpose: '图片、视频与工作流', description: '浏览视觉示例与精选提示词，探索视频运镜、Skills 和工作流指引，再带到自己的创作工具中使用。', state: studio.preview ? '本机预览' : studio.href ? '独立产品' : studio.invalid ? '地址配置有误' : '待配置发布地址', account: '内容首版已完成；不直接运行模型或自动安装 Skills', destination: studio.href, external: true, action: '打开魔法工坊' },
    { id: 'synaroute', name: 'SynaRoute', purpose: '本地 API 路由与多模型协同', description: '桌面工具保持独立。先在魔法 API 选择密钥，再使用已有的 SynaRoute 导入功能。', state: '桌面工具', account: '复用 API 配置，不等于已接通账号单点登录', destination: '/keys', action: '配置 SynaRoute' }
  ]
}

export function familyLoginTarget(product: FamilyProduct): { path: string; query: { redirect: string } } {
  return { path: '/login', query: { redirect: product.id === 'api' ? '/dashboard' : '/dashboard?product=' + encodeURIComponent(product.id) } }
}
