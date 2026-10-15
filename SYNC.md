# 同步上游 & 构建部署 Runbook

2026-10-15发布补充：用户已要求更新测试，分组门槛／缓存／福利本轮提交推送origin/main并核验同SHA CI／安全／GHCR；CI独立PG18与CUSTOM_GROUP_ACCESS_TEST_DSN须随接缝保留。固定镜像与终态以HANDOFF最新记录为准，生产由用户备份后自行更新；下方未授权／未提交状态为历史。

2026-10-15后续补充优先：分组门槛已加入独立事务通知缓存，并追加按指定日期的多时段福利倍率。257仍未发布，最终版本包含活动JSON及缓存通知触发器，不可用本地早期257替代；保留Wire监听退出清理、默认用户倍率缓存标记、批量生图活动快照、复制事务及当前倍率展示接缝。规则／月充值热请求不再逐次回源；退款通知异步、30秒TTL兜底，监听异常回源拒绝失败。最终证据见output/group-access-cache-20261015/verification.json及HANDOFF，未提交部署；下方40／42项为初版历史。

2026-10-15 分组开放规则源码完成、未发布：新增257迁移，两个门槛独立开关和分组到期后的同平台目标自动路由，保留原退款；完整接缝见CUSTOMIZATIONS及deploy/GROUP_ACCESS_RULES.md。40项专项含PG18、Linux58包unit、前端42项与构建／lint／Linux embed已通过，证据output/group-access-20261015/verification.json；本轮尚未提交或跑远端CI，没有本功能镜像标签，不把下方旧标签当作包含新功能。发布后须核验对应SHA CI／安全／GHCR，备份后同版升级并验证退款与到期账单；生产由用户部署。

2026-10-14全工作区发布终态：59项保留改动已汇合上游0.2.16并推送6406511c687e59b2afd37cf24da3db0d6011bea5，含248/249、工坊清理限流、智力效率、依赖/更新脚本；同SHA CI38522649457/安全38522649572/GHCR38522649396全绿，Go unit/integration/专用PG18/Redis8/lint及embed编译通过。正式固定镜像ghcr.io/engineemomo/sub2apicust:sha-6406511（linux/amd64），摘要sha256:9608581fa3c7e8c9b36e368e7815f58376c9c3ee9d2c48f449e0412f6b2244bb，registry/源码标签/构建日志验真。用户备份数据库/配置/app_data后`./update.sh sha-6406511`，前后端同版；生产与本机8080未更新。最终交接文档skip-ci不产生新镜像，历史59项未发布状态由本条覆盖，证据见HANDOFF和output/workspace-release-20261014。

2026-10-14 上游0.2.16发布终态：功能fe6323ca643af95cd779e8c4700b1bc2583eb29a已推origin/main，包含upstream/main=9c2cca9e627291bc719bada3d57de23b625a4180与正式v0.2.16；同SHA CI38517873853（unit/integration、专用PG18/Redis8、lint0、前端47文件774项含IP管理16项）、安全38517873829及GHCR38517873878全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-fe6323c（linux/amd64），摘要sha256:3740fc0c7b541500a4ba50f94a59696a4cb3dcf52335dca0578275547bb7660d；registry/revision/source、源码树605c0fbae241ffb0bda14d14798dfe9cd5465a0c及二进制0.2.16-custom.fe6323ca6交叉验真。Ent生成无差异、Go embed、定制专项与4189源码61条/正式镜像25条真实HTTP验收通过，覆盖登录、管理员权限、通配符/网段封禁、永久默认与在线更新拒绝。210个自有源码/迁移保持，主工作区main已快进；其他59个未提交文件已备份恢复，依赖仅合入Vue3.5.43并整理原锁文件，付款/248/249/智力效率等其他会话工作仍未发布。3个测试容器、测试卷/网络及工具镜像已清理，原8080镜像b311d7372950与原卷保留，8080/PG/Redis当前仍停止，未把原8080当0.2.16验收。生产由用户备份后执行`./update.sh sha-fe6323c`，生产真实IP及共享网段误拦率仍待；此前sha-bbee35d及本轮“待发布”条目是阶段记录。证据output/upstream-sync-20261014/verified-release.json、formal-smoke.json、original-work-restored.json与cleanup.json。

