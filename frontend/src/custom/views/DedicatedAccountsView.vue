<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { dedicatedAPI, dedicatedDisplayName, effectiveStatus, remainingQuota, type DedicatedView } from '@/custom/dedicated/api'
import { dedicatedCopy } from '@/custom/dedicated/copy'

const { locale } = useI18n()
const copy = computed(() => dedicatedCopy[locale.value.startsWith('zh') ? 'zh' : 'en'])
const accounts = ref<DedicatedView[]>([])
const busy = ref(false)
const error = ref(false)
const page = ref(1)
const now = ref(Date.now())
let generation = 0
let timer: ReturnType<typeof setInterval> | undefined
function date(value: string | null) {
  if (!value || !Number.isFinite(Date.parse(value))) return copy.value.never
  return new Date(value).toLocaleString(locale.value)
}
async function load(target = page.value) {
  const current = ++generation
  busy.value = true
  error.value = false
  accounts.value = []
  try {
    const result = await dedicatedAPI.mine(target)
    if (current === generation) { accounts.value = result; page.value = target; now.value = Date.now() }
  } catch {
    if (current === generation) error.value = true
  } finally {
    if (current === generation) busy.value = false
  }
}
onMounted(() => { void load(); timer = setInterval(() => { now.value = Date.now() }, 30_000) })
onBeforeUnmount(() => { generation++; clearInterval(timer) })
</script>

<template>
  <AppLayout>
    <template #page-actions><button class="btn btn-secondary" type="button" :disabled="busy" @click="load()">{{ busy ? copy.loading : copy.refresh }}</button></template>
    <div class="mofa-dedicated" :aria-busy="busy">
      <p class="mofa-dedicated-note">{{ copy.statusHint }}</p>
      <p v-if="busy" role="status">{{ copy.loading }}</p>
      <div v-else-if="error" class="mofa-dedicated-empty" role="alert"><p>{{ copy.loadError }}</p><button class="btn btn-secondary" @click="load()">{{ copy.refresh }}</button></div>
      <div v-else-if="!accounts.length" class="mofa-dedicated-empty"><h2>{{ copy.empty }}</h2><p>{{ copy.emptyHint }}</p></div>
      <section v-for="account in accounts" :key="account.id" class="mofa-dedicated-account" :aria-labelledby="'dedicated-' + account.id">
        <header>
          <div><h2 :id="'dedicated-' + account.id">{{ account.label }}</h2><p>{{ account.platform === 'anthropic' ? 'Claude' : account.platform === 'openai' ? 'ChatGPT / Codex' : copy.title }} · {{ copy.group }} {{ dedicatedDisplayName(account.group_name, account.group_id) }}</p></div>
          <span class="mofa-dedicated-status" :data-state="effectiveStatus(account, now)">{{ copy.statuses[effectiveStatus(account, now)] || copy.unknown }}</span>
        </header>
        <dl class="mofa-dedicated-facts">
          <div><dt>{{ copy.expires }}</dt><dd>{{ date(account.expires_at) }}</dd></div>
          <div><dt>{{ copy.lastUsed }}</dt><dd>{{ date(account.last_used_at) }}</dd></div>
          <div><dt>{{ copy.sampled }}</dt><dd>{{ date(account.sampled_at) }}</dd></div>
        </dl>
        <div v-if="['available', 'rate_limited', 'unavailable'].includes(effectiveStatus(account, now))" class="mofa-dedicated-quotas">
          <div v-for="window in account.windows" :key="window.key" class="mofa-dedicated-quota">
            <div><h3>{{ copy.windows[window.key] || window.key }}</h3><strong>{{ remainingQuota(window, account.sampled_at, now) === null ? copy.unknown : copy.remaining + ' ' + remainingQuota(window, account.sampled_at, now)!.toFixed(1) + '%' }}</strong></div>
            <progress v-if="remainingQuota(window, account.sampled_at, now) !== null" :value="remainingQuota(window, account.sampled_at, now)!" max="100" :aria-label="(copy.windows[window.key] || window.key) + ' ' + copy.remaining" />
            <p>{{ copy.resets }}: {{ date(window.resets_at) }}</p>
          </div>
          <p v-if="!account.windows.length">{{ copy.noQuota }}</p>
        </div>
        <footer><p>{{ copy.keyHint }}</p><RouterLink to="/keys" class="btn btn-secondary">{{ copy.keys }}</RouterLink></footer>
      </section>
      <nav v-if="!busy && (page > 1 || accounts.length === 50)" class="mofa-dedicated-pagination" :aria-label="copy.page">
        <button type="button" class="btn btn-secondary" :disabled="page === 1" @click="load(page - 1)">{{ copy.previous }}</button><span>{{ copy.page }} {{ page }}</span><button type="button" class="btn btn-secondary" :disabled="accounts.length < 50" @click="load(page + 1)">{{ copy.next }}</button>
      </nav>
      <aside class="mofa-dedicated-note"><p>{{ copy.quotaHint }}</p><p>{{ copy.contact }}</p></aside>
    </div>
  </AppLayout>
</template>
