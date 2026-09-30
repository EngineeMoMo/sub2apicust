import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'

type NavigationGuard = (
  to: Record<string, any>,
  from: Record<string, any>,
  next: ReturnType<typeof vi.fn>
) => Promise<void>

const routerHarness = vi.hoisted(() => ({
  guard: null as NavigationGuard | null,
  scroll: null as null | ((to: { path: string; hash: string }, from: unknown, savedPosition: unknown) => unknown),
}))

const authStore = vi.hoisted(() => ({
  checkAuth: vi.fn(),
  isAuthenticated: true,
  isAdmin: false,
  isSimpleMode: false,
  hasPendingAuthSession: false,
}))

const appStore = vi.hoisted(() => ({
  siteName: 'Sub2API',
  backendModeEnabled: false,
  publicSettingsLoaded: false,
  cachedPublicSettings: null as null | {
    payment_enabled?: boolean
    risk_control_enabled?: boolean
    subscription_enabled?: boolean
    custom_menu_items?: []
  },
  fetchPublicSettings: vi.fn(),
}))

vi.mock('vue-router', () => ({
  createWebHistory: vi.fn(() => ({})),
  createRouter: vi.fn(options => {
    routerHarness.scroll = options.scrollBehavior
    return {
      beforeEach: vi.fn((guard: NavigationGuard) => {
        routerHarness.guard = guard
      }),
      afterEach: vi.fn(),
      onError: vi.fn(),
    }
  }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStore,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStore,
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => ({ customMenuItems: [] }),
}))

vi.mock('@/stores/adminCompliance', () => ({
  useAdminComplianceStore: () => ({
    initialized: true,
    fetchStatus: vi.fn(),
    requireAcknowledgement: vi.fn(),
  }),
}))

vi.mock('@/composables/useNavigationLoading', () => ({
  useNavigationLoadingState: () => ({
    startNavigation: vi.fn(),
    endNavigation: vi.fn(),
    isLoading: { value: false },
  }),
}))

vi.mock('@/composables/useRoutePrefetch', () => ({
  useRoutePrefetch: () => ({
    triggerPrefetch: vi.fn(),
    cancelPendingPrefetch: vi.fn(),
    resetPrefetchState: vi.fn(),
  }),
}))

function createDeferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise
  })
  return { promise, resolve }
}

function runGuard(meta: Record<string, unknown>, path: string) {
  if (!routerHarness.guard) {
    throw new Error('router guard was not registered')
  }

  const next = vi.fn()
  const navigation = routerHarness.guard(
    {
      path,
      fullPath: path,
      name: 'FeatureRoute',
      params: {},
      meta: { requiresAuth: true, ...meta },
    },
    {},
    next
  )
  return { navigation, next }
}

