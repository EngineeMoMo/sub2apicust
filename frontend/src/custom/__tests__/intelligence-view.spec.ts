import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import IntelligenceView from '@/custom/views/IntelligenceView.vue'
import IntelligenceComparison from '@/custom/components/IntelligenceComparison.vue'
import { parseIntelligence } from '@/custom/intelligence/data'

vi.mock('@/components/layout/AppLayout.vue', () => ({
  default: defineComponent({ template: '<main><slot name="page-actions" /><slot /></main>' })
}))

const source = {
  schema: 2, type: 'distributed_intelligence_efficiency', mode: 'equal_latest_3', source_updated_at: '2026-09-26T18:07:30+08:00',
  points: [
    { model: 'gpt-test', harness: 'codex', effort: 'high', iq: 90, passed: 60, valid_tasks: 100, average_minutes: 10, average_price_usd: 0.1, price_aggregation: 'median' },
    { model: 'claude-test', harness: 'codex', effort: 'max', iq: 150, passed: 1, valid_tasks: 1, average_minutes: null, average_price_usd: 0.2 },
    { model: 'gpt-test', harness: 'dsh', effort: 'high', iq: 0, passed: 0, valid_tasks: 50, average_minutes: 0, average_price_usd: 0 }
  ]
}

describe('智力效率页面交互', () => {
  let wrapper: VueWrapper | undefined
  const fetchMock = vi.fn()
  beforeEach(() => {
    vi.stubGlobal('fetch', fetchMock)
    fetchMock.mockReset().mockResolvedValue({ ok: true, json: async () => source })
  })
  afterEach(() => { wrapper?.unmount(); vi.unstubAllGlobals() })
  async function render(locale = 'zh') {
    wrapper = mount(IntelligenceView, { global: { plugins: [createI18n({ legacy: false, locale, messages: {} })] } })
    await flushPromises()
    return wrapper
  }
  it('呈现数据、模型图标和独立环境行，缺失不是零分', async () => {
    const page = await render()
    expect(page.findAll('tbody tr')).toHaveLength(3)
    expect(page.findAll('.mofa-intel-model-icon')).toHaveLength(3)
    expect(page.findAll('.mofa-intel-score strong').map(cell => cell.text())).toEqual(['90', '150', '0'])
    expect(page.findAll('.mofa-intel-missing').length).toBeGreaterThan(0)
    expect(page.find('.mofa-intel-sample').text()).toBe('低样本')
    expect(page.text()).toContain('不是统计可信度保证')
  })
  it('支持搜索、环境与样本过滤，以及清空筛选', async () => {
    const page = await render()
    await page.get('input[type=search]').setValue('gpt')
    expect(page.findAll('tbody tr')).toHaveLength(2)
    await page.get('select').setValue('dsh')
    expect(page.findAll('tbody tr')).toHaveLength(1)
    await page.get('input[type=search]').setValue('missing')
    expect(page.text()).toContain('没有符合条件的数据')
    await page.findAll('button').find(button => button.text() === '清除筛选')!.trigger('click')
    await page.get('input[type=checkbox]').setValue(true)
    expect(page.findAll('tbody tr')).toHaveLength(2)
    expect(page.text()).not.toContain('claude-test')
  })
  it('效率对比直接展示模型与数值，筛选档位并排序，不隐藏缺失指标所在的模型', async () => {
    const page = await render()
    await page.findAll('button').find(button => button.text() === '效率对比')!.trigger('click')
    expect(page.findAll('.mofa-intel-compare-table tbody tr')).toHaveLength(2)
    expect(page.find('canvas').exists()).toBe(false)
    const controls = page.findAll('.mofa-intel-compare-controls select')
    await controls[0].setValue('')
    expect(page.findAll('.mofa-intel-compare-table tbody tr')).toHaveLength(3)
    expect(page.get('.mofa-intel-compare-table tbody tr').text()).toContain('claude-test')
    expect(page.get('.mofa-intel-compare-table tbody tr').text()).toContain('—')
    await controls[1].setValue('average_minutes')
    expect(page.get('.mofa-intel-compare-table tbody tr').text()).toContain('dsh')
    await controls[1].setValue('average_price_usd')
    expect(page.get('.mofa-intel-compare-table tbody tr').text()).toContain('dsh')
    expect(page.text()).toContain('三个指标单位不同')
  })
  it('矩阵默认源顺序，可按名称或指定档位分数排序', async () => {
    const page = await render()
    const sort = page.get('.mofa-intel-sort-controls select')
    expect(page.get('.mofa-intel-matrix tbody tr').text()).toContain('gpt-test')
    expect(page.text()).toContain('不是智力排名')
    await sort.setValue('name')
    expect(page.get('.mofa-intel-matrix tbody tr').text()).toContain('claude-test')
    await sort.setValue('effort:high')
    const rows = page.findAll('.mofa-intel-matrix tbody tr')
    expect(rows[0].text()).toContain('90')
    expect(rows[2].text()).toContain('claude-test')
    await sort.setValue('source')
    expect(page.findAll('.mofa-intel-matrix tbody tr')[1].text()).toContain('claude-test')
  })
  it('说明区只保留指定两段，保留手动刷新', async () => {
    const page = await render()
    expect(page.text()).not.toContain('查看源站')
    expect(page.text()).not.toContain('30 分钟')
    expect(page.find('a[href^="https://codexradar.com"]').exists()).toBe(false)
    expect(page.get('.mofa-intel-notes').findAll('p').map(paragraph => paragraph.text())).toEqual([
      'IQ 是源站评测指标，不是人类智商；不同环境、样本数和推理档位会影响可比性。',
      '少于 30 个有效样本标记为低样本；30 只是展示筛选阈值，不是统计可信度保证。缺失数据不按 0 分处理。'
    ])
    expect(page.get('.mofa-intel-notes').find('h2').exists()).toBe(false)
    expect(page.text()).not.toContain('数据来源：Codex Radar')
    expect(page.findAll('button').some(button => button.text() === '刷新数据')).toBe(true)
  })
  it('网络失败保留表格并提示；首屏失败可重试', async () => {
    fetchMock.mockRejectedValueOnce(new Error('offline'))
    const page = await render()
    expect(page.get('[role=alert]').text()).toContain('数据源暂时不可用')
    await page.findAll('button').find(button => button.text() === '刷新数据')!.trigger('click')
    await flushPromises()
    expect(page.findAll('tbody tr')).toHaveLength(3)
    fetchMock.mockRejectedValueOnce(new Error('offline'))
    await page.findAll('button').find(button => button.text() === '刷新数据')!.trigger('click')
    await flushPromises()
    expect(page.get('[role=alert]').text()).toContain('仍显示上次成功获取的数据')
    expect(page.findAll('tbody tr')).toHaveLength(3)
  })
  it('英文界面提供等价信息且移除外链与频率提示', async () => {
    const page = await render('en')
    expect(page.text()).toContain('Choose models with evidence.')
    expect(page.text()).toContain('Small sample')
    expect(page.get('.mofa-intel-notes').findAll('p').map(paragraph => paragraph.text())).toEqual([
      'IQ is a source benchmark metric, not human IQ. Harness, sample size and effort affect comparability.',
      'Fewer than 30 valid samples are marked as small samples. The threshold is a display filter, not a statistical confidence guarantee. Missing data is not treated as zero.'
    ])
    expect(page.get('.mofa-intel-notes').find('h2').exists()).toBe(false)
    expect(page.text()).not.toContain('Data source: Codex Radar')
    expect(page.text()).not.toContain('30 minutes')
    expect(page.find('a[href^="https://codexradar.com"]').exists()).toBe(false)
  })
  it('对比分页保留整组缩放比例，筛选后回第一页', async () => {
    const normalized = parseIntelligence(source).points[0]
    const points = Array.from({ length: 13 }, (_, index) => ({ ...normalized, model: `model-${index}`, iq: index, average_price_usd: 0 }))
    wrapper = mount(IntelligenceComparison, { props: { points }, global: { plugins: [createI18n({ legacy: false, locale: 'zh', messages: {} })] } })
    expect(wrapper.findAll('tbody tr')).toHaveLength(12)
    const buttons = wrapper.findAll('button')
    expect(buttons[0].attributes('disabled')).toBeDefined()
    await buttons[1].trigger('click')
    expect(wrapper.findAll('tbody tr')).toHaveLength(1)
    expect(wrapper.get('tbody tr').text()).toContain('model-12')
    expect(wrapper.get('.mofa-intel-metric-iq .mofa-intel-compare-bar > span').attributes('style')).toContain('100%')
    expect(wrapper.get('.mofa-intel-metric-average_price_usd .mofa-intel-compare-bar > span').attributes('style')).toContain('0%')
    await wrapper.setProps({ points: points.slice(0, 2) })
    expect(wrapper.findAll('tbody tr')).toHaveLength(2)
    expect(wrapper.get('tbody tr').text()).toContain('model-0')
    expect(wrapper.findAll('button')[1].attributes('disabled')).toBeDefined()
  })
})
