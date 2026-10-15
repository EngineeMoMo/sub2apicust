import { inject, onMounted, onUnmounted, provide, ref, watch, type InjectionKey, type Ref } from 'vue'
import { getGroupRateSchedules, type GroupRateSchedules } from './api'

interface RateDisplay {
  schedules: Ref<GroupRateSchedules['schedules']>
  now: Ref<number>
  ready: Ref<boolean>
  enabled: Ref<boolean>
}

const rateDisplayKey: InjectionKey<RateDisplay> = Symbol('groupRateDisplay')

export function provideGroupRateDisplay(enabled: Ref<boolean>) {
  const schedules = ref<GroupRateSchedules['schedules']>({})
  const now = ref(0)
  const ready = ref(false)
  let serverTime = 0
  let receivedAt = 0
  let version = 0
  let disposed = false
  let clockTimer: ReturnType<typeof setInterval> | undefined
  let refreshTimer: ReturnType<typeof setInterval> | undefined

  async function reload() {
    const requestVersion = ++version
    if (!enabled.value || disposed) return
    try {
      const data = await getGroupRateSchedules()
      if (disposed || requestVersion !== version) return
      const instant = Date.parse(data.server_time)
      if (!Number.isFinite(instant)) throw new Error('Invalid server time')
      schedules.value = data.schedules
      serverTime = instant
      receivedAt = performance.now()
      now.value = serverTime
      ready.value = true
    } catch {
      if (!disposed && requestVersion === version) ready.value = false
    }
  }

  const refreshVisible = () => {
    if (document.visibilityState !== 'hidden') void reload()
  }

  provide(rateDisplayKey, { schedules, now, ready, enabled })
  watch(enabled, () => { ready.value = false; void reload() })
  onMounted(() => {
    clockTimer = setInterval(() => { now.value = serverTime + performance.now() - receivedAt }, 1000)
    refreshTimer = setInterval(refreshVisible, 30_000)
    document.addEventListener('visibilitychange', refreshVisible)
  })
  onUnmounted(() => {
    disposed = true
    version++
    clearInterval(clockTimer)
    clearInterval(refreshTimer)
    document.removeEventListener('visibilitychange', refreshVisible)
  })
  return { reload }
}

export function useGroupRateDisplay() {
  return inject(rateDisplayKey, null)
}
