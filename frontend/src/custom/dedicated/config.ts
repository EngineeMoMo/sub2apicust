const configIssues: Record<string, [string, string]> = {
  account_missing: ['所选账号不存在或已删除，账号编号', 'The selected account is missing or deleted. Account IDs'],
  group_missing: ['所选分组不存在或已删除，分组编号', 'The selected group is missing or deleted. Group IDs'],
  group_inactive: ['所选分组未启用，分组编号', 'The selected group is disabled. Group IDs'],
  group_not_exclusive: ['所选分组未开启专属分组开关；名称中含专属不等于已开启，分组编号', 'Enable the exclusive-group setting; naming the group exclusive does not enable it. Group IDs'],
  platform_mismatch: ['所选分组与账号的平台不一致，分组编号', 'The group and account platforms do not match. Group IDs'],
  group_not_standard: ['所选分组不是标准分组，分组编号', 'The selected group is not a standard group. Group IDs'],
  fallback_route: ['所选分组仍配置了备用路由或无效请求备用路由，分组编号', 'Remove both fallback routes from the group. Group IDs'],
  account_type: ['账号类型不支持包号；Claude需OAuth或Setup Token，ChatGPT/Codex需OAuth，账号编号', 'Unsupported account type. Claude requires OAuth or Setup Token; ChatGPT/Codex requires OAuth. Account IDs'],
  shadow_account: ['所选账号是影子账号，账号编号', 'A shadow account cannot be assigned. Account IDs'],
  account_has_shadows: ['所选账号仍有未删除的影子账号，影子账号编号', 'The account still has shadow accounts. Shadow account IDs'],
  account_not_in_group: ['所选账号尚未加入所选分组，账号编号', 'The account has not joined the selected group. Account IDs'],
  account_other_groups: ['所选账号还关联其他分组，其他分组编号', 'The account also belongs to other groups. Other group IDs'],
  account_unsafe_groups: ['账号关联的其他分组不符合标准专属分组要求，请检查专属开关、状态、平台和备用路由，分组编号', 'Another account group must be active, exclusive, standard, on the same platform and without fallback routes. Group IDs'],
  related_group_other_accounts: ['账号关联的其他专属分组还包含别的账号，分组编号', 'Another linked exclusive group contains a different account. Group IDs'],
  group_other_accounts: ['所选分组还包含其他账号，其他账号编号', 'The group also contains other accounts. Other account IDs'],
  no_members: ['未选择包号成员，分组编号', 'No assignment members were selected. Group IDs'],
  other_authorized_users: ['账号关联的专属分组还授权给包号名单以外的用户；同批共用者请全部加入同一条包号，用户编号', 'Users outside the assignment still have access to a linked exclusive group. Add all intended members to the same assignment. User IDs'],
  group_subscriptions: ['所选分组仍有未删除的订阅记录，分组编号', 'The group still has subscription records. Group IDs'],
  member_unavailable: ['包号成员未启用、不存在或已删除，用户编号', 'An assignment member is disabled, missing or deleted. User IDs'],
  member_not_authorized: ['包号成员尚未获得所选分组授权，用户编号', 'An assignment member does not have access to the selected group. User IDs'],
  non_member_keys: ['账号关联的专属分组仍有非包号成员的未停用密钥，密钥所属用户编号', 'Users outside the assignment still own enabled keys in a linked exclusive group. Key owner IDs']
}

export function dedicatedConfigMessage(error: unknown, locale: string): string | null {
  const candidate = error as { response?: { data?: { metadata?: unknown } }; metadata?: unknown } | null
  const metadata = candidate?.response?.data?.metadata ?? candidate?.metadata
  if (!metadata || typeof metadata !== 'object') return null
  const { config_issue: issue, resource_ids: resourceIDs } = metadata as Record<string, unknown>
  if (typeof issue !== 'string' || !Object.prototype.hasOwnProperty.call(configIssues, issue)) return null
  const message = configIssues[issue]![locale.startsWith('zh') ? 0 : 1]
  return typeof resourceIDs === 'string' && /^[1-9]\d*(, [1-9]\d*){0,9}$/.test(resourceIDs) ? `${message}: ${resourceIDs}` : message
}
