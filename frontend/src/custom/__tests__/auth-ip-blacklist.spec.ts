import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import AuthIPBlacklistView from '@/custom/views/AuthIPBlacklistView.vue'
import { authIPBlacklistAPI, formatAuthIPTime, type AuthIPEntry } from '@/custom/authIpBlacklist'
import { customRoutes } from '@/custom/routes'

const api = vi.hoisted(() => ({ get: vi.fn(), put: vi.fn(), post: vi.fn() }))
vi.mock('@/api/client', () => ({ apiClient: api }))
const config = { enabled: false, threshold: 10, window_seconds: 7200, block_seconds: 900, block_mode: 'timed', version: 1 }
const entry: AuthIPEntry = {
  id: 8, ip: '61.120.9.25', generation: 3, blocked: true,
  blocked_at: '2026-10-13T12:32:32Z', trigger_path: '/api/v1/auth/login', trigger_count: 10,
  window_seconds: 7200, first_event_at: '2026-10-13T11:12:39Z', last_event_at: '2026-10-13T12:32:32Z', rule_version: 2
}
function render() {
  return mount(AuthIPBlacklistView, { global: { stubs: {
    AppLayout: { template: '<main><slot /></main>' },
    ConfirmDialog: { props: ['show', 'message'], emits: ['confirm', 'cancel'], template: '<div v-if="show"><p>{{ message }}</p><button data-testid="confirm" @click="$emit(\'confirm\')">确认移除</button></div>' }
  } } })
}
beforeEach(() => {
  vi.resetAllMocks()
  api.get.mockImplementation(async (url: string) => ({ data: url.endsWith('/config') ? config : { items: [entry], total: 1, page: 1, page_size: 20, pages: 1 } }))
  api.put.mockResolvedValue({ data: { ...config, version: 2 } })
  api.post.mockResolvedValue({ data: { ...entry, blocked: false, generation: 4 } })
})

