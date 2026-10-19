<template>
  <AppLayout>
    <template #page-actions>
      <button class="btn btn-primary" type="button" :disabled="loading" @click="refresh">
        <Icon name="refresh" size="sm" aria-hidden="true" />{{ loading ? copy.refreshing : copy.refresh }}
      </button>
    </template>
    <div class="mofa-intel" :aria-busy="loading">
      <section class="mofa-intel-overview">
        <div class="mofa-intel-intro">
          <IntelligenceIcon class="mofa-intel-emblem" />
          <div><h2>{{ copy.intro }}</h2><p>{{ copy.summary }}</p></div>
        </div>
        <dl class="mofa-intel-stats">
          <div><dt>{{ copy.models }}</dt><dd>{{ snapshot ? modelCount : '—' }}</dd></div>
          <div><dt>{{ copy.points }}</dt><dd>{{ snapshot?.points.length ?? '—' }}</dd></div>
          <div><dt>{{ copy.samples }}</dt><dd>{{ snapshot ? weakCount : '—' }}</dd></div>
        </dl>
      </section>
      <div class="mofa-intel-freshness" aria-live="polite">
        <span>{{ copy.updated }}：<time :datetime="snapshot?.checkedAt">{{ date(snapshot?.checkedAt) }}</time></span>
        <span>{{ copy.fetched }}：{{ date(fetchedAt) }}</span>
      </div>
      <p v-if="failed" class="mofa-intel-alert" role="alert">{{ snapshot ? copy.retained : copy.error }}</p>
      <p v-else-if="snapshot?.refreshFailed" class="mofa-intel-alert" role="status">{{ copy.sourceFailed }}</p>
      <p v-if="snapshot && unreadCount && !snapshot.refreshFailed" class="mofa-intel-alert" role="status">{{ copy.partial }} · {{ unreadCount }} / {{ snapshot.points.length }}</p>
      <div v-if="!snapshot" class="mofa-intel-empty" role="status">
        <IntelligenceIcon class="mofa-intel-emblem" /><p>{{ loading ? copy.loading : copy.error }}</p>
        <button v-if="!loading" class="btn btn-secondary" type="button" @click="refresh">{{ copy.refresh }}</button>
      </div>
      <template v-else>
        <div class="mofa-intel-toolbar">
          <label class="mofa-intel-search"><span class="sr-only">{{ copy.search }}</span><Icon name="search" size="sm" aria-hidden="true" /><input v-model="search" type="search" :placeholder="copy.search" /></label>
          <label class="mofa-intel-check"><input v-model="robust" type="checkbox" />{{ copy.robust }}</label>
          <div class="mofa-intel-switch" role="group" :aria-label="copy.title">
            <button type="button" :aria-pressed="view === 'matrix'" @click="view = 'matrix'">{{ copy.matrix }}</button>
            <button type="button" :aria-pressed="view === 'chart'" @click="view = 'chart'">{{ copy.chart }}</button>
          </div>
        </div>
        <section class="mofa-intel-panel">
          <header class="mofa-intel-panel-heading"><div><h2>{{ view === 'matrix' ? copy.matrix : copy.chart }}</h2><p>{{ view === 'matrix' ? copy.matrixHint : copy.chartHint }}</p></div><span>{{ filtered.length }} / {{ snapshot.points.length }}</span></header>
          <div v-if="view === 'matrix'" class="mofa-intel-sort-controls">
            <label class="mofa-intel-select">
              <span>{{ copy.sort }}</span>
              <select v-model="matrixSort">
                <option value="source">{{ copy.sourceOrder }}</option>
                <option value="name">{{ copy.nameOrder }}</option>
                <option v-for="effort in efforts" :key="effort" :value="`effort:${effort}`">{{ effortName(effort) }} · {{ copy.scoreOrder }}</option>
              </select>
            </label>
            <p>{{ copy.orderHint }}</p>
          </div>
          <div v-if="!filtered.length" class="mofa-intel-empty"><p>{{ copy.empty }}</p><button class="btn btn-secondary" type="button" @click="reset">{{ copy.reset }}</button></div>
          <div v-else-if="view === 'matrix'" class="mofa-intel-table-wrap" tabindex="0" role="region" :aria-label="copy.matrix">
            <table class="mofa-intel-matrix">
              <caption class="sr-only">{{ copy.matrixHint }}</caption>
              <thead><tr><th scope="col">{{ copy.model }}</th><th v-for="effort in efforts" :key="effort" scope="col">{{ effortName(effort) }}<small>{{ effort }}</small></th></tr></thead>
              <tbody><tr v-for="row in rows" :key="row.key">
                <th scope="row"><div class="mofa-intel-model"><span class="mofa-intel-model-icon"><ModelIcon :model="row.model" size="24px" aria-hidden="true" /></span><span>{{ modelLabel(row.model) }}<small>{{ row.model }}</small></span></div></th>
                <td v-for="effort in efforts" :key="effort">
                  <div v-if="row.cells[effort]" class="mofa-intel-score" :class="{ 'is-small-sample': weak(row.cells[effort]) }">
                    <strong><span class="sr-only">{{ copy.score }} </span>{{ formatScore(row.cells[effort].score) }}</strong>
                    <span v-if="row.cells[effort].score !== null" class="mofa-intel-bar" aria-hidden="true"><span :style="{ width: `${(row.cells[effort].score ?? 0) / snapshot.scoreMax * 100}%` }" /></span>
                    <small><span class="mofa-intel-pair">{{ copy.coverageShort }} {{ coverageText(row.cells[effort]) }}</span> · <span class="mofa-intel-pair">{{ copy.runsShort }} {{ row.cells[effort].graded_runs ?? '—' }}</span></small>
                    <small class="mofa-intel-price"><span class="sr-only">{{ copy.cost }} </span>{{ formatPrice(row.cells[effort].average_price_usd) }}</small>
                    <small class="mofa-intel-time"><span class="sr-only">{{ copy.time }} </span>{{ formatMinutes(row.cells[effort].average_minutes, copy.minutes) }}</small>
                    <span v-for="flag in pointFlags(row.cells[effort], copy)" :key="flag" class="mofa-intel-sample">{{ flag }}</span>
                  </div>
                  <span v-else class="mofa-intel-missing" :aria-label="copy.missing">—</span>
                </td>
              </tr></tbody>
            </table>
          </div>
          <div v-else class="mofa-intel-chart-body">
            <div class="mofa-intel-compare-controls">
              <label class="mofa-intel-select">
                <span>{{ copy.compareEffort }}</span>
                <select v-model="compareEffort">
                  <option value="">{{ copy.allEfforts }}</option>
                  <option v-for="effort in efforts" :key="effort" :value="effort">{{ effortName(effort) }} / {{ effort }}</option>
                </select>
              </label>
              <label class="mofa-intel-select">
                <span>{{ copy.compareSort }}</span>
                <select v-model="compareMetric">
                  <option value="score">{{ copy.scoreOrder }}</option>
                  <option value="average_minutes">{{ copy.timeOrder }}</option>
                  <option value="average_price_usd">{{ copy.costOrder }}</option>
                </select>
              </label>
            </div>
            <IntelligenceComparison v-if="comparisonPoints.length" :points="comparisonPoints" :required-tasks="snapshot.requiredTasks" :score-max="snapshot.scoreMax" />
            <div v-else class="mofa-intel-empty">
              <p>{{ copy.empty }}</p>
              <button class="btn btn-secondary" type="button" @click="compareEffort = ''">{{ copy.allEfforts }}</button>
            </div>
          </div>
        </section>
      </template>
      <aside class="mofa-intel-notes"><p>{{ copy.priceNote }}</p><p>{{ copy.note }}</p><p>{{ copy.sampleNote }}</p></aside>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import IntelligenceIcon from '@/custom/components/IntelligenceIcon.vue'
