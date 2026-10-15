<template>
  <button type="button" data-testid="group-access-policy" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400" @click="open">
    <Icon name="clock" size="sm" />
    <span class="text-xs">{{ text('action') }}</span>
  </button>
  <BaseDialog v-if="show" :show="show" :title="`${text('title')} · ${group.name}`" width="wide" :close-on-escape="!saving" :show-close-button="!saving" @close="close">
    <div v-if="loading" class="py-8 text-center text-sm text-gray-500" role="status">{{ text('loading') }}</div>
    <div v-else-if="!policy" class="space-y-4">
      <p role="alert" class="text-sm text-red-600">{{ error }}</p>
      <button type="button" class="btn btn-primary" @click="load">{{ text('retry') }}</button>
    </div>
    <form v-else class="space-y-5" @submit.prevent="save">
      <p class="text-sm text-gray-500 dark:text-gray-400">{{ text('description') }}</p>
      <fieldset :disabled="saving" class="space-y-5 disabled:opacity-60">
        <section class="space-y-3 rounded-xl border border-gray-200 p-4 dark:border-dark-600">
          <div class="flex items-center justify-between gap-4">
            <span class="font-medium">{{ text('monthly') }}</span>
            <Toggle v-model="policy.monthly_recharge_enabled" :aria-label="text('monthly')" data-testid="monthly-toggle" />
          </div>
          <p class="text-xs leading-relaxed text-gray-500 dark:text-gray-400">{{ text('monthlyHint') }}</p>
          <label v-if="policy.monthly_recharge_enabled" class="block space-y-2 text-sm">
            <span>{{ text('amount') }}</span>
            <input v-model.number="policy.monthly_recharge_min" data-testid="monthly-min" type="number" min="0.01" max="1000000000000" step="0.01" required class="input w-full" />
          </label>
        </section>
        <section class="space-y-3 rounded-xl border border-gray-200 p-4 dark:border-dark-600">
          <div class="flex items-center justify-between gap-4">
            <span class="font-medium">{{ text('balance') }}</span>
            <Toggle v-model="policy.balance_enabled" :aria-label="text('balance')" data-testid="balance-toggle" />
          </div>
          <p class="text-xs leading-relaxed text-gray-500 dark:text-gray-400">{{ text('balanceHint') }}</p>
          <label v-if="policy.balance_enabled" class="block space-y-2 text-sm">
            <span>{{ text('amount') }}</span>
            <input v-model.number="policy.balance_min" data-testid="balance-min" type="number" min="0.01" max="1000000000000" step="0.01" required class="input w-full" />
          </label>
        </section>
        <p class="rounded-lg bg-primary-50 p-3 text-xs leading-relaxed text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">{{ text('refundHint') }}</p>
        <section class="space-y-3 rounded-xl border border-gray-200 p-4 dark:border-dark-600">
          <div class="flex items-center justify-between gap-4">
            <span class="font-medium">{{ text('schedule') }}</span>
            <Toggle v-model="rateScheduleEnabled" :aria-label="text('schedule')" data-testid="schedule-toggle" />
          </div>
          <p class="text-xs leading-relaxed text-gray-500 dark:text-gray-400">{{ text('scheduleHint') }}</p>
          <template v-if="rateScheduleEnabled">
            <div v-for="(period, index) in ratePeriods" :key="index" class="space-y-2 rounded-lg border border-gray-100 p-3 dark:border-dark-600" data-testid="rate-period">
              <label class="block space-y-1 text-sm">
                <span>{{ text('periodStart') }}</span>
                <input v-model="period.starts_at" type="datetime-local" step="1" required class="input w-full" :data-testid="`period-start-${index}`" />
              </label>
              <label class="block space-y-1 text-sm">
                <span>{{ text('periodEnd') }}</span>
                <input v-model="period.ends_at" type="datetime-local" step="1" required class="input w-full" :data-testid="`period-end-${index}`" />
              </label>
              <label class="block space-y-1 text-sm">
                <span>{{ text('periodRate') }}</span>
                <input v-model.number="period.rate_multiplier" type="number" min="0" max="1000" step="any" required class="input w-full" :data-testid="`period-rate-${index}`" />
              </label>
              <button type="button" class="btn text-sm" @click="ratePeriods.splice(index, 1)">{{ text('removePeriod') }}</button>
            </div>
            <div class="flex flex-wrap gap-2">
              <button type="button" class="btn text-sm" :disabled="ratePeriods.length >= 32" data-testid="add-period" @click="ratePeriods.push({ starts_at: '', ends_at: '', rate_multiplier: 0.1 })">{{ text('addPeriod') }}</button>
              <button type="button" class="btn text-sm" data-testid="seed-periods" @click="seedPeriods">{{ text('examplePeriods') }}</button>
            </div>
          </template>
        </section>
        <section class="space-y-3 rounded-xl border border-gray-200 p-4 dark:border-dark-600">
          <div class="flex items-center justify-between gap-4">
            <span class="font-medium">{{ text('expiry') }}</span>
            <Toggle v-model="expiryEnabled" :disabled="group.subscription_type === 'subscription' && !expiryEnabled" :aria-label="text('expiry')" data-testid="expiry-toggle" />
          </div>
          <p class="text-xs leading-relaxed text-gray-500 dark:text-gray-400">{{ text('expiryHint') }}</p>
          <template v-if="expiryEnabled">
            <label class="block space-y-2 text-sm">
              <span>{{ text('expiresAt') }}</span>
              <input v-model="expiryInput" data-testid="expires-at" type="datetime-local" step="1" required class="input w-full" />
            </label>
            <label class="block space-y-2 text-sm">
              <span>{{ text('fallback') }}</span>
              <select v-model="fallbackId" data-testid="fallback-group" required class="input w-full">
                <option :value="null" disabled>{{ text('selectFallback') }}</option>
                <option v-if="fallbackId && !fallbackGroups.some(candidate => candidate.id === fallbackId)" :value="fallbackId" disabled>#{{ fallbackId }} · {{ text('unavailable') }}</option>
                <option v-for="candidate in fallbackGroups" :key="candidate.id" :value="candidate.id">{{ candidate.name }} · #{{ candidate.id }}</option>
              </select>
            </label>
            <p class="text-xs leading-relaxed text-gray-500 dark:text-gray-400">{{ text('fallbackHint') }}</p>
          </template>
        </section>
      </fieldset>
      <p v-if="error" role="alert" class="text-sm text-red-600">{{ error }}</p>
      <div class="flex justify-end gap-3 border-t border-gray-100 pt-4 dark:border-dark-600">
        <button type="button" class="btn" :disabled="saving" @click="close">{{ text('close') }}</button>
        <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? text('saving') : text('save') }}</button>
      </div>
    </form>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { AdminGroup } from '@/types'
