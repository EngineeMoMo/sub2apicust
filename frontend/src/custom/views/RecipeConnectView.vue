<template>
  <div class="mofa-recipe-connect-shell">
    <header class="mofa-recipe-connect-header">
      <RouterLink to="/dashboard" class="mofa-recipe-connect-brand"><img :src="app.siteLogo || logo" width="40" height="40" alt="魔法家族标识"><span>{{ app.siteName }}</span></RouterLink>
      <BrandThemeToggle />
    </header>
    <main class="mofa-recipe-connect" aria-labelledby="recipe-connect-title">
      <h1 id="recipe-connect-title">为魔法配方选择模型</h1>
      <p>已通过本站账户登录。选择已有密钥和模型，将配置带回配方页。</p>
      <p v-if="error" class="mofa-recipe-connect-error" role="alert">{{ error }}</p>
      <template v-if="request">
        <dl class="mofa-recipe-connect-destination"><div><dt>连接用途</dt><dd>{{ request.kind === 'image' ? '生图模型' : '文字模型' }}</dd></div><div><dt>接收配置的页面</dt><dd>{{ request.origin }}</dd></div></dl>
        <div v-if="!finished">
          <p v-if="loadingKeys" role="status">正在读取你的密钥…</p>
          <p v-else-if="!keys.length">没有可用的密钥。请先在控制台创建密钥并绑定可用分组，再回来刷新。</p>
          <form v-if="keys.length" @submit.prevent="connect">
            <label for="recipe-key">选择密钥与分组</label>
            <select id="recipe-key" v-model="keyID" :disabled="sending" @change="selectKey"><option value="">请选择密钥</option><option v-for="key in keys" :key="key.id" :value="String(key.id)">{{ key.name }} · {{ key.group?.name || '未命名分组' }}</option></select>
            <p class="mofa-recipe-connect-hint">只列出启用、未到期、未耗尽配额且分组有效的密钥，不显示密钥原文。</p>
            <p v-if="loadingModels" role="status">正在读取所选密钥的模型目录…</p>
            <label for="recipe-model">选择模型</label>
            <select id="recipe-model" v-model="model" :disabled="!models.length || loadingModels || sending"><option value="">请选择模型</option><option v-for="name in models" :key="name" :value="name">{{ name }}</option></select>
            <p class="mofa-recipe-connect-hint">目录来自所选密钥。{{ request.kind === 'image' ? '生图请选支持 Images API 的模型；目录不代表每个模型都有生图能力。' : '实际可用性与费用以分组和模型权限为准。' }}</p>
            <template v-if="request.kind === 'text'"><label for="recipe-protocol">文字接口格式</label><select id="recipe-protocol" v-model="protocol" :disabled="sending"><option value="chat">Chat Completions</option><option value="responses">Responses（Codex 等）</option></select></template>
            <p class="mofa-recipe-connect-hint">应用会把所选 API 密钥、接口地址和模型名发送到上方配方页面；账号密码与登录令牌留在本站。配置只在配方当前页面有效。</p>
            <button class="mofa-recipe-connect-primary" type="submit" :disabled="!model || !selectedKey || loadingModels || sending">{{ sending ? '正在等待配方页接收…' : '将所选配置用于魔法配方' }}</button>
          </form>
          <div class="mofa-recipe-connect-actions"><button type="button" :disabled="loadingKeys || sending" @click="refreshKeys">刷新密钥列表</button><a href="/keys" target="_blank" rel="noopener noreferrer">打开密钥管理</a></div>
        </div>
        <p v-else role="status">配置已带回魔法配方。请回到原页面，检查提示词后运行；本次连接未调用生成模型。</p>
      </template>
      <p v-else>请从魔法配方的“模型设置 → 登录魔法 API 并选择配置”进入本页。</p>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { keysAPI } from '@/api/keys'
import { buildGatewayUrl } from '@/api/url'
import type { ApiKey } from '@/types'
import BrandThemeToggle from '@/custom/components/BrandThemeToggle.vue'
import logo from '@/custom/assets/mofa-mark-flat.png'
import { connectionRequest, gatewayBase, loadModels, selectedConnection, usableKey, type RecipeConnectionRequest } from '@/custom/recipes/connect'

