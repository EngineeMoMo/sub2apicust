import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick, reactive } from 'vue'
import { createI18n } from 'vue-i18n'
import WorkspaceHeading from '@/custom/components/WorkspaceHeading.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'

const route = reactive({ name: 'Keys', params: { id: 'example' }, meta: { title: 'Fallback', titleKey: 'keys.title', descriptionKey: 'keys.description' } })
const app = reactive({ cachedPublicSettings: { subscription_enabled: false, custom_menu_items: [] as { id: string; label: string }[] } })
const auth = reactive({ isAdmin: false })
const admin = reactive({ customMenuItems: [{ id: 'example', label: 'Admin page' }] })
vi.mock('vue-router', () => ({ useRoute: () => route }))
vi.mock('@/stores', () => ({ useAppStore: () => app, useAuthStore: () => auth }))
vi.mock('@/stores/adminSettings', () => ({ useAdminSettingsStore: () => admin }))

function mountHeading() {
  const i18n = createI18n({
    legacy: false,
    locale: 'zh',
    messages: {
      zh: { keys: { title: () => 'API 密钥', description: () => '管理密钥' }, nav: { recharge: () => '充值' }, purchase: { rechargeDescription: () => '账户充值' } },
      en: { keys: { title: () => 'API keys', description: () => 'Manage your keys' } }
    }
  })
  const wrapper = mount(WorkspaceHeading, {
    global: { plugins: [i18n] },
    slots: { actions: '<button>原操作</button>' }
  })
  return { wrapper, i18n }
}

beforeEach(() => {
  route.name = 'Keys'
  auth.isAdmin = false
  app.cachedPublicSettings.custom_menu_items = []
})

describe('工作台展示结构', () => {
  it('正文只有一个一级标题，语言切换后标题和描述同步', async () => {
    const { wrapper, i18n } = mountHeading()
    expect(wrapper.findAll('h1')).toHaveLength(1)
    expect(wrapper.get('h1').text()).toBe('API 密钥')
    expect(wrapper.get('button').text()).toBe('原操作')
    i18n.global.locale.value = 'en'
    await nextTick()
    expect(wrapper.get('h1').text()).toBe('API keys')
    expect(wrapper.get('p').text()).toBe('Manage your keys')
    wrapper.unmount()
  })

  it('不向普通用户读取管理员自定义菜单名称', async () => {
    route.name = 'CustomPage'
    const { wrapper } = mountHeading()
    expect(wrapper.get('h1').text()).toBe('API 密钥')
    auth.isAdmin = true
    await nextTick()
    expect(wrapper.get('h1').text()).toBe('Admin page')
    app.cachedPublicSettings.custom_menu_items = [{ id: 'example', label: 'Public page' }]
    await nextTick()
    expect(wrapper.get('h1').text()).toBe('Public page')
    wrapper.unmount()
  })

  it('购买页面仍按上游计费模式解析标题', () => {
    route.name = 'PurchaseSubscription'
    const { wrapper } = mountHeading()
    expect(wrapper.get('h1').text()).toBe('充值')
    expect(wrapper.get('p').text()).toBe('账户充值')
    wrapper.unmount()
  })

  it('筛选、表格、分页在同一面板，保留操作和移动模式', async () => {
    const width = window.innerWidth
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 1440 })
    const wrapper = mount(TablePageLayout, {
      slots: { actions: '<button>创建</button>', filters: '<input aria-label="搜索">', table: '<table><tbody><tr><td>真实插槽</td></tr></tbody></table>', pagination: '<nav>分页</nav>' }
    })
    try {
      expect(wrapper.get('.mofa-data-panel input').attributes('aria-label')).toBe('搜索')
      expect(wrapper.get('.mofa-data-panel table').text()).toBe('真实插槽')
      expect(wrapper.get('.mofa-data-panel nav').text()).toBe('分页')
      expect(wrapper.get('.mofa-table-actions button').text()).toBe('创建')
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: 360 })
      window.dispatchEvent(new Event('resize'))
      await nextTick()
      expect(wrapper.get('.table-page-layout').classes()).toContain('mobile-mode')
    } finally {
      wrapper.unmount()
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: width })
    }
  })
})
