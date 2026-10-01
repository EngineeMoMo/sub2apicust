import { apiClient } from '@/api/client'

export const studioCategories = ['年轻人像', '时尚肖像', 'Cosplay', '动漫二次元', '科技机甲', '动物自然', '奇幻风景', '电商产品', '场景插画', '海报社媒', '空间设计']
export const studioStyles = ['写实摄影', '电影感', '二次元', '3D手作', '概念设计', '平面海报', '水彩', '水墨', '像素', '美漫', '剪纸', '复古未来']
export const statusLabels: Record<string, string> = { pending: '待审核', published: '已公开', rejected: '已退回', unpublished: '已下架', withdrawn: '已撤回' }
export interface StudioSubmission {
  id: string
  title: string
  author: string
  category: string
  style: string
  media: 'image' | 'video'
  prompt: string
  prompt_kind: 'actual' | 'reference'
  model: string
  notes: string
  status: string
  reason: string
  created_at: string
}
export function validateStudioFile(file: File | undefined, media: 'image' | 'video'): string {
  if (!file) return '请先选择作品文件。'
  const allowed = media === 'image' ? ['image/png', 'image/jpeg', 'image/webp'] : ['video/mp4']
  if (!allowed.includes(file.type)) return media === 'image' ? '图片仅支持 PNG、JPEG、WebP。' : '视频仅支持 MP4。'
  if (!file.size || file.size > (media === 'image' ? 12 : 30) * 1024 * 1024) return media === 'image' ? '图片最多12MB，不能是空文件。' : '视频最多30MB，不能是空文件。'
  return ''
}
export const studioAPI = {
  async list(review = false, signal?: AbortSignal) {
    return (await apiClient.get<StudioSubmission[]>(review ? '/admin/studio/submissions' : '/studio/submissions', { signal })).data
  },
  async submit(body: FormData) {
    return (await apiClient.post<StudioSubmission>('/studio/submissions', body, { headers: { 'Content-Type': 'multipart/form-data' }, timeout: 120000 })).data
  },
  async preview(id: string, review: boolean, signal: AbortSignal) {
    if (!/^[a-f0-9]{32}$/.test(id)) throw new Error('无效投稿编号。')
    return (await apiClient.get<Blob>((review ? '/admin/studio/submissions/' : '/studio/submissions/') + id + '/media', { responseType: 'blob', signal })).data
  },
  async withdraw(id: string) {
    return (await apiClient.post<StudioSubmission>('/studio/submissions/' + id + '/withdraw')).data
  },
  async review(entry: StudioSubmission, action: 'approve' | 'reject' | 'unpublish', reason: string, confirmed: boolean) {
    return (await apiClient.post<StudioSubmission>('/admin/studio/submissions/' + entry.id + '/review', {
      action, expected_status: entry.status, reason, review_confirmed: confirmed
    })).data
  }
}
