import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import {
  EFFORTS, formatMinutes, formatPrice, formatScore, INTELLIGENCE_ENDPOINT, lowCoverage, matrixRows, modelLabel, parseIntelligence, pointFlags,
  REFRESH_INTERVAL, REQUEST_TIMEOUT, sortEfficiencyPoints, sortMatrixRows, type IntelligencePoint
} from '@/custom/intelligence/data'
import { useIntelligence } from '@/custom/intelligence/useIntelligence'
import { intelligenceCopy } from '@/custom/intelligence/copy'
import { customRoutes } from '@/custom/routes'

const client = vi.hoisted(() => ({ get: vi.fn() }))
vi.mock('@/api/client', () => ({ apiClient: client }))

// 与 2026-10-12 实取 gpt-6.1-sol/low 经后端代读后的字段一致。
const point: IntelligencePoint = {
  model: 'gpt-6.1-sol', effort: 'low', state: 'ok', score: 81.75471971832992, coverage: 64, score_status: 'complete', graded_runs: 83,
  average_price_usd: 23.478868 / 71, price_samples: 71, average_minutes: 40180886 / 82 / 60000, duration_samples: 82,
  statistics_samples: 83, statistics_state: 'available', read_at: '2026-10-12T10:00:00Z'
}
const payload = (points: unknown[] = [point]) => ({
  benchmark: 'radar-bench-final64-20261004', catalog_version: 'dradar-config64', score_max: 150, required_tasks: 64,
  checked_at: '2026-10-12T10:00:00Z', refresh_failed: false, points: points.map(item => ({ ...(item as object) }))
})

