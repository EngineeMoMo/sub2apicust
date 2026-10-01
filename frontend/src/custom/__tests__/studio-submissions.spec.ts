import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { reactive } from 'vue'
import StudioSubmissionsView from '@/custom/views/StudioSubmissionsView.vue'
import { customRoutes } from '@/custom/routes'
import { validateStudioFile } from '@/custom/studio/api'

const harness = vi.hoisted(() => ({
  auth: { isAdmin: false, user: { id: 1 } }, route: { path: '/tools/studio/submit' },
  list: vi.fn(), submit: vi.fn(), preview: vi.fn(), review: vi.fn(), withdraw: vi.fn()
}))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => harness.auth }))
vi.mock('vue-router', () => ({ useRoute: () => harness.route }))
vi.mock('@/custom/studio/api', async importOriginal => ({
  ...await importOriginal<typeof import('@/custom/studio/api')>(),
  studioAPI: { list: harness.list, submit: harness.submit, preview: harness.preview, review: harness.review, withdraw: harness.withdraw }
}))
let wrapper: VueWrapper | undefined
const entry = { id: 'a'.repeat(32), title: '原创测试作品', author: '署名', category: '奇幻风景', style: '水彩', media: 'image', prompt_kind: 'actual', prompt: '投稿人实际使用的完整提示词', model: '自己的模型', notes: '', status: 'pending', reason: '', created_at: '2026-10-01' }
beforeEach(() => {
  harness.auth = reactive({ isAdmin: false, user: { id: 1 } }); harness.route = reactive({ path: '/tools/studio/submit' })
  harness.list.mockReset().mockResolvedValue([]); harness.submit.mockReset().mockResolvedValue(entry)
  harness.preview.mockReset().mockResolvedValue(new Blob(['image'], { type: 'image/png' }))
  harness.review.mockReset().mockResolvedValue({ ...entry, status: 'published' }); harness.withdraw.mockReset().mockResolvedValue({ ...entry, status: 'withdrawn' })
  Object.defineProperty(URL, 'createObjectURL', { value: vi.fn(() => 'blob:test-local'), configurable: true })
  Object.defineProperty(URL, 'revokeObjectURL', { value: vi.fn(), configurable: true })
})
afterEach(() => { wrapper?.unmount(); vi.restoreAllMocks() })
function page() {
  wrapper = mount(StudioSubmissionsView, { global: { stubs: { AppLayout: { template: '<main><slot /></main>' }, RouterLink: true } } })
  return wrapper
}
function button(text: string) { return wrapper!.findAll('button').find(node => node.text() === text)! }
async function chooseFile() {
  const input = wrapper!.get('input[type="file"]')
  let files = [new File(['image'], 'art.png', { type: 'image/png' })]
  Object.defineProperty(input.element, 'files', { get: () => files, configurable: true })
  Object.defineProperty(input.element, 'value', { get: () => '', set: (value: string) => { if (!value) files = [] }, configurable: true })
  await input.trigger('change')
}
describe('工坊投稿与审核组件（模拟账号/API，不是生产验收）', () => {
  it('投稿/审核路由复用主站登录和管理员守卫', () => {
    expect(customRoutes.find(route => route.name === 'StudioSubmissions')?.meta).toMatchObject({ requiresAuth: true })
    expect(customRoutes.find(route => route.name === 'StudioReview')?.meta).toMatchObject({ requiresAuth: true, requiresAdmin: true })
  })
  it('文件类型与体积校验拒绝SVG、空文件和超限', () => {
    expect(validateStudioFile(undefined, 'image')).toContain('选择')
    expect(validateStudioFile(new File(['svg'], 'art.svg', { type: 'image/svg+xml' }), 'image')).toContain('仅支持')
    expect(validateStudioFile(new File([], 'art.png', { type: 'image/png' }), 'image')).toContain('空文件')
    expect(validateStudioFile(new File([new Uint8Array(12 * 1024 * 1024 + 1)], 'art.png', { type: 'image/png' }), 'image')).toContain('最多12MB')
    expect(validateStudioFile(new File(['mp4'], 'film.mp4', { type: 'video/mp4' }), 'video')).toBe('')
  })
  it('选文件仅本地预览，不立即上传；清空input不会丢失本次文件', async () => {
    page(); await chooseFile()
    expect(wrapper!.text()).toContain('art.png')
    expect(wrapper!.get('img').attributes('src')).toBe('blob:test-local')
    expect(harness.submit).not.toHaveBeenCalled()
    expect(button('提交审核').attributes('disabled')).toBeDefined()
    wrapper!.unmount(); wrapper = undefined; expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:test-local')
  })
  it('视频投稿显示待开放，强制切换类型也不会上传', async () => {
    page()
    expect(wrapper!.get('option[value="video"]').attributes('disabled')).toBeDefined()
    expect(wrapper!.text()).toContain('视频投稿待开放')
    await wrapper!.findAll('select')[0].setValue('video')
    await wrapper!.get('form').trigger('submit')
    expect(wrapper!.get('[role="alert"]').text()).toContain('视频投稿待开放')
    expect(harness.submit).not.toHaveBeenCalled()
  })
  it('旧视频可私有查看与撤回，但不能恢复新投稿入口', async () => {
    harness.list.mockResolvedValue([{ ...entry, media: 'video', status: 'rejected' }])
    harness.preview.mockResolvedValue(new Blob(['video'], { type: 'video/mp4' }))
    page(); await button('我的投稿').trigger('click'); await flushPromises()
    await wrapper!.get('.mofa-studio-submission-row').trigger('click'); await flushPromises()
    expect(wrapper!.get('video').attributes('src')).toBe('blob:test-local')
    await button('修改后重新投稿').trigger('click')
    expect(wrapper!.text()).toContain('视频投稿待开放')
    expect(wrapper!.find('form').exists()).toBe(false)
    expect(harness.submit).not.toHaveBeenCalled()
    await button('撤回作品').trigger('click'); await button('确认撤回').trigger('click'); await flushPromises()
    expect(harness.withdraw).toHaveBeenCalledOnce()
  })
  it('确认两项授权后提交multipart，成功提示待审核而非自动公开', async () => {
    page(); await chooseFile()
    await wrapper!.get('input[minlength="2"]').setValue('原创图像')
    await wrapper!.get('input[maxlength="40"]').setValue('作者')
    const selects = wrapper!.findAll('select'); await selects[1].setValue('奇幻风景'); await selects[2].setValue('水彩')
    await wrapper!.findAll('input[maxlength="80"]')[1].setValue('真实模型')
    await wrapper!.get('textarea[minlength="10"]').setValue('这是一份完整的实际生成提示词')
    for (const checkbox of wrapper!.findAll('input[type="checkbox"]')) await checkbox.setValue(true)
    await wrapper!.get('form').trigger('submit'); await flushPromises()
    const body = harness.submit.mock.calls[0][0] as FormData
    expect(body.get('file')).toBeInstanceOf(File); expect(body.get('rights_confirmed')).toBe('true'); expect(body.get('publish_consent')).toBe('true')
    expect(wrapper!.text()).toContain('状态为待审核'); expect(harness.review).not.toHaveBeenCalled()
  })
  it('管理员必须加载私有素材并确认审核；只勾选或只预览都不能发布', async () => {
    harness.auth.isAdmin = true; harness.route.path = '/admin/studio/submissions'; harness.list.mockResolvedValue([entry])
    page(); await flushPromises(); await wrapper!.get('.mofa-studio-submission-row').trigger('click'); await flushPromises()
    expect(harness.preview).toHaveBeenCalledWith(entry.id, true, expect.any(AbortSignal))
    const approve = button('审核通过并公开'); expect(approve.attributes('disabled')).toBeDefined()
    await wrapper!.get('input[type="checkbox"]').setValue(true); expect(approve.attributes('disabled')).toBeDefined()
    await wrapper!.get('img').trigger('load'); expect(approve.attributes('disabled')).toBeUndefined()
    await approve.trigger('click'); await flushPromises()
    expect(harness.review).toHaveBeenCalledWith(expect.objectContaining({ status: 'pending' }), 'approve', '', true)
    expect(wrapper!.text()).toContain('已通过审核并公开')
  })
  it('账号切换清空文件、授权和私人提示词，迟到的响应不会带回旧记录', async () => {
    page(); await chooseFile(); await wrapper!.get('textarea[minlength="10"]').setValue('前账号的私人提示词')
    harness.auth.user.id = 2; await flushPromises()
    expect(wrapper!.find('img').exists()).toBe(false); expect(wrapper!.get('textarea[minlength="10"]').element.value).toBe('')
    expect(URL.revokeObjectURL).toHaveBeenCalled(); expect(wrapper!.text()).not.toContain('art.png')
  })
  it('上传返回前切换账号，旧响应不能显示成功或读取新账号目录', async () => {
    let finish: (value: typeof entry) => void = () => {}
    harness.submit.mockReturnValue(new Promise(resolve => { finish = resolve }))
    page(); await chooseFile()
    for (const checkbox of wrapper!.findAll('input[type="checkbox"]')) await checkbox.setValue(true)
    await wrapper!.get('form').trigger('submit')
    expect(harness.submit).toHaveBeenCalledOnce()
    harness.auth.user.id = 2; await flushPromises(); finish(entry); await flushPromises()
    expect(wrapper!.text()).not.toContain('已提交'); expect(harness.list).not.toHaveBeenCalled()
  })
})
