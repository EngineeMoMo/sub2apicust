import { beforeEach, describe, expect, it, vi } from 'vitest'
import { completeSubscriptionAssignment, prepareSubscriptionAssignment } from '../subscriptionAssignment'

const request = { user_id: 11, group_id: 7, plan_id: 71, validity_days: 3 }

describe('single-user subscription assignment keys', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
    localStorage.setItem('auth_user', JSON.stringify({ id: 777 }))
    completeSubscriptionAssignment(prepareSubscriptionAssignment(request))
  })

  it('preserves the key for identical requests and canonicalizes property order', () => {
    const first = prepareSubscriptionAssignment(request)
    const retry = prepareSubscriptionAssignment({ validity_days: 3, plan_id: 71, group_id: 7, user_id: 11 })
    expect(retry.key).toBe(first.key)
    expect(retry.request).toEqual(first.request)
  })

  it('starts a new operation after confirmed success', () => {
    const first = prepareSubscriptionAssignment(request)
    completeSubscriptionAssignment(first)
    expect(prepareSubscriptionAssignment(request).key).not.toBe(first.key)
  })

  it('separates users, groups, plans, durations and administrators', () => {
    const first = prepareSubscriptionAssignment(request)
    for (const changed of [{ user_id: 12 }, { group_id: 8 }, { plan_id: 72 }, { validity_days: 7 }]) {
      const other = prepareSubscriptionAssignment({ ...request, ...changed })
      expect(other.key).not.toBe(first.key)
      completeSubscriptionAssignment(other)
    }
    localStorage.setItem('auth_user', JSON.stringify({ id: 778 }))
    const otherAdmin = prepareSubscriptionAssignment(request)
    expect(otherAdmin.key).not.toBe(first.key)
    completeSubscriptionAssignment(otherAdmin)
  })

  it('recovers the pending key from session storage after module reload', async () => {
    const first = prepareSubscriptionAssignment(request)
    vi.resetModules()
    const reloaded = await import('../subscriptionAssignment')
    const retry = reloaded.prepareSubscriptionAssignment(request)
    expect(retry.key).toBe(first.key)
    reloaded.completeSubscriptionAssignment(retry)
    completeSubscriptionAssignment(first)
  })
})
