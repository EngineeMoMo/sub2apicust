export interface PreviewSection {
  id: string
  label: string
  icon: 'grid' | 'key' | 'chart' | 'clock' | 'creditCard' | 'document' | 'book'
  description: string
  target: string
  action: string
  columns: string[]
}

export const previewSections: PreviewSection[] = [
  { id: 'overview', label: '控制台概览', icon: 'grid', description: '先熟悉工作空间，再决定如何接入模型服务。', target: '/dashboard', action: '进入我的控制台', columns: [] },
  { id: 'plans', label: '订阅套餐', icon: 'creditCard', description: '比较当前在售套餐的价格、有效期与权益。', target: '/purchase?tab=subscription', action: '购买订阅', columns: [] },
  { id: 'keys', label: 'API 密钥', icon: 'key', description: '创建和管理调用凭证，将模型服务接入你信任的工具。', target: '/keys', action: '创建 API Key', columns: ['密钥名称', '密钥', '状态', '操作'] },
  { id: 'usage', label: '用量记录', icon: 'chart', description: '了解使用记录的查看方式；真实调用与费用仅向账户本人展示。', target: '/usage', action: '查看我的用量', columns: ['调用时间', '模型', '用量', '费用'] },
  { id: 'subscriptions', label: '我的订阅', icon: 'clock', description: '查看自己购买的套餐、有效期与已用额度。', target: '/subscriptions', action: '查看我的订阅', columns: ['订阅', '状态', '有效期', '已用额度'] },
  { id: 'balance', label: '余额与充值', icon: 'creditCard', description: '充值关联你的账户，确认金额和支付方式后再下单。', target: '/purchase', action: '账户充值', columns: ['账户余额', '充值金额', '支付方式'] },
  { id: 'orders', label: '我的订单', icon: 'document', description: '核对自己的充值与订阅订单，查看支付及处理状态。', target: '/orders', action: '查看我的订单', columns: ['订单号', '类型', '金额', '状态'] },
  { id: 'guide', label: '接入教程', icon: 'book', description: '了解 Claude Code、Codex 和常用客户端的接入步骤。', target: '/guide', action: '阅读接入教程', columns: [] },
  { id: 'faq', label: '常见问题', icon: 'book', description: '查看入门、充值、接入与排错说明。', target: '/faq', action: '阅读常见问题', columns: [] }
]

export function previewPath(section: string): string {
  return section === 'overview' ? '/preview' : `/preview/${section}`
}

export function isPreviewPath(path: string): boolean {
  return path === '/preview' || path.startsWith('/preview/')
}