import IntelligenceComparison from '@/custom/components/IntelligenceComparison.vue'
import { intelligenceCopy } from '@/custom/intelligence/copy'
import {
  EFFORTS, formatMinutes, formatPrice, formatScore, lowCoverage, matrixRows, modelLabel, pointFlags, sortEfficiencyPoints, sortMatrixRows,
  type EfficiencyMetric, type IntelligencePoint, type MatrixSort
} from '@/custom/intelligence/data'
import { useIntelligence } from '@/custom/intelligence/useIntelligence'

const { locale } = useI18n()
const copy = computed(() => intelligenceCopy[locale.value.startsWith('zh') ? 'zh' : 'en'])
const { snapshot, loading, failed, fetchedAt, refresh } = useIntelligence()
const search = ref('')
const robust = ref(false)
const view = ref<'matrix' | 'chart'>('matrix')
const matrixSort = ref<MatrixSort>('source')
const compareEffort = ref('high')
const compareMetric = ref<EfficiencyMetric>('score')
const modelCount = computed(() => new Set(snapshot.value?.points.map(point => point.model)).size)
const weakCount = computed(() => snapshot.value?.points.filter(lowCoverage).length ?? 0)
const unreadCount = computed(() => snapshot.value?.points.filter(point => point.state !== 'ok').length ?? 0)
const efforts = computed(() => [...EFFORTS, ...new Set(snapshot.value?.points.map(point => point.effort).filter(effort => !EFFORTS.includes(effort)))])
const filtered = computed(() => {
  const query = search.value.trim().toLowerCase()
  return (snapshot.value?.points ?? []).filter(point =>
    (point.model.toLowerCase().includes(query) || modelLabel(point.model).toLowerCase().includes(query)) &&
    (!robust.value || (point.state !== 'unavailable' && !lowCoverage(point)))
  )
})
const rows = computed(() => sortMatrixRows(matrixRows(filtered.value), matrixSort.value))
const comparisonPoints = computed(() => sortEfficiencyPoints(
  filtered.value.filter(point => !compareEffort.value || point.effort === compareEffort.value), compareMetric.value
))
function effortName(effort: string) {
  const names: Record<string, string> = { ultra: copy.value.ultra, max: copy.value.max, xhigh: copy.value.xhigh, high: copy.value.high, medium: copy.value.medium, low: copy.value.lowEffort }
  return Object.prototype.hasOwnProperty.call(names, effort) ? names[effort] : effort
}
// 读取失败或覆盖不足：分数条转为弱化色，避免少数题目的高分看起来像可靠排名。
function weak(point: IntelligencePoint) {
  return point.state === 'unavailable' || lowCoverage(point)
}
function coverageText(point: IntelligencePoint) {
  return point.coverage === null ? '—' : `${point.coverage}/${snapshot.value?.requiredTasks}`
}
function date(value: string | number | null | undefined) {
  return value == null ? '—' : new Date(value).toLocaleString(locale.value, { hour12: false })
}
function reset() { search.value = ''; robust.value = false }
</script>
