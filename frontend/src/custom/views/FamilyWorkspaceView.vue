<template>
  <AppLayout>
    <section class="mofa-product-workspace" :aria-label="product === 'recipes' ? '魔法配方工作区' : '魔法工坊工作区'" :aria-busy="!loaded">
      <div v-if="!loaded" class="mofa-product-workspace-loading" role="status">
        <span class="mofa-product-workspace-spinner" aria-hidden="true" />
        <span>正在打开{{ product === 'recipes' ? '魔法配方' : '魔法工坊' }}…</span>
      </div>
      <iframe :key="product" ref="frame" :src="source" :title="product === 'recipes' ? '魔法配方' : '魔法工坊'" referrerpolicy="no-referrer" @load="initialize" />
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { keysAPI } from '@/api/keys'
import { buildGatewayUrl } from '@/api/url'
import { gatewayBase, loadModels } from '@/custom/recipes/connect'
import { acceptsWorkspaceMessage, WorkspaceConnections } from '@/custom/family/workspace'

const route = useRoute(), router = useRouter(), auth = useAuthStore(), app = useAppStore()
const product = computed(() => route.path === '/tools/studio' ? 'studio' : 'recipes')
// 工坊公开页的“登录控制台生成”会带上条目编号，只接受目录里的编号格式，转成子页 hash。
const studioItem = computed(() => {
  const item = route.query?.item
  return product.value === 'studio' && typeof item === 'string' && /^[a-z0-9-]{1,80}$/.test(item) ? item : ''
})
const source = computed(() => '/' + product.value + '/?embedded=1' + (studioItem.value ? '#item=' + studioItem.value : ''))
const frame = ref<HTMLIFrameElement>(), loaded = ref(false)
let nonce = '', epoch = 0, observer: MutationObserver | undefined
const connections = new WorkspaceConnections({
  list: (page, signal) => keysAPI.list(page, 100, { status: 'active' }, { signal }),
  models: loadModels,
  base: () => gatewayBase(app.apiBaseUrl || '', buildGatewayUrl('/v1'))
})
const theme = () => document.documentElement.classList.contains('dark') ? 'dark' : 'light'
function send(type: string, details: Record<string, unknown> = {}) {
  frame.value?.contentWindow?.postMessage({ type, nonce, ...details }, location.origin)
}
function initialize() {
  connections.clear(); epoch++
  nonce = Array.from(crypto.getRandomValues(new Uint8Array(32)), byte => byte.toString(16).padStart(2, '0')).join('')
  loaded.value = true
  send('mofa-host-init', { theme: theme() })
}
async function receive(event: MessageEvent) {
  if (!acceptsWorkspaceMessage(event, frame.value?.contentWindow, location.origin, nonce)) return
  const { id, action, payload } = event.data
  if (!auth.isAuthenticated) { connections.clear(); send('mofa-host-session-ended'); return }
  if (action === 'theme') {
    if (!['dark', 'light'].includes(payload.theme)) return
    document.documentElement.classList.toggle('dark', payload.theme === 'dark')
    localStorage.setItem('theme', payload.theme)
    send('mofa-host-response', { id, result: null })
    return
  }
  if (action === 'family') { send('mofa-host-response', { id, result: null }); void router.push('/family'); return }
  if (action === 'studio-submit' && product.value === 'studio') { send('mofa-host-response', { id, result: null }); void router.push('/tools/studio/submit'); return }
  if (action === 'manage-keys') { send('mofa-host-response', { id, result: null }); void router.push('/keys'); return }
  // 工坊只开放生图连接；文字连接仍只给配方。
  if (product.value === 'studio' && payload.kind !== 'image') { send('mofa-host-response', { id, error: '魔法工坊只支持生图连接。' }); return }
  const started = epoch
  try {
    const result = await connections.execute(action, payload)
    if (started === epoch && auth.isAuthenticated) send('mofa-host-response', { id, result })
  } catch (problem) {
    if (started === epoch) send('mofa-host-response', { id, error: problem instanceof Error ? problem.message : '配置读取失败，请重试。' })
  }
}
watch(product, () => { loaded.value = false; nonce = ''; epoch++; connections.clear() })
watch(() => auth.isAuthenticated, loggedIn => { if (!loggedIn) { epoch++; connections.clear(); send('mofa-host-session-ended') } })
onMounted(() => {
  window.addEventListener('message', receive)
  observer = new MutationObserver(() => send('mofa-host-theme', { theme: theme() }))
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
})
onUnmounted(() => { epoch++; connections.clear(); observer?.disconnect(); window.removeEventListener('message', receive) })
</script>
