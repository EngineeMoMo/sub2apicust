# Kiro 原生接入（本地Docker，精确计量/账单待验收）

## 2026-10-24 计量兼容修复（当前状态）

六个参考仓人工巡检均与锁定SHA相同，没有新提交/Release变化；既有kirocc已经提供metering及估算回退设计，本轮局部适配见[来源台账](../tools/kiro/ports-20261024.json)。当前本机版本0.2.17-custom.kiro-usage-20261024-1092d0243ba0，镜像/运行Image dd27a780e6af，三服务healthy。

实际OAuth账号testkiro、Sonnet4.5受控单次生成返回OK.，测试完成成功，同时明确warning、usage_source=missing、billing_ready=false；事件计数为正文2、context/metadata/metering各1。有限计数没有原始payload，不能认定所有Kiro都缺精确量或推断未测模型权限。账号保留verified，正式请求仍要求精确Token；估算模式三协议/流式及非流式已做隔离模拟回归，未改实号模式或实际收费验证。

Go41顶层/128含子项、前端5文件152项、变更eslint/类型、镜像内前端/i18n3项/Family3顶层25含子项/Linux embed通过；原数据/设置/317迁移/配置与卷保持。完整证据及新备份output/kiro-usage-fix-20261024/verification.json；回退before-kiro-usage-20261024保留上一版。不自动估算收费、不创建定时任务，未提交推送或生产部署。

## 2026-10-24 修复前诊断（历史）

用户截图中OAuth账号testkiro、模型claude-sonnet-4-5-20250929已输出OK.，再报“未返回可核实的Token用量”。本机原日志14:30:57/14:31:12确认该错误，Haiku也有同错；部分其他模型返回上游连接失败，状态/正文未录，须分别查证。

本站解码器只接收metadataEvent.tokenUsage中的精确输入/输出字段，流末HasUsage=false主动报错；测试已先转发正文，所以会先显示响应再失败。正式三协议网关也复用该解码器，仍会被阻断。参考项目有估算回退，本仓未启用；没有实际事件结构时，不知道是本响应未提供精确用量还是解析字段尚未兼容，不能由截图推断所有Kiro响应都无Token。

本轮仅只读诊断，未改测试/网关/收费规则，未新发模型请求或部署。下一步须取最小脱敏事件结构、审阅可验证的用量并区分“能生成”与“计量就绪”，不能把提示改绿当完整修复，也不擅自估算收费。证据output/kiro-usage-diagnosis-20261024/diagnosis.json；下方历史离线回归及本机部署不构成实号/账单验收。

更新：2026-10-23，Asia/Shanghai。已完成本机 Docker 部署，尚未提交推送或部署生产；本机最新版本/证据以文末及 HANDOFF 四/五为准。参考来源与人工巡检见 [KIRO_REFERENCE_TRACKING.md](KIRO_REFERENCE_TRACKING.md)，逐项适配台账见 [ports-20261023.json](../tools/kiro/ports-20261023.json)。

## 添加账号

管理员进入“账号管理 → 添加账号 → Kiro”，选择 OAuth 凭据导入或 Kiro API Key，粘贴 JSON 或导入一个 JSON 文件。JSON 可以是凭据对象，也可以放在 credentials 子对象内；不支持批量账号数组。名称、分组、代理、并发、优先级、倍率、到期时间沿用原账号设置。

### API Key 从哪里获取

