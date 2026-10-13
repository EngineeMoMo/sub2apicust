<template>
  <button
    type="button"
    class="mofa-theme-toggle"
    :aria-label="isDark ? copy.light : copy.dark"
    :title="isDark ? copy.light : copy.dark"
    @click="toggleTheme"
  >
    <Icon :name="isDark ? 'sun' : 'moon'" size="md" />
  </button>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { useBrandCopy } from '@/custom/brand/copy'

const copy = useBrandCopy()
const isDark = ref(document.documentElement.classList.contains('dark'))

type ViewTransitionDocument = Document & { startViewTransition?: (update: () => void) => { finished: Promise<void> } }

function applyTheme(dark: boolean) {
  isDark.value = dark
  document.documentElement.classList.toggle('dark', dark)
  localStorage.setItem('theme', dark ? 'dark' : 'light')
}

// [CUSTOM] 支持 View Transitions 时从按钮位置圆形扩散切换；不支持或减少动效时同步切换。
function toggleTheme(event: MouseEvent) {
  const next = !document.documentElement.classList.contains('dark')
  const doc = document as ViewTransitionDocument
  const reduced = typeof window.matchMedia === 'function' && window.matchMedia('(prefers-reduced-motion: reduce)').matches
  if (typeof doc.startViewTransition !== 'function' || reduced) {
    applyTheme(next)
    return
  }
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  const x = rect.left + rect.width / 2
  const y = rect.top + rect.height / 2
  const radius = Math.hypot(Math.max(x, window.innerWidth - x), Math.max(y, window.innerHeight - y))
  const root = document.documentElement
  root.style.setProperty('--theme-x', `${x}px`)
  root.style.setProperty('--theme-y', `${y}px`)
  root.style.setProperty('--theme-r', `${Math.ceil(radius)}px`)
  root.classList.add('mofa-theme-transition')
  doc.startViewTransition(() => applyTheme(next)).finished.finally(() => root.classList.remove('mofa-theme-transition'))
}
</script>
