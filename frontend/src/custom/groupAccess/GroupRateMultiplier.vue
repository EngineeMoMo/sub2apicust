<template>
  <div class="text-sm text-gray-700 dark:text-gray-300" :title="labels.hint">
    <span data-testid="current-rate">{{ known ? `${currentRate}x` : '—' }}</span>
    <div v-if="active || !known" class="text-xs text-gray-500 dark:text-gray-400" data-testid="rate-detail">
      {{ active ? labels.promotion : labels.unavailable }} · {{ labels.base }} {{ group.rate_multiplier }}x
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AdminGroup } from '@/types'
import { useGroupRateDisplay } from './rateDisplay'

const props = defineProps<{ group: Pick<AdminGroup, 'id' | 'rate_multiplier'> }>()
const display = useGroupRateDisplay()
const { locale } = useI18n()
const known = computed(() => !display || !display.enabled.value || display.ready.value)
const active = computed(() => {
  if (!display?.enabled.value || !display.ready.value) return undefined
  const schedule = display.schedules.value[props.group.id]
  if (!schedule?.rate_schedule_enabled) return undefined
  return schedule.rate_schedule?.find(period => Date.parse(period.starts_at) <= display.now.value && display.now.value < Date.parse(period.ends_at))
})
const currentRate = computed(() => active.value?.rate_multiplier ?? props.group.rate_multiplier)
const labels = computed(() => locale.value.startsWith('zh')
  ? { promotion: '分时福利', base: '基础', unavailable: '活动倍率未读取', hint: '当前分组倍率；用户专属倍率、高峰及独立图片计费仍按原规则执行。' }
  : { promotion: 'Promotion', base: 'Base', unavailable: 'Schedule unavailable', hint: 'Current group rate; user overrides, peak and independent image pricing still apply.' })
</script>