2026-10-14 同步上游0.2.16：用户要求更新开源fork，目标upstream/main=9c2cca9e6（正式v0.2.16=c3cb8588ce6a5aef4dc29fa4f71ac91f3c0bdf3d，主线随后仅版本号/赞助信息更新）；相对711bfc744新增235条上游提交。merge-tree演算后在独立工作树合并，14处文本冲突逐项解决，保留Go1.27.2、x/image0.45.0、SheetJS0.20.3、DOMPurify3.4.16及认证IP/包号CI存储；采用Vue3.5.43和上游完整13平台配额测试。210个自有源码/迁移文件哈希保持，原6平台fixture已由上游动态清单取代。前端47文件774项、lint/类型/生产构建；Docker内Ent生成无差异、Go embed、PG18/Redis8专项及4189隔离61条真实HTTP验收通过。原8080应用未替换、其他会话未提交改动保留，生产未部署；本轮功能SHA/CI/安全/GHCR终态及固定镜像以HANDOFF最新记录为准，不能沿用sha-bbee35d的门禁证明0.2.16。证据output/upstream-sync-20261014。

2026-10-14 网段防护发布终态：功能bbee35dbcc30525bae354dc1d270f90b5591fab7已推origin/main；同SHA CI38508986548（unit/integration、专用PG18/Redis8、lint0、前端38文件504项含IP管理16项）、安全38508986630、GHCR38508986565全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-bbee35d（linux/amd64），摘要sha256:0dd2c3067bf73f3b5308ee72b7e27d22014ff71aa618fb7c02404bfda7ac9e22；28文件源码哈希、registry/revision/source及构建日志交叉验真。支持手动通配符/CIDR及自动同网段不同IP共享明确失败计数，默认/24和/64，单IP/前缀/期限可配置。8080于2026-10-14 11:56完成备份更新及三服务healthy验收、原数据/卷保持，254/255生效、自动仍false、permanent/version5、名单0；本机保留原智力效率，正式发布排除其他会话改动。独立测试环境已清理，生产由用户备份后`./update.sh sha-bbee35d`；同网段正常用户也可能受限，生产真实IP和误拦率仍待验证。证据output/auth-ip-network-20261014/verified-release.json、local-verified.json及HANDOFF。 新加入默认永久封禁，只有管理员可调整规则/期限及解封；已有临时名单保留原到期时间。复查旧自动单IP名单解封的网段计数边界，修复前真实PG回归失败、修复后8组网段回归及全量unit通过；最终45条真实HTTP断言通过。最终8080镜像sha256:b311d7372950873015c08c68567d36719ad37457693a7a1915b5ca4839b2a3c1、容器bd9fd0be85449daeb20028f49a3e29495d4337891e1ca9e18f6cf627f2e6d3e6、chunk assets/AuthIPBlacklistView-DwaftuX8.js；255前备份667706字节、目录1283行、SHA256cec4821389c3cf8bec26426184695a0e6922066f8be26d6c3dfd273c621e6ffb与before-auth-ip-permanent-20261014镜像保留。最新日志backend-permanent-final/frontend-permanent-final/build-permanent-release/build-permanent-local.log；旧aade433及下方timed/version4/44条为阶段记录，最终更新使用sha-bbee35d。 2026-10-14 13:08接续会话只读检查确认Docker引擎未运行，当前8080健康未复测；CI/GHCR及远端HEAD已重新核验，不沿用旧发布结果。

