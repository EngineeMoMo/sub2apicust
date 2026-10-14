<template>
  <div v-if="tools && Object.keys(tools).length" data-testid="tool-costs" class="space-y-1">
    <div v-for="(tool, name) in tools" :key="name" class="space-y-1">
      <div class="flex items-center justify-between gap-4">
        <span class="text-gray-400">{{ name === 'web_search' ? t('usage.webSearchTool') : name }} × {{ tool.calls }}</span>
        <span class="font-medium text-white">${{ tool.unit_price.toFixed(8) }} {{ t('usage.perToolCall') }}</span>
      </div>
      <div class="flex items-center justify-between gap-4">
        <span class="text-gray-400">{{ t('usage.toolOriginalCost') }}</span>
        <span class="font-medium text-white">${{ tool.total_cost.toFixed(8) }}</span>
      </div>
      <div class="flex items-center justify-between gap-4">
        <span class="text-gray-400">{{ t('usage.toolActualCost') }}</span>
        <span class="font-medium text-green-400">${{ tool.actual_cost.toFixed(8) }}</span>
      </div>
    </div>
    <p class="text-xs text-gray-400">{{ t('usage.toolIncluded') }}</p>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { UsageLog } from '@/types'

defineProps<{ tools?: UsageLog['tool_usage'] }>()
const { t } = useI18n()
</script>
