# 敏感认证 IP 黑名单

2026-10-14已发布：功能07bcef204d09349e0574b51a5b8cfdd6266288a3、固定镜像sha-07bcef2（linux/amd64），同SHA CI（真实PG／Redis、unit／integration、lint、前端497项）／安全／GHCR全部成功，镜像摘要与revision标签交叉验真。备份后在部署目录执行`./update.sh sha-07bcef2`，生产由用户更新。详细证据见AUTH_IP_BLACKLIST_REVIEW及HANDOFF；下方10-13验证为本地历史。

本功能源码在2026-10-13实施。用户先要求保护正常使用，后允许“封禁找管理员解除，或封禁时间可配置”。当前只统计可识别的认证失败，后台可选择临时或永久封禁；成功操作不计数。默认临时15分钟、自动拉黑关闭。发布和生产验证状态以HANDOFF最新记录为准，旧CI结果不代表包含本功能。

## 规则与时间

- 默认同一IP最近120分钟累计10次可识别的认证失败，第10次返回`403`、`reason=IP_BLOCKED`。封禁方式可选`timed`（临时，1–1440分钟，到期解除）或`permanent`（永久，须管理员移除）；默认timed15分钟。次数、统计周期、方式及时长均可在后台调整。
- 成功请求不计数，包括正常登录、注册、发验证码、两步验证的正常中间步骤、Passkey 和 OAuth 操作。计入明确的 `401 INVALID_CREDENTIALS`、`401 PASSKEY_VERIFICATION_FAILED`，以及无业务 reason 的 `400 Invalid request` / `Invalid passkey response`。未知错误保守排除，不把所有 400/401 都当攻击。切换受保护接口不会清零。
- 人机验证／验证码校验失败、TOTP 错误、邮件或表单业务校验、令牌过期／失效、429 已有限流、500/503 服务端故障不计数；原有单接口限流继续保留。成功操作不触发本名单，但仍受原有分钟限流约束。
- 每次统计 `(当前时刻－window_seconds, 当前时刻]`；恰好落在起点的事件排除，横跨整点仍累计，未来分值不计入。
- 请求进入时从 Redis `TIME` 取得 UTC 时间，完成后仅将符合条件的失败事件按进入时间加入 Lua 滑动窗口。处理完成时已过窗口边界的事件不加入。同毫秒并发请求使用唯一成员，不会合并。认证响应短暂缓冲以在第 10 次失败时替换为受控拒绝，缓冲最大 256KiB，较大／流式响应原样透传。
- 数据库和 API 时间保留 UTC；页面按浏览器所在时区转换一次，页面明确标注时区。审计日志时间依旧是请求处理结束的时间，可能略晚于计数时间。
- 120分钟是统计周期，15分钟是默认临时封禁期限。临时期限从PostgreSQL写入封禁的时间计算；到期从活跃列表排除，下次请求原子恢复、递增状态版本并重新计数。永久记录`blocked_until=NULL`，应用重启、Redis清空或等待都不会解除，须管理员移除。关闭开关只停止新增封禁，已有记录保持原期限；切换模式或时长只影响新封禁，不把已有临时记录变为永久，也不自动解除永久记录。
- 规则保存后配置版本递增，之后的新请求使用新计数键；管理员解封后该 IP 状态版本递增，下一次从新窗口计数。旧计数键自行过期，旧在途请求通过数据库版本检查不能重新写回旧封禁。
- 功能开启后统计新请求，历史审计不回灌黑名单。

OAuth 登录启动和回调 GET 只检查已有封禁，不计数。刷新 Token、登出、促销码与邀请码检查不受本名单影响；微信支付 OAuth 保持原行为。正常模型调用和管理接口不接入此防护，已有管理员会话可从被封 IP 进入菜单解封。

原有单接口每分钟限流保留。所有敏感请求先读取 PostgreSQL 状态；数据库故障返回受控 `503 AUTH_IP_GUARD_UNAVAILABLE`，Redis 计数故障返回 `429 AUTH_IP_COUNTER_UNAVAILABLE`（Retry-After 60 秒）。故障不会生成永久黑名单。Redis 清空和应用重启不会解除数据库已有封禁。

## 管理

菜单“IP 黑名单”，路径 `/admin/ip-blacklist`，复用现有管理员鉴权、面板限流、审计和合规确认。

- 自动拉黑开关、触发次数（1–1000整数）、统计周期（1–1440整分钟）、封禁方式。临时显示时长（1–1440整分钟），永久显示管理员解除说明，切回临时保留已有配置时长。默认10次/120分钟、timed15分钟，配置读取失败时禁止保存默认值。
- 活跃名单分页、精确 IPv4/IPv6 搜索；IPv4 映射 IPv6 会归一为 IPv4。展示封禁时间、触发接口、次数、周期、窗口首末请求时间。
- 移除须确认并附当前 `generation`。过期页面返回 `409 IP_BLACKLIST_CONFLICT`，刷新后重新确认，不静默覆盖最新状态。
- 规则修改须附 `version`，冲突同样 409。自动封禁、规则修改、解封走现有审计服务，事件名为 `auth.ip_blacklist.block`、`admin.ip_blacklist.config.update`、`admin.ip_blacklist.unblock`。认证 POST 的后续拒绝记 `auth.ip_blacklist.reject`。

API 前缀 `/api/v1/admin/ip-blacklist`：

