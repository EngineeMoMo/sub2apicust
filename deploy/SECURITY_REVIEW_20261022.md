# 2026-10-22 网站防护审查

后续状态：用户随后授权修复，已完成本地补修与定向验证，见 `SECURITY_HARDENING_20261022.md`。下文保留审查时的现场快照；生产未部署、管理员绑定/step-up启用与反代配置仍待，不将历史缺口误写成已在线修复。

审查对象：`https://ai.mofamilys.com` 的匿名只读响应，以及当前本地工作树的认证、管理员权限、支付和部署代码。线上抽查时间以证据响应的 `Date` 为准（2026-10-22，约北京时间 10:00）。当前本地含未发布注册 IP 限额和订阅兑换修改；**不能把本地源码行为直接当作生产已部署行为**。

本轮只审查、隔离复现和记录，没有修改业务实现、线上配置、数据库或实际订单，没有提交、推送或部署。未进行爆破、压测或生产写请求。下列 P1/P2 表示建议处理顺序，不是已证实的漏洞评级。

## 已核实：优先加固项

| 顺序 | 事实与取证 | 建议 | 结论边界 |
| --- | --- | --- | --- |
| P1 管理员二次验证 | 线上公开设置 `totp_enabled=false`。本地 `backend/internal/handler/auth_handler.go:265` 的密码登录 TOTP 挑战受全站开关与用户绑定状态共同控制。 | 开启全站 TOTP 功能，并逐一为实际管理员绑定、验证登录；保留恢复方案。核验是否启用敏感操作二次验证。 | 仅打开全站功能不会自动强制所有管理员完成绑定。未登录生产管理员，也不知道 OAuth 上游是否另有 MFA。 |
| P1 支付敏感操作二次验证覆盖 | `backend/internal/server/routes/payment.go:96` 的管理组挂了管理员鉴权、审计与合规确认，没有传入或挂载 `stepUpAuth`。退款、支付配置、通道增改删、人工收款/释放的 handler 也未调用 `EnforceStepUp`。 | 为退款、收款密钥/通道变更、人工记账/释放等接入已有 step-up；同步评估其他敏感设置的服务端门控。机器凭证应单独设计最小权限。 | 已认证管理员的敏感操作加固缺口；**不是匿名可退款或已证实支付越权**。既有审计和人工确认不能等同二次身份验证。生产精确代码版本未知。 |
| P2 HTTPS 策略 | HTTPS 首页 200，有 CSP 等响应头，但没有 `Strict-Transport-Security`；HTTP `/login` 返回 200、正文 `OK`，没有跳转。见 `https-home.headers`、`http-login.headers` 与 `.body`。 | 反代增加 HTTP 到 HTTPS 的 301/308 跳转，并在 HTTPS 响应部署 HSTS。`includeSubDomains` 和 preload 需先确认全部相关子域支持 HTTPS。 | HTTP 当前是两字节占位响应，**未发现它在返回登录页面，不能据此声称密码已明文传输**。本轮只证明所抽查响应缺少该策略。 |
| P2 公开订单接口限流 | `payment.go:80` 的 `/payment/public/orders/resolve` 没挂单独限流。本地用实际路由注册、miniredis 和空 JSON 连续请求 60 次，均到达 handler 参数校验（400），没有 429。旧 `/orders/verify` 有 20 次/分钟/IP 限制，但现有回归也证实 Redis 故障时放行。 | 对恢复入口补 IP/全局请求预算；合法凭证通过后再考虑按订单限制主动查单频率。给支付查询的 Redis 故障场景配置边缘兜底额度，避免全面影响付款中的正常用户。 | 恢复凭证先验签再读库；合法待处理/过期订单可能触发主动查单。没有有效凭证不能任意触发该查单逻辑。未测生产请求洪泛，不知道边缘是否已有额外限流。 |

### 支付二次验证的源码定位

- 路由：`backend/internal/server/routes/payment.go:96`、`:109`、`:112`、`:114` 及订单/通道子组。
- 退款/通道/配置 handler：`backend/internal/handler/admin/payment_handler.go:232`、`:425`、`:442`、`:463`、`:502`。
- 人工释放/记账：`backend/internal/handler/admin/custom_payment_collection.go:37`、`:73`。
- 现有可复用门控：`backend/internal/server/middleware/step_up.go`，开启后要求真人 JWT 会话、用户已绑定 TOTP 和当前会话有效授权；禁止管理员 API Key。
- 管理账号导出、备份等其他路径已有 step-up：`backend/internal/server/routes/admin.go`。本轮没有发现需要从零另建一套二次验证系统的证据。

### 公开订单入口的验证边界

`backend/internal/service/payment_resume_lookup.go:12` 先校验签名凭证，再根据订单 ID 取库并核对用户/通道/支付类型；待处理或过期订单才进入主动查单。签名校验见 `payment_resume_service.go:466`。隔离复现使用空 JSON，不触库、不请求支付商、不接触线上；证明的是该路由缺少请求频率门控，不能将它解释为可伪造订单凭证、无限真实查单或修改金额。

## 已核实：现有保护和验证结果