describe('feature route guard', () => {
  beforeAll(async () => {
    await import('@/router')
  })

  it.each(['/guide', '/help/guide', '/preview/guide'])('%s 的教程目录滚动到章节', path => {
    expect(routerHarness.scroll?.({ path, hash: '#guide-codex' }, {}, null)).toEqual({ el: '#guide-codex' })
  })

  it('教程外页面及未知锚点保留置顶，浏览器返回优先恢复原位置', () => {
    expect(routerHarness.scroll?.({ path: '/faq', hash: '#guide-codex' }, {}, null)).toEqual({ top: 0 })
    expect(routerHarness.scroll?.({ path: '/guide', hash: '#unknown' }, {}, null)).toEqual({ top: 0 })
    expect(routerHarness.scroll?.({ path: '/guide', hash: '#guide-codex' }, {}, { top: 120 })).toEqual({ top: 120 })
  })

  beforeEach(() => {
    appStore.backendModeEnabled = false
    authStore.isAuthenticated = true
    authStore.isAdmin = false
    authStore.isSimpleMode = false
    appStore.publicSettingsLoaded = false
    appStore.cachedPublicSettings = null
    appStore.fetchPublicSettings.mockReset()
  })

  // [CUSTOM] 使用实际注册的路由守卫验证游客入口。
  it.each(['/home', '/brand', '/family', '/api', '/plans', '/faq', '/guide', '/preview', '/preview/keys', '/preview/plans', '/preview/faq', '/preview/guide'])('匿名可访问 %s', async path => {
    authStore.isAuthenticated = false
    appStore.publicSettingsLoaded = true
    const { navigation, next } = runGuard({ requiresAuth: false }, path)
    await navigation
    expect(next).toHaveBeenCalledWith()
  })

  it.each(['/home', '/brand', '/family', '/api', '/plans', '/faq', '/guide', '/preview', '/preview/keys', '/preview/plans', '/preview/faq', '/preview/guide'])('冷启动确认后台模式后拦截 %s', async path => {
    authStore.isAuthenticated = false
    appStore.fetchPublicSettings.mockImplementation(async () => {
      appStore.backendModeEnabled = true
      appStore.publicSettingsLoaded = true
    })
    const { navigation, next } = runGuard({ requiresAuth: false }, path)
    await navigation
    expect(next).toHaveBeenCalledWith('/login')
  })

  it.each(['/keys', '/usage', '/orders', '/purchase'])('个人页 %s 仍要求登录', async path => {
    authStore.isAuthenticated = false
    const { navigation, next } = runGuard({}, path)
    await navigation
    expect(next).toHaveBeenCalledWith({ path: '/login', query: { redirect: path } })
  })

  it.each(['/family', '/api'])('后台模式中已登录非管理员不能访问 %s', async path => {
    appStore.backendModeEnabled = true
    appStore.publicSettingsLoaded = true
    const { navigation, next } = runGuard({ requiresAuth: false }, path)
    await navigation
    expect(next).toHaveBeenCalledWith('/login')
  })

  it('有效站内会话可直接进入配方选择页，未登录时保留完整回跳目标', async () => {
    const path = '/connect/recipes?origin=https%3A%2F%2Frecipes.test&nonce=abc&kind=text'
    appStore.publicSettingsLoaded = true
    let attempt = runGuard({}, path)
    await attempt.navigation
    expect(attempt.next).toHaveBeenCalledWith()
    authStore.isAuthenticated = false
    attempt = runGuard({}, path)
    await attempt.navigation
    expect(attempt.next).toHaveBeenCalledWith({ path: '/login', query: { redirect: path } })
  })

  it('waits for the first public-settings request before deciding payment access', async () => {
    const deferred = createDeferred<{ payment_enabled: boolean }>()
    appStore.fetchPublicSettings.mockImplementation(async () => {
      const settings = await deferred.promise
      appStore.cachedPublicSettings = settings
      appStore.publicSettingsLoaded = true
      return settings
    })

    const { navigation, next } = runGuard({ requiresPayment: true }, '/purchase')

    await vi.waitFor(() => expect(appStore.fetchPublicSettings).toHaveBeenCalledTimes(1))
    expect(next).not.toHaveBeenCalled()

    deferred.resolve({ payment_enabled: true })
    await navigation
    expect(next).toHaveBeenCalledOnce()
    expect(next).toHaveBeenCalledWith()
  })

  it.each([
    ['payment', { requiresPayment: true }, '/purchase'],
    ['risk control', { requiresRiskControl: true }, '/admin/risk-control'],
    ['subscription', { requiresSubscription: true }, '/subscriptions'],
  ])('does not treat a failed %s settings load as explicitly disabled', async (_name, meta, path) => {
    authStore.isAdmin = meta.requiresRiskControl === true
    appStore.fetchPublicSettings.mockResolvedValue(null)

    const { navigation, next } = runGuard(meta, path)
    await navigation

    expect(appStore.publicSettingsLoaded).toBe(false)
    expect(next).toHaveBeenCalledOnce()
    expect(next).toHaveBeenCalledWith()
  })

  it.each([
    ['payment', { requiresPayment: true }, { payment_enabled: false }, '/dashboard'],
    [
      'risk control',
      { requiresRiskControl: true },
      { risk_control_enabled: false },
      '/admin/settings',
    ],
    ['subscription', { requiresSubscription: true }, { subscription_enabled: false }, '/dashboard'],
  ])('redirects when loaded settings explicitly disable %s', async (_name, meta, settings, target) => {
    authStore.isAdmin = meta.requiresRiskControl === true
    appStore.cachedPublicSettings = settings
    appStore.publicSettingsLoaded = true

    const { navigation, next } = runGuard(meta, '/feature')
    await navigation

    expect(appStore.fetchPublicSettings).not.toHaveBeenCalled()
    expect(next).toHaveBeenCalledOnce()
    expect(next).toHaveBeenCalledWith(target)
  })
})

describe('subscription route guard (opt-out flag)', () => {
  beforeEach(() => {
    authStore.isAdmin = false
    authStore.isSimpleMode = false
    appStore.publicSettingsLoaded = true
    appStore.fetchPublicSettings.mockReset()
  })

  it.each([
    ['missing key', {}],
    ['explicit true', { subscription_enabled: true }],
  ])('lets /subscriptions through when the flag is %s', async (_name, settings) => {
    appStore.cachedPublicSettings = settings

    const { navigation, next } = runGuard({ requiresSubscription: true }, '/subscriptions')
    await navigation

    expect(next).toHaveBeenCalledOnce()
    expect(next).toHaveBeenCalledWith()
  })

  it('sends admins to the admin dashboard when subscriptions are disabled', async () => {
    authStore.isAdmin = true
    appStore.cachedPublicSettings = { subscription_enabled: false }

    const { navigation, next } = runGuard({ requiresSubscription: true }, '/subscriptions')
    await navigation

    expect(next).toHaveBeenCalledWith('/admin/dashboard')
  })
})
