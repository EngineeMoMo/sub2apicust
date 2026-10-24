import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const zh = {
  usageMode: 'Token 计量方式', usagePreserve: '保持账号原有计量方式',
  usageVerified: '精确计量（默认；缺少用量时拒绝正式请求）',
  usageEstimated: '允许估算计费（缺少精确用量时按估算 Token 扣费）',
  usageWarning: '旧账号保持原设置。估算模式按本站模型价格和倍率扣费，缓存读写按 0 计，不代表 Kiro 精确用量或账单；响应会标记 kiro_usage_source。连接测试成功不等于已取得精确用量。',
  type: 'Kiro 凭据类型', oauth: 'OAuth 凭据导入', apiKey: 'Kiro API Key',
  credentials: '凭据 JSON', file: '从 JSON 文件导入',
  hint: '支持 Kiro 凭据 JSON（含 refreshToken 或 refresh_token）。IDC 需同时包含 clientId 和 clientSecret。认证地区 auth_region 与运行地区 runtime_region 分别配置，默认 us-east-1。',
  apiKeyHint: '将 Kiro 官方生成的 ksk_ 密钥放入 JSON 的 api_key 字段；runtime_region 默认 us-east-1。这里填写上游 Kiro 密钥，与本站发给用户的 API Key 不同。',
  helpTitle: '凭据 / API Key 从哪里获取？',
  apiKeySource: '登录 Kiro 网页端，进入 Account → Settings → API Keys，创建密钥后立即复制。完整密钥只在创建时显示一次。',
  apiKeyPlans: 'API Key 需要 Pro、Pro+、Pro Max 或 Power 套餐。组织账号还需管理员允许创建 API Key。',
  apiKeyLink: '打开 Kiro 官方 API Keys',
  oauthSource: '先在 Kiro IDE 中登录自己的账号，再导入该客户端保存的凭据 JSON。常见缓存位置：',
  oauthWindows: 'Windows', oauthUnix: 'macOS / Linux',
  oauthExtra: 'Builder ID / IAM Identity Center 还需对应客户端注册文件中的 clientId 和 clientSecret，合并到同一个 JSON 再导入。clientIdHash 不能代替 clientId。不同版本的缓存位置可能不同，CLI 的 SQLite 数据库不能直接作为 JSON 导入。',
  oauthLink: '查看 Kiro 官方登录说明',
  oauthImportOnly: '本站当前提供凭据导入；首次登录授权需在 Kiro 客户端完成。',
  editHint: '留空保留现有凭据。填写 JSON 仅更新提供的字段，密钥不会从服务端回显。',
  invalid: '请填写有效的 Kiro 凭据 JSON，并检查重复字段。',
  required: '请填写 Kiro 凭据 JSON。',
  capabilities: '支持 Messages、Chat Completions、Responses 的文本、base64 图片与函数工具；默认要求精确用量；只有显式允许估算后才按估算 Token 计费。生成长度与采样参数由 Kiro 决定，max_tokens 不能作为费用硬上限；暂不支持 stop 或强制调用工具。'
}
const en: typeof zh = {
  usageMode: 'Token accounting mode', usagePreserve: 'Keep the existing account mode',
  usageVerified: 'Verified usage (default; reject requests with missing usage)',
  usageEstimated: 'Allow estimated billing (charge estimated tokens when exact usage is missing)',
  usageWarning: 'Existing accounts keep their mode. Estimates are charged at this site’s model prices and multiplier, with zero cache read/write discounts; they are not exact Kiro usage or invoices. Responses identify kiro_usage_source. A successful connection test does not verify exact usage.',
  type: 'Kiro credential type', oauth: 'Import OAuth credentials', apiKey: 'Kiro API key',
  credentials: 'Credentials JSON', file: 'Import a JSON file',
  hint: 'Paste Kiro credential JSON with refreshToken or refresh_token. IDC also requires clientId and clientSecret. Set auth_region and runtime_region separately; both default to us-east-1.',
  apiKeyHint: 'Put the official Kiro ksk_ key in the JSON api_key field; runtime_region defaults to us-east-1. This upstream key is different from API keys issued to users by this site.',
  helpTitle: 'Where do I get credentials or an API key?',
  apiKeySource: 'Sign in to Kiro Web, open Account → Settings → API Keys, create a key and copy it immediately. The full key is shown only once at creation.',
  apiKeyPlans: 'API keys require Pro, Pro+, Pro Max or Power. Organization accounts also need their administrator to allow API key creation.',
  apiKeyLink: 'Open official Kiro API Keys',
  oauthSource: 'Sign in to your account in Kiro IDE, then import the credential JSON saved by that client. Common cache locations:',
  oauthWindows: 'Windows', oauthUnix: 'macOS / Linux',
  oauthExtra: 'Builder ID / IAM Identity Center also needs clientId and clientSecret from the matching client registration file, merged into the same JSON. clientIdHash does not replace clientId. Cache locations may vary by client version; CLI SQLite databases cannot be imported as JSON.',
  oauthLink: 'Read the official Kiro sign-in guide',
  oauthImportOnly: 'This site imports credentials. Complete the initial sign-in authorization in your Kiro client.',
  editHint: 'Leave blank to keep current credentials. JSON updates only supplied fields. Saved secrets are never displayed.',
  invalid: 'Enter valid Kiro credential JSON and check for conflicting fields.',
  required: 'Enter Kiro credential JSON.',
  capabilities: 'Supports Messages, Chat Completions and Responses with text, base64 images and function tools. Exact usage is required by default; estimated token billing needs explicit opt-in. Kiro controls output length and sampling, so max_tokens is not a spending cap. Stop sequences and forced tool choice are unsupported.'
}
export function useKiroText() {
  const { locale } = useI18n()
  return computed(() => (locale?.value ?? 'zh-CN').startsWith('zh') ? zh : en)
}
