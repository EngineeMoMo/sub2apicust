export type KiroAccountType = 'oauth' | 'apikey'

const aliases: Record<string, string> = {
  accessToken: 'access_token', refreshToken: 'refresh_token', clientId: 'client_id',
  clientSecret: 'client_secret', profileArn: 'profile_arn', profileARN: 'profile_arn',
  expiresAt: 'expires_at', authMethod: 'auth_method', region: 'auth_region',
  ssoRegion: 'auth_region', sso_region: 'auth_region', apiRegion: 'runtime_region', api_region: 'runtime_region'
}
const acceptedKeys = new Set([
  'access_token', 'refresh_token', 'client_id', 'client_secret', 'profile_arn',
  'expires_at', 'auth_method', 'auth_region', 'runtime_region', 'api_key', 'model_mapping', 'usage_mode'
])

export function parseKiroCredentials(text: string): Record<string, unknown> {
  const document: unknown = JSON.parse(text)
  if (!document || typeof document !== 'object' || Array.isArray(document)) throw new Error('invalid_kiro_json')
  const outer = document as Record<string, unknown>
  const source = outer.credentials && typeof outer.credentials === 'object' && !Array.isArray(outer.credentials)
    ? outer.credentials as Record<string, unknown> : outer
  const credentials: Record<string, unknown> = {}
  for (const [key, value] of Object.entries(source)) {
    const canonical = aliases[key] ?? key
    if (!acceptedKeys.has(canonical)) continue
    if (Object.prototype.hasOwnProperty.call(credentials, canonical) && credentials[canonical] !== value) throw new Error('conflicting_kiro_field')
    credentials[canonical] = value
  }
  if (Object.keys(credentials).length === 0) throw new Error('empty_kiro_credentials')
  if (credentials.usage_mode !== undefined && !['verified', 'estimated'].includes(credentials.usage_mode as string)) throw new Error('invalid_kiro_usage_mode')
  return credentials
}
