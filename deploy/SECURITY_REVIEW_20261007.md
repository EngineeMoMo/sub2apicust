# 2026-10-07 项目安全抽查与版本可见性

## 范围与证据

源码基线 `18c27bb9e`，应用 VERSION 为 `0.2.13`；用户截图显示 `v0.2.13-custom.80c30c825`。截图不是服务器部署取证，本轮未连接生产。

本轮人工抽查管理认证、用户 API Key 所有权、包号成员隔离、订单读取／校验、支付回调及公开版本数据；另执行前端生产依赖在线审计。不是逐行全仓审计、渗透测试或绝对安全证明。

本地原始证据位于 `output/security-review-20261007/`（gitignored）：`audit.json`、`advisories-summary.json`、`frontend-tests.log`、`go-tests.log`、`security-tests.log`、`typecheck.log`、`lint.log`、`build.log`、`go-build.log`。

## 已核实的问题

### P3：普通用户与匿名入口暴露精确部署版本（本轮已修）

- `frontend/src/components/common/VersionBadge.vue` 原有非管理员静态版本分支。
- `backend/internal/handler/setting_handler.go:GetPublicSettings` 原返回 `h.version`；`backend/internal/service/setting_public.go:GetPublicSettingsForInjection` 原将 `s.version` 注入共享 HTML。
- 影响：无须管理员权限即可获得补丁版本和定制提交标识，便利版本指纹识别；未发现该信息本身导致越权执行。
- 修复：组件根节点仅在管理员身份渲染，删除非管理员分支；两个公开出口保留兼容字段 `version: ""`，不返回实际部署版本。管理员仍通过原管理接口读取版本及更新状态。即使浏览器有旧公开版本或管理员缓存，普通用户也不会渲染徽标。
- 这不承诺隐藏全部技术栈指纹或清除部署前已获取的历史版本信息。

### 生产依赖图原有 17 条已知安全告警（用户要求修复后已清零）

修复前 `pnpm audit --prod --json` 返回：**0 critical、0 high、13 moderate、4 low**。命令退出 1 表示有告警，不是扫描失败。用户随后要求修复，现已移除无源码导入的 `@lobehub/icons` 及其间接依赖链，DOMPurify升级并锁定到3.4.16，最低版本与override同步。复扫 `audit-fixed.json`：**所有严重度均为0，退出0**；没有增加或延长安全豁免。下表保留修复前证据。

| 包 | 告警数 | 审计给出的修复版本 | 主要风险 |
| --- | ---: | --- | --- |
| yaml | 中危 1 | >=1.10.3 | 深层嵌套解析栈溢出 |
| uuid | 中危 2 | 11.x >=11.1.1；13.x >=13.0.1 | 指定缓冲区时边界检查缺失 |
| mermaid | 中危 8、低危 1 | >=11.16.1 覆盖本轮命中 | HTML/CSS 注入、原型污染、图表拒绝服务 |
| decode-uri-component | 中危 1 | >=0.5.0 | 畸形输入解码拒绝服务 |
| colord | 中危 1 | >=2.9.4 | 畸形颜色字符串处理过慢 |
| dompurify | 低危 2 | >=3.4.16 | 特定 IN_PLACE／hook 使用下的 XSS |
| katex | 低危 1 | >=0.18.2 | 已存在原型污染时绕过信任限制 |

取证：`audit.json` 的 advisories 包含每条公告 URL、依赖路径、受影响与修复版本；大部分路径经 `@lobehub/icons -> @lobehub/ui` 引入。`frontend/src` 搜索未找到 Mermaid/KaTeX 业务调用，`@lobehub/icons` 命中为提取图标路径的注释；DOMPurify 有实际字符串净化调用。图标仍使用仓库内已有SVG路径。修复后的锁文件不再包含lobehub／Mermaid／KaTeX／uuid／colord／decode-uri-component依赖，DOMPurify为3.4.16。本轮未证实修复前告警在最终产物中可被攻击者触发；移除或升级受影响依赖不依赖此推断。

## 已核实的保护

- `server/routes/admin.go` 在管理路由组挂载 AdminAuth；`server/middleware/admin_auth.go` 的 JWT 路径校验签名、数据库用户启用状态、TokenVersion、会话绑定及管理员角色。UI 隐藏不是服务端授权依据。
- `handler/api_key_handler.go:GetByID` 检查 `key.UserID != subject.UserID`；修改与删除传入认证用户 ID。
- `service/custom_dedicated_accounts.go:View` 使用用户限定查询及成员／有效期准入；本轮执行包号未认证与其他所有者等现有回归。
- `service/payment_order.go:GetOrder`、`payment_order_lifecycle.go:VerifyOrderByOutTradeNo` 检查订单所属用户。
- `server/routes/payment.go` 的匿名旧订单查询有 IP 限流，`handler/payment_handler.go` 仅返回状态和时间等精简信息；该旧接口依设计无需登录，Redis 故障时限流放行，是保留的运营边界，不等同已发现订单写入越权。
- `handler/payment_webhook_handler.go` 先通过 provider 验签再处理通知；本轮执行现有支付宝及 Webhook 定向回归。
- 登录、注册、验证码等入口在 `server/routes/auth.go` 配置 Redis 故障时拒绝请求的限流。

这些是所检查路径的代码与测试证据，不代表支付渠道、所有接口和生产配置均已验证。

## 验证与未知

- 用户要求修复后的最终前端验证：**368文件／2928项全部通过**，包括新增SVG安全2项和版本权限3项；`pnpm run build`（含i18n、vue-tsc、Vite）成功，相关文件ESLint通过。日志`all-tests-fixed.log`、`build-fixed.log`、`lint-fixed.log`。
- 安装恢复：初始沙箱清理旧node_modules耗时异常，终止后自动审批允许沙箱外安装；发现Vue包入口缺失后，以`pnpm install --force --frozen-lockfile --ignore-scripts`重新导入完整依赖，恢复后才记录测试成功。日志`dependency-install-retry.log`与`dependency-repair.log`；锁文件未因修复本地安装再次改变。
- `pnpm audit --prod --json` 的`audit-fixed.json`为0告警；审计依据当前漏洞库，不涵盖未知漏洞或生产配置。
- 新前端产物完成后再次执行便携Go1.27.0的`go build -tags embed`成功、退出0，证据`go-build-fixed.log`；未重新跑全量后端测试。

- 前端真实组件测试及侧栏回归：2 文件、13 项通过，包括旧缓存不展示、管理员展示、展开详情后退出管理员身份隐藏。新增权限回归已接入 Makefile 关键测试。
- `vue-tsc --noEmit`、两处前端文件 ESLint、Vite 生产构建通过。
- Go：公开设置／注入 schema 定向测试 3 包通过；管理员认证、包号 handler、支付宝／Webhook／公开订单定向测试 4 包通过；`go build -tags embed` 成功。
- 本轮使用仓库 output 内已有便携 Go 1.27.0；日志有遥测文件权限警告，但上述命令实际退出 0。没有安装宿主工具链。
- 初始前端命令遇到 Windows 临时目录短路径权限／PATH 问题，新测试夹具先后缺少 i18n 和 Clipboard 隔离，均修正后重跑通过。生产构建仍有既有大 chunk 警告，不影响退出码。
- 未重跑全量 Linux unit／integration、后端 govulncheck、真实 PG/Redis、浏览器真实账号或生产测试。既有 80c30c825 的安全 CI 不能算成本轮未提交修改的 CI。
- 未提交／推送／部署；页面实际生效需后续同版发布及用户更新。真实线上是否存在其他漏洞、依赖告警可利用性，本轮不知道。
