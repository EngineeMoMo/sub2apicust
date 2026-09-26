import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { EFFORTS, INTELLIGENCE_SOURCE, lowSample, matrixRows, parseIntelligence, REFRESH_INTERVAL, sortMatrixRows, sortEfficiencyPoints } from '@/custom/intelligence/data'
import { useIntelligence } from '@/custom/intelligence/useIntelligence'
import { intelligenceCopy } from '@/custom/intelligence/copy'
import { customRoutes } from '@/custom/routes'

const point = { model: 'test-model', effort: 'high', harness: 'codex', iq: 0, passed: 0, valid_tasks: 50, average_minutes: 1, average_price_usd: 0.5, price_aggregation: 'median' }
const payload = () => ({ schema: 2, type: 'distributed_intelligence_efficiency', mode: 'equal_latest_3', source_updated_at: '2026-09-26T18:07:30+08:00', points: [{ ...point }] })

describe('智力效率数据契约', () => {
  it('矩阵按指定档位降序、缺失最后、同分保留原序且不修改原数组', () => {
    const rows = matrixRows([
      { ...point, model: 'missing', iq: null },
      { ...point, model: 'zero', iq: 0 },
      { ...point, model: 'second', iq: 90 },
      { ...point, model: 'first', iq: 90 },
      { ...point, model: 'other-effort', effort: 'max', iq: 150 }
    ])
    expect(sortMatrixRows(rows, 'effort:high').map(row => row.model)).toEqual(['second', 'first', 'zero', 'missing', 'other-effort'])
    expect(sortMatrixRows(rows, 'source').map(row => row.model)).toEqual(['missing', 'zero', 'second', 'first', 'other-effort'])
    expect(sortMatrixRows(rows, 'name').map(row => row.model)).toEqual(['first', 'missing', 'other-effort', 'second', 'zero'])
  })
  it('对比按费用或耗时升序，零值正常且缺失最后', () => {
    const points = [
      { ...point, model: 'missing', average_minutes: null, average_price_usd: null },
      { ...point, model: 'positive', average_minutes: 3, average_price_usd: 3 },
      { ...point, model: 'zero', average_minutes: 0, average_price_usd: 0 }
    ]
    expect(sortEfficiencyPoints(points, 'average_minutes').map(point => point.model)).toEqual(['zero', 'positive', 'missing'])
    expect(sortEfficiencyPoints(points, 'average_price_usd').map(point => point.model)).toEqual(['zero', 'positive', 'missing'])
    expect(points[0].model).toBe('missing')
  })
  it('保留源分数、聚合方式和源更新时间，不重算 IQ', () => {
    const parsed = parseIntelligence(payload())
    expect(parsed.points[0]).toEqual(point)
    expect(parsed.sourceUpdatedAt).toBe(payload().source_updated_at)
  })
  it('保留缺失与零值的区别', () => {
    const data = payload()
    const parsed = parseIntelligence({ ...data, points: [{ ...point, iq: null, average_minutes: null }] })
    expect(parsed.points[0].iq).toBeNull()
    expect(parsed.points[0].average_minutes).toBeNull()
    expect(parseIntelligence(data).points[0].iq).toBe(0)
  })
  it.each([NaN, Infinity, -1, '100'])('拒绝非法指标 %s', iq => {
    expect(() => parseIntelligence({ ...payload(), points: [{ ...point, iq }] })).toThrow()
  })
  it.each([
    { schema: 3 }, { source_updated_at: 'invalid' }, { points: [] }, { points: [point, point] },
    { points: [{ ...point, passed: 51 }] }, { points: [{ ...point, model: '' }] }
  ])('拒绝不兼容数据 %o', override => {
    expect(() => parseIntelligence({ ...payload(), ...override })).toThrow()
  })
  it('不同评测环境独立成行，未知档位也不丢失', () => {
    const data = parseIntelligence({ ...payload(), points: [point, { ...point, harness: 'dsh' }, { ...point, effort: 'future' }] })
    const rows = matrixRows(data.points)
    expect(rows).toHaveLength(2)
    expect(rows[0].cells.future).toBeDefined()
    expect(rows[1].harness).toBe('dsh')
    expect(EFFORTS).toEqual(['ultra', 'max', 'xhigh', 'high', 'medium', 'low'])
  })
  it('少量及未知样本单独标记，不把高分当成可信排名', () => {
    expect(lowSample({ ...point, iq: 150, valid_tasks: 1 })).toBe(true)
    expect(lowSample({ ...point, valid_tasks: null })).toBe(true)
    expect(lowSample({ ...point, valid_tasks: 30 })).toBe(false)
  })
  it('中英文文案键一致且入口需要认证', () => {
    expect(Object.keys(intelligenceCopy.zh).sort()).toEqual(Object.keys(intelligenceCopy.en).sort())
    expect(customRoutes.find(route => route.path === '/intelligence')?.meta?.requiresAuth).toBe(true)
  })
})

