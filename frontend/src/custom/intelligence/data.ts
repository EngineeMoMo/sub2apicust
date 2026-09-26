export const INTELLIGENCE_SOURCE = 'https://codexradar.com/data/intelligence-efficiency.json'
export const REFRESH_INTERVAL = 30 * 60 * 1000
export const LOW_SAMPLE_THRESHOLD = 30
export const EFFORTS = ['ultra', 'max', 'xhigh', 'high', 'medium', 'low']

export interface IntelligencePoint {
  model: string
  effort: string
  harness: string
  iq: number | null
  passed: number | null
  valid_tasks: number | null
  average_minutes: number | null
  average_price_usd: number | null
  price_aggregation: string | null
}

export interface IntelligenceSnapshot {
  sourceUpdatedAt: string
  mode: string
  points: IntelligencePoint[]
}

function record(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('invalid-data')
  return value as Record<string, unknown>
}

function text(value: unknown): string {
  if (typeof value !== 'string' || !value.trim() || value.length > 200) throw new Error('invalid-data')
  return value
}

function metric(value: unknown): number | null {
  if (value === null || value === undefined) return null
  if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) throw new Error('invalid-data')
  return value
}

export function parseIntelligence(value: unknown): IntelligenceSnapshot {
  const source = record(value)
  if (source.schema !== 2 || source.type !== 'distributed_intelligence_efficiency') throw new Error('invalid-schema')
  const sourceUpdatedAt = text(source.source_updated_at)
  if (!Number.isFinite(Date.parse(sourceUpdatedAt))) throw new Error('invalid-date')
  if (!Array.isArray(source.points) || source.points.length === 0 || source.points.length > 5000) throw new Error('invalid-points')
  const keys = new Set<string>()
  const points = source.points.map((value): IntelligencePoint => {
    const point = record(value)
    const parsed = {
      model: text(point.model), effort: text(point.effort), harness: text(point.harness),
      iq: metric(point.iq), passed: metric(point.passed), valid_tasks: metric(point.valid_tasks),
      average_minutes: metric(point.average_minutes), average_price_usd: metric(point.average_price_usd),
      price_aggregation: point.price_aggregation == null ? null : text(point.price_aggregation)
    }
    const key = JSON.stringify([parsed.model, parsed.harness, parsed.effort])
    if (keys.has(key)) throw new Error('duplicate-point')
    if (parsed.passed !== null && parsed.valid_tasks !== null && parsed.passed > parsed.valid_tasks) throw new Error('invalid-count')
    keys.add(key)
    return parsed
  })
  return { sourceUpdatedAt, mode: text(source.mode), points }
}

export function lowSample(point: IntelligencePoint): boolean {
  return point.valid_tasks === null || point.valid_tasks < LOW_SAMPLE_THRESHOLD
}

export function matrixRows(points: IntelligencePoint[]) {
  const rows = new Map<string, { key: string; model: string; harness: string; cells: Record<string, IntelligencePoint> }>()
  for (const point of points) {
    const key = JSON.stringify([point.model, point.harness])
    const row = rows.get(key) ?? { key, model: point.model, harness: point.harness, cells: Object.create(null) }
    row.cells[point.effort] = point
    rows.set(key, row)
  }
  return [...rows.values()]
}

export type MatrixSort = 'source' | 'name' | `effort:${string}`
export type EfficiencyMetric = 'iq' | 'average_minutes' | 'average_price_usd'

function compareMetric(first: number | null, second: number | null, descending: boolean) {
  if (first === null) return second === null ? 0 : 1
  if (second === null) return -1
  return descending ? second - first : first - second
}

export function sortMatrixRows(rows: ReturnType<typeof matrixRows>, sort: MatrixSort) {
  if (sort === 'source') return rows
  return [...rows].sort((first, second) => {
    if (sort === 'name') return first.model.localeCompare(second.model) || first.harness.localeCompare(second.harness)
    const effort = sort.slice('effort:'.length)
    return compareMetric(first.cells[effort]?.iq ?? null, second.cells[effort]?.iq ?? null, true)
  })
}

export function sortEfficiencyPoints(points: IntelligencePoint[], metric: EfficiencyMetric) {
  return [...points].sort((first, second) => compareMetric(first[metric], second[metric], metric === 'iq'))
}