describe('IP 黑名单管理', () => {
  it('保存以分钟输入、以秒提交，并附带配置版本', async () => {
    const wrapper = render()
    await flushPromises()
    expect(wrapper.get('[data-testid="minutes"]').element).toHaveProperty('value', '120')
    await wrapper.get('[data-testid="minutes"]').setValue('180')
    await wrapper.get('[data-testid="enabled"]').setValue(true)
    await wrapper.findAll('form')[0].trigger('submit')
    await flushPromises()
    expect(api.put).toHaveBeenCalledWith('/admin/ip-blacklist/config', { enabled: true, threshold: 10, window_seconds: 10800, block_seconds: 900, block_mode: 'timed', version: 1 })
    expect(wrapper.get('[data-testid="cooldown"]').element).toHaveProperty('value', '15')
    expect(wrapper.text()).toContain('成功操作不计数')
    expect(wrapper.text()).toContain('已有临时封禁到期恢复')
  })
  it('永久模式须管理员解除，并保持临时时长供以后切回', async () => {
    api.put.mockImplementation(async (_url: string, value: object) => ({ data: { ...value, version: 2 } }))
    const wrapper = render()
    await flushPromises()
    await wrapper.get('[data-testid="cooldown"]').setValue('')
    await wrapper.get('[data-testid="block-mode"]').setValue('permanent')
    expect(wrapper.find('[data-testid="cooldown"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('永久封禁：管理员移除')
    await wrapper.findAll('form')[0].trigger('submit')
    await flushPromises()
    expect(api.put).toHaveBeenLastCalledWith('/admin/ip-blacklist/config', { ...config, block_mode: 'permanent' })
    expect(wrapper.text()).toContain('修改规则不会改变已有记录的期限')
    await wrapper.get('[data-testid="block-mode"]').setValue('timed')
    expect(wrapper.get('[data-testid="cooldown"]').element).toHaveProperty('value', '15')
    await wrapper.get('[data-testid="cooldown"]').setValue('45')
    await wrapper.findAll('form')[0].trigger('submit')
    await flushPromises()
    expect(api.put).toHaveBeenLastCalledWith('/admin/ip-blacklist/config', { ...config, block_seconds: 2700, version: 2 })
  })
  it('重新打开时正确载入永久模式和临时记录期限', async () => {
    api.get.mockImplementation(async (url: string) => ({ data: url.endsWith('/config') ? { ...config, block_mode: 'permanent' } : { items: [{ ...entry, blocked_until: '2026-10-13T12:47:32Z' }], total: 1, page: 1, page_size: 20, pages: 1 } }))
    const wrapper = render()
    await flushPromises()
    expect(wrapper.get('[data-testid="block-mode"]').element).toHaveProperty('value', 'permanent')
    expect(wrapper.find('[data-testid="cooldown"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('须联系管理员移除')
    expect(wrapper.text()).toContain('到期：')
  })
  it('配置读取失败时禁止保存默认值', async () => {
    api.get.mockRejectedValue({ status: 503, message: '防护数据库暂时不可用' })
    const wrapper = render()
    await flushPromises()
    expect(wrapper.get('[data-testid="save"]').attributes('disabled')).toBeDefined()
    await wrapper.findAll('form')[0].trigger('submit')
    expect(api.put).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('防护数据库暂时不可用')
  })
  it('精确搜索和分页传递到服务器', async () => {
    api.get.mockImplementation(async (url: string, options?: { params: { page: number } }) => ({ data: url.endsWith('/config') ? config : { items: [entry], total: 21, page: options?.params.page ?? 1, page_size: 20, pages: 2 } }))
    const wrapper = render()
    await flushPromises()
    await wrapper.get('[data-testid="search"]').setValue(' 61.120.9.25 ')
    await wrapper.findAll('form')[1].trigger('submit')
    await flushPromises()
    expect(api.get).toHaveBeenLastCalledWith('/admin/ip-blacklist/entries', { params: { ip: '61.120.9.25', page: 1, page_size: 20 } })
    await wrapper.findAll('button').find(button => button.text() === '下一页')!.trigger('click')
    await flushPromises()
    expect(api.get).toHaveBeenLastCalledWith('/admin/ip-blacklist/entries', { params: { ip: '61.120.9.25', page: 2, page_size: 20 } })
  })
  it('确认解封携带状态版本，409刷新名单并要求再次确认', async () => {
    api.post.mockRejectedValue({ status: 409 })
    const wrapper = render()
    await flushPromises()
    await wrapper.findAll('button').find(button => button.text() === '移除')!.trigger('click')
    expect(api.post).not.toHaveBeenCalled()
    await wrapper.get('[data-testid="confirm"]').trigger('click')
    await flushPromises()
    expect(api.post).toHaveBeenCalledWith('/admin/ip-blacklist/entries/8/unblock', { generation: 3 })
    expect(wrapper.text()).toContain('名单状态已变化')
    expect(wrapper.find('[data-testid="confirm"]').exists()).toBe(false)
    expect(api.get).toHaveBeenCalledTimes(3)
  })
  it('配置409读取当前版本后再保存', async () => {
    api.put.mockRejectedValue({ status: 409 })
    const wrapper = render()
    await flushPromises()
    await wrapper.findAll('form')[0].trigger('submit')
    await flushPromises()
    expect(wrapper.text()).toContain('请重新确认后保存')
    expect(api.get).toHaveBeenCalledTimes(3)
  })
  it('解封成功立即刷新，界面标注时区并显示首末时间', async () => {
    const wrapper = render()
    await flushPromises()
    expect(wrapper.text()).toContain('页面时间：')
    expect(wrapper.text()).toContain('10 次 / 120 分钟')
    await wrapper.findAll('button').find(button => button.text() === '移除')!.trigger('click')
    await wrapper.get('[data-testid="confirm"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('61.120.9.25 已解封')
    expect(api.get).toHaveBeenCalledTimes(3)
  })
  it('路由要求管理员，时间转换只做一次', async () => {
    const route = customRoutes.find(record => record.path === '/admin/ip-blacklist')
    expect(route?.meta).toMatchObject({ requiresAuth: true, requiresAdmin: true })
    expect(formatAuthIPTime(entry.blocked_at, 'Asia/Shanghai')).toContain('20:32:32')
    expect(formatAuthIPTime('invalid', 'Asia/Shanghai')).toBe('—')
    await authIPBlacklistAPI.list('', 1)
    expect(api.get).toHaveBeenCalledWith('/admin/ip-blacklist/entries', { params: { ip: '', page: 1, page_size: 20 } })
  })
})
