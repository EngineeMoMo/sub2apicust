import { describe, expect, it } from 'vitest'
import { dedicatedConfigMessage } from '@/custom/dedicated/config'

describe('包号配置失败原因', () => {
  it('保留归一化API错误中的实际冲突编号，提供中英文说明', () => {
    const error = { metadata: { config_issue: 'account_other_groups', resource_ids: '55, 66' } }
    expect(dedicatedConfigMessage(error, 'zh-CN')).toBe('所选账号还关联其他分组，其他分组编号: 55, 66')
    expect(dedicatedConfigMessage(error, 'en')).toContain('Other group IDs: 55, 66')
  })
  it('Axios原始响应也能显示缺少授权的成员编号', () => {
    expect(dedicatedConfigMessage({ response: { data: { metadata: { config_issue: 'member_not_authorized', resource_ids: '11' } } } }, 'zh')).toContain('包号成员尚未获得所选分组授权，用户编号: 11')
  })
  it.each([null, {}, { metadata: { config_issue: '__proto__' } }, { metadata: { config_issue: 'unknown', resource_ids: 'secret' } }])('旧版或未知诊断回退原提示，不直接展示原始错误', error => {
    expect(dedicatedConfigMessage(error, 'zh')).toBeNull()
  })
  it('非法编号不展示原始字段内容', () => {
    expect(dedicatedConfigMessage({ metadata: { config_issue: 'non_member_keys', resource_ids: 'Token-secret' } }, 'zh')).not.toContain('Token-secret')
  })
})
