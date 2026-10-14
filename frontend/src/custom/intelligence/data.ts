// 本站后端代读的 Codex Radar「雷达Bench」公开汇总（源站接口无 CORS，浏览器不直连源站）。
export const INTELLIGENCE_ENDPOINT = '/intelligence'
export const REFRESH_INTERVAL = 10 * 60 * 1000 // 与服务端缓存周期一致
export const REQUEST_TIMEOUT = 60_000 // 首次读取源站需逐个档位查询，服务端最多等 45 秒
export const LOW_COVERAGE_THRESHOLD = 40 // 与源站页面「数据不足」阈值一致
export const EFFORTS = ['ultra', 'max', 'xhigh', 'high', 'medium', 'low']
const STATES = ['ok', 'retained', 'unavailable'] as const

export interface IntelligencePoint {
  model: string
  effort: string
  state: (typeof STATES)[number]
  score: number | null
  coverage: number | null
  score_status: string
  graded_runs: number | null
  average_price_usd: number | null
  price_samples: number | null
  average_minutes: number | null
  duration_samples: number | null
  statistics_samples: number | null
  statistics_state: string
  read_at: string
}

export interface IntelligenceSnapshot {
  benchmark: string
  catalogVersion: string
  scoreMax: number
  requiredTasks: number
  checkedAt: string
  refreshFailed: boolean
  points: IntelligencePoint[]
}

function record(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('invalid-data')
  return value as Record<string, unknown>
}

function text(value: unknown, required = true): string {
  if (typeof value !== 'string' || value.length > 200 || (required && !value.trim())) throw new Error('invalid-data')
  return value
}

function metric(value: unknown): number | null {
  if (value === null || value === undefined) return null
  if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) throw new Error('invalid-data')
  return value
}

function count(value: unknown): number | null {
  const parsed = metric(value)
  if (parsed !== null && !Number.isInteger(parsed)) throw new Error('invalid-data')
  return parsed
}

export function parseIntelligence(value: unknown): IntelligenceSnapshot {
  const source = record(value)
  const scoreMax = metric(source.score_max)
  const requiredTasks = count(source.required_tasks)
  const refreshFailed = source.refresh_failed
  if (!scoreMax || !requiredTasks || typeof refreshFailed !== 'boolean') throw new Error('invalid-schema')
  const checkedAt = text(source.checked_at)
  if (!Number.isFinite(Date.parse(checkedAt))) throw new Error('invalid-date')
  if (!Array.isArray(source.points) || source.points.length === 0 || source.points.length > 200) throw new Error('invalid-points')
  const keys = new Set<string>()
  const points = source.points.map((value): IntelligencePoint => {
    const point = record(value)
    const state = STATES.find(item => item === point.state)
    if (!state) throw new Error('invalid-state')
    const parsed: IntelligencePoint = {
      model: text(point.model), effort: text(point.effort), state,
      score: metric(point.score), coverage: count(point.coverage), score_status: text(point.score_status, false),
      graded_runs: count(point.graded_runs), average_price_usd: metric(point.average_price_usd), price_samples: count(point.price_samples),
      average_minutes: metric(point.average_minutes), duration_samples: count(point.duration_samples),
      statistics_samples: count(point.statistics_samples), statistics_state: text(point.statistics_state, false), read_at: text(point.read_at, false)
    }
    if ((parsed.score !== null && parsed.score > scoreMax) || (parsed.coverage !== null && parsed.coverage > requiredTasks)) throw new Error('invalid-range')
    const key = JSON.stringify([parsed.model, parsed.effort])
    if (keys.has(key)) throw new Error('duplicate-point')
    keys.add(key)
    return parsed
  })
  return { benchmark: text(source.benchmark), catalogVersion: text(source.catalog_version), scoreMax, requiredTasks, checkedAt, refreshFailed, points }
}

// 已读取但覆盖题数不足（含 0 题）；读取失败不算在内。
export function lowCoverage(point: IntelligencePoint): boolean {
  return point.coverage !== null && point.coverage < LOW_COVERAGE_THRESHOLD
}

// 与源站卡片同名：gpt-6.1-sol → GPT-6.1 Sol。
export function modelLabel(model: string): string {
  return model.replace(/^gpt-/, 'GPT-').replace(/-([a-z]+)$/, (_, name: string) => ` ${name[0].toUpperCase()}${name.slice(1)}`)
}

// 与源站页面相同精度：分数 1 位、费用 2 位、耗时 1 位；缺失显示 —，不当作 0 或免费。
export function formatScore(value: number | null): string {
  return value === null ? '—' : value.toFixed(1)
}

export function formatPrice(value: number | null): string {
  if (value === null) return '—'
  return value > 0 && value < 0.005 ? '<$0.01' : `$${value.toFixed(2)}`
}

export function formatMinutes(value: number | null, unit: string): string {
  return value === null ? '—' : `${value.toFixed(1)}${unit}`
}

export interface FlagLabels { unreadPoint: string; retainedPoint: string; noGrades: string; low: string; statsInvalid: string }

// 格内标注：读取失败、本轮未更新、无有效判分／数据不足、统计来源异常。
export function pointFlags(point: IntelligencePoint, labels: FlagLabels): string[] {
  if (point.state === 'unavailable') return [labels.unreadPoint]
  const flags: string[] = []
  if (point.state === 'retained') flags.push(labels.retainedPoint)
  if (point.coverage === 0) flags.push(labels.noGrades)
  else if (lowCoverage(point)) flags.push(labels.low)
  if (point.statistics_state === 'invalid_source_statistics') flags.push(labels.statsInvalid)
  return flags
}

export function matrixRows(points: IntelligencePoint[]) {
  const rows = new Map<string, { key: string; model: string; cells: Record<string, IntelligencePoint> }>()
  for (const point of points) {
    const row = rows.get(point.model) ?? { key: point.model, model: point.model, cells: Object.create(null) }
    row.cells[point.effort] = point
    rows.set(point.model, row)
  }
  return [...rows.values()]
}

export type MatrixSort = 'source' | 'name' | `effort:${string}`
export type EfficiencyMetric = 'score' | 'average_minutes' | 'average_price_usd'

function compareMetric(first: number | null, second: number | null, descending: boolean) {
  if (first === null) return second === null ? 0 : 1
  if (second === null) return -1
  return descending ? second - first : first - second
}

export function sortMatrixRows(rows: ReturnType<typeof matrixRows>, sort: MatrixSort) {
  if (sort === 'source') return rows
  return [...rows].sort((first, second) => {
    if (sort === 'name') return first.model.localeCompare(second.model)
    const effort = sort.slice('effort:'.length)
    return compareMetric(first.cells[effort]?.score ?? null, second.cells[effort]?.score ?? null, true)
  })
}

export function sortEfficiencyPoints(points: IntelligencePoint[], metric: EfficiencyMetric) {
  return [...points].sort((first, second) => compareMetric(first[metric], second[metric], metric === 'score'))
}
