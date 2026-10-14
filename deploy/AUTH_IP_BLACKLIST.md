# 敏感认证 IP 黑名单

2026-10-14 上游0.2.16发布终态：功能fe6323ca643af95cd779e8c4700b1bc2583eb29a已推origin/main，包含upstream/main=9c2cca9e627291bc719bada3d57de23b625a4180与正式v0.2.16；同SHA CI38517873853（unit/integration、专用PG18/Redis8、lint0、前端47文件774项含IP管理16项）、安全38517873829及GHCR38517873878全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-fe6323c（linux/amd64），摘要sha256:3740fc0c7b541500a4ba50f94a59696a4cb3dcf52335dca0578275547bb7660d；registry/revision/source、源码树605c0fbae241ffb0bda14d14798dfe9cd5465a0c及二进制0.2.16-custom.fe6323ca6交叉验真。Ent生成无差异、Go embed、定制专项与4189源码61条/正式镜像25条真实HTTP验收通过，覆盖登录、管理员权限、通配符/网段封禁、永久默认与在线更新拒绝。210个自有源码/迁移保持，主工作区main已快进；其他59个未提交文件已备份恢复，依赖仅合入Vue3.5.43并整理原锁文件，付款/248/249/智力效率等其他会话工作仍未发布。3个测试容器、测试卷/网络及工具镜像已清理，原8080镜像b311d7372950与原卷保留，8080/PG/Redis当前仍停止，未把原8080当0.2.16验收。生产由用户备份后执行`./update.sh sha-fe6323c`，生产真实IP及共享网段误拦率仍待；此前sha-bbee35d及本轮“待发布”条目是阶段记录。证据output/upstream-sync-20261014/verified-release.json、formal-smoke.json、original-work-restored.json与cleanup.json。

2026-10-14 同步上游0.2.16：用户要求更新开源fork，目标upstream/main=9c2cca9e6（正式v0.2.16=c3cb8588ce6a5aef4dc29fa4f71ac91f3c0bdf3d，主线随后仅版本号/赞助信息更新）；相对711bfc744新增235条上游提交。merge-tree演算后在独立工作树合并，14处文本冲突逐项解决，保留Go1.27.2、x/image0.45.0、SheetJS0.20.3、DOMPurify3.4.16及认证IP/包号CI存储；采用Vue3.5.43和上游完整13平台配额测试。210个自有源码/迁移文件哈希保持，原6平台fixture已由上游动态清单取代。前端47文件774项、lint/类型/生产构建；Docker内Ent生成无差异、Go embed、PG18/Redis8专项及4189隔离61条真实HTTP验收通过。原8080应用未替换、其他会话未提交改动保留，生产未部署；本轮功能SHA/CI/安全/GHCR终态及固定镜像以HANDOFF最新记录为准，不能沿用sha-bbee35d的门禁证明0.2.16。证据output/upstream-sync-20261014。

2026-10-14 网段防护发布终态：功能bbee35dbcc30525bae354dc1d270f90b5591fab7已推origin/main；同SHA CI38508986548（unit/integration、专用PG18/Redis8、lint0、前端38文件504项含IP管理16项）、安全38508986630、GHCR38508986565全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-bbee35d（linux/amd64），摘要sha256:0dd2c3067bf73f3b5308ee72b7e27d22014ff71aa618fb7c02404bfda7ac9e22；28文件源码哈希、registry/revision/source及构建日志交叉验真。支持手动通配符/CIDR及自动同网段不同IP共享明确失败计数，默认/24和/64，单IP/前缀/期限可配置。8080于2026-10-14 11:56完成备份更新及三服务healthy验收、原数据/卷保持，254/255生效、自动仍false、permanent/version5、名单0；本机保留原智力效率，正式发布排除其他会话改动。独立测试环境已清理，生产由用户备份后`./update.sh sha-bbee35d`；同网段正常用户也可能受限，生产真实IP和误拦率仍待验证。证据output/auth-ip-network-20261014/verified-release.json、local-verified.json及HANDOFF。 新加入默认永久封禁，只有管理员可调整规则/期限及解封；已有临时名单保留原到期时间。复查旧自动单IP名单解封的网段计数边界，修复前真实PG回归失败、修复后8组网段回归及全量unit通过；最终45条真实HTTP断言通过。最终8080镜像sha256:b311d7372950873015c08c68567d36719ad37457693a7a1915b5ca4839b2a3c1、容器bd9fd0be85449daeb20028f49a3e29495d4337891e1ca9e18f6cf627f2e6d3e6、chunk assets/AuthIPBlacklistView-DwaftuX8.js；255前备份667706字节、目录1283行、SHA256cec4821389c3cf8bec26426184695a0e6922066f8be26d6c3dfd273c621e6ffb与before-auth-ip-permanent-20261014镜像保留。最新日志backend-permanent-final/frontend-permanent-final/build-permanent-release/build-permanent-local.log；旧aade433及下方timed/version4/44条为阶段记录，最终更新使用sha-bbee35d。 2026-10-14 13:08接续会话只读检查确认Docker引擎未运行，当前8080健康未复测；CI/GHCR及远端HEAD已重新核验，不沿用旧发布结果。

