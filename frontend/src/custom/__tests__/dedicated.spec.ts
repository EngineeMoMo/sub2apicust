import { describe, expect, it } from 'vitest'
import { dedicatedErrorCode, effectiveStatus, remainingQuota, type DedicatedView, type DedicatedWindow } from '@/custom/dedicated/api'
import { customRoutes } from '@/custom/routes'

const now = Date.parse('2026-09-28T08:00:00Z')
const sample = new Date(now - 60_000).toISOString()
const window: DedicatedWindow = { key: 'five_hour', remaining_percent: 65, resets_at: new Date(now + 60_000).toISOString(), stale: false }

describe('包号额度展示边界', () => {
  it('新鲜数据展示精确剩余比例，包括0%', () => {
    expect(remainingQuota(window, sample, now)).toBe(65)
    expect(remainingQuota({ ...window, remaining_percent: 0 }, sample, now)).toBe(0)
  })
  it.each([null, NaN, Infinity, -1, 101])('无效比例 %s 不是零额度', value => {
    expect(remainingQuota({ ...window, remaining_percent: value }, sample, now)).toBeNull()
  })
  it.each([null, 'bad', new Date(now - 16 * 60_000).toISOString(), new Date(now + 120_000).toISOString()])('无效采样时间 %s 不显示比例', value => {
    expect(remainingQuota(window, value, now)).toBeNull()
  })
  it('窗口已重置不推测剩余100%', () => {
    expect(remainingQuota({ ...window, resets_at: new Date(now).toISOString() }, sample, now)).toBeNull()
    expect(remainingQuota({ ...window, stale: true }, sample, now)).toBeNull()
    expect(remainingQuota({ ...window, resets_at: 'invalid' }, sample, now)).toBeNull()
  })
  it('页面停留期间包号到期也更新状态', () => {
    const account = { status: 'available', expires_at: new Date(now).toISOString() } as DedicatedView
    expect(effectiveStatus(account, now)).toBe('expired')
    expect(effectiveStatus({ ...account, status: 'revoked' }, now)).toBe('revoked')
    expect(effectiveStatus({ ...account, expires_at: 'invalid' }, now)).toBe('unavailable')
  })
  it('用户路由需登录、管理路由需管理员', () => {
    const user = customRoutes.find(route => route.name === 'DedicatedAccounts')!
    const admin = customRoutes.find(route => route.name === 'AdminDedicatedAccounts')!
    expect(user.meta?.requiresAuth).toBe(true)
    expect(admin.meta?.requiresAuth).toBe(true)
    expect(admin.meta?.requiresAdmin).toBe(true)
  })
  it('只提取机器错误码，不展示后端原始报错', () => {
    expect(dedicatedErrorCode({ response: { data: { reason: 'DEDICATED_ACCOUNT_CONFIG', message: 'secret' } } })).toBe('DEDICATED_ACCOUNT_CONFIG')
    expect(dedicatedErrorCode({ code: 'DEDICATED_ACCOUNT_CONFLICT' })).toBe('DEDICATED_ACCOUNT_CONFLICT')
    expect(dedicatedErrorCode(new Error('secret'))).toBe('')
  })
})