const route = useRoute()
const app = useAppStore()
const auth = useAuthStore()
const request = ref<RecipeConnectionRequest>()
const keys = ref<ApiKey[]>([])
const keyID = ref('')
const model = ref('')
const models = ref<string[]>([])
const protocol = ref<'chat' | 'responses'>('chat')
const base = ref('')
const error = ref('')
const loadingKeys = ref(false)
const loadingModels = ref(false)
const sending = ref(false)
const finished = ref(false)
const selectedKey = computed(() => keys.value.find(key => String(key.id) === keyID.value))
let keyLoad: AbortController | undefined
let modelLoad: AbortController | undefined
let modelTimer: ReturnType<typeof setTimeout> | undefined
let sendTimer: ReturnType<typeof setTimeout> | undefined
let alive = true

async function refreshKeys() {
  if (!request.value || !auth.isAuthenticated || sending.value) return
  modelLoad?.abort()
  modelLoad = undefined
  clearTimeout(modelTimer)
  keyLoad?.abort()
  keyLoad = new AbortController()
  const operation = keyLoad
  keys.value = []
  keyID.value = ''
  model.value = ''
  models.value = []
  loadingModels.value = false
  loadingKeys.value = true
  error.value = ''
  try {
    const all: ApiKey[] = []
    for (let page = 1; ; page++) {
      const result = await keysAPI.list(page, 100, { status: 'active' }, { signal: operation.signal })
      if (!alive || operation.signal.aborted) return
      all.push(...result.items)
      if (page >= result.pages || !result.items.length) break
    }
    keys.value = all.filter(key => usableKey(key) && !(request.value?.kind === 'image' && key.group?.allow_image_generation === false))
  } catch { if (alive && !operation.signal.aborted) error.value = '密钥列表读取失败，请检查登录状态后重试。' }
  finally { if (alive && keyLoad === operation) loadingKeys.value = false }
}

async function selectKey() {
  modelLoad?.abort()
  clearTimeout(modelTimer)
  modelLoad = new AbortController()
  const operation = modelLoad
  const key = selectedKey.value
  models.value = []
  model.value = ''
  error.value = ''
  loadingModels.value = false
  if (!key || !usableKey(key)) return
  protocol.value = key.group?.platform === 'openai' ? 'responses' : 'chat'
  loadingModels.value = true
  modelTimer = setTimeout(() => operation.abort(), 30000)
  try {
    const result = await loadModels(base.value, key.key, operation.signal)
    if (!alive || modelLoad !== operation || operation.signal.aborted) return
    models.value = result
    if (!result.length) error.value = '此密钥没有返回可选模型，请检查分组配置或更换密钥。'
  } catch (problem) {
    if (alive && modelLoad === operation) error.value = operation.signal.aborted ? '模型目录读取超时，请重试或更换密钥。' : (problem as Error).message
  } finally { if (alive && modelLoad === operation) { clearTimeout(modelTimer); loadingModels.value = false } }
}

function connect() {
  if (!request.value || !selectedKey.value || !window.opener || sending.value || !auth.isAuthenticated) return
  try {
    const config = selectedConnection(request.value, selectedKey.value, models.value, model.value, base.value, protocol.value)
    error.value = ''
    sending.value = true
    window.opener.postMessage({ type: 'mofa-recipes-connection', nonce: request.value.nonce, config }, request.value.origin)
    sendTimer = setTimeout(() => { sending.value = false; error.value = '配方页面未确认接收，请回到原页面重新发起连接。' }, 10000)
  } catch (problem) { sending.value = false; error.value = (problem as Error).message }
}

function received(event: MessageEvent) {
  if (!request.value || !sending.value || event.source !== window.opener || event.origin !== request.value.origin) return
  if (event.data?.type !== 'mofa-recipes-applied' || event.data.nonce !== request.value.nonce) return
  clearTimeout(sendTimer)
  sending.value = false
  finished.value = true
  keys.value = []
  keyID.value = ''
  models.value = []
}

onMounted(async () => {
  window.addEventListener('message', received)
  try {
    if (!window.opener) throw new Error('未找到原配方页面，请从配方页重新发起连接。')
    request.value = connectionRequest(route.query, window.location.origin, String(import.meta.env.VITE_MAGIC_RECIPES_ORIGIN || '').trim())
    await app.fetchPublicSettings()
    if (!alive) return
    base.value = gatewayBase(app.apiBaseUrl, buildGatewayUrl('/v1'))
    await refreshKeys()
  } catch (problem) { request.value = undefined; error.value = (problem as Error).message }
})
onUnmounted(() => {
  alive = false
  keyLoad?.abort()
  modelLoad?.abort()
  clearTimeout(modelTimer)
  clearTimeout(sendTimer)
  window.removeEventListener('message', received)
})
</script>