2026-10-14 网段防护本轮：同步须保留新增254迁移、custom_auth_ip_target/network及真实存储回归，自有管理页的自动范围/前缀和手动通配符/CIDR。250–253不改；254保留开关和旧名单、version递增，默认network、IPv4 /24/IPv6 /64，新请求累计失败，不回灌历史。前后端须同版，旧页面PUT缺新参数400应刷新；既有认证、菜单、Makefile/CI接缝不扩大。原卷与备份保留，正式SHA及发布验真见HANDOFF；本机8080已验证，生产由用户部署。


2026-10-14 手动封禁发布终态：065eee935独立推送，同SHA CI/安全/GHCR全部成功，固定sha-065eee9（linux/amd64），摘要8da3604012c2a2c121dd47a7fa22542c785b5c77c60335177498a40058807d38。保留自有manual服务/仓储/测试及253迁移、前端入口/原因/来源和nullable统计；旧250–252不改，只保护认证，重复有效封禁409，自动开关不影响名单。8080已备份更新且三服务健康，原业务计数/规则/PG/Redis/挂载保留；正式镜像排除其他会话源码，生产由用户`./update.sh sha-065eee9`，取证HANDOFF及AUTH_IP_BLACKLIST_REVIEW。

2026-10-14 管理员手动封禁：同步保留自有custom_auth_ip_manual_block服务/仓储/测试、253迁移及管理页POST entries入口。旧250–252不改；manual来源、必填原因、空统计、共享IP锁/generation、重复有效封禁409须保留。只保护现有认证，自动开关不影响手动名单。新SHA验证/发布见HANDOFF，本轮不带其他会话修改，不以旧sha-07bcef2验收手动新增。

> 2026-10-14 IP防护发布终态：07bcef204已推origin/main，同SHA CI（真实PG18／Redis8、unit／integration、lint、前端497项）／安全／GHCR成功；固定镜像sha-07bcef2（linux/amd64），摘要sha256:81bcf46cd8e1008f058c296f09d32af059d0e7fadefa66cf355b522885d5bc13已交叉核验。用户备份后`./update.sh sha-07bcef2`，保留250/251/252和登记接缝，前后端同版。默认关闭、成功不计数，永久／临时可选，只保护认证；生产真实IP和共享出口验证待用户完成。其他会话248/249及智力效率未发布。本条覆盖下方历史未发布状态，证据见HANDOFF和AUTH_IP_BLACKLIST_REVIEW。

> 2026-10-13 23:31 IP防护在8080运行7f423b3890c9，尚未发布生产：成功操作不计数，默认10次可识别失败/120分钟、临时15分钟，开关关闭；用户最新授权增加可选永久封禁，须管理员移除。保留250/251/252迁移及CUSTOMIZATIONS登记的router/auth.go/AppSidebar/Makefile接缝、失败分类／mode版本／期限保留回归。API block_mode必须明确timed/permanent，前后端同版；规则切换不改变已有记录期限。同步不恢复全请求计数。发布核验新SHA CI/安全/镜像及Cloudflare→Nginx真实IP，见deploy/AUTH_IP_BLACKLIST.md和HANDOFF。其他会话248/249未带入本机；下方官网发布镜像不包含IP防护，生产由用户部署。

> 2026-10-13最终发布：官网与登录页动效已推送f8928e20d，同SHA CI（含真实PG的unit/integration）、安全、GHCR全部成功，固定镜像`ghcr.io/engineemomo/sub2apicust:sha-f8928e2`（摘要`sha256:cd8c1667f943c0a6706742dcc3c0e9a1c203e75ba53dd28ec528da9f920e3d6a`）。这一版只有前端改动，没有迁移。用户备份后执行`./update.sh sha-f8928e2`，生产未代部署。本条覆盖下方10-10的更新命令。

