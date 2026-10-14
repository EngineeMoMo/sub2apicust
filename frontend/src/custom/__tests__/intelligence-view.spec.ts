import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import IntelligenceView from '@/custom/views/IntelligenceView.vue'
import IntelligenceComparison from '@/custom/components/IntelligenceComparison.vue'
import { parseIntelligence } from '@/custom/intelligence/data'

const client = vi.hoisted(() => ({ get: vi.fn() }))
vi.mock('@/api/client', () => ({ apiClient: client }))
vi.mock('@/components/layout/AppLayout.vue', () => ({
  default: defineComponent({ template: '<main><slot name="page-actions" /><slot /></main>' })
}))

const base = { state: 'ok', score_status: 'provisional', statistics_state: 'available', read_at: '2026-10-12T10:00:00Z' }
// 取值来自 2026-10-12 实取数据：sol/low 为截图同格，luna/high 是 1 题即 150 分的低覆盖示例，astra/high 模拟本轮读取失败。
const source = (overrides: Record<string, unknown> = {}) => ({
  benchmark: 'radar-bench-final64-20261004', catalog_version: 'dradar-config64', score_max: 150, required_tasks: 64,
  checked_at: '2026-10-12T10:00:00Z', refresh_failed: false,
  points: [
    { ...base, model: 'gpt-6.1-sol', effort: 'low', score: 81.75471971832992, coverage: 64, score_status: 'complete', graded_runs: 83, average_price_usd: 23.478868 / 71, price_samples: 71, average_minutes: 40180886 / 82 / 60000, duration_samples: 82, statistics_samples: 83 },
    { ...base, model: 'gpt-6.1-sol', effort: 'high', score: 106.06, coverage: 50, graded_runs: 51, average_price_usd: 0.48, price_samples: 51, average_minutes: 17.4, duration_samples: 51, statistics_samples: 51 },
    { ...base, model: 'gpt-5.6-luna', effort: 'high', score: 150, coverage: 1, graded_runs: 1, average_price_usd: 0.11, price_samples: 1, average_minutes: 4.4, duration_samples: 1, statistics_samples: 1 },
    { ...base, model: 'gpt-5.6-luna', effort: 'low', score: 0, coverage: 6, graded_runs: 6, average_price_usd: 0, price_samples: 6, average_minutes: 1.4, duration_samples: 6, statistics_samples: 6 },
    { model: 'gpt-6-astra', effort: 'high', state: 'unavailable', score: null, coverage: null, score_status: '', graded_runs: null, average_price_usd: null, price_samples: null, average_minutes: null, duration_samples: null, statistics_samples: null, statistics_state: '', read_at: '' }
  ],
  ...overrides
})