2026-10-14 最新默认规则：用户要求加入名单默认永久封禁，只有管理员能调整。新增255迁移将自动规则及数据库默认值设为permanent并递增version；手动新增也默认永久。管理员可以选择临时封禁及期限、调整规则和解除名单；已有临时记录保留到期时间。254已应用内容不再修改，发布及验真见HANDOFF最新状态。

2026-10-14 网段防护本轮：手动支持完整IP、末尾连续通配符和CIDR；自动默认按网段累计明确认证失败并封禁整段，可选单IP及前缀。254迁移保留自动开关和旧名单、递增version，默认IPv4 /24、IPv6 /64。本机8080已备份更新并通过真实HTTP验收；正式本轮SHA/CI/GHCR及生产边界见HANDOFF最新状态，以下旧发布标签不含网段功能。

2026-10-14 手动封禁发布终态：管理员手动新增已随065eee935/sha-065eee9发布，同SHA CI（真实PG/Redis、unit/integration/lint0、前端502项含IP管理14项）、安全/GHCR成功。备份后`./update.sh sha-065eee9`，刷新后台“IP 黑名单 → 手动新增 IP”，填写完整IP、原因并选择永久或临时。253同版部署，重复有效封禁409不覆盖，自动关闭也生效；只限制当前敏感认证。8080已备份更新，生产未代部署；镜像摘要/标签验真及本机验证见HANDOFF和AUTH_IP_BLACKLIST_REVIEW，旧07bcef204版本不含手动新增。

2026-10-14已发布：功能07bcef204d09349e0574b51a5b8cfdd6266288a3、固定镜像sha-07bcef2（linux/amd64），同SHA CI（真实PG／Redis、unit／integration、lint、前端497项）／安全／GHCR全部成功，镜像摘要与revision标签交叉验真。备份后在部署目录执行`./update.sh sha-07bcef2`，生产由用户更新。详细证据见AUTH_IP_BLACKLIST_REVIEW及HANDOFF；下方10-13验证为本地历史。

本功能源码在2026-10-13实施。用户先要求保护正常使用，后允许“封禁找管理员解除，或封禁时间可配置”。当前只统计可识别的认证失败，后台可选择临时或永久封禁；成功操作不计数。默认临时15分钟、自动拉黑关闭。发布和生产验证状态以HANDOFF最新记录为准，旧CI结果不代表包含本功能。

## 规则与时间

- 默认同一网段最近120分钟累计10次可识别的认证失败，第10次返回`403`、`reason=IP_BLOCKED`并永久封禁整个网段；IPv4默认/24（如210.230.206.*），IPv6默认/64。管理员可改为每个IP分开累计，或调整IPv4前缀1–32、IPv6前缀1–128；前缀越小范围越大。默认`permanent`，须管理员移除；管理员可选择`timed`（临时，1–1440分钟，到期解除），临时时长默认值保留15分钟。次数、统计周期、范围、方式及时长均只允许管理员调整。
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

管理员可点“手动新增 IP”，填写完整IPv4/IPv6、末尾连续通配符（210.230.206.*、210.230.*.*、210.*.*.*）或IPv4/IPv6 CIDR，必填原因（1–500字），选择永久或临时封禁（1–1440整分钟），确认后立即生效。通配符按CIDR存储，例如210.230.206.*变为210.230.206.0/24；页面确认显示IPv4实际范围与地址数。拒绝全网/0、全星号、中间星号、无效地址。手动封禁独立于自动开关和统计规则；只限制当前敏感认证接口，模型调用、已认证用户接口和管理员接口保持当前范围。列表标注来源及原因，手动记录不产生失败次数和窗口首末时间，统计项显示“—”。