describe('智力效率数据契约', () => {
  it('保留后端字段，展示精度与源站页面一致', () => {
    const parsed = parseIntelligence(payload())
    expect(parsed.points[0]).toEqual(point)
    expect(parsed).toMatchObject({ scoreMax: 150, requiredTasks: 64, checkedAt: '2026-10-12T10:00:00Z', refreshFailed: false })
    expect([formatScore(point.score), formatPrice(point.average_price_usd), formatMinutes(point.average_minutes, '分钟')]).toEqual(['81.8', '$0.33', '8.2分钟'])
    expect(modelLabel('gpt-6.1-sol')).toBe('GPT-6.1 Sol')
    expect(modelLabel('gpt-5.6-luna')).toBe('GPT-5.6 Luna')
  })
  it('零分、零费用照常显示，缺失显示破折号而不是 0 或免费', () => {
    expect([formatScore(0), formatScore(null)]).toEqual(['0.0', '—'])
    expect([formatPrice(0), formatPrice(0.004), formatPrice(null)]).toEqual(['$0.00', '<$0.01', '—'])
    expect(formatMinutes(null, '分钟')).toBe('—')
    const parsed = parseIntelligence(payload([{ ...point, state: 'unavailable', score: null, coverage: null, graded_runs: null, average_price_usd: null, score_status: '', statistics_state: '', read_at: '' }]))
    expect(parsed.points[0].score).toBeNull()
    expect(parsed.points[0].average_price_usd).toBeNull()
  })
  it.each([NaN, Infinity, -1, '100', 150.5])('拒绝非法分数 %s', score => {
    expect(() => parseIntelligence(payload([{ ...point, score }]))).toThrow()
  })
  it.each([
    { refresh_failed: 'no' }, { score_max: 0 }, { checked_at: 'invalid' }, { points: [] }, { points: [point, point] },
    { points: [{ ...point, state: 'stale' }] }, { points: [{ ...point, coverage: 65 }] }, { points: [{ ...point, coverage: 10.5 }] }, { points: [{ ...point, model: '' }] }
  ])('拒绝不兼容数据 %o', override => {
    expect(() => parseIntelligence({ ...payload(), ...override })).toThrow()
  })
  it('数据不足、无有效判分、读取失败与统计异常分别标注', () => {
    const labels = intelligenceCopy.zh
    expect(pointFlags(point, labels)).toEqual([])
    expect(pointFlags({ ...point, coverage: 39 }, labels)).toEqual(['数据不足'])
    expect(pointFlags({ ...point, coverage: 0, score: null }, labels)).toEqual(['无有效判分'])
    expect(pointFlags({ ...point, state: 'retained', statistics_state: 'invalid_source_statistics' }, labels)).toEqual(['本轮未更新', '统计来源异常'])
    expect(pointFlags({ ...point, state: 'unavailable', coverage: null }, labels)).toEqual(['读取失败'])
    expect(lowCoverage({ ...point, coverage: 40 })).toBe(false)
    expect(lowCoverage({ ...point, coverage: null })).toBe(false)
  })
  it('矩阵按模型成行；按指定档位降序、缺失最后、同分保留原序且不修改原数组', () => {
    const rows = matrixRows([
      { ...point, model: 'missing', effort: 'high', score: null },
      { ...point, model: 'zero', effort: 'high', score: 0 },
      { ...point, model: 'second', effort: 'high', score: 90 },
      { ...point, model: 'first', effort: 'high', score: 90 },
      { ...point, model: 'other-effort', effort: 'max', score: 150 },
      { ...point, model: 'zero', effort: 'future', score: 1 }
    ])
    expect(rows).toHaveLength(5)
    expect(rows[1].cells.future).toBeDefined()
    expect(sortMatrixRows(rows, 'effort:high').map(row => row.model)).toEqual(['second', 'first', 'zero', 'missing', 'other-effort'])
    expect(sortMatrixRows(rows, 'source').map(row => row.model)).toEqual(['missing', 'zero', 'second', 'first', 'other-effort'])
    expect(sortMatrixRows(rows, 'name').map(row => row.model)).toEqual(['first', 'missing', 'other-effort', 'second', 'zero'])
    expect(EFFORTS).toEqual(['ultra', 'max', 'xhigh', 'high', 'medium', 'low'])
  })
  it('对比按费用或耗时升序，零值正常且缺失最后', () => {
    const points = [
      { ...point, model: 'missing', average_minutes: null, average_price_usd: null },
      { ...point, model: 'positive', average_minutes: 3, average_price_usd: 3 },
      { ...point, model: 'zero', average_minutes: 0, average_price_usd: 0 }
    ]
    expect(sortEfficiencyPoints(points, 'average_minutes').map(item => item.model)).toEqual(['zero', 'positive', 'missing'])
    expect(sortEfficiencyPoints(points, 'average_price_usd').map(item => item.model)).toEqual(['zero', 'positive', 'missing'])
    expect(points[0].model).toBe('missing')
  })
  it('中英文文案键一致且入口需要认证', () => {
    expect(Object.keys(intelligenceCopy.zh).sort()).toEqual(Object.keys(intelligenceCopy.en).sort())
    expect(customRoutes.find(route => route.path === '/intelligence')?.meta?.requiresAuth).toBe(true)
  })
})
describe('智力效率同步生命周期', () => {
  let wrapper: ReturnType<typeof mount> | undefined
  let api: ReturnType<typeof useIntelligence>
  beforeEach(() => {
    vi.useFakeTimers()
    client.get.mockReset().mockResolvedValue({ data: payload() })
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('visible')
  })
  afterEach(() => { wrapper?.unmount(); wrapper = undefined; vi.restoreAllMocks(); vi.useRealTimers() })
  async function start() {
    wrapper = mount(defineComponent({ setup() { api = useIntelligence(); return () => null } }))
    await flushPromises()
  }
  it('首屏读取本站认证接口（不再直连第三方），10 分钟刷新，卸载后不轮询', async () => {
    await start()
    expect(client.get).toHaveBeenCalledWith(INTELLIGENCE_ENDPOINT, expect.objectContaining({ timeout: REQUEST_TIMEOUT }))
    expect(client.get.mock.calls[0][1].signal).toHaveProperty('aborted', false)
    expect(api.snapshot.value?.points).toHaveLength(1)
    await vi.advanceTimersByTimeAsync(REFRESH_INTERVAL - 1)
    expect(client.get).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(1)
    expect(client.get).toHaveBeenCalledTimes(2)
    wrapper?.unmount(); wrapper = undefined
    await vi.advanceTimersByTimeAsync(REFRESH_INTERVAL)
    expect(client.get).toHaveBeenCalledTimes(2)
  })
  it('失败或格式异常保留旧数据与旧获取时间，重试可恢复', async () => {
    await start()
    const snapshot = api.snapshot.value
    const fetchedAt = api.fetchedAt.value
    client.get.mockResolvedValueOnce({ data: { schema: 99 } })
    await api.refresh()
    expect(api.failed.value).toBe(true)
    expect(api.snapshot.value).toBe(snapshot)
    expect(api.fetchedAt.value).toBe(fetchedAt)
    client.get.mockRejectedValueOnce({ status: 503, message: '评测数据源暂时不可用' })
    await api.refresh()
    expect(api.snapshot.value).toBe(snapshot)
    await api.refresh()
    expect(api.failed.value).toBe(false)
  })
  it('隐藏页不轮询，重新可见且过期时刷新', async () => {
    await start()
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('hidden')
    await vi.advanceTimersByTimeAsync(REFRESH_INTERVAL)
    expect(client.get).toHaveBeenCalledTimes(1)
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('visible')
    document.dispatchEvent(new Event('visibilitychange'))
    await flushPromises()
    expect(client.get).toHaveBeenCalledTimes(2)
  })
  it('请求中去重，卸载时中断请求', async () => {
    client.get.mockImplementation((_url, options) => new Promise((_resolve, reject) => {
      options.signal.addEventListener('abort', () => reject(new Error('aborted')))
    }))
    await start()
    void api.refresh()
    expect(client.get).toHaveBeenCalledTimes(1)
    const signal = client.get.mock.calls[0][1].signal as AbortSignal
    wrapper?.unmount(); wrapper = undefined
    await flushPromises()
    expect(signal.aborted).toBe(true)
  })
  it('超时显示错误，不伪造数据', async () => {
    client.get.mockRejectedValue({ code: 'ECONNABORTED', message: 'timeout of 60000ms exceeded' })
    await start()
    expect(api.failed.value).toBe(true)
    expect(api.loading.value).toBe(false)
    expect(api.snapshot.value).toBeNull()
  })
})