| 方法和路径 | 参数 | 返回 |
| --- | --- | --- |
| GET `/config` | 无 | enabled / threshold / window_seconds / block_seconds / block_mode / version |
| PUT `/config` | 完整配置，包含当前version和明确的block_mode | 新配置及递增版本 |
| GET `/entries` | page / page_size（最多100）/ ip（完整地址或空） | items / total / page / page_size / pages |
| POST `/entries/:id/unblock` | `{ "generation": 当前状态版本 }` | 已解封记录及新状态版本 |

迁移250创建两张自有表`custom_auth_ip_blacklist_config`、`custom_auth_ip_blacklist`；251添加`block_seconds=900`、可空`blocked_until`并递增版本，使旧全请求计数失效；252添加`block_mode='timed'`及模式约束，再次递增版本。250/251已应用内容不修改。API必须提交`block_mode=timed/permanent`；非法或遗漏模式400，避免旧客户端静默将永久模式降为临时。旧前端需刷新，前后端同版部署。两种模式下block_seconds保留有效临时时长；永久模式不使用该时长。解封或到期保留状态行并递增版本，阻止旧请求重封；规则改变或解封不需要重启。

## 上线核验

用户确认拓扑为 Cloudflare → Nginx。生产真实配置尚未取证，开启前由部署者完成以下核验：

1. 备份数据库、部署配置和app_data，记录当前和本轮发布镜像摘要。以本轮提交的CI/GHCR结果为准；前后端同版部署，迁移250/251/252须成功应用。
2. 按 [边缘安全说明](EDGE_SECURITY.md) 核对 `server.trusted_proxies`、`security.client_ip_from_forwarded_headers` 兼容开关、转发头白名单和会话绑定 IP 来源。不要信任任意客户端传入的 X-Forwarded-For / CF-Connecting-IP。
3. Nginx 仅从验证过的 Cloudflare 链路接受客户端 IP，覆盖传向应用的转发头；源站仅允许所需代理来源，应用端口不对公网开放。开启 Nginx real_ip 后须核对传入应用的整条转发链，不将公网或 Cloudflare 网段无条件配置为“所有请求可信”。
4. 用两个真实出口各发一次受控敏感请求，对照应用审计 IP、边缘记录和实际出口，确认不是共同 CDN / Nginx 地址，且伪造头无法改变统计身份。测试 IP 必须可由已有管理员会话解封。
5. 保持默认关闭，通过审计观察流量再选择模式并开启。默认10次可识别失败/2小时、临时15分钟；可选择永久／管理员解除。正常成功操作和验证码错误不累计，共享出口仍共同累计可识别失败，不能承诺零误拦或把失败等同确证恶意。永久模式下误拦须人工解除；生产合理阈值和误拦率未知，可后台调整。
6. 检查同 IP 混合接口到阈值拒绝、管理员解封即生效、模型调用不受影响，以及本轮审计记录。生产由用户部署，不自动拉黑截图历史 IP。

回退到不含防护的旧应用镜像会失去本功能的拦截；数据库名单仍保留，但旧代码不会执行检查。不要为回退删除表或重置状态版本。

## 验证与维护

截图回放采用以下北京时间，Redis 输入转换为 UTC，不对截图额外加减 8 小时：

`2026-10-13 19:12:39、19:19:01、19:35:39、19:43:23、19:57:26、20:05:11、20:19:10、20:21:41、20:31:28、20:32:32`。

跨度 1 小时 19 分 53 秒，计数器回放中第 10 次在 20:32:32 达阈值。新规则还要求这些记录属于可识别的认证失败；仅凭截图 HTTP 400 无法核实响应 reason/正文，不能承诺这 10 条在新分类下全部计入。保护功能只观察部署后的新请求，历史截图自身不会触发拉黑。

前端定向验证：`pnpm test:run src/custom/__tests__/auth-ip-blacklist.spec.ts`、`pnpm typecheck`、`pnpm build`。后端：`go test -tags=unit ./...`、`go build -tags embed ./cmd/server`；本机没有宿主 Go，本轮使用隔离 Docker Go 1.27.2 工具链。

2026-10-13 23:31本地检查通过：Docker Go1.27.2的13项专项、真实PG18/Redis8、全量unit、运行镜像embed构建；前端5文件55项、类型、定向ESLint和Vite构建通过。实际运行镜像在8081专用库验证永久封禁跨应用重启和Redis清空、管理员解除、两种模式切换保留旧期限、临时45分钟及受控到期。8080更新到7f423b3890c9、三服务healthy；原数据、设置和卷保持，迁移252应用、规则version=3，开关关闭、名单0条。证据`output/auth-ip-local-20261013`中的mode-*及测试/构建日志。2026-10-14独立发布07bcef204及sha-07bcef2，远端同SHA CI／安全／GHCR／lint通过，证据`output/auth-ip-release-20261014`，生产未代部署；本地候选不作为已发布镜像。

真实存储测试只连接隔离测试服务：`AUTH_IP_TEST_POSTGRES_DSN`（专用数据库，测试会清空自有黑名单状态）、`AUTH_IP_TEST_REDIS_ADDR`（专用 Redis，测试清空 DB 3）。不要指向生产或已有业务数据库。定向命令为 `go test -tags=unit ./internal/service ./internal/repository ./internal/server/middleware ./internal/server/routes -run '^TestAuthIP' -count=1 -v`。

同步上游需保留四个接缝：router 注册新模块并传防护、auth.go 审计后执行可选防护、AppSidebar 管理员菜单、Makefile 关键前端回归中的 auth-ip-blacklist.spec.ts。自有路由及页面位于 custom 叠加层；接缝及测试清单见 CUSTOMIZATIONS。新增迁移编号若上游或其他会话发生占用，发布前重新协调，不修改已经应用过的迁移内容。