目标已有有效封禁或被更宽网段完全覆盖时新增返回 `409 IP_BLACKLIST_ALREADY_BLOCKED`，不覆盖原期限；需先解除原规则再新增。已有窄规则时允许新增更宽范围并保留窄规则；解除一条规则后其他有效覆盖仍拦截。自动封禁、手动新增、到期和解封共享目标事务锁、策略共享/独占锁、generation及相关重叠规则摘要，旧在途请求和旧页面不能覆盖新的管理员决定。手动新增事件 `admin.ip_blacklist.manual_block` 写入现有审计，含管理员身份、规范IP/网段、原因、状态版本和到期时间。

- 自动拉黑开关、触发次数（1–1000整数）、统计周期（1–1440整分钟）、封禁方式。临时显示时长（1–1440整分钟），永久显示管理员解除说明，切回临时保留已有配置时长。默认10次/120分钟、永久封禁；只有管理员能调整，配置读取失败时禁止保存默认值。
- 活跃名单分页；完整IP搜索返回覆盖它的规则，通配符/CIDR搜索返回重叠规则。IPv4映射IPv6归一为IPv4，网段按网络起点归一，/32和/128显示单IP。展示封禁时间、触发接口、次数、周期、窗口首末请求时间。
- 移除须确认并附当前 `generation`。过期页面返回 `409 IP_BLACKLIST_CONFLICT`，刷新后重新确认，不静默覆盖最新状态。
- 规则修改须附 `version`，冲突同样 409。自动封禁、规则修改、解封走现有审计服务，事件名为 `auth.ip_blacklist.block`、`admin.ip_blacklist.config.update`、`admin.ip_blacklist.unblock`。认证 POST 的后续拒绝记 `auth.ip_blacklist.reject`。

API 前缀 `/api/v1/admin/ip-blacklist`：

| 方法和路径 | 参数 | 返回 |
| --- | --- | --- |
| GET `/config` | 无 | enabled / threshold / window_seconds / block_seconds / block_mode / aggregation_mode / network_ipv4_prefix / network_ipv6_prefix / version |
| PUT `/config` | 完整配置，包含当前version、明确的block_mode和aggregation_mode（ip/network）及两个前缀 | 新配置及递增版本；旧页面缺少范围/前缀400须刷新 |
| GET `/entries` | page / page_size（最多100）/ ip（完整IP、通配符、CIDR或空） | items / total / page / page_size / pages |
| POST `/entries` | ip / block_mode / block_seconds / reason；永久时长0，临时时长为60–86400整分钟秒数 | 手动封禁记录及generation；重复有效封禁409 |
| POST `/entries/:id/unblock` | `{ "generation": 当前状态版本 }` | 已解封记录及新状态版本 |

迁移250创建两张自有表`custom_auth_ip_blacklist_config`、`custom_auth_ip_blacklist`；251添加`block_seconds=900`、可空`blocked_until`并递增版本，使旧全请求计数失效；252添加`block_mode='timed'`及模式约束，再次递增版本。250/251已应用内容不修改。API必须提交`block_mode=timed/permanent`；非法或遗漏模式400，避免旧客户端静默将永久模式降为临时。旧前端需刷新，前后端同版部署。两种模式下block_seconds保留有效临时时长；永久模式不使用该时长。解封或到期保留状态行并递增版本，阻止旧请求重封；规则改变或解封不需要重启。

## 上线核验

254允许INET规范网段并新增GiST包含索引与网段规则字段，默认network/24/64；保留旧开关和名单、version递增使旧计数失效。旧250–253不改，不从历史审计回灌封禁。网段内不同IP共用滑动窗口，成功不计数；切换范围、管理员变更相关覆盖规则、解封及到期后开启新计数。首次仍默认关闭；已开启的部署升级会保留开关并切换网段默认值，应核验范围后使用。

255将默认自动规则切换为永久、version再次递增；保留开关、阈值、网段范围和所有已有名单的到期时间，不将既有临时记录改为永久。升级前管理员可先关闭自动；升级后由管理员确认规则再开启。旧自动单IP名单解封或到期也会使相应网段窗口失效，避免刚解封就沿用旧网段失败次数再次封禁。

迁移253添加 `source=auto/manual` 和原因，旧记录归类为auto；手动记录 count/window 为0、首末时间为NULL，来源约束阻止伪造自动统计。保留250–252原内容，前后端同版更新。旧 `sha-07bcef2` 不含手动新增界面或接口；应用迁移253后不要直接回退旧程序处理曾手动封禁的IP，旧自动写入未清理manual来源，会受到新数据约束拒绝。

用户确认拓扑为 Cloudflare → Nginx。生产真实配置尚未取证，开启前由部署者完成以下核验：