- 线上公开配置：注册已开启、邮箱验证已开启、Turnstile 已开启、邮箱白名单 13 项；邀请码关闭、注册邮箱域名配额关闭。公开 DTO 未暴露注册 IP 限额，**无法以字段缺失判断线上开关或版本**。
- 匿名 `GET /api/v1/admin/settings` 与 `GET /api/v1/user/profile` 均为 401，没有返回管理员设置或个人数据。
- HTTPS 首页含 nonce CSP、`frame-ancestors 'none'`、`base-uri 'self'`、`form-action 'self'`、`X-Frame-Options: DENY`、`X-Content-Type-Options: nosniff` 和 Referrer-Policy。CSP 的宽泛 `connect-src https:` 可后续盘点收紧，但本轮未证实 XSS。
- 本地认证路由已有 Redis 故障拒绝的限流配置；邮箱验证码已有原子尝试次数上限（5 次）；支付回调先验签再调用支付通知处理。这些是已存在的代码保护，不应重复描述成尚未实现。
- 本轮 Go 1.27.3 离线定向回归：54 个顶层测试、106 个含子测试 pass 事件，失败/跳过均为 0。范围为管理员 token version、step-up、CORS、安全头、公开订单限流、支付回调提供商验签、恢复凭证/订单匹配和返回 URL 校验。
- 另有 1 个隔离路由复现通过：上述 60 次空 JSON 请求。合计 55 个顶层测试、107 个含子测试 pass 事件。复现通过表示确认缺口，**不表示已经补修**。临时测试已移出业务目录，源文件留在证据目录。
- 本轮 `pnpm audit --prod --json` 返回 0，当前本地锁文件的生产依赖 info/low/moderate/high/critical 全部为 0。仅覆盖本地锁文件与当前公告库，不代表线上镜像全量依赖无漏洞。本轮未重跑完整 Go 单测、govulncheck 或正式 SHA 的 CI。
- 仓库已有 `.github/workflows/security-scan.yml`，包含 push/PR/每周审计，不能把本轮建议表述成“目前完全没有安全扫描”。

## 推断：针对 VPN 批量注册的后续策略

已核实邀请码当前关闭；本地 IP 限额已实现但未发布。轮换公网 IP 不会命中同一个 IP 唯一名额，因此无法由这项限制单独保证一人一号。是否正在遭受该类批量注册，需要实际注册日志、邮箱验证/验证码结果、失败率和注册后消耗来确认；本轮没有这些生产证据。

如果以防薅注册赠额为主要目标，建议先评估启用已有受控一次性邀请码；同时为新账号赠额和高成本能力设置观察期、额度/并发限制，再根据日志决定全站注册速率、域名配额和风险升级验证。具体额度应由正常用户量决定。不能把代理/VPN标记直接等同恶意，也不建议以伪造 MAC 或容易重置的浏览器标识作唯一身份。

可信代理链需要单独核验：本地注册 IP 防护依赖 `server.trusted_proxies` 的真实直连来源和转发头清洗。配置错误可能把所有用户归为代理 IP；盲目信任头则可能接受伪造来源。CDN/WAF 有无代理检测或源站白名单，本轮不知道。

## 推断：其他可排期的加固

本地 `frontend/src/stores/auth.ts` 将 access/refresh token 存入 localStorage。可评估 HttpOnly、Secure 的刷新会话和对应 CSRF/跨站部署设计，以降低未来脚本注入时的会话暴露面；这涉及认证协议和兼容性，本轮没有发现可利用 XSS，也没有修改存储方案。CSP 来源列表可在盘点验证码和支付依赖后收紧。

## 不知道：需要服务器或管理员现场证据

- 生产镜像精确 SHA、注册 IP 限额是否部署、实际 `step_up_enabled`、管理员个人 TOTP 绑定与管理员 API Key 使用情况。
- WAF/CDN 规则、可信代理 CIDR、源站绕过路径、真实跨 IP 注册/攻击情况。
- SSH/防火墙、PostgreSQL/Redis 是否仅内网可达、补丁状态、备份完整性和实际恢复演练。
- 实际注册赠额和自动发放规则；单纯公开设置不足以确认是否存在可薅取的奖励。

没有这些证据，不能宣称网站整体安全、已经被入侵，或上述所有本地缺口都能在当前生产复现。

## 证据位置与复核

`output/security-review-20261022/verification.json` 记录范围、实际测试数量与限制；目录中的 HTTP 头/正文、`public-flags.json`、`pnpm-audit.json`、Go JSONL 和隔离复现源码保留原始取证。该目录属于本地输出，不包含新的生产发布物。

定向回归命令（在 `backend`，使用已核实 Go 1.27.3，`GOTOOLCHAIN=local`、`GOPROXY=off`）：

```text
go test -tags=unit -json ./internal/server/middleware ./internal/server/routes ./internal/handler ./internal/service -run 'Test(AdminAuthJWT|EnforceStepUp|CORS_|SecurityHeaders|GenerateNonce|GetNonceFromContext|EnhanceCSPPolicy|AddToDirective|PublicOrderVerifyRateLimit|VerifyNotificationWithProviders|ResolveOrderPublicByResumeToken|GetPublicOrderByResumeToken|VerifyOrderPublic|PaymentResumeTokenRoundTrip|CreateTokenRejectsMissingSigningKey|ParseTokenRejects|WeChatPaymentResumeTokenRoundTrip|CreateWeChatPaymentResumeTokenRejects|ParseWeChatPaymentResumeTokenRejects|CanonicalizeReturnURL)' -count=1
```

临时复现源 `resolve-route-reproduction.go.txt` 归档后已从 `backend/internal/server/routes` 移除；如需复核，在该测试包恢复为临时 `_test.go` 后执行同名测试并移除。未经后续明确修复与部署，不得把本报告写成加固已上线。