在 [Kiro 网页端 API Keys](https://app.kiro.dev/settings/api-keys) 登录自己的账号，进入 Account → Settings → API Keys 创建密钥，立即复制完整的 `ksk_` 密钥（只显示一次）。官方文档列出的支持套餐为 Pro、Pro+、Pro Max、Power，Free 不支持创建；组织账号也受管理员策略限制。随后在本站表单选择“Kiro API Key”，把密钥放入下方 JSON 的 `api_key` 字段；这里与本站给用户下发的 API Key 是两种不同密钥。

官方来源：[认证与 API Key 说明](https://kiro.dev/docs/getting-started/authentication/)、[AWS 的入口说明](https://aws.amazon.com/blogs/devops/accelerating-development-workflows-with-kiro-cli-as-a-pre-commit-and-git-hook-agent/)，人工核查日期/图标下载摘要见 `tools/kiro/ui-sources-20261023.json`。有密钥不等于本站代理已用该真实账号验收，运行地区/模型/账单仍须实际确认。

### OAuth 凭据从哪里获取

先在 Kiro IDE 中登录自己的账号，再导入客户端保存的 JSON。参考项目记录的常见位置：Windows 为 `%USERPROFILE%\.aws\sso\cache\kiro-auth-token.json`，macOS/Linux 为 `~/.aws/sso/cache/kiro-auth-token.json`。这是客户端缓存惯例，不是官方对所有版本的保证；CLI 可能用 SQLite，不能直接作为 JSON 导入。

Builder ID / IAM Identity Center 还须合并对应注册 JSON 的 `clientId` / `clientSecret`；`clientIdHash` 只用于定位注册文件，不能当 `clientId`。本站没有首次浏览器或设备码授权入口，首次登录须在客户端完成；只由账号所有者读取自己的凭据，不把密钥发到聊天或受版本控制的文件。路径及注册文件说明参考 [kiro-gateway 作者 README](https://github.com/jwadow/kiro-gateway)，官方登录方式见 [Kiro 认证文档](https://kiro.dev/docs/getting-started/authentication/)。

Social OAuth 示例（示例值不能使用）：


```json
{
  "auth_method": "social",
  "refresh_token": "YOUR_REFRESH_TOKEN",
  "profile_arn": "YOUR_PROFILE_ARN",
  "auth_region": "us-east-1",
  "runtime_region": "us-east-1"
}
```

不要把真实密钥写入受版本控制的文件。

IDC / Builder ID 示例：


```json
{
  "auth_method": "idc",
  "refresh_token": "YOUR_REFRESH_TOKEN",
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET",
  "profile_arn": "YOUR_PROFILE_ARN",
  "auth_region": "us-east-1",
  "runtime_region": "us-east-1"
}
```

静态 API Key 示例：


```json
{
  "api_key": "ksk_YOUR_KIRO_API_KEY",
  "runtime_region": "us-east-1"
}
```

- 支持 accessToken / refreshToken / clientId / clientSecret / profileArn / expiresAt 等 camelCase 与 snake_case。region / ssoRegion 指认证地区，apiRegion 指运行地区；同一个 JSON 内别名冲突会拒绝。
- 认证地区和运行地区分别配置，默认均为 us-east-1。实际可用运行地区须用该账号验证，不根据登录地区自动推断。
- 仅导入 access_token 时须同时提供 expires_at 和 profile_arn；到期后无法自动续期。expires_at 支持 Unix 秒、毫秒或 RFC3339。
- IDC 刷新须有 client_id / client_secret / refresh_token，通常还须从账号凭据提供 profile_arn。Social 刷新如果未返回 profileArn，也须补齐；不编造 profile。
- 编辑时留空保留凭据，提供的字段覆盖对应字段，其余保留。替换 refresh_token 且未提供新的 access_token / expires_at / profile_arn 时会清除对应旧值，避免使用旧身份继续请求。管理响应不回显访问令牌、刷新令牌、API Key 或 client_secret。
- 运行端点固定为 runtime.<运行地区>.kiro.dev，表单不接受自定义 Base URL。设置账号代理后，认证刷新和生成都走该代理；代理缺失、失效或停用时报错。

## 分组与调用

建立 Kiro 分组并绑定账号，再给用户的 API Key 选择该分组；也可依照现有组合分组规则选中 Kiro 账号。账户调度、专属账号限制、分组鉴权、日/月净充值门槛、余额门槛、并发及本站费用结算沿用原链路。本轮没有修改这些已有业务规则。

支持本站 Messages、Chat Completions、Responses 三个入口的流式和非流式输出。默认模型目录取自已固定的参考代码，包括 claude-sonnet-4-6、claude-opus-4-6、claude-haiku-4-5-20251001 及 4.5 系列；目录不代表账号有全部模型权限。可在凭据中设置 model_mapping；未知模型原样送给 Kiro，不随意裁剪自定义模型的日期后缀。

后台手动刷新、批量刷新、连接测试均进入 Kiro 服务；停用/错误账号允许管理员主动连接测试，但普通网关请求仍检查最新账号状态与调度冷却。后台定期刷新 OAuth 令牌属于账号运行能力，与用户取消的“参考项目更新巡检”无关。

## 支持范围与明确限制

- 文本、文本系统提示、base64 的 PNG/JPEG/GIF/WebP 图片、函数工具、多轮工具调用和文本工具结果。单张解码图片限 4 MiB，工具输入限 1 MiB；单 EventStream 帧限 4 MiB、总响应限 32 MiB，请求最长 5 分钟。
- 拒绝 URL 图片、工具结果图片、非文本 system、内置服务工具、stop_sequences、强制 tool_choice、助手预填和结构化输出约束，避免静默丢弃。
- Responses 为无状态兼容：每次提交完整历史；previous_response_id、conversation、store=true、background=true 均拒绝。Chat 仅支持一个候选回复。
- Kiro 生成长度和采样采用上游默认行为。本地接收的 max_tokens / max_output_tokens、temperature、top_p、top_k 及 thinking.budget_tokens 不保证被 Kiro 执行，不能作为费用硬上限；本轮没有用字符近似截断冒充 Token 限制。output_config.effort 可映射到参考协议字段，实际模型行为待实测。
- 没有网站内的 Kiro 设备码登录、Social 浏览器授权、账号余额查询、在线模型发现或 count_tokens。首次获取凭据仍由账号所有者在现有 Kiro 客户端/授权流程完成。连接测试会发起真实生成，若账号可用可能产生上游费用。

## 用量与刷新策略

优先接收metadataEvent.tokenUsage的uncachedInputTokens/outputTokens/cacheReadInputTokens/cacheWriteInputTokens；同时兼容meteringEvent.inputTokens/outputTokens。metadata优先，按最近有效快照覆盖，不累加重复快照；全零占位及只有credits的事件不当作精确Token。未缓存输入与缓存读写分开记录，13/3/2/17夹具保持不变。

计量方式默认verified：缺少可识别的精确输入/输出用量时正式请求返回错误，流式不发送成功终止事件；连接测试可成功，并另发黄色warning及billing_ready=false。收到精确字段后仍按本站模型价格/倍率结算，不代表已与Kiro账单对账。

管理员接受估算后，可在账号管理→编辑Kiro账号→Token计量方式中选择“允许估算计费”并保存；编辑默认保留已有方式，服务端不会回显秘密。仅更新模式的等效JSON为`{"usage_mode":"estimated"}`，旧凭据由服务端合并保留。恢复精确策略选择verified。已有账号不会自动改为estimated。

estimated只在完整有效生成但没有精确Token时回退：输入复用本站tiktoken估算，输出估算可见正文/思考及工具名与参数；缓存读写为0，不模拟缓存折扣、不按context百分比或credits换美元。按本站价格和倍率对估算Token收费，估算不是Kiro精确用量或账单。redacted_thinking、CRC损坏、截断、非法计量及空响应仍报错；已有精确Token优先，不能被估算覆盖。

三协议非流式JSON与流末事件增加kiro_usage_source（metadata/metering/estimated）；非流式另有X-Sub2API-Kiro-Usage-Source头，请求响应带X-Sub2API-Kiro-Usage-Policy，估算记录不含正文/密钥的结构化日志。客户端断连仍在有限时长内继续读取上游以收集用量；估算策略不改变原生成长度、采样或max_tokens限制。

OAuth 刷新共用现有互斥/分布式协调。401/403 最多强制刷新一次。轮换写入使用“完整旧凭据 + 代理 ID + 平台/类型”的 PostgreSQL CAS，更新凭据与调度 outbox 在同一语句内；并发管理员修改获保留。保存失败不返回新 token，并终止后台本轮刷新重试；后续周期仍可能需要重新导入被上游消费的旧 refresh_token。

## 验证及未完成的验收

已通过Go定向46顶层/119含子项（零跳过）、真实PG 9组CAS、前端关键70文件998项、类型检查、生产构建、Windows embed及巡检工具24项离线回归。改动范围eslint零错误，1条原有warning。全量后端仍有Command Code/Ollama两项失败，已在未修改cafcd2bac基线复现；不把定向结果当作全量通过。验证命令、失败记录与源文件哈希见[本地验证摘要](../tools/kiro/native-verification-20261023.json)及output/kiro-native-20261023/verification.json。已实际运行独立 kiro_native_test 数据库的 CAS、代理变更、并发轮换、outbox 失败回滚回归，以及三协议真实服务入口的模拟上游测试；不是只测试协议转换函数。

已取得本机真实OAuth账号Sonnet4.5最小文本生成证据；精确Token和Kiro账单、实号估算结算、图片/工具多轮、其他模型权限与地区仍未验收。未提交推送或部署生产，发布时须核验同一业务SHA的远程CI，不能复用其他功能的旧CI或镜像。

## 本机部署记录（2026-10-23）

本机访问 `http://127.0.0.1:8080`，登录后进入“账号管理 → 添加账号 → Kiro”导入已有凭据JSON或文件。网站内首次浏览器授权仍不在本轮范围。

版本 `0.2.17-custom.kiro-local-20261023-661725e62a56`，固定镜像 `sub2apicust:kiro-local-20261023-661725e62a56`，沿用 `sub2apicust-local` 项目与原 `.env` / override / 三个数据卷。三服务healthy，管理员真实HTTP登录、版本及Kiro筛选接口通过；HTTP账号管理脚本确有Kiro按钮及JSON/文件字段。原数据计数保持，迁移304→317，含已有日充值门槛261且默认关闭。这里只完成本地部署与入口验证，未用真实Kiro凭据调用模型。

备份及证据在 `output/kiro-local-docker-20261023/`，摘要 `verification.json`、源码锁 `source-manifest.json`、备份 `database.dump`（668744字节，pg_restore目录校验通过）和配置/app_data/Redis副本。旧镜像标签 `sub2apicust:before-kiro-20261023` 保留。需要镜像回退时先核对本机现场，再把旧镜像tag为 `sub2apicust:local-theme`，沿用原compose仅重建应用；不要自动恢复数据库或删卷。

启动日志另有历史价格库远程哈希连接0.0.0.0失败；本地256模型价格库已加载，网络配置未改。参考项目更新仍只人工巡检。

## 图标与获取说明修正（2026-10-23，本机当前版本）

当前版本 `0.2.17-custom.kiro-help-20261023-ae5ad68369d1`，固定镜像 `sub2apicust:kiro-help-20261023-ae5ad68369d1`，容器/镜像 Image 同为 `sha256:9c12017f81977fabbf37694ce1973129e0c2b5b66adb8162cc3195be6bb1be92`。官网图标通过本地 SVG 打包内联到共用 PlatformIcon，创建按钮已接入；凭据表单新增按类型显示的获取帮助，新建默认展开、编辑默认折叠。浏览器可用 Ctrl+F5 更新页面资源。

定向5文件144项、变更lint、前端类型/生产构建、i18n3项、Family3顶层/25含子项及Linux embed通过；三服务healthy，实际HTTP脚本与官网SVG几何、官方链接/提示、管理员登录/版本及Kiro筛选均验证。原业务计数、281项设置摘要、317项迁移、PG/Redis容器及三个数据卷保持；本轮无后端/依赖/迁移变动。没有可复用的浏览器后台会话，因此未声称后台真实视觉验收，也未做真实Kiro模型调用/账单验收。

新证据/备份 `output/kiro-ui-help-20261023/verification.json`、`database.dump`（695345字节，pg_restore目录校验通过）及原配置/数据文件副本。回退标签 `sub2apicust:before-kiro-help-20261023` 保留上一版Kiro镜像；更早备份/镜像不覆盖，不自动恢复数据库。构建后仅整理了 `frontend/src/custom/UPGRADE.md` 说明位置，源码清单差异单列，运行源码无变化。未提交推送/生产部署，参考源仅人工巡检。
