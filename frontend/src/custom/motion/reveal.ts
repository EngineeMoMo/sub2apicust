// [CUSTOM] 官网滚动出现：共享一个 IntersectionObserver，失败或减少动效时内容始终可见。
import type { Directive } from 'vue'

const ROOT_CLASS = 'mofa-motion'
const REVEALED_CLASS = 'is-revealed'
let observer: IntersectionObserver | null = null

/** 浏览器支持 IO 且用户未开启“减少动态效果”时才启用动效。 */
export function motionAllowed(): boolean {
  if (typeof window === 'undefined' || typeof window.IntersectionObserver !== 'function') return false
  if (typeof window.matchMedia !== 'function') return true
  return !window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

function sharedObserver(): IntersectionObserver {
  if (observer) return observer
  observer = new IntersectionObserver(entries => {
    for (const entry of entries) {
      if (!entry.isIntersecting) continue
      entry.target.classList.add(REVEALED_CLASS)
      observer?.unobserve(entry.target)
    }
  }, { rootMargin: '0px 0px -8% 0px', threshold: 0 })
  return observer
}

/**
 * v-reveal：元素首次进入视口时加上 is-revealed。
 * 绑定数字表示依次出现的序号（每级 70ms，最多 8 级）。
 * 隐藏初始态只在 html.mofa-motion 下生效，未启用时元素保持原样可见。
 */
export const vReveal: Directive<HTMLElement, number | undefined> = {
  mounted(element, binding) {
    element.classList.add('mofa-reveal')
    if (typeof binding.value === 'number' && binding.value > 0) {
      element.style.setProperty('--reveal-delay', `${Math.min(binding.value, 8) * 70}ms`)
    }
    if (!motionAllowed()) return
    document.documentElement.classList.add(ROOT_CLASS)
    sharedObserver().observe(element)
  },
  beforeUnmount(element) {
    observer?.unobserve(element)
  }
}

/** 仅供测试重置模块状态。 */
export function resetRevealForTest(): void {
  observer?.disconnect()
  observer = null
  document.documentElement.classList.remove(ROOT_CLASS)
}