import { useAppStore } from '@/stores/app'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import { fromBeijingDateTime, getGroupAccessPolicy, normalizeGroupRateSchedule, saveGroupAccessPolicy, toBeijingDateTime, validGroupThreshold, type GroupAccessPolicy, type GroupRatePeriod } from './api'
import { groupAccessMessages } from './messages'

const props = defineProps<{ group: AdminGroup }>()
const emit = defineEmits<{ saved: [] }>()
const { locale } = useI18n()
const text = (key: keyof typeof groupAccessMessages.zh) => groupAccessMessages[(locale?.value ?? 'zh').startsWith('zh') ? 'zh' : 'en'][key]
const appStore = useAppStore()
const show = ref(false)
const loading = ref(false)
const saving = ref(false)
const policy = ref<GroupAccessPolicy | null>(null)
const candidates = ref<AdminGroup[]>([])
const error = ref('')
const expiryEnabled = ref(false)
const expiryInput = ref('')
const fallbackId = ref<number | null>(null)
const rateScheduleEnabled = ref(false)
const ratePeriods = ref<GroupRatePeriod[]>([])
let requestVersion = 0

const fallbackGroups = computed(() => candidates.value.filter(candidate => candidate.id !== props.group.id && candidate.status === 'active' && candidate.platform === props.group.platform && candidate.subscription_type !== 'subscription'))

