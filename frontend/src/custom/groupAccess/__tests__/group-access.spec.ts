import { describe, expect, it } from 'vitest'
import { fromBeijingDateTime, toBeijingDateTime, validGroupThreshold } from '../api'

describe('分组开放规则的时间和金额', () => {
  it('将北京时间保存为绝对时间，与浏览器时区无关', () => {
    expect(fromBeijingDateTime('2026-11-01T00:00')).toBe('2026-10-31T16:00:00.000Z')
    expect(fromBeijingDateTime('2026-10-15T12:34:56')).toBe('2026-10-15T04:34:56.000Z')
    expect(toBeijingDateTime('2026-10-31T16:00:00Z')).toBe('2026-11-01T00:00:00')
  })

  it.each(['', '2026-02-30T12:00', '2026-13-01T12:00', '2026-10-15', '2026-10-15T25:00'])('拒绝无效到期时间 %s', value => {
    expect(fromBeijingDateTime(value)).toBeNull()
  })

  it.each([NaN, Infinity, -1, 0.001, 1e13])('拒绝无效门槛 %s', amount => {
    expect(validGroupThreshold(amount, true)).toBe(false)
    expect(validGroupThreshold(amount, false)).toBe(false)
  })

  it('关闭时允许零，开启时必须大于零，支持两位小数', () => {
    expect(validGroupThreshold(0, false)).toBe(true)
    expect(validGroupThreshold(0, true)).toBe(false)
    expect(validGroupThreshold(0.01, true)).toBe(true)
    expect(validGroupThreshold(100, true)).toBe(true)
  })
})
