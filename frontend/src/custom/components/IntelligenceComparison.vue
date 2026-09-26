<template>
  <div class="mofa-intel-comparison">
    <p class="mofa-intel-scale-note">{{ copy.scaleNote }}</p>
    <table class="mofa-intel-compare-table">
      <caption class="sr-only">{{ copy.chart }} · {{ copy.passed }}</caption>
      <thead>
        <tr>
          <th scope="col">{{ copy.model }}</th>
          <th v-for="metric in metrics" :key="metric.key" scope="col">
            {{ metric.label }}<small>{{ metric.direction }}</small>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="point in visiblePoints" :key="JSON.stringify([point.model, point.harness, point.effort])" :class="{ 'is-small-sample': lowSample(point) }">
          <th scope="row">
            <div class="mofa-intel-model">
              <span class="mofa-intel-model-icon"><ModelIcon :model="point.model" size="24px" aria-hidden="true" /></span>
              <div>
                <span>{{ point.model }}</span>
                <small>{{ point.harness }} / {{ point.effort }}</small>
                <small>{{ copy.passed }}：{{ format(point.passed) }} / {{ format(point.valid_tasks) }}</small>
                <span v-if="lowSample(point)" class="mofa-intel-sample">{{ point.valid_tasks === null ? copy.unknown : copy.low }}</span>
              </div>
            </div>
          </th>
          <td v-for="metric in metrics" :key="metric.key" :class="`mofa-intel-metric-${metric.key}`">
            <span class="mofa-intel-mobile-metric" aria-hidden="true">{{ metric.label }}</span>
            <strong>{{ format(point[metric.key], metric.key === 'average_price_usd' ? 4 : 2) }}</strong>
            <span class="mofa-intel-compare-bar" aria-hidden="true">
              <span :style="{ width: `${(point[metric.key] ?? 0) / maxima[metric.key] * 100}%` }" />
            </span>
            <small v-if="metric.key === 'average_price_usd' && point.price_aggregation">{{ copy.aggregation }}：{{ point.price_aggregation }}</small>
          </td>
        </tr>
      </tbody>
    </table>
    <nav v-if="points.length" class="mofa-intel-pagination" :aria-label="copy.shown">
      <span aria-live="polite">{{ copy.shown }} {{ (page - 1) * PAGE_SIZE + 1 }}–{{ Math.min(page * PAGE_SIZE, points.length) }} / {{ points.length }}</span>
      <div>
        <button class="btn btn-secondary" type="button" :disabled="page <= 1" @click="page--">{{ copy.previous }}</button>
        <button class="btn btn-secondary" type="button" :disabled="page >= pageCount" @click="page++">{{ copy.next }}</button>
      </div>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import ModelIcon from '@/components/common/ModelIcon.vue'
import { intelligenceCopy } from '@/custom/intelligence/copy'
import { lowSample, type EfficiencyMetric, type IntelligencePoint } from '@/custom/intelligence/data'

const props = defineProps<{ points: IntelligencePoint[] }>()
const { locale } = useI18n()
const copy = computed(() => intelligenceCopy[locale.value.startsWith('zh') ? 'zh' : 'en'])
const PAGE_SIZE = 12
const page = ref(1)
const pageCount = computed(() => Math.max(1, Math.ceil(props.points.length / PAGE_SIZE)))
const visiblePoints = computed(() => props.points.slice((page.value - 1) * PAGE_SIZE, page.value * PAGE_SIZE))
const metrics = computed<{ key: EfficiencyMetric; label: string; direction: string }[]>(() => [
  { key: 'iq', label: copy.value.score, direction: copy.value.higher },
  { key: 'average_minutes', label: copy.value.time, direction: copy.value.lower },
  { key: 'average_price_usd', label: copy.value.cost, direction: copy.value.lower }
])
const maxima = computed(() => ({
  iq: Math.max(Number.EPSILON, ...props.points.map(point => point.iq ?? 0)),
  average_minutes: Math.max(Number.EPSILON, ...props.points.map(point => point.average_minutes ?? 0)),
  average_price_usd: Math.max(Number.EPSILON, ...props.points.map(point => point.average_price_usd ?? 0))
}))
watch(() => props.points, () => { page.value = 1 })
function format(value: number | null, maximumFractionDigits = 2) {
  return value === null ? '—' : value.toLocaleString(locale.value, { maximumFractionDigits })
}
</script>
