import type { AssignSubscriptionRequest } from '@/types'

const pendingKeys = new Map<string, string>()

export interface SubscriptionAssignmentOperation {
  request: AssignSubscriptionRequest
  key: string
  storageKey: string
}

export function prepareSubscriptionAssignment(request: AssignSubscriptionRequest): SubscriptionAssignmentOperation {
  let adminId: number | string = 'unknown'
  try {
    const user = JSON.parse(localStorage.getItem('auth_user') ?? 'null') as { id?: unknown } | null
    if (typeof user?.id === 'number' && Number.isSafeInteger(user.id) && user.id > 0) adminId = user.id
  } catch {
    adminId = 'unknown'
  }
  const payload: AssignSubscriptionRequest = {
    user_id: request.user_id,
    group_id: request.group_id,
    plan_id: request.plan_id,
    validity_days: request.validity_days
  }
  const storageKey = `sub2api:admin:subscription-assign:${adminId}:${JSON.stringify(payload)}`
  let key = pendingKeys.get(storageKey)
  try {
    key ??= sessionStorage.getItem(storageKey) ?? undefined
  } catch {
    key = pendingKeys.get(storageKey)
  }
  if (!key) {
    const requestId = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`
    key = `subscription-assign-${adminId}-${requestId}`
  }
  pendingKeys.set(storageKey, key)
  const operation = { request: payload, key, storageKey }
  try {
    sessionStorage.setItem(storageKey, key)
  } catch {
    return operation
  }
  return operation
}

export function completeSubscriptionAssignment(operation: SubscriptionAssignmentOperation) {
  pendingKeys.delete(operation.storageKey)
  try {
    sessionStorage.removeItem(operation.storageKey)
  } catch {
    return
  }
}
