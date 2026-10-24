<template>
  <div class="space-y-3" data-testid="kiro-credentials-fields">
    <div v-if="!editing">
      <label class="input-label" for="kiro-account-type">{{ text.type }}</label>
      <select id="kiro-account-type" class="input" :value="accountType" @change="changeType">
        <option value="oauth">{{ text.oauth }}</option>
        <option value="apikey">{{ text.apiKey }}</option>
      </select>
    </div>
    <details :open="!editing" class="rounded-lg border border-gray-200 px-3 py-2 dark:border-dark-600" data-testid="kiro-credential-help">
      <summary class="cursor-pointer text-sm font-medium text-gray-700 dark:text-gray-200">{{ text.helpTitle }}</summary>
      <div v-if="accountType === 'apikey'" class="mt-2 space-y-2 text-sm text-gray-600 dark:text-gray-400">
        <p>{{ text.apiKeySource }}</p>
        <p>{{ text.apiKeyPlans }}</p>
        <a href="https://app.kiro.dev/settings/api-keys" target="_blank" rel="noopener noreferrer" class="text-primary-600 hover:underline dark:text-primary-400">{{ text.apiKeyLink }}</a>
      </div>
      <div v-else class="mt-2 space-y-2 text-sm text-gray-600 dark:text-gray-400">
        <p>{{ text.oauthSource }}</p>
        <p>{{ text.oauthWindows }}: <code class="break-all">%USERPROFILE%\.aws\sso\cache\kiro-auth-token.json</code></p>
        <p>{{ text.oauthUnix }}: <code class="break-all">~/.aws/sso/cache/kiro-auth-token.json</code></p>
        <p>{{ text.oauthExtra }}</p>
        <p>{{ text.oauthImportOnly }}</p>
        <a href="https://kiro.dev/docs/getting-started/authentication/" target="_blank" rel="noopener noreferrer" class="text-primary-600 hover:underline dark:text-primary-400">{{ text.oauthLink }}</a>
      </div>
    </details>
    <label class="input-label" for="kiro-usage-mode">{{ text.usageMode }}</label>
    <select id="kiro-usage-mode" class="input" :value="usageMode" :disabled="!canChangeUsageMode" @change="changeUsageMode">
      <option v-if="editing" value="preserve">{{ text.usagePreserve }}</option>
      <option value="verified">{{ text.usageVerified }}</option>
      <option value="estimated">{{ text.usageEstimated }}</option>
    </select>
    <p class="input-hint">{{ text.usageWarning }}</p>
    <label class="input-label" for="kiro-credentials-json">{{ text.credentials }}</label>
    <textarea
      id="kiro-credentials-json" class="input font-mono text-sm" rows="7" autocomplete="off"
      spellcheck="false" :value="modelValue" :placeholder="placeholder"
      @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    />
    <label class="input-label" for="kiro-credentials-file">{{ text.file }}</label>
    <input id="kiro-credentials-file" type="file" accept=".json,application/json" class="input" @change="importFile" />
    <p class="input-hint">{{ editing ? text.editHint : accountType === 'apikey' ? text.apiKeyHint : text.hint }}</p>
    <p class="input-hint">{{ text.capabilities }}</p>
    <p v-if="fileError" class="text-sm text-red-600" role="alert">{{ text.invalid }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useKiroText } from './messages'
import { parseKiroCredentials, type KiroAccountType } from './credentials'

const props = withDefaults(defineProps<{
  modelValue: string
  accountType: KiroAccountType
  editing?: boolean
}>(), { editing: false })
const emit = defineEmits<{
  'update:modelValue': [value: string]
  'update:accountType': [value: KiroAccountType]
}>()
const text = useKiroText()
const fileError = ref(false)
const canChangeUsageMode = computed(() => {
  if (!props.modelValue.trim()) return true
  try { parseKiroCredentials(props.modelValue); return true } catch { return false }
})
const usageMode = computed(() => {
  if (props.modelValue.trim()) {
    try {
      const mode = parseKiroCredentials(props.modelValue).usage_mode
      if (mode === 'verified' || mode === 'estimated') return mode
    } catch { return props.editing ? 'preserve' : 'verified' }
  }
  return props.editing ? 'preserve' : 'verified'
})
function changeUsageMode(event: Event) {
  const mode = (event.target as HTMLSelectElement).value
  const credentials = props.modelValue.trim() ? parseKiroCredentials(props.modelValue) : {}
  if (mode === 'preserve') delete credentials.usage_mode
  else credentials.usage_mode = mode
  emit('update:modelValue', Object.keys(credentials).length ? JSON.stringify(credentials, null, 2) : '')
}
const placeholder = computed(() => props.accountType === 'apikey'
  ? '{"api_key":"ksk_…","runtime_region":"us-east-1"}'
  : '{"refreshToken":"…","authMethod":"social","auth_region":"us-east-1","runtime_region":"us-east-1"}')

function changeType(event: Event) {
  emit('update:accountType', (event.target as HTMLSelectElement).value as KiroAccountType)
  emit('update:modelValue', '')
}
async function importFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  fileError.value = false
  try {
    if (file.size > 1024 * 1024) throw new Error('too_large')
    const content = await file.text()
    parseKiroCredentials(content)
    emit('update:modelValue', content)
  } catch {
    fileError.value = true
  } finally {
    input.value = ''
  }
}
</script>
