export function safeGuestRedirect(value: unknown, fallback = '/dashboard'): string {
  if (typeof value !== 'string' || !value.startsWith('/') || value.startsWith('//')) return fallback
  try {
    const decoded = decodeURIComponent(value)
    if (decoded.includes('\\') || [...decoded].some(character => character.charCodeAt(0) <= 32) || decoded.startsWith('//')) return fallback
    const url = new URL(value, 'https://local.invalid')
    if (url.origin !== 'https://local.invalid' || url.pathname.startsWith('//') || /^\/(login|register)(\/|$)/.test(url.pathname)) return fallback
    return `${url.pathname}${url.search}${url.hash}`
  } catch {
    return fallback
  }
}

export function selectedPublicPlanID(value: unknown): number | null {
  if (typeof value !== 'string' || !/^[1-9]\d*$/.test(value)) return null
  const id = Number(value)
  return Number.isSafeInteger(id) ? id : null
}
