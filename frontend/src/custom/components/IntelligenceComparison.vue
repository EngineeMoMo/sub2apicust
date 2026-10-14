<template>
  <div class="mofa-intel-comparison">
    <p class="mofa-intel-scale-note">{{ copy.scaleNote }}</p>
    <table class="mofa-intel-compare-table">
      <caption class="sr-only">{{ copy.chart }} · {{ copy.chartHint }}</caption>
      <thead>
        <tr>
          <th scope="col">{{ copy.model }}</th>
          <th v-for="metric in metrics" :key="metric.key" scope="col">
            {{ metric.label }}<small>{{ metric.direction }}</small>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="point in visiblePoints" :key="JSON.stringify([point.model, point.effort])" :class="{ 'is-small-sample': point.state === 'unavailable' || lowCoverage(point) }">
          <th scope="row">
            <div class="mofa-intel-model">
              <span class="mofa-intel-model-icon"><ModelIcon :model="point.model" size="24px" aria-hidden="true" /></span>
              <div>
                <span>{{ modelLabel(point.model) }}</span>
                <small>{{ point.model }} / {{ point.effort }}</small>
                <small>{{ copy.coverage }}：{{ point.coverage === null ? '—' : `${point.coverage}/${requiredTasks}` }} · {{ copy.runs }}：{{ point.graded_runs ?? '—' }}</small>
                <span v-for="flag in pointFlags(point, copy)" :key="flag" class="mofa-intel-sample">{{ flag }}</span>
              </div>
            </div>
          </th>
          <td v-for="metric in metrics" :key="metric.key" :class="`mofa-intel-metric-${metric.key}`">
            <span class="mofa-intel-mobile-metric" aria-hidden="true">{{ metric.label }}</span>
            <strong>{{ metric.format(point[metric.key]) }}</strong>
            <span class="mofa-intel-compare-bar" aria-hidden="true">
              <span :style="{ width: `${(point[metric.key] ?? 0) / maxima[metric.key] * 100}%` }" />
            </span>
            <small v-if="metric.samples && point.state !== 'unavailable'">{{ sampleText(point[metric.samples], point.statistics_samples) }}</small>
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
import { formatMinutes, formatPrice, formatScore, lowCoverage, modelLabel, pointFlags, type EfficiencyMetric, type IntelligencePoint } from '@/custom/intelligence/data'

const props = defineProps<{ points: IntelligencePoint[]; requiredTasks: number; scoreMax: number }>()
const { locale } = useI18n()
const copy = computed(() => intelligenceCopy[locale.value.startsWith('zh') ? 'zh' : 'en'])
const PAGE_SIZE = 12
const page = ref(1)
const pageCount = computed(() => Math.max(1, Math.ceil(props.points.length / PAGE_SIZE)))
const visiblePoints = computed(() => props.points.slice((page.value - 1) * PAGE_SIZE, page.value * PAGE_SIZE))
const metrics = computed<{ key: EfficiencyMetric; label: string; direction: string; format: (value: number | null) => string; samples?: 'duration_samples' | 'price_samples' }[]>(() => [
  { key: 'score', label: copy.value.score, direction: copy.value.higher, format: formatScore },
  { key: 'average_minutes', label: copy.value.time, direction: copy.value.lower, format: value => formatMinutes(value, copy.value.minutes), samples: 'duration_samples' },
  { key: 'average_price_usd', label: copy.value.cost, direction: copy.value.lower, format: formatPrice, samples: 'price_samples' }
])
// 雷达分按满分缩放；耗时、费用按整组（跨分页）最大值缩放。
const maxima = computed(() => ({
  score: props.scoreMax,
  average_minutes: Math.max(Number.EPSILON, ...props.points.map(point => point.average_minutes ?? 0)),
  average_price_usd: Math.max(Number.EPSILON, ...props.points.map(point => point.average_price_usd ?? 0))
}))
watch(() => props.points, () => { page.value = 1 })
function sampleText(known: number | null, total: number | null) {
  return known === null || total === null ? copy.value.noStatistics : `${copy.value.knownSamples} ${known}/${total}`
}
</script>