describe('智力效率同步生命周期', () => {
  let wrapper: ReturnType<typeof mount> | undefined
  let api: ReturnType<typeof useIntelligence>
  const fetchMock = vi.fn()
  beforeEach(() => {
    vi.useFakeTimers()
    vi.stubGlobal('fetch', fetchMock)
    fetchMock.mockReset().mockResolvedValue({ ok: true, json: async () => payload() })
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('visible')
  })
  afterEach(() => { wrapper?.unmount(); wrapper = undefined; vi.restoreAllMocks(); vi.unstubAllGlobals(); vi.useRealTimers() })
  async function start() {
    wrapper = mount(defineComponent({ setup() { api = useIntelligence(); return () => null } }))
    await flushPromises()
  }
  it('首屏请求不带凭据，30 分钟刷新，卸载后不轮询', async () => {
    await start()
    expect(fetchMock).toHaveBeenCalledWith(INTELLIGENCE_SOURCE, expect.objectContaining({ credentials: 'omit', referrerPolicy: 'no-referrer' }))
    await vi.advanceTimersByTimeAsync(REFRESH_INTERVAL - 1)
    expect(fetchMock).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(1)
    expect(fetchMock).toHaveBeenCalledTimes(2)
    wrapper?.unmount(); wrapper = undefined
    await vi.advanceTimersByTimeAsync(REFRESH_INTERVAL)
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })
  it('失败或格式异常保留旧数据与旧获取时间，重试可恢复', async () => {
    await start()
    const snapshot = api.snapshot.value
    const fetchedAt = api.fetchedAt.value
    fetchMock.mockResolvedValueOnce({ ok: true, json: async () => ({ schema: 99 }) })
    await api.refresh()
    expect(api.failed.value).toBe(true)
    expect(api.snapshot.value).toBe(snapshot)
    expect(api.fetchedAt.value).toBe(fetchedAt)
    fetchMock.mockResolvedValueOnce({ ok: false })
    await api.refresh()
    expect(api.snapshot.value).toBe(snapshot)
    await api.refresh()
    expect(api.failed.value).toBe(false)
  })
  it('隐藏页不轮询，重新可见且过期时刷新', async () => {
    await start()
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('hidden')
    await vi.advanceTimersByTimeAsync(REFRESH_INTERVAL)
    expect(fetchMock).toHaveBeenCalledTimes(1)
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('visible')
    document.dispatchEvent(new Event('visibilitychange'))
    await flushPromises()
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })
  it('请求中去重，15 秒超时中断并显示错误', async () => {
    fetchMock.mockImplementation((_url, options) => new Promise((_resolve, reject) => {
      options.signal.addEventListener('abort', () => reject(new Error('aborted')))
    }))
    await start()
    void api.refresh()
    expect(fetchMock).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(15_000)
    expect(api.failed.value).toBe(true)
    expect(api.loading.value).toBe(false)
    expect(api.snapshot.value).toBeNull()
  })
  it('卸载时中断请求', async () => {
    fetchMock.mockImplementation((_url, options) => new Promise((_resolve, reject) => {
      options.signal.addEventListener('abort', () => reject(new Error('aborted')))
    }))
    await start()
    const signal = fetchMock.mock.calls[0][1].signal as AbortSignal
    wrapper?.unmount(); wrapper = undefined
    await flushPromises()
    expect(signal.aborted).toBe(true)
  })
})