1. 备份数据库、部署配置和app_data，记录当前和本轮发布镜像摘要。以本轮提交的CI/GHCR结果为准；前后端同版部署，迁移250/251/252须成功应用。
2. 按 [边缘安全说明](EDGE_SECURITY.md) 核对 `server.trusted_proxies`、`security.client_ip_from_forwarded_headers` 兼容开关、转发头白名单和会话绑定 IP 来源。不要信任任意客户端传入的 X-Forwarded-For / CF-Connecting-IP。
3. Nginx 仅从验证过的 Cloudflare 链路接受客户端 IP，覆盖传向应用的转发头；源站仅允许所需代理来源，应用端口不对公网开放。开启 Nginx real_ip 后须核对传入应用的整条转发链，不将公网或 Cloudflare 网段无条件配置为“所有请求可信”。
4. 用两个真实出口各发一次受控敏感请求，对照应用审计 IP、边缘记录和实际出口，确认不是共同 CDN / Nginx 地址，且伪造头无法改变统计身份。测试 IP 必须可由已有管理员会话解封。
5. 保持默认关闭，由管理员通过审计观察流量再选择范围/期限并开启。默认同网段10次可识别失败/2小时、永久封禁；管理员可调整为临时。正常成功操作和验证码错误不累计，但封禁会限制网段内所有地址的敏感认证，包括共享网段的正常用户。不能承诺零误拦或把同网段全部IP等同恶意；合理前缀、阈值和误拦率未知，只允许管理员后台调整。
6. 检查同网段不同IP及混合接口共同到阈值拒绝，相邻网段正常；管理员解封即生效、模型调用不受影响，核对本轮审计。生产由用户部署，不自动拉黑截图历史IP或网段。

回退到不含防护的旧应用镜像会失去本功能的拦截；数据库名单仍保留，但旧代码不会执行检查。不要为回退删除表或重置状态版本。

## 验证与维护

截图回放采用以下北京时间，Redis 输入转换为 UTC，不对截图额外加减 8 小时：

`2026-10-13 19:12:39、19:19:01、19:35:39、19:43:23、19:57:26、20:05:11、20:19:10、20:21:41、20:31:28、20:32:32`。

跨度 1 小时 19 分 53 秒，计数器回放中第 10 次在 20:32:32 达阈值。新规则还要求这些记录属于可识别的认证失败；仅凭截图 HTTP 400 无法核实响应 reason/正文，不能承诺这 10 条在新分类下全部计入。保护功能只观察部署后的新请求，历史截图自身不会触发拉黑。

前端定向验证：`pnpm test:run src/custom/__tests__/auth-ip-blacklist.spec.ts`、`pnpm typecheck`、`pnpm build`。后端：`go test -tags=unit ./...`、`go build -tags embed ./cmd/server`；本机没有宿主 Go，本轮使用隔离 Docker Go 1.27.2 工具链。

2026-10-13 23:31本地检查通过：Docker Go1.27.2的13项专项、真实PG18/Redis8、全量unit、运行镜像embed构建；前端5文件55项、类型、定向ESLint和Vite构建通过。实际运行镜像在8081专用库验证永久封禁跨应用重启和Redis清空、管理员解除、两种模式切换保留旧期限、临时45分钟及受控到期。8080更新到7f423b3890c9、三服务healthy；原数据、设置和卷保持，迁移252应用、规则version=3，开关关闭、名单0条。证据`output/auth-ip-local-20261013`中的mode-*及测试/构建日志。2026-10-14独立发布07bcef204及sha-07bcef2，远端同SHA CI／安全／GHCR／lint通过，证据`output/auth-ip-release-20261014`，生产未代部署；本地候选不作为已发布镜像。

真实存储测试只连接隔离测试服务：`AUTH_IP_TEST_POSTGRES_DSN`（专用数据库，测试会清空自有黑名单状态）、`AUTH_IP_TEST_REDIS_ADDR`（专用Redis，测试清空DB0/3）。不要指向生产或已有业务数据库。定向命令为 `go test -tags=unit ./internal/service ./internal/repository ./internal/server/middleware ./internal/server/routes -run '^TestAuthIP' -count=1 -v`。

同步上游需保留四个接缝：router 注册新模块并传防护、auth.go 审计后执行可选防护、AppSidebar 管理员菜单、Makefile 关键前端回归中的 auth-ip-blacklist.spec.ts。自有路由及页面位于 custom 叠加层；接缝及测试清单见 CUSTOMIZATIONS。新增迁移编号若上游或其他会话发生占用，发布前重新协调，不修改已经应用过的迁移内容。
