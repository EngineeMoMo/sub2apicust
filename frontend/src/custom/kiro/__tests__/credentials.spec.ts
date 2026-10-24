import { describe, expect, it } from 'vitest'
import { parseKiroCredentials } from '../credentials'

describe('Kiro credential import', () => {
  it.each(['verified', 'estimated'])('accepts the explicit accounting mode %s', mode => {
    expect(parseKiroCredentials(JSON.stringify({ usage_mode: mode }))).toEqual({ usage_mode: mode })
  })
  it.each(['auto', '', null, true, 1])('rejects an unsupported accounting mode %s', mode => {
    expect(() => parseKiroCredentials(JSON.stringify({ refreshToken: 'private', usage_mode: mode }))).toThrow()
  })
  it('canonicalizes social and IDC exports without retaining unrelated secrets', () => {
    expect(parseKiroCredentials(JSON.stringify({
      refreshToken: 'refresh', clientId: 'client', clientSecret: 'secret', region: 'eu-west-1',
      apiRegion: 'us-east-1', cookie: 'discard', password: 'discard'
    }))).toEqual({ refresh_token: 'refresh', client_id: 'client', client_secret: 'secret', auth_region: 'eu-west-1', runtime_region: 'us-east-1' })
  })
  it('accepts account export containers and model mappings', () => {
    expect(parseKiroCredentials('{"credentials":{"api_key":"key","model_mapping":{"public":"native"}}}'))
      .toEqual({ api_key: 'key', model_mapping: { public: 'native' } })
  })
  it.each(['[]', 'null', 'true', '{bad', '{}', '{"cookie":"secret"}', '{"refreshToken":"a","refresh_token":"b"}'])('rejects invalid, empty or conflicting imports: %s', input => {
      expect(() => parseKiroCredentials(input)).toThrow()
    })
})
