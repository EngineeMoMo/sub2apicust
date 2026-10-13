<template>
  <div class="mofa-brand-home mofa-public">
    <a class="mofa-skip-link" href="#public-main">跳到正文</a>
    <header class="mofa-brand-mast" :class="{ 'is-scrolled': scrolled, 'is-hidden': hidden, 'is-menu-open': menuOpen }" @focusin="onFocusIn" @focusout="onFocusOut">
      <router-link to="/home" class="mofa-brand-lockup">
        <img :src="siteLogo" :alt="app.siteName" width="42" height="42" decoding="async" fetchpriority="high">
        <span>{{ app.siteName }}<small>多模型 API 服务</small></span>
      </router-link>
      <button ref="menuButton" type="button" class="mofa-nav-toggle" aria-controls="public-nav" :aria-expanded="menuOpen" @click="menuOpen = !menuOpen">
        <Icon :name="menuOpen ? 'x' : 'menu'" size="md" aria-hidden="true" />
        <span class="sr-only">{{ menuOpen ? '关闭菜单' : '打开菜单' }}</span>
      </button>
      <nav id="public-nav" ref="nav" class="mofa-brand-nav" aria-label="官网导航" @click="closeFromLink">
        <router-link to="/family">家族首页</router-link>
        <router-link to="/api">魔法 API</router-link>
        <router-link to="/plans">订阅套餐</router-link>
        <router-link to="/guide">接入教程</router-link>
        <router-link to="/faq">常见问题</router-link>
        <router-link v-if="!auth.isAuthenticated" to="/preview">游客预览</router-link>
        <router-link v-if="app.cachedPublicSettings?.model_plaza_enabled && (auth.isAuthenticated || !app.cachedPublicSettings?.model_plaza_require_auth)" to="/model-plaza">模型广场</router-link>
        <BrandThemeToggle />
        <router-link v-if="auth.isAuthenticated" :to="auth.isAdmin ? '/admin/dashboard' : '/dashboard'" class="btn btn-primary">进入控制台</router-link>
        <template v-else>
          <router-link to="/login" class="btn btn-secondary">登录</router-link>
          <router-link v-if="app.cachedPublicSettings?.registration_enabled" to="/register" class="btn btn-primary">注册</router-link>
        </template>
      </nav>
    </header>
    <main id="public-main" class="mofa-public-main" tabindex="-1"><slot /></main>
    <footer class="mofa-brand-footer">
      <span>© {{ new Date().getFullYear() }} {{ app.siteName }}</span>
      <router-link to="/faq">使用前有疑问？查看常见问题</router-link>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { Router } from 'vue-router'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import Icon from '@/components/icons/Icon.vue'
import BrandThemeToggle from '@/custom/components/BrandThemeToggle.vue'
import { prefetchRoute, whenIdle } from '@/custom/motion/prefetch'

const app = useAppStore()
const auth = useAuthStore()
const siteLogo = computed(() => sanitizeUrl(app.siteLogo || '', { allowRelative: true, allowDataUrl: true }) || '/logo.svg')
// 测试中未安装路由时为 undefined，预取直接跳过。
const router = getCurrentInstance()?.appContext.config.globalProperties.$router as Router | undefined
const PREFETCH_PATHS = ['/plans', '/guide', '/faq', '/api', '/family', '/login']

const scrolled = ref(false)
const hidden = ref(false)
const menuOpen = ref(false)
const menuButton = ref<HTMLButtonElement>()
const nav = ref<HTMLElement>()
let scrollFrame = 0
let lastY = 0
let cancelIdle: (() => void) | undefined

// 向下滚动收起页头、向上滚动出现：路由按锚点把章节滚到视口顶部时，标题不被吸顶页头遮住。
function readScroll() {
  scrollFrame = 0
  const y = Math.max(0, window.scrollY)
  scrolled.value = y > 8
  if (menuOpen.value || y < 120) {
    hidden.value = false
    lastY = y
  } else if (Math.abs(y - lastY) > 12) {
    hidden.value = y > lastY
    lastY = y
  }
}
function onScroll() {
  if (!scrollFrame) scrollFrame = requestAnimationFrame(readScroll)
}
function closeFromLink(event: MouseEvent) {
  if ((event.target as HTMLElement | null)?.closest('a')) menuOpen.value = false
}
function onKeydown(event: KeyboardEvent) {
  if (event.key !== 'Escape' || !menuOpen.value) return
  menuOpen.value = false
  menuButton.value?.focus()
}
function onFocusIn() {
  hidden.value = false
}
function onFocusOut(event: FocusEvent) {
  const next = event.relatedTarget as Node | null
  if (menuOpen.value && next && !(event.currentTarget as HTMLElement).contains(next)) menuOpen.value = false
}
function onOutside(event: PointerEvent) {
  const target = event.target as Node | null
  if (target && !nav.value?.contains(target) && !menuButton.value?.contains(target)) menuOpen.value = false
}
watch(menuOpen, open => {
  if (open) document.addEventListener('pointerdown', onOutside)
  else document.removeEventListener('pointerdown', onOutside)
})

onMounted(() => {
  void app.fetchPublicSettings().catch(() => {})
  lastY = Math.max(0, window.scrollY)
  readScroll()
  window.addEventListener('scroll', onScroll, { passive: true })
  document.addEventListener('keydown', onKeydown)
  cancelIdle = whenIdle(() => PREFETCH_PATHS.forEach(path => prefetchRoute(router, path)))
})
onBeforeUnmount(() => {
  if (scrollFrame) cancelAnimationFrame(scrollFrame)
  window.removeEventListener('scroll', onScroll)
  document.removeEventListener('keydown', onKeydown)
  document.removeEventListener('pointerdown', onOutside)
  cancelIdle?.()
})
</script>