> 2026-10-10最终发布：包号职责收敛及安全补修已推送origin/main，功能提交`026d9ab2e78e244858e9a57fafe3297e193434ac`。同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38025165503)、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38025165599)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38025165523)全部成功，含Go unit（独立PostgreSQL）、integration、lint、前端及脚本检查。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-026d9ab`（linux/amd64），摘要`sha256:93d974774096864fcf60a67050586c01fbf7bab27f6f6eb7f349e0e6aa96adeb`。Go1.27.2、x/net0.60.0与x/tools0.50.0已生效，govulncheck零可达漏洞（模块中2项未调用漏洞，不宣称全部依赖零漏洞）；golangci-lint2.14.0通过，旧HTTP/2接口仅有五文件/明确符号限定的弃用兼容规则，迁移另列待办。证据output/security-release-20261010与output/dedicated-release-20261010。用户备份后执行`./update.sh sha-026d9ab`；生产未代部署，真实请求/到期/跨组计费待用户验收。本条覆盖下方历史未发布或失败候选记录，后续纯文档提交不产生新镜像。


> 2026-10-12 官网动效同步注意：
> - vite.config.ts 的 manualChunks 要保留 `/@airwallex/` → `vendor-airwallex`，否则支付 SDK 会回到每页首屏加载；插件数组保留 `mofaAppShell()`。
> - 上游若改语言包入口或公开路由，核对 `custom/shell/app-shell.mjs`。
> - 本机改了 vite.config.ts 后，先用 `npx vue-tsc -b tsconfig.node.json` 重新生成被忽略的 vite.config.js，再运行 `pnpm dev`。

> 2026-10-11 漏洞修复（未发布）同步注意：
> - 迁移 248/249 整体替换 247/245 的同名触发函数，不能修改已发布的 247/245。若上游也改了这两个触发函数，以 248/249 为最终版本再合并。
> - 前端已升级到 vite 6、vitest 4、typescript-eslint 8、eslint-plugin-vue 10，并加入 `pnpm.overrides`；上游仍是 vite 5 和 vitest 2。合并 `frontend/package.json` 时保留本 fork 的版本和 overrides，然后执行 `pnpm install` 重新生成锁文件，不要手工解决锁文件冲突。
> - 上游接缝：`.eslintrc.cjs`（`plugin:vue/essential` 和三条规则）、`.eslintignore`、`UsageView.spec.ts` 的 Blob mock 写法。新增的上游测试如果用 `vi.fn` 加箭头函数模拟构造函数，在 vitest 4 下会失败，需要改成 function。
> - 详见 CUSTOMIZATIONS 顶部新节。

> 2026-10-10未发布包号职责收敛：保留归属/有效期/免扣，独立限额与billing-policy接口已移除。同步时不要恢复历史策略入口，保留免扣凭证与原生网关限制；接缝见CUSTOMIZATIONS和deploy/DEDICATED_ENTITLEMENT_ONLY.md。

2026-10-08发布终态：收款限制与人工开通及两项审查修复已推送575467a5f，同SHA CI（含unit／integration／lint）、安全、GHCR全部成功；固定镜像sha-575467a。更新及证据见HANDOFF和deploy/UPDATE_GUIDE，下方本轮待发布记录为历史；生产由用户备份后更新。

2026-10-08审查补修接缝：保留Alipay.CancelPayment的SDK响应成功码校验及交易不存在错误；不得恢复忽略业务失败即释放额度。保留人工续费到公开目录的入口，运行custom_alipay_collection_test.go及plan-stock.spec.ts；本轮仍未发布。

2026-10-08最新源码：收款限制与人工开通已完成但未发布，新增247迁移及默认关闭的50／1000人民币共享限额、可配置快捷充值／自定义开关、套餐contact_admin与联系说明，人工发放继续扣库存。迁移、接缝、证据与待办以HANDOFF最新状态、CUSTOMIZATIONS及deploy/PAYMENT_COLLECTION_LIMITS.md为准；旧sha-70e66cf不含此功能，不能沿用旧CI结论。发布前须核对历史付款链接及线下到账，前后端同版，生产由用户部署。

2026-10-08已发布库存与帮助定制：最终70e66cf73的CI（unit／integration／lint／前端与脚本）、安全、GHCR全绿，固定镜像sha-70e66cf。后续同步保留下列245／246库存事务与帮助接缝，真实PG测试不得跳过冒充通过；生产更新须先备份并核对历史人工归属。详见HANDOFF最新终态，下方未发布说明为历史。

2026-10-08帮助页同步注意：保留AppSidebar的FaqIcon／GuideIcon两种语义图标、router教程新旧锚点白名单；自有层GuideToolSteps和六节教程、FAQ分类筛选／16题与预览问号图标。合并后运行guide／guest／guest-preview／family-home测试并检查窄屏浅深色；官方文档会变，技术步骤更新须重新查证。当前未发布。

2026-10-08同轮库存补充：保留246迁移／归属账本、SubscriptionService内权益与扣减同事务及过期幂等重查、付款私有来源不重扣、管理员plan_id选择、我的订阅库存读取／售罄保护。新增TestCustomSubscriptionStockPostgres，真实PG覆盖人工与购买混合争抢、回滚、历史回填与重跑；前端plan-stock与管理员交互回归须保留。历史归属不确定部分上线前核对，详见库存说明。最终216前端／421后端定向及构建通过，未发布。

2026-10-08新增库存接缝待发布：同步保留245迁移的原子条件更新与订单触发器、Ent两字段、套餐CRUD默认／遗漏语义、晚付与失败订阅保护、四类目录／结账库存投影及前端售罄入口。运行TestCustomPlanStock（真实PG环境变量）与plan-stock.spec.ts；不要只用Ent Schema.Create替代迁移。详见CUSTOMIZATIONS和deploy/SUBSCRIPTION_PLAN_STOCK.md，尚未推送／部署。

2026-10-07最新发布终态：功能80c30c8250d7de59da55b964251f1c437bcf6ddd已推origin/main，同SHA CI（unit／integration）、安全扫描与GHCR均成功，固定镜像sha-80c30c8。包含周／月按开通计时、244迁移、支付宝新订单官方直码及SheetJS0.20.3补修；更新前备份，更新后必须新建支付订单真机验收。生产由用户部署，证据与步骤见HANDOFF四／五和deploy/UPDATE_GUIDE；本条覆盖下方历史未发布状态。

2026-10-07发布终态：用户授权后已推送功能4bc5abbb5，同SHA CI（unit／integration）、安全扫描和GHCR全部成功，固定镜像sha-4bc5abb；发布与更新步骤以HANDOFF第四节和deploy/UPDATE_GUIDE最新节为准，覆盖下方历史未发布状态。生产由用户部署。

2026-10-06本地同步v0.2.13：稳定tag3040209f2＋随后0.2.13 VERSION元数据；合并保留WAP扫码、游客目录与上游verify限流。新增余额在途预占必须跳过已认证包号凭证，否则零余额包号并发被误拒；接缝与回归见CUSTOMIZATIONS。新老241迁移按完整文件名并存，独立PG新库／升级／幂等通过，禁止重命名既有已应用迁移。本轮尚未推送或部署，完整integration已按用户授权补跑通过（含安全审计专用PG／Redis补测），6项既有／外部依赖跳过及新SHA CI仍待，不沿用历史CI结果。

2026-10-06支付宝安全补强同步要求：保留CUSTOMIZATIONS登记的官方通知严格字段／配置公钥／总额、查询订单绑定、精确分金额及新扫码HTTPS／可信站点规则；不得恢复金额回退或用Referer扩大可信Host。路由Guard在限流前禁止缓存，302正文为空。运行TestCustomAlipay及支付履约回归，真实仓储并发测试也要保留；发布与真实扫码边界见deploy/ALIPAY_SECURITY_REVIEW.md。生产由用户部署，不沿用历史CI结论。

2026-10-06我的订阅提示源码完成未发布：同步保留SubscriptionsView的custom卡片与独立列表／重置时间读取、原progress接口真实嵌套结构和服务端resets_at，不恢复本地固定小时推算；15秒共享显示时钟、标签返回更新与卸载清理、过期实际日期及最后周期结束提示保持。样式仅theme.css，原平台语义色保持；Makefile增加subscription-timing.spec.ts，本轮27文件399项、类型／lint／Vite及深浅／手机浏览器模拟夹具通过。未改后端或实际部署，真实账号及本次代码发布门禁待用户要求发布后核验，不能借旧CI／镜像说明线上已改；详见CUSTOMIZATIONS、custom/UPGRADE和HANDOFF。

2026-10-06电脑端支付宝 WAP 扫码已完成源码，未发布。同步时保留默认关闭的桌面专用开关、官方实例 WAP 分支、订单分钟期限与 qrcode snapshot、32字节令牌免登录短入口及官方目标校验；手机两项旧配置、通知幂等、个人订单认证不得覆盖。按CUSTOMIZATIONS最新节逐处检查，运行新 `TestCustomAlipayDesktopWap` 与Makefile前端关键回归。前后端同版构建，无新迁移；关闭设置只影响新订单，旧二维码仍依订单期限校验。Windows全量两项已知失败不算CI成功，发布时核验本次实际代码SHA的unit／integration、安全和镜像；真实App／小额实付待用户按 [验收说明](deploy/ALIPAY_DESKTOP_WAP.md) 检查。已有更新脚本专项保留，生产由用户部署。

2026-10-02配方图片打开修复已发布：功能4bf1436fb同SHA CI／安全／GHCR全绿，固定镜像sha-4bf1436，摘要见HANDOFF，含上一版包号免扣。同步保留自有site的dialog原图按钮、Base64下载入口与材料／连接失效清理；不要恢复data地址的新窗口链接。运行配方四组Node测试和build-products.test.mjs，并检查独立及同源内嵌页的打开／Esc／窄屏。此轮无新上游接缝或CSP变更，生产由用户备份更新。

2026-10-01包号免扣版已发布：功能acd73d716的CI（unit／integration、真实PG、lint、前端372项）、安全与GHCR全绿，使用固定镜像sha-acd73d7，详见[更新手册](deploy/UPDATE_GUIDE.md)。同步上游须保留迁移243、私有准入凭证、独立参考计量、共享限制与WS帧保护，并实际执行带DEDICATED_TEST_POSTGRES_DSN的PG回归；普通计费与已有多人多组隔离一并验。生产由用户备份后更新，更新脚本专项仍未提交，不用纯文档提交SHA猜镜像。

本 fork 的目标：**保留定制的同时能同步上游 sub2api 更新。**
定制策略为「配置 + 新增功能」，尽量不改上游核心逻辑。

- 上游（只读）：`upstream` → https://github.com/Wei-Shaw/sub2api.git
- 你的私有仓：`origin` → https://github.com/EngineeMoMo/sub2apicust.git
- 定制清单：见 [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md)

---

2026-10-07新增版本权限接缝：同步后保留VersionBadge仅管理员渲染，以及setting_handler.go／setting_public.go两个公开version空串出口。运行新增custom_version_visibility_test.go及VersionBadge.visibility.spec.ts（已接Makefile）；管理员版本仍走受保护接口。后续已移除未使用@lobehub/icons链并升级DOMPurify3.4.16，保留锁文件与svg-sanitization.spec.ts关键回归；生产依赖审计全0、全量2928项通过。源码未发布，见HANDOFF及安全报告。

2026-10-07本轮已正式发布：07f0ee69e已推送，同SHA CI（unit／integration）、安全、GHCR全成功，固定sha-07f0ee6，详情见HANDOFF最新终态。上段源码未发布为历史状态；后续文档skip-ci提交不替代功能镜像。

## 0. 一次性初始化（已完成的部分打勾）

- [x] `git clone` 上游为基线
- [x] `git remote rename origin upstream`（上游只读）
- [x] `git remote set-url --push upstream DISABLED_NO_PUSH_TO_UPSTREAM`（防误推上游）
- [x] `git config rerere.enabled true`（记住冲突解法）
- [x] 建私有仓并设为 `origin`（已完成：EngineeMoMo/sub2apicust，已 `push -u origin main`）：
  ```bash
  # 方式一：已有私有空仓
  git remote add origin <你的私有仓地址>
  git push -u origin main

  # 方式二：用 gh 一条命令建私有仓并推送（需已登录 gh）
  gh repo create <名字> --private --source=. --remote=origin --push
  ```
- [ ] 在 GitHub 私有仓 Settings → Actions 确认 Actions 已启用（GHCR 构建用内置 `GITHUB_TOKEN`，无需额外密钥）

---

## 1. 日常：同步上游更新

```bash
git fetch upstream
git checkout -b sync-test main          # 先在测试分支试合，别直接动 main
git merge upstream/main
```

- 有冲突：解冲突（rerere 会自动套用你以前的解法），然后 `git commit`。
- 解完后**打开 [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md)，逐条核对定制是否还在**。
- 跑自检（见 CUSTOMIZATIONS.md 末尾清单）：build + unit test 通过。

验证通过后合回 main 并推送：

```bash
git checkout main
git merge sync-test
git branch -d sync-test
git push origin main                     # push 会触发 CI 构建镜像（见下）
```

## 2. 构建镜像（CI 自动 / 手动）

`.github/workflows/custom-image.yml` 会在 **push 到 main** 或 **手动触发** 时，
用根 `Dockerfile` 构建并推到 `ghcr.io/<你的用户名或组织>/<仓库名>`。

标签规则：
- `latest` —— main 最新
- `sha-xxxxxxx` —— 每个 commit（可精确回滚）
- 打 tag `v*` 时 —— 该版本号

手动触发：私有仓 → Actions → “Build custom image (GHCR)” → Run workflow。

> GHCR 私有镜像默认需要登录才能 pull。首次在服务器上：
> ```bash
> echo <GHCR_PAT> | docker login ghcr.io -u <你的用户名> --password-stdin
> ```
> PAT 需要 `read:packages` 权限。

## 3. 部署 / 升级（服务器上）

服务器的部署目录里放 `deploy/docker-compose.yml`（或 `.local.yml`）+ 你的
`.env` + `docker-compose.override.yml`（override 把 image 指向你的 GHCR 镜像，
见 `deploy/docker-compose.override.yml.example`）。

```bash
cd <部署目录>
docker compose pull                      # 拉你 CI 刚推的定制镜像
docker compose up -d
docker compose logs -f sub2api           # 冒烟
```

## 4. 🔴 红线（务必遵守）

1. **App 内「在线更新」已被关闭**（`DISABLE_ONLINE_UPDATE=true`，在 `docker-compose.override.yml` 里）。
   它原本拉的是**上游**预编译产物，会覆盖你的定制镜像。关闭后：前端不再出现「立即更新」按钮，
   后端也拒绝应用/回滚。定制版一律走「本 runbook 的 merge + CI 构建 + compose pull」升级。
   如需把更新检查指向自己的仓库，设 `UPDATE_GITHUB_REPO=your-org/your-repo`（且需你自己发 Release）。
2. **密钥不进本 fork。** `.env` / `config.yaml` 里的 `JWT_SECRET`、`POSTGRES_PASSWORD`、
   `TOTP_ENCRYPTION_KEY` 等已被 gitignore；单独放私有 ops 仓或加密备份。
3. **Go 版本硬钉 1.27.0**（`backend/go.mod` + 3 个 Dockerfile + CI 断言）。上游升级时随 merge
   带过来即可；**你自己若改 Dockerfile，记得同步这几处**，否则 build 时才炸。
4. **改 ent/schema 要 `go generate ./ent` 并提交生成码；加前端依赖要提交 `pnpm-lock.yaml`。**
