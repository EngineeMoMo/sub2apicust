import { onBeforeUnmount, onMounted, ref, shallowRef } from 'vue'
import { INTELLIGENCE_SOURCE, parseIntelligence, REFRESH_INTERVAL, type IntelligenceSnapshot } from './data'

export function useIntelligence() {
  const snapshot = shallowRef<IntelligenceSnapshot | null>(null)
  const loading = ref(false)
  const failed = ref(false)
  const fetchedAt = ref<number | null>(null)
  const now = ref(Date.now())
  let controller: AbortController | null = null
  let disposed = false
  let timer: ReturnType<typeof setInterval> | undefined

  async function refresh() {
    if (loading.value || disposed) return
    loading.value = true
    controller = new AbortController()
    const timeout = setTimeout(() => controller?.abort(), 15_000)
    try {
      const response = await fetch(INTELLIGENCE_SOURCE, {
        signal: controller.signal, credentials: 'omit', referrerPolicy: 'no-referrer', cache: 'no-cache'
      })
      if (!response.ok) throw new Error('source-unavailable')
      const parsed = parseIntelligence(await response.json())
      if (disposed) return
      snapshot.value = parsed
      fetchedAt.value = Date.now()
      now.value = Date.now()
      failed.value = false
    } catch {
      if (!disposed) failed.value = true
    } finally {
      clearTimeout(timeout)
      if (!disposed) loading.value = false
      controller = null
    }
  }

  function checkFreshness() {
    now.value = Date.now()
    if (document.visibilityState === 'hidden') return
    if (fetchedAt.value === null || now.value - fetchedAt.value >= REFRESH_INTERVAL) void refresh()
  }

  onMounted(() => {
    void refresh()
    timer = setInterval(() => {
      now.value = Date.now()
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

  return { snapshot, loading, failed, fetchedAt, now, refresh }
}
