import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

export const brandCopy = {
  zh: {
    headline: '让每一次连接，',
    highlight: '都更从容。',
    introduction: '从 API 密钥到用量记录，把模型服务的接入与管理，放在一个清晰的工作空间。',
    workspace: '进入控制台',
    login: '登录账户',
    home: '品牌首页',
    light: '切换浅色模式',
    dark: '切换深色模式',
    accessTitle: '从接入开始',
    accessDescription: '创建并管理 API 密钥，按需设置使用额度与有效期。',
    accessAction: '管理 API 密钥',
    usageTitle: '让使用可见',
    usageDescription: '查看请求记录、Token 用量与消费明细，了解每一次调用。',
    usageAction: '查看使用记录',
    sectionTitle: '为日常使用，留一份从容',
    sectionNote: '接入 · 管理 · 查看',
    identity: '同一个标志，熟悉的连接',
    footer: '连接模型，也连接你的工作。',
    access: '模型接入',
    control: '清晰管理',
    accountNote: '使用你的账户，继续访问工作空间。'
  },
  en: {
    headline: 'Every connection,',
    highlight: 'a little more effortless.',
    introduction: 'Bring API keys, usage records and model access together in one clear workspace.',
    workspace: 'Open console',
    login: 'Sign in',
    home: 'Brand home',
    light: 'Switch to light mode',
    dark: 'Switch to dark mode',
    accessTitle: 'Start with access',
    accessDescription: 'Create API keys and manage their spending limits and expiration dates.',
    accessAction: 'Manage API keys',
    usageTitle: 'See your usage',
    usageDescription: 'Review requests, token usage and cost details for your model calls.',
    usageAction: 'View usage records',
    sectionTitle: 'A calmer space for everyday work',
    sectionNote: 'Access · Manage · Review',
    identity: 'The same mark. A familiar connection.',
    footer: 'Connecting models to your work.',
    access: 'MODEL ACCESS',
    control: 'CLEAR CONTROL',
    accountNote: 'Sign in with your account to continue to your workspace.'
  }
} as const

export function useBrandCopy() {
  const { locale } = useI18n()
  return computed(() => locale.value.startsWith('zh') ? brandCopy.zh : brandCopy.en)
}