async function load() {
  const version = ++requestVersion
  const groupId = props.group.id
  loading.value = true
  policy.value = null
  error.value = ''
  try {
    const [loaded, groups] = await Promise.all([getGroupAccessPolicy(groupId), adminAPI.groups.getAll(props.group.platform)])
    if (version !== requestVersion || !show.value || props.group.id !== groupId) return
    policy.value = loaded
    candidates.value = groups
    expiryEnabled.value = loaded.expires_at !== null
    expiryInput.value = toBeijingDateTime(loaded.expires_at)
    fallbackId.value = loaded.fallback_group_id
    rateScheduleEnabled.value = loaded.rate_schedule_enabled ?? false
    ratePeriods.value = (loaded.rate_schedule ?? []).map(period => ({ ...period, starts_at: toBeijingDateTime(period.starts_at), ends_at: toBeijingDateTime(period.ends_at) }))
  } catch {
    if (version === requestVersion) error.value = text('loadFailed')
  } finally {
    if (version === requestVersion) loading.value = false
  }
}

function open() {
  show.value = true
  void load()
}

function seedPeriods() {
  const date = (ratePeriods.value[0]?.starts_at || toBeijingDateTime(new Date().toISOString())).slice(0, 10)
  ratePeriods.value = [
    { starts_at: `${date}T00:00:00`, ends_at: `${date}T18:00:00`, rate_multiplier: 0.1 },
    { starts_at: `${date}T18:00:00`, ends_at: `${date}T20:30:00`, rate_multiplier: 0.15 },
    { starts_at: `${date}T20:30:00`, ends_at: `${date}T23:00:00`, rate_multiplier: 0.1 },
  ]
  if (props.group.subscription_type !== 'subscription') {
    expiryEnabled.value = true
    expiryInput.value = `${date}T23:00:00`
  }
}

function close() {
  if (saving.value) return
  show.value = false
  requestVersion++
}

function serverError(cause: unknown): string {
  const message = (cause as { response?: { data?: { message?: unknown } } })?.response?.data?.message
  return typeof message === 'string' && message ? message : text('saveFailed')
}

async function save() {
  if (!policy.value || loading.value || saving.value) return
  const current = policy.value
  if (!validGroupThreshold(current.monthly_recharge_min, current.monthly_recharge_enabled) || !validGroupThreshold(current.balance_min, current.balance_enabled)) {
    error.value = text('invalidAmount')
    return
  }
  const expiry = expiryEnabled.value ? fromBeijingDateTime(expiryInput.value) : null
  const schedule = normalizeGroupRateSchedule(ratePeriods.value)
  if (!schedule || (rateScheduleEnabled.value && !schedule.length) || (rateScheduleEnabled.value && expiry && schedule.some(period => Date.parse(period.ends_at) > Date.parse(expiry)))) {
    error.value = text('invalidSchedule')
    return
  }
  if (expiryEnabled.value && (!expiry || !fallbackGroups.value.some(candidate => candidate.id === fallbackId.value))) {
    error.value = text('invalidExpiry')
    return
  }
  saving.value = true
  error.value = ''
  const version = requestVersion
  const groupId = props.group.id
  try {
    await saveGroupAccessPolicy(groupId, { ...current, group_id: groupId, expires_at: expiry, fallback_group_id: expiryEnabled.value ? fallbackId.value : null, rate_schedule_enabled: rateScheduleEnabled.value, rate_schedule: schedule })
    if (version !== requestVersion || props.group.id !== groupId) return
    appStore.showSuccess(text('saved'))
    emit('saved')
    show.value = false
    requestVersion++
  } catch (cause) {
    if (version === requestVersion) error.value = serverError(cause)
  } finally {
    saving.value = false
  }
}

watch(() => props.group.id, () => { if (show.value) void load() })
onUnmounted(() => { requestVersion++ })
</script>
