import { beforeEach, describe, expect, it, vi } from 'vitest'
import { studioAPI } from '@/custom/studio/api'
const client = vi.hoisted(() => ({ get: vi.fn(), post: vi.fn() }))
vi.mock('@/api/client', () => ({ apiClient: client }))
beforeEach(() => { client.get.mockReset().mockResolvedValue({ data: [] }); client.post.mockReset().mockResolvedValue({ data: { status: 'pending' } }) })
describe('投稿API使用认证客户端，不把Token放入媒体URL', () => {
  it('用户与管理员目录、blob预览走不同受保护入口', async () => {
    await studioAPI.list(false); await studioAPI.list(true)
    expect(client.get.mock.calls.map(call => call[0])).toEqual(['/studio/submissions', '/admin/studio/submissions'])
    const operation = new AbortController(), id = 'a'.repeat(32)
    await studioAPI.preview(id, false, operation.signal)
    expect(client.get).toHaveBeenLastCalledWith('/studio/submissions/' + id + '/media', { responseType: 'blob', signal: operation.signal })
    await expect(studioAPI.preview('../private', true, operation.signal)).rejects.toThrow('无效')
  })
  it('上传采用multipart，审核带旧状态及明确确认，不自动调用审核', async () => {
    const body = new FormData(); await studioAPI.submit(body)
    expect(client.post).toHaveBeenCalledTimes(1)
    expect(client.post).toHaveBeenCalledWith('/studio/submissions', body, expect.objectContaining({ timeout: 120000 }))
    await studioAPI.review({ id: 'a'.repeat(32), status: 'pending' } as Parameters<typeof studioAPI.review>[0], 'approve', '', true)
    expect(client.post).toHaveBeenLastCalledWith('/admin/studio/submissions/' + 'a'.repeat(32) + '/review', { action: 'approve', expected_status: 'pending', reason: '', review_confirmed: true })
  })
})
