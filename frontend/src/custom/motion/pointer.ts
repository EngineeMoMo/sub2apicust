// [CUSTOM] 精细指针下的光斑与轻倾斜；触屏、减少动效或测试环境不注册任何监听。
import type { Directive } from 'vue'

type PointerState = { frame: number; x: number; y: number; move: (event: PointerEvent) => void; leave: () => void }
const states = new WeakMap<HTMLElement, PointerState>()

export function finePointer(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false
  return window.matchMedia('(hover: hover) and (pointer: fine)').matches
    && !window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

/**
 * 写入 --mx/--my（元素内 0–100%）与 --tilt-x/--tilt-y（±1）。
 * 样式层决定如何使用这些变量，指令本身不改布局。
 */
function bind(element: HTMLElement, activeClass: string) {
  if (!finePointer()) return
  const state: PointerState = {
    frame: 0,
    x: 0,
    y: 0,
    move(event) {
      // 同一帧内多次移动只计算一次，且使用最后一次的位置。
      state.x = event.clientX
      state.y = event.clientY
      if (state.frame) return
      state.frame = requestAnimationFrame(() => {
        state.frame = 0
        const rect = element.getBoundingClientRect()
        if (!rect.width || !rect.height) return
        const x = (state.x - rect.left) / rect.width
        const y = (state.y - rect.top) / rect.height
        element.style.setProperty('--mx', `${(x * 100).toFixed(1)}%`)
        element.style.setProperty('--my', `${(y * 100).toFixed(1)}%`)
        element.style.setProperty('--tilt-x', (x * 2 - 1).toFixed(3))
        element.style.setProperty('--tilt-y', (y * 2 - 1).toFixed(3))
        element.classList.add(activeClass)
      })
    },
    leave() {
      if (state.frame) cancelAnimationFrame(state.frame)
      state.frame = 0
      element.classList.remove(activeClass)
      element.style.setProperty('--tilt-x', '0')
      element.style.setProperty('--tilt-y', '0')
    }
  }
  element.addEventListener('pointermove', state.move, { passive: true })
  element.addEventListener('pointerleave', state.leave, { passive: true })
  states.set(element, state)
}

function unbind(element: HTMLElement) {
  const state = states.get(element)
  if (!state) return
  if (state.frame) cancelAnimationFrame(state.frame)
  element.removeEventListener('pointermove', state.move)
  element.removeEventListener('pointerleave', state.leave)
  states.delete(element)
}

/** v-pointer：光斑与倾斜共用同一组变量，元素在指针内时带 is-pointing。 */
export const vPointer: Directive<HTMLElement> = {
  mounted(element) { bind(element, 'is-pointing') },
  beforeUnmount(element) { unbind(element) }
}