describe('智力效率页面交互', () => {
  let wrapper: VueWrapper | undefined
  beforeEach(() => { client.get.mockReset().mockImplementation(async () => ({ data: source() })) })
  afterEach(() => { wrapper?.unmount(); wrapper = undefined })
  async function render(locale = 'zh') {
    wrapper = mount(IntelligenceView, { global: { plugins: [createI18n({ legacy: false, locale, messages: {} })] } })
    await flushPromises()
    return wrapper
  }
  const button = (page: VueWrapper, label: string) => page.findAll('button').find(item => item.text() === label)!
  const strip = (value: string, label: string) => value.replace(label, '').trim()

  it('按模型成行，每格展示雷达分、覆盖、判分题次、费用与耗时；0 分不是缺失', async () => {
    const page = await render()
    const rows = page.findAll('.mofa-intel-matrix tbody tr')
    expect(rows.map(row => row.get('th').text())).toEqual(['GPT-6.1 Solgpt-6.1-sol', 'GPT-5.6 Lunagpt-5.6-luna', 'GPT-6 Astragpt-6-astra'])
    expect(page.findAll('.mofa-intel-model-icon')).toHaveLength(3)
    expect(page.findAll('.mofa-intel-score strong').map(cell => strip(cell.text(), '雷达分'))).toEqual(['106.1', '81.8', '150.0', '0.0', '—'])
    expect(page.findAll('.mofa-intel-price').map(cell => strip(cell.text(), '平均费用'))).toEqual(['$0.48', '$0.33', '$0.11', '$0.00', '—'])
    expect(page.findAll('.mofa-intel-time').map(cell => strip(cell.text(), '平均耗时'))).toEqual(['17.4分钟', '8.2分钟', '4.4分钟', '1.4分钟', '—'])
    expect(rows[0].text()).toContain('覆盖 64/64 · 判分 83')
    expect(rows[2].text()).toContain('覆盖 — · 判分 —')
    expect(page.findAll('.mofa-intel-sample').map(chip => chip.text())).toEqual(['数据不足', '数据不足', '读取失败'])
    expect(page.findAll('.mofa-intel-missing')).toHaveLength(13)
    expect(page.findAll('.mofa-intel-stats dd').map(cell => cell.text())).toEqual(['3', '5', '2'])
    expect(page.text()).toContain('部分组合本轮未能读取，已在格内标注 · 1 / 5')
    expect(page.find('[role=alert]').exists()).toBe(false)
  })
  it('支持按名称或展示名搜索、覆盖题数过滤，以及清空筛选', async () => {
    const page = await render()
    const search = page.get('input[type=search]')
    await search.setValue('luna')
    expect(page.findAll('tbody tr')).toHaveLength(1)
    await search.setValue('6.1 sol')
    expect(page.get('tbody tr').text()).toContain('GPT-6.1 Sol')
    await search.setValue('missing')
    expect(page.text()).toContain('没有符合条件的数据')
    await button(page, '清除筛选').trigger('click')
    expect(page.findAll('tbody tr')).toHaveLength(3)
    await page.get('input[type=checkbox]').setValue(true)
    expect(page.findAll('tbody tr')).toHaveLength(1)
    expect(page.text()).not.toContain('GPT-5.6 Luna')
    expect(page.text()).not.toContain('GPT-6 Astra')
  })
  it('效率对比默认高档位、按雷达分排序；可切全部档位并按耗时或费用排序，不隐藏缺失指标的组合', async () => {
    const page = await render()
    await button(page, '效率对比').trigger('click')
    const rows = () => page.findAll('.mofa-intel-compare-table tbody tr')
    expect(rows()).toHaveLength(3)
    expect(page.find('canvas').exists()).toBe(false)
    expect(rows()[0].text()).toContain('gpt-5.6-luna / high')
    expect(rows()[0].text()).toContain('数据不足')
    expect(rows()[2].text()).toContain('读取失败')
    expect(rows()[2].text()).toContain('—')
    const controls = page.findAll('.mofa-intel-compare-controls select')
    await controls[0].setValue('')
    expect(rows()).toHaveLength(5)
    expect(page.text()).toContain('已知样本 71/83')
    await controls[1].setValue('average_minutes')
    expect(rows()[0].text()).toContain('gpt-5.6-luna / low')
    expect(rows()[4].text()).toContain('gpt-6-astra / high')
    await controls[1].setValue('average_price_usd')
    expect(rows()[0].text()).toContain('$0.00')
    expect(page.text()).toContain('三个指标单位不同')
  })
  it('矩阵默认源顺序，可按名称或指定档位雷达分排序', async () => {
    const page = await render()
    const sort = page.get('.mofa-intel-sort-controls select')
    const rows = () => page.findAll('.mofa-intel-matrix tbody tr')
    expect(rows()[0].text()).toContain('GPT-6.1 Sol')
    expect(page.text()).toContain('不是智力排名')
    await sort.setValue('name')
    expect(rows()[0].text()).toContain('GPT-5.6 Luna')
    await sort.setValue('effort:low')
    expect(rows().map(row => row.get('th').text())).toEqual(['GPT-6.1 Solgpt-6.1-sol', 'GPT-5.6 Lunagpt-5.6-luna', 'GPT-6 Astragpt-6-astra'])
    await sort.setValue('effort:high')
    expect(rows()[0].text()).toContain('GPT-5.6 Luna')
    expect(rows()[2].text()).toContain('GPT-6 Astra')
    await sort.setValue('source')
    expect(rows()[1].text()).toContain('GPT-5.6 Luna')
  })
  it('说明费用口径、雷达分算法与数据不足阈值，不含外链与刷新频率', async () => {
    const page = await render()
    const notes = page.get('.mofa-intel-notes').findAll('p').map(paragraph => paragraph.text())
    expect(notes).toHaveLength(3)
    expect(notes[0]).toContain('按官方标准 API 价格折算，不是本站套餐价')
    expect(notes[0]).toContain('不是免费')
    expect(notes[1]).toContain('百分制成绩 × 1.5（满分 150）')
    expect(notes[1]).toContain('未覆盖的题不按 0 分计')
    expect(notes[1]).toContain('不是人类智商')
    expect(notes[2]).toContain('覆盖不足 40 题标记为数据不足')
    expect(page.get('.mofa-intel-notes').find('h2').exists()).toBe(false)
    expect(page.find('a[href]').exists()).toBe(false)
    expect(page.text()).not.toContain('数据来源：Codex Radar')
    expect(page.text()).not.toContain('10 分钟')
    expect(page.text()).not.toContain('30 分钟')
    expect(page.findAll('button').some(item => item.text() === '刷新数据')).toBe(true)
  })
  it('网络失败保留表格并提示；首屏失败可重试；源站本轮失败单独说明', async () => {
    client.get.mockRejectedValueOnce(new Error('offline'))
    const page = await render()
    expect(page.get('[role=alert]').text()).toContain('数据源暂时不可用')
    await button(page, '刷新数据').trigger('click')
    await flushPromises()
    expect(page.findAll('tbody tr')).toHaveLength(3)
    client.get.mockRejectedValueOnce(new Error('offline'))
    await button(page, '刷新数据').trigger('click')
    await flushPromises()
    expect(page.get('[role=alert]').text()).toContain('仍显示上次成功获取的数据')
    expect(page.findAll('tbody tr')).toHaveLength(3)
    client.get.mockResolvedValueOnce({ data: source({ refresh_failed: true }) })
    await button(page, '刷新数据').trigger('click')
    await flushPromises()
    expect(page.find('[role=alert]').exists()).toBe(false)
    expect(page.text()).toContain('本轮读取数据源失败，显示的是上次成功读取的数据。')
  })
  it('英文界面提供等价信息且不含外链', async () => {
    const page = await render('en')
    expect(page.text()).toContain('Choose models with evidence.')
    expect(page.text()).toContain('Insufficient data')
    expect(page.text()).toContain('Read failed')
    expect(page.findAll('.mofa-intel-time').map(cell => strip(cell.text(), 'Mean duration'))[0]).toBe('17.4 min')
    expect(page.get('.mofa-intel-notes').findAll('p')[1].text()).toContain('percentage score × 1.5 (max 150)')
    expect(page.find('a[href]').exists()).toBe(false)
  })
  it('对比分页保留整组缩放比例，雷达分按满分 150 绘制，筛选后回第一页', async () => {
    const normalized = parseIntelligence(source()).points[0]
    const points = Array.from({ length: 13 }, (_, index) => ({ ...normalized, model: `model-${index}`, score: index === 12 ? 150 : index, average_price_usd: 0 }))
    wrapper = mount(IntelligenceComparison, { props: { points, requiredTasks: 64, scoreMax: 150 }, global: { plugins: [createI18n({ legacy: false, locale: 'zh', messages: {} })] } })
    expect(wrapper.findAll('tbody tr')).toHaveLength(12)
    const buttons = wrapper.findAll('button')
    expect(buttons[0].attributes('disabled')).toBeDefined()
    await buttons[1].trigger('click')
    expect(wrapper.findAll('tbody tr')).toHaveLength(1)
    expect(wrapper.get('tbody tr').text()).toContain('model-12')
    expect(wrapper.get('.mofa-intel-metric-score .mofa-intel-compare-bar > span').attributes('style')).toContain('100%')
    expect(wrapper.get('.mofa-intel-metric-average_price_usd .mofa-intel-compare-bar > span').attributes('style')).toContain('0%')
    await wrapper.setProps({ points: points.slice(0, 2) })
    expect(wrapper.findAll('tbody tr')).toHaveLength(2)
    expect(wrapper.get('tbody tr').text()).toContain('model-0')
    expect(wrapper.findAll('button')[1].attributes('disabled')).toBeDefined()
  })
})
