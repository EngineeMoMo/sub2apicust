import { describe, expect, it } from 'vitest'
import { normalizeGroupRateSchedule } from '../api'

describe('一次性福利时段校验', () => {
  it('北京时间连续时段正确转换为UTC并排序', () => {
    expect(normalizeGroupRateSchedule([
      { starts_at: '2026-11-01T18:00', ends_at: '2026-11-01T20:30', rate_multiplier: 0.15 },
      { starts_at: '2026-11-01T12:00', ends_at: '2026-11-01T18:00', rate_multiplier: 0.1 },
    ])).toEqual([
      { starts_at: '2026-11-01T04:00:00.000Z', ends_at: '2026-11-01T10:00:00.000Z', rate_multiplier: 0.1 },
      { starts_at: '2026-11-01T10:00:00.000Z', ends_at: '2026-11-01T12:30:00.000Z', rate_multiplier: 0.15 },
    ])
  })
  it.each([
    { starts_at: '2026-11-01T18:00', ends_at: '2026-11-01T17:00', rate_multiplier: 0.1 },
    { starts_at: '2026-11-01T18:00', ends_at: '2026-11-01T20:00', rate_multiplier: -1 },
    { starts_at: '2026-11-01T18:00', ends_at: '2026-11-01T20:00', rate_multiplier: Number.NaN },
    { starts_at: 'invalid', ends_at: '2026-11-01T20:00', rate_multiplier: 0.1 },
  ])('拒绝非法时间或倍率：%j', period => { expect(normalizeGroupRateSchedule([period])).toBeNull() })
  it('拒绝重叠时段', () => {
    const period = { starts_at: '2026-11-01T18:00', ends_at: '2026-11-01T20:30', rate_multiplier: 0.1 }
    expect(normalizeGroupRateSchedule([period, period])).toBeNull()
  })
})
