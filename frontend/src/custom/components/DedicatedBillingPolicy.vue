<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { dedicatedAPI, dedicatedErrorCode, type DedicatedBillingPolicy } from '@/custom/dedicated/api'

const props = defineProps<{ bindingId: number }>()
const emit = defineEmits<{ close: [] }>()
const { locale } = useI18n()
const zh = computed(() => locale.value.startsWith('zh'))
const policy = ref<DedicatedBillingPolicy>()
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const saved = ref(false)
const errorElement = ref<HTMLElement>()
let generation = 0
async function load() {
  const current = ++generation
  loading.value = true
  policy.value = undefined
  error.value = ''
  saved.value = false
  try { const result = await dedicatedAPI.billingPolicy(props.bindingId); if (current === generation) policy.value = result }
  catch { if (current === generation) error.value = zh.value ? '读取失败，请刷新后重试。' : 'Could not load. Refresh and retry.' }
  finally { if (current === generation) loading.value = false }
}
async function save() {
  if (!policy.value || saving.value) return
  const current = generation
  saving.value = true
  saved.value = false
  error.value = ''
  try {
    const result = await dedicatedAPI.saveBillingPolicy(props.bindingId, { ...policy.value })
    if (current === generation) { policy.value = result; saved.value = true }
  } catch (failure) {
    if (current === generation) {
      error.value = dedicatedErrorCode(failure) === 'DEDICATED_ACCOUNT_STALE'
        ? zh.value ? '限制已被其他操作修改。请刷新读取最新值后再保存。' : 'The limits changed. Refresh before saving.'
        : zh.value ? '保存失败，请检查字段范围后重试。' : 'Could not save. Check the field ranges and retry.'
      await nextTick()
      errorElement.value?.focus()
    }
  } finally { if (current === generation) saving.value = false }
}
watch(() => props.bindingId, () => { saving.value = false; void load() }, { immediate: true })
onBeforeUnmount(() => { generation++ })
</script>

<template>
  <form class="mofa-dedicated-form" @submit.prevent="save">
    <h2>{{ zh ? '包号共享使用限制' : 'Shared assignment limits' }} #{{ bindingId }}</h2>
    <p>{{ zh ? '所有成员、专属组和密钥共用以下限制。有效包号使用实扣0，普通分组仍正常计费。修改限制、续期或新建密钥不会清空用量。' : 'All members, linked groups and keys share these limits. Valid dedicated usage is billed at zero; public groups retain normal billing. Editing, renewal and new keys do not reset usage.' }}</p>
    <p>{{ zh ? '模型范围沿用分组模型白名单。每日请求次数按UTC零点重置；已准入的请求包括上游失败请求，不退次数。' : 'Group model allowlists still apply. Daily requests reset at midnight UTC. Admitted requests, including upstream failures, consume a request.' }}</p>
    <p v-if="loading" role="status">{{ zh ? '正在读取…' : 'Loading…' }}</p>
    <p v-if="error" ref="errorElement" class="mofa-dedicated-feedback" tabindex="-1" role="alert">{{ error }}</p>
    <p v-if="saved" role="status">{{ zh ? '使用限制已保存。' : 'Limits saved.' }}</p>
    <fieldset v-if="policy" :disabled="saving">
      <div class="mofa-dedicated-form-grid">
        <label class="mofa-dedicated-field"><span>{{ zh ? '共享并发（1–200）' : 'Shared concurrency (1–200)' }}</span><input v-model.number="policy.concurrency_limit" type="number" min="1" max="200" step="1" required /></label>
        <label class="mofa-dedicated-field"><span>{{ zh ? '每分钟请求数（1–10000）' : 'Requests per minute (1–10000)' }}</span><input v-model.number="policy.rpm_limit" type="number" min="1" max="10000" step="1" required /></label>
        <label class="mofa-dedicated-field"><span>{{ zh ? '每日请求上限（0不限制）' : 'Daily requests (0 disables limit)' }}</span><input v-model.number="policy.daily_request_limit" type="number" min="0" max="1000000" step="1" required /></label>
        <label class="mofa-dedicated-field"><span>{{ zh ? '请求体上限（字节）' : 'Maximum request body (bytes)' }}</span><input v-model.number="policy.max_body_bytes" type="number" min="1024" max="33554432" step="1" required /><small>{{ zh ? '默认2097152字节（2 MiB），不等于Token上限。' : 'Default: 2097152 bytes (2 MiB). This is not a token limit.' }}</small></label>
      </div>
      <label class="mofa-dedicated-restore"><input v-model="policy.allow_images" type="checkbox" />{{ zh ? '允许包号生图（仍取决于账号和分组权限）' : 'Allow image generation (account and group permissions still apply)' }}</label>
      <div class="mofa-dedicated-actions"><button type="submit" class="btn btn-primary">{{ saving ? zh ? '正在保存…' : 'Saving…' : zh ? '保存使用限制' : 'Save limits' }}</button><button type="button" class="btn btn-secondary" @click="load">{{ zh ? '刷新限制' : 'Refresh limits' }}</button><button type="button" class="btn btn-secondary" @click="emit('close')">{{ zh ? '关闭' : 'Close' }}</button></div>
    </fieldset>
    <div v-else-if="!loading" class="mofa-dedicated-actions"><button type="button" class="btn btn-secondary" @click="load">{{ zh ? '重试' : 'Retry' }}</button><button type="button" class="btn btn-secondary" @click="emit('close')">{{ zh ? '关闭' : 'Close' }}</button></div>
  </form>
</template>
