// [CUSTOM] 官网空闲时预取相邻公开页的路由块；只下载代码，不发起数据请求。
import type { Router } from 'vue-router'

const prefetched = new Set<string>()
type Loader = () => Promise<unknown>

function constrainedNetwork(): boolean {
  const connection = (navigator as Navigator & { connection?: { saveData?: boolean; effectiveType?: string } }).connection
  return Boolean(connection?.saveData) || /(^|-)2g$/.test(connection?.effectiveType || '')
}

export function prefetchRoute(router: Router | undefined, path: string): void {
  if (!router || prefetched.has(path) || constrainedNetwork()) return
  prefetched.add(path)
  let matched
  try { matched = router.resolve(path).matched } catch { return }
  for (const record of matched) {
    for (const component of Object.values(record.components || {})) {
      // 懒加载路由是无参函数；已解析的组件对象（含函数式组件声明）跳过。
      if (typeof component === 'function' && component.length === 0 && !('render' in component) && !('setup' in component)) {
        try { void Promise.resolve((component as Loader)()).catch(() => prefetched.delete(path)) }
        catch { prefetched.delete(path) }
      }
    }
  }
}

export function whenIdle(task: () => void): () => void {
  if (typeof window.requestIdleCallback === 'function') {
    const handle = window.requestIdleCallback(task, { timeout: 3000 })
    return () => window.cancelIdleCallback(handle)
  }
  const handle = window.setTimeout(task, 1200)
  return () => window.clearTimeout(handle)
}

export function resetPrefetchForTest(): void {
  prefetched.clear()
}
