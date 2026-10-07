import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { reactive, nextTick } from 'vue'
import VersionBadge from '../VersionBadge.vue'

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copied: false, copyToClipboard: vi.fn() })
}))
const mocks = vi.hoisted(() => ({ fetchVersion: vi.fn(), auth: { isAdmin: false } }))
vi.mock('@/stores', () => ({
  useAuthStore: () => mocks.auth,
  useAppStore: () => ({
    currentVersion: '0.2.13-custom.cached', latestVersion: '', hasUpdate: false,
    versionLoading: false, buildType: 'source', fetchVersion: mocks.fetchVersion
  })
}))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key })
}))

describe('[CUSTOM] 版本仅管理员可见', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.auth = reactive({ isAdmin: false })
  })

  it.each(['普通用户', '游客'])('%s 不展示公开旧版本或管理员缓存，也不请求版本接口', () => {
    const wrapper = mount(VersionBadge, { props: { version: '0.2.13-custom.old-public' } })
    expect(wrapper.text()).toBe('')
    expect(wrapper.find('button').exists()).toBe(false)
    expect(mocks.fetchVersion).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('管理员读取受保护接口，退出管理员身份后隐藏已展开详情', async () => {
    mocks.auth.isAdmin = true
    const wrapper = mount(VersionBadge)
    expect(wrapper.text()).toContain('v0.2.13-custom.cached')
    expect(mocks.fetchVersion).toHaveBeenCalledWith(false)
    await wrapper.get('button').trigger('click')
    expect(wrapper.text()).toContain('version.currentVersion')
    mocks.auth.isAdmin = false
    await nextTick()
    expect(wrapper.text()).toBe('')
    expect(wrapper.find('button').exists()).toBe(false)
    wrapper.unmount()
  })
})
