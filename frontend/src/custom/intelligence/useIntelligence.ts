import { onBeforeUnmount, onMounted, ref, shallowRef } from 'vue'
import { apiClient } from '@/api/client'
import { INTELLIGENCE_ENDPOINT, parseIntelligence, REFRESH_INTERVAL, REQUEST_TIMEOUT, type IntelligenceSnapshot } from './data'

export function useIntelligence() {
  const snapshot = shallowRef<IntelligenceSnapshot | null>(null)
  const loading = ref(false)
  const failed = ref(false)
  const fetchedAt = ref<number | null>(null)
  let controller: AbortController | null = null
  let disposed = false
  let timer: ReturnType<typeof setInterval> | undefined

  // 读取本站接口（需登录）；刷新只读服务端缓存，不会每次都请求源站。
  async function refresh() {
    if (loading.value || disposed) return
    loading.value = true
    controller = new AbortController()
    try {
      const { data } = await apiClient.get(INTELLIGENCE_ENDPOINT, { signal: controller.signal, timeout: REQUEST_TIMEOUT })
      const parsed = parseIntelligence(data)
      if (disposed) return
      snapshot.value = parsed
      fetchedAt.value = Date.now()
      failed.value = false
    } catch {
      if (!disposed) failed.value = true
    } finally {
      if (!disposed) loading.value = false
      controller = null
    }
  }

  function checkFreshness() {
    if (document.visibilityState === 'hidden') return
    if (fetchedAt.value === null || Date.now() - fetchedAt.value >= REFRESH_INTERVAL) void refresh()
  }

  onMounted(() => {
    void refresh()
    timer = setInterval(() => {
      if (document.visibilityState !== 'hidden') void refresh()
    }, REFRESH_INTERVAL)
    document.addEventListener('visibilitychange', checkFreshness)
  })

  onBeforeUnmount(() => {
    disposed = true
    controller?.abort()
    clearInterval(timer)
    document.removeEventListener('visibilitychange', checkFreshness)
  })

  return { snapshot, loading, failed, fetchedAt, refresh }
}
