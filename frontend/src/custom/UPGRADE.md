<!-- 2026-10-07：周/月开通即计时、支付宝官方直码接缝优先于下方历史短码说明。 -->

2026-10-15后续补充优先：开放规则弹窗增加可独立关闭的一次性多时段福利倍率，最多32段北京时间起止日期；「三段福利示例」预填0.1至18:00、0.15至20:30、0.1至23:00及23:00到期，仍须选择兜底。时段外恢复基础倍率，用户专属／高峰／独立图片定价保留原规则；源API可选分组和模型广场显示当时基础倍率，不能只改展示而漏掉请求／批量生图快照及用户倍率缓存接缝。资格使用独立事件缓存，原认证与计费缓存不变；前后端须同版发布，最终证据output/group-access-cache-20261015/verification.json，未提交部署。

2026-10-15 分组开放规则源码新增，尚未发布：保留 GroupsView 操作栏的 `GroupAccessPolicyButton` 入口与 `custom/groupAccess/`；两门槛独立开关、同时开启 AND，金额为 USD 账户额度；北京时间到期时间须带同平台目标。前后端同版，257 迁移及绑定／认证／模型广场／WebSocket／复制事务接缝不可丢；用户原绑定不批改，实际请求与计费采用有效目标，退款后每次新请求重新查资格。说明见 [GROUP_ACCESS_RULES](../../../deploy/GROUP_ACCESS_RULES.md)，验证终态以 HANDOFF 为准。

2026-10-14 Claude工具附加费新增：Anthropic/composite分组设置基础搜索单价，按实际成功调用叠加到Token费；用户与管理员共用费用明细，XLSX/CSV导出保留工具快照。新增256迁移及后端原生JSON/SSE/协议转换接缝，前后端须同版。普通客户端工具不收费，旧账单不补扣；最终功能d3a2c346352e7eca5aacc197785fd366f90f83b6已推送origin/main，同SHA CI/安全/GHCR全部成功，固定ghcr.io/engineemomo/sub2apicust:sha-d3a2c34及摘要/revision/source已验真；生产及本机8080未部署，交接skip-ci不生成新镜像；配置/覆盖边界见[CLAUDE_TOOL_BILLING](../../../deploy/CLAUDE_TOOL_BILLING.md)，全部接缝与验证终态见CUSTOMIZATIONS/HANDOFF。

2026-10-14全工作区发布终态：59项保留改动已推送6406511c6，智力效率雷达Bench、收款待确认/人工释放、工坊清理与配额及工具链接缝一并发布，保留上游0.2.16和认证IP防护。同SHA CI/安全/GHCR全部成功，正式固定sha-6406511及registry摘要/源码标签验真；本地409文件3279项、类型/构建，Node20 CI关键49文件790项通过，源码lint0错误26警告。前后端同版，248/249须备份后升级；生产及本机8080未部署此镜像，视觉/真实业务待验收。交接skip-ci不生成新镜像，证据与待办见HANDOFF最新状态，覆盖下方59项未发布记录。

2026-10-14 上游0.2.16发布终态：功能fe6323ca643af95cd779e8c4700b1bc2583eb29a已推origin/main，包含upstream/main=9c2cca9e627291bc719bada3d57de23b625a4180与正式v0.2.16；同SHA CI38517873853（unit/integration、专用PG18/Redis8、lint0、前端47文件774项含IP管理16项）、安全38517873829及GHCR38517873878全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-fe6323c（linux/amd64），摘要sha256:3740fc0c7b541500a4ba50f94a59696a4cb3dcf52335dca0578275547bb7660d；registry/revision/source、源码树605c0fbae241ffb0bda14d14798dfe9cd5465a0c及二进制0.2.16-custom.fe6323ca6交叉验真。Ent生成无差异、Go embed、定制专项与4189源码61条/正式镜像25条真实HTTP验收通过，覆盖登录、管理员权限、通配符/网段封禁、永久默认与在线更新拒绝。210个自有源码/迁移保持，主工作区main已快进；其他59个未提交文件已备份恢复，依赖仅合入Vue3.5.43并整理原锁文件，付款/248/249/智力效率等其他会话工作仍未发布。3个测试容器、测试卷/网络及工具镜像已清理，原8080镜像b311d7372950与原卷保留，8080/PG/Redis当前仍停止，未把原8080当0.2.16验收。生产由用户备份后执行`./update.sh sha-fe6323c`，生产真实IP及共享网段误拦率仍待；此前sha-bbee35d及本轮“待发布”条目是阶段记录。证据output/upstream-sync-20261014/verified-release.json、formal-smoke.json、original-work-restored.json与cleanup.json。

2026-10-14 同步上游0.2.16：用户要求更新开源fork，目标upstream/main=9c2cca9e6（正式v0.2.16=c3cb8588ce6a5aef4dc29fa4f71ac91f3c0bdf3d，主线随后仅版本号/赞助信息更新）；相对711bfc744新增235条上游提交。merge-tree演算后在独立工作树合并，14处文本冲突逐项解决，保留Go1.27.2、x/image0.45.0、SheetJS0.20.3、DOMPurify3.4.16及认证IP/包号CI存储；采用Vue3.5.43和上游完整13平台配额测试。210个自有源码/迁移文件哈希保持，原6平台fixture已由上游动态清单取代。前端47文件774项、lint/类型/生产构建；Docker内Ent生成无差异、Go embed、PG18/Redis8专项及4189隔离61条真实HTTP验收通过。原8080应用未替换、其他会话未提交改动保留，生产未部署；本轮功能SHA/CI/安全/GHCR终态及固定镜像以HANDOFF最新记录为准，不能沿用sha-bbee35d的门禁证明0.2.16。证据output/upstream-sync-20261014。

2026-10-14 网段防护发布终态：功能bbee35dbcc30525bae354dc1d270f90b5591fab7已推origin/main；同SHA CI38508986548（unit/integration、专用PG18/Redis8、lint0、前端38文件504项含IP管理16项）、安全38508986630、GHCR38508986565全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-bbee35d（linux/amd64），摘要sha256:0dd2c3067bf73f3b5308ee72b7e27d22014ff71aa618fb7c02404bfda7ac9e22；28文件源码哈希、registry/revision/source及构建日志交叉验真。支持手动通配符/CIDR及自动同网段不同IP共享明确失败计数，默认/24和/64，单IP/前缀/期限可配置。8080于2026-10-14 11:56完成备份更新及三服务healthy验收、原数据/卷保持，254/255生效、自动仍false、permanent/version5、名单0；本机保留原智力效率，正式发布排除其他会话改动。独立测试环境已清理，生产由用户备份后`./update.sh sha-bbee35d`；同网段正常用户也可能受限，生产真实IP和误拦率仍待验证。证据output/auth-ip-network-20261014/verified-release.json、local-verified.json及HANDOFF。 新加入默认永久封禁，只有管理员可调整规则/期限及解封；已有临时名单保留原到期时间。复查旧自动单IP名单解封的网段计数边界，修复前真实PG回归失败、修复后8组网段回归及全量unit通过；最终45条真实HTTP断言通过。最终8080镜像sha256:b311d7372950873015c08c68567d36719ad37457693a7a1915b5ca4839b2a3c1、容器bd9fd0be85449daeb20028f49a3e29495d4337891e1ca9e18f6cf627f2e6d3e6、chunk assets/AuthIPBlacklistView-DwaftuX8.js；255前备份667706字节、目录1283行、SHA256cec4821389c3cf8bec26426184695a0e6922066f8be26d6c3dfd273c621e6ffb与before-auth-ip-permanent-20261014镜像保留。最新日志backend-permanent-final/frontend-permanent-final/build-permanent-release/build-permanent-local.log；旧aade433及下方timed/version4/44条为阶段记录，最终更新使用sha-bbee35d。 2026-10-14 13:08接续会话只读检查确认Docker引擎未运行，当前8080健康未复测；CI/GHCR及远端HEAD已重新核验，不沿用旧发布结果。

2026-10-14 网段防护本轮：自有AuthIPBlacklistView/authIpBlacklist及spec新增自动范围ip/network、IPv4/IPv6前缀、通配符/CIDR输入与覆盖范围确认。接口PUT必须提交aggregation_mode及两个前缀，旧缓存页面需刷新；永久/临时及原认证范围保留，成功不计数。同段不同IP共同累计失败并封整段，页面明确共享网段正常用户也受限；解除单条不解除其他覆盖规则。16项管理页回归、类型/lint/Vite及8080新chunk BznUWgMC已验证，正式发布状态见HANDOFF，不修改theme或混入其他会话依赖升级。


2026-10-14 手动封禁发布终态：065eee935/sha-065eee9已发布，同SHA前端类型/38文件502项含IP管理14项、后端CI/安全/GHCR通过。AuthIPBlacklistView的手动新增、永久/临时、原因、确认、manual来源与空统计、重复409保留，253同版部署，旧250–252不改。8080备份更新、真实管理员登录及chunk AuthIPBlacklistView-CoRkhMZ9.js已HTTP核验；原配置关闭自动拉黑，手动名单独立生效，只限制敏感认证。未发布其他会话智力效率/依赖升级；生产由用户更新，证据及本机区别见HANDOFF。

2026-10-14 管理员手动封禁：自有AuthIPBlacklistView/TS API/spec增加“手动新增 IP”、永久/临时、必填原因、确认及manual来源/原因展示；手动统计显示—，nullable首末时间不转换为1970。POST entries与253迁移同版部署，重复有效封禁409刷新而不覆盖，自动关闭也生效；仅保护现有认证范围。验证/发布见HANDOFF，旧sha-07bcef2不含入口。

2026-10-14 IP防护已随07bcef204发布，固定镜像sha-07bcef2；同SHA前端类型／38文件497项含管理9项、后端CI／安全／GHCR通过。前后端及250/251/252同版部署，旧缓存页面PUT缺block_mode返回400需刷新；默认关闭，永久与临时方式可选，成功操作不计数。此发布未包含其他会话智力效率改接或依赖升级，生产由用户更新；具体摘要和证据见HANDOFF。

2026-10-13 23:31 IP防护最新修订：只累计可识别失败，成功／验证码／业务校验／限流及系统故障排除；默认10次/120分钟、临时15分钟、开关关闭。用户授权可选永久封禁，须管理员移除。保留custom/authIpBlacklist.ts、custom/views/AuthIPBlacklistView.vue、custom/routes.ts管理员路由及AppSidebar的[CUSTOM]入口。block_mode=timed/permanent下拉选择；timed显示1–1440分钟时长，permanent隐藏时长并保留已存配置供切回。API window_seconds/block_seconds为秒，PUT必须明确block_mode；列表blocked_until空值显示永久封禁／管理员移除，规则修改不改已有期限。加载失败禁止保存默认值，version/generation冲突409刷新重试，时间标注浏览器时区。回归auth-ip-blacklist.spec.ts（9项），与后端250/251/252和Begin/Failed guard同版发布，旧客户端缺mode返回400需刷新。8080运行7f423b3890c9，55项定向／类型／lint／构建通过，生产未发布；见HANDOFF及deploy/AUTH_IP_BLACKLIST_REVIEW.md，不恢复全请求计数。

2026-10-12 官网动效，同步时保留：
- `PublicLayout`：同一个 `<nav>` 加菜单按钮，吸顶、向下滚动收起，以及空闲预取。
- `BrandPanel` 的装饰轨道；`FamilyHomeView` 的滑动指示条和 `data-product`；`PublicPlansView` 的骨架与 `role="status"`。
- theme.css 末尾的动效层：所有动画都在 `no-preference` 内，不用 `!important`。
- vite.config.ts 里的 `mofaAppShell()` 和 `vendor-airwallex` 规则。
- 合并后运行 motion、public-nav、app-shell、stripeLazyLoading、family-home、guest、brand、upgrade-contract 这几个 spec，并在浏览器里检查手机菜单、教程锚点和深色模式。

2026-10-08发布终态：收款限制与人工开通及两项审查修复已推送575467a5f，同SHA CI（含unit／integration／lint）、安全、GHCR全部成功；固定镜像sha-575467a。更新及证据见HANDOFF和deploy/UPDATE_GUIDE，下方本轮待发布记录为历史；生产由用户备份后更新。

2026-10-08审查补修：SubscriptionCard人工续费转公开`/plans`，不能重新指向受payment_enabled限制的`/purchase`；在线续费仍保持原分组路由。plan-stock回归覆盖支付关闭时展示联系说明，未发布。

2026-10-08最新源码：收款限制与人工开通已完成但未发布，新增247迁移及默认关闭的50／1000人民币共享限额、可配置快捷充值／自定义开关、套餐contact_admin与联系说明，人工发放继续扣库存。迁移、接缝、证据与待办以HANDOFF最新状态、CUSTOMIZATIONS及deploy/PAYMENT_COLLECTION_LIMITS.md为准；旧sha-70e66cf不含此功能，不能沿用旧CI结论。发布前须核对历史付款链接及线下到账，前后端同版，生产由用户部署。

2026-10-08最新终态：本页库存／帮助／生图开通说明已随70e66cf73发布，同SHA CI／安全／GHCR成功，镜像sha-70e66cf。下方未发布为历史记录，后续同步仍须保留各接缝，生产由用户验收。

2026-10-08帮助补充：AppSidebar分别使用questionCircle／book，预览图标类型同步；PublicGuideView六节分类／GuideToolSteps与新旧锚点、PublicFaqView分类搜索、16题和theme.css响应式样式须保留。回归见guide／guest／guest-preview；来源核对及验证范围见GUEST_PORTAL。

2026-10-08库存补充：保留自有SubscriptionStockPicker、SubscriptionCard库存及加载／失败／售罄禁用；管理员SubscriptionsView显式选择plan_id，用户SubscriptionsView独立加载库存，types/index请求字段。自有plan-stock与subscription-timing，以及两组管理员分配交互测试已覆盖。样式仍只在theme.css。后端245／246与非购买事务扣减必须同版发布。

2026-10-08套餐库存：保留PlanEditDialog限额默认-1／0停售／遗漏保留、AdminPaymentPlansView占用及剩余、SubscriptionPlanCard与PaymentView深链接／续费售罄保护、PublicPlansView游客限制、两类API类型和中英文文案；样式仅theme.css的mofa-stock-*。plan-stock.spec.ts进入Makefile关键测试。后端245正式迁移触发器是并发权威；用户界面不可替代库存裁决。详见仓库deploy/SUBSCRIPTION_PLAN_STOCK.md，源码未发布。
# 雾钛青：上游同步核对清单

## 2026-10-06 同步 v0.2.13

- 订阅custom卡片／主题与SynaRoute接线保留；设置新增TypeSafe及充值赠金／折扣字段，与桌面支付宝开关共存。
- settings.authSourceDefaults.spec.ts包含六个平台，UsageView.spec.ts保留用户CSV Billing Type列；两项纳入Makefile关键回归。全量365文件2921项通过，不能只凭旧关键子集说明同步成功。

## 2026-10-06 我的订阅时间提示

- 审查修复须保留：列表独立结束 loading，progress 慢响应不阻塞卡片；重试／卸载使旧请求结果失效。保留慢响应补齐、延迟失败与旧请求成功／失败不覆盖新结果的四项回归。

- `views/user/SubscriptionsView.vue` 的custom卡片装配、原列表API与AppLayout保持；共享时钟只更新显示，时间读取失败不能隐藏订阅／用量，列表失败不能误报暂无订阅。暂停／撤销不露续费，有效／过期续费仍携原分组到purchase。
- 保留 `custom/subscriptions/timing.ts` 读取 `/subscriptions/progress` 的 `{subscription, progress}`＋`progress.{daily,weekly,monthly}.resets_at` 真实契约。不要沿旧前端 `SubscriptionProgress` 类型猜响应，或恢复window_start加24／168／720小时的本地算法；后端日历日与周／月旧锚点以服务端返回为准。
- 倒计时15秒刷新、visibilitychange及时更新及卸载清理保持；窗口已到重置但无新读数只显示使用后更新，不能本地清用量／伪造下一周期。订阅先到期／日卡提示额度结束，已过期仍显示实际到期日期、不显示未来重置时间。
- 新样式只位于theme.css的mofa-subscription作用域，雾钛青tokens与原平台语义徽标保留。回归 `custom/__tests__/subscription-timing.spec.ts`（21项），Makefile应包括该文件（本轮27文件399项）；类型／相关lint／Vite与1440／900／390深浅浏览器夹具已通过，真实账户与生产仍待发布后验收。

## 2026-10-06 电脑端支付宝 WAP 扫码

- 保留 `AlipayDesktopWapSetting.vue` 与 SettingsView 的 import／支付区装配、默认 false、保存字段；中英文 label／帮助及 Toggle 的 aria-labelledby／aria-describedby 共同保留。不改 theme.css。
- 管理员读写 `payment_alipay_desktop_wap_qrcode`、支付配置 `alipay_desktop_wap_qrcode` 与后端存储 `ALIPAY_DESKTOP_WAP_QRCODE` 对齐；后端遗漏字段保留旧值，前端不要借此改动两个手机选项。
- 结账和恢复继续用原 `qr_code`＋`payment_mode=qrcode` 路径；二维码内容为带随机令牌的本站短入口。订单保存原 WAP URL 与实际模式，前端不得因当前开关关闭而把旧电脑订单改成跳转，也不得将手机标记为新桌面流程。
- 保留后端 public/alipay/wap/:id 的免登录令牌授权、no-store／no-referrer、分钟期限与官方网关校验；个人订单接口仍认证。完整后端接缝见根CUSTOMIZATIONS，无新迁移。
- 回归 `src/custom/__tests__/alipay-desktop-wap.spec.ts`、SettingsView、原 PaymentView／PaymentResult、paymentFlow／paymentStatus；Makefile关键列表应含新项。已通过26文件378项、类型／lint／生产构建，实际支付宝App扫码和实付尚未验证，不能用jsdom恢复或302夹具代替。启用与真机验收见根deploy/ALIPAY_DESKTOP_WAP.md。

- 配方生成图片的“打开图片”保留页内原生dialog，避免Base64 data新窗口空跳转；原图下载仅PNG／JPEG／WebP Base64，远程图提示长按／右键保存。保留Esc／关闭回焦点、Tab循环、换材料／配方／连接／退出清理及图片加载错误，独立与主站外置CSP共用。不改母站theme.css或放宽CSP；回归配方ui.test.cjs、model.test.cjs和build-products.test.mjs。

- 2026-10-01 最新要求：默认浅色且保留用户已保存偏好。同步 `main.ts`、`AppSidebar.vue`、`HomeView.vue`、`KeyUsageView.vue` 的 `[CUSTOM]` 初始化，不恢复无偏好时跟随系统深色。内置产品走认证路由 `/tools/recipes`、`/tools/studio`，正文中的同源 iframe 跟随宿主主题；显式独立外部地址仍兼容新窗口。不要把 `.mofa-product-workspace` 改成与 AppLayout 根节点重复的 `.mofa-workspace`。
- 保留 `FamilyWorkspaceView`、`workspace.ts`、共享 `family-runtime/host-client.js` 的来源／窗口／nonce、目录时效、权限复查、退出取消；配方同页选择仅在该宿主下启用，独立模式仍用原选择页。共享运行时须由 Docker COPY、两产品 builder 接入。对子页的同源访问是信任关系，不是安全沙箱；后期跨站拆分须重新明确 CORS、SSO与退出方案，不能传 JWT 到产品 URL。
- 保留 Vite `dev-products.mjs` 的精确开发目录映射与测试；旧 gitignored `vite.config.js` 可抢先加载，预览使用 `vite --config vite.config.ts`。Go 的产品 CSP／SAMEORIGIN 只允许同源嵌入，mp4 文件仍严格白名单并验证 Range206；不开放 PNG原图／来源文档或通配资源。工坊片段署名与许可证不要删，不能称为模型实测作品。回归新增 `family-workspace*`、`family-dev-products`、host-client、配方UI、工坊UI及 TestFamily。

- 一体化产品：保留 vite.config.ts 的 [CUSTOM] 打包器 import／config 钩子、Docker 两项源码 COPY／web embed 测试、Dockerignore 模板与测试例外／本机产物排除，以及 embed_on.go 两条 [CUSTOM] 静态分流。默认 /recipes/／/studio/ 先由白名单静态处理，不能回落为主站 SPA 或误改成必须配置外部地址；配方脚本／样式外置，当前来源选择仍须 nonce／窗口／用途授权。回归 build-products.test.mjs、family-products／home／switcher、配方 connector 与 go test -tags embed -run '^TestFamily' -v ./internal/web；仅镜像构建限定本轮测试，完整web旧logo.png夹具失败仍需另行维护，不冒充全量通过。详见 deploy/FAMILY_INTEGRATED.md。

- 家族首页图标：保留AppSidebar.vue的FamilyHomeIcon定制import与`/family`条目的皇冠组件；不要重新复用DashboardIcon。图形留在custom/components/FamilyHomeIcon.vue，20px用户主菜单与16px管理员个人区继承原尺寸／currentColor；仪表盘图标、路由和登录边界不改。回归upgrade-contract.spec.ts及上游AppSidebar.spec.ts。

- 包号保存400诊断与多专属组：保留custom/dedicated/config.ts白名单错误转换及AdminDedicatedAccountsView.showError，handler保留metadata。后端custom_dedicated_groups.go保证同批成员多个专属组的保存／认证／选号／历史隔离查询一致，成员移除事务清原账号关联组，不误删无关组。运行dedicated-config／dedicated-view、原UserAllowedGroupsModal及TestCustomDedicatedConfigPostgres：合法多专属组通过，关联公开池／名单外授权仍拒绝，两位共组须在同一记录选全；此前严格单组规则已被用户新要求替代。

## 2026-10-01 家族截图修正接缝

- SynaRoute入口与当前态共同保留：family/products.ts为`/keys?product=synaroute`，FamilyProductSwitcher.vue仅在该上下文选中SynaRoute，普通密钥页仍选API；样式位于theme.css并使用紫色product-accent。工坊子路由也保持工坊当前态。
- 自有magic-recipes／magic-studio的embedded吸顶栏必须有不透明主题背景与层级；配方设置scroll-margin-top:88px及目录top:88px配套，避免正文叠字与设置标题遮挡。独立页面行为保留。
- 配方workspace选择不自动开分组权限；有效Key均被生图权限筛掉时显示具体原因，刷新／用途变化清旧模型目录，无选项下拉禁用。相关用例在family-workspace.spec.ts与magic-recipes/site/ui.test.cjs；真实账号响应仍待取证。
- 工坊最新47图及来源见PROVENANCE-CURATED-20261001，公开白名单打包只保留运行WebP，撤下midnight-editorial／coral-sneaker WebP，保留历史PNG与记录。升级后运行build-products.test.mjs与工坊core／ui测试核对完整图／prompt／旧链接；视频仍待开放。

## 2026-09-30 魔法配方登录选择（源码完成，尚未发布）

- 保留custom/routes.ts的受保护 /connect/recipes 与RecipeConnectView.vue、recipes/connect.ts；复用既有router登录守卫及LoginView的safeGuestRedirect，不另写密码收集或登录令牌转交。
- keysAPI.list按登录主体读取本人已有Key并分页；/v1/models只用所选Key查询目录，目录不是模型能力实测。不自动创建／改绑密钥、改额度或开分组生图权限；结果只经用户显式点击带回配方。
- 生产默认只允许同源接收页面；独立部署在原站构建环境配置VITE_MAGIC_RECIPES_ORIGIN为精确origin，无路径／通配。本机站仅额外允许4178本机来源；不要将生产任意来源加入允许名单。env配置不进仓。
- 回传必须验证窗口、来源、随机nonce、用途、五分钟期限及一次性状态；精确targetOrigin与接收确认保留。原站登录JWT、密码不出原站；配方API Key仅内存，刷新清除。COOP隔离导致opener丢失时走明确错误或手动连接，不关闭安全保护。
- 新样式只在theme.css的mofa-recipe-connect作用域；独立配方仍用自己的site/theme.css。99项原站定向与51项独立原型测试、类型／相关lint／构建通过；真实账号闭环、CORS及新增手机控件仍待。只刷新了自己空白测试页，用户旧页材料未动；两端尚未部署8080镜像或生产。

## 2026-09-30 包号专项修复（尚未发布）

- 编辑原样传回`expected_updated_at`，保留后端时间精度；409要求刷新重开。恢复撤销必须明确勾选，发送`reactivate=true`；不得恢复为无条件保存激活。
- 列表`config_status`区分结构异常、成员警告、配置通过；配置通过不是上游实时可用。缺失字段显示未读取，不能默认健康。
- `backend/internal/service/api_key_service.go`新增[CUSTOM]接缝必须保留：所有显式／隐式有效状态写入前重新验证专属资格；请求按当前成员校验，不用其他成员或退出者旧Key状态拒绝全组。保存继续严格验证目标配置，保留独占与历史隔离。
- 新增样式只在theme.css（恢复checkbox尺寸及诊断文案换行）；前后端一起发布，旧页面缺版本被拒绝时刷新。

## 2026-09-30 未发布变更检查

- **2026-09-30 已部署本机Docker（优先于本轮未部署记录）**：用户授权本机测试。原三容器处于停止状态，保留原卷启动Postgres／Redis，先完成597905字节数据库备份与pg_restore目录校验、旧镜像标签sub2apicust:before-console-20260930，再代理预检200并构建。前端类型／国际化3测／构建及Go embed成功；仅应用容器重建，运行镜像与标签一致be2f26f98e8c3dee62d765a8b78e6cd33021dc3b6f618c76085e310180b234b8，三服务healthy，health=ok。迁移242两列和三个索引确认存在；数据库／Redis容器ID及app_data卷不变，用户2／密钥0／组3／套餐1／订单0／包号0／设置280计数保持。备份逐项比较只有Claude版本2.1.283→2.1.285与Codex版本0.158.0→0.159.2两键自动同步，09:44启动日志证实，其余278键不变。HTTP入口index-Ckx_MqIV.js及三个新页面chunk验证通过，账户与管理员接口匿名401；浏览器8080官网正常打开并留给用户测试，尚未做带数据管理员交互验收。证据output/console-deploy-20260930，备份SHA256=1D900EC14FB9915B5E912F0A84778271CEB149178B7CC526E3A9F55411EE037E。未提交推送或操作生产；迁移后的回退需兼顾数据库，不能直接跑旧单用户版本。

- 保留AppSidebar的 `/help/faq` CUSTOM接缝；登录后官网两处 `/preview` 入口隐藏，公开FAQ不受影响，控制台帮助保留侧栏。
- AdminDedicatedAccountsView要显示全部成员名称及账号／组名；编辑不锁定账号／分组；仅撤销后允许确认删除。用户页只显示本人可见信息。
- 后端admin DELETE路由继续走认证／合规／审计；242迁移与custom成员校验、旧隔离标记一起发布，不拆开。按 `deploy/DEDICATED_TROUBLESHOOTING.md` 做备份及回退评估，不直接跑旧单用户版本。
- 智力效率费用是数据源平均任务成本，不能标成本站售价或Token报价；缺失不是免费，分数缺失无分数条（2026-10-12 起分数为雷达分，见下方新节）。
- 本轮139定向前端＋3国际化、类型／lint／build、Docker定向Go／embed与临时PostgreSQL测试通过；未全量Go或远端CI，未更新8080及生产，管理页真实浏览器视觉验收仍待。

本文件只描述维护操作，不授权提交、推送或部署。定制降低冲突面，但不保证未来零冲突。

**发布基线（2026-09-28）**：fa45b2023为游客功能通过全部CI的代码提交，GHCR标签sha-fa45b20已发布；Go单测／集成、前端、lint与安全扫描均由该SHA重新验证。生产升级命令见deploy/UPDATE_GUIDE.md；不要把后续仅文档[skip ci]提交的SHA当作镜像标签。

**最新：2026-09-28游客适配修正**：8080镜像ea6e3a5ecc3c，三服务健康，证据output/preview-responsive-20260928；theme.css游客工作区使用满宽和对称弹性边距，不能恢复1280px上限。升级时检查320／390／768／1024／1440／2031／2560宽度下的左右留白、顶栏操作和字段排列，不只检查横向溢出。

**最新运行状态（2026-09-28游客控制台预览）**：8080已按后续用户授权运行ecb3961845e8；入口index-C4UWqsIx.js和GuestPreviewView-D-VurzQx.js确认新预览已打包。三服务健康，业务计数／设置摘要／数据卷保持，浏览器入口和登录取消／真实套餐通过；证据output/preview-20260928-175007，回退sub2apicust:before-preview-20260928-175007。此条优先于下方旧镜像与未部署描述，生产未改。

**最新运行状态（2026-09-28游客官网）**：本机8080运行镜像`e7a582285cbd`，入口`index-C6kd8wBQ.js`，游客套餐／FAQ真实HTTP和浏览器通过；下方旧镜像记录均为历史。Go embed编译通过，Go单测及完整支付联调仍待。回退镜像`sub2apicust:before-guest-20260928-171257`，数据库备份及验收证据在output/guest-20260928-171257。生产未改。

## 展示层边界

- 2026-09-28新增游客控制台预览：保留首页主CTA、公共导航、custom/routes的 `/preview/:section?` 及router/index.ts后台模式的isPreviewPath接缝。预览不挂载真实账户布局；公开套餐／FAQ保留embedded支持。回归guest-preview.spec.ts与feature-access.spec.ts。179项定向测试及构建通过，本轮新预览仅4175验收，尚未进入上方8080运行镜像。

- 推广清理：保留AppHeader、HomeView、KeyUsageView的 [CUSTOM] 移除标记；ProxyAdBanner保持空兼容组件。回归运行 pnpm exec vitest run src/custom/__tests__/promotion-removal.spec.ts；不要误删客服配置、GitHub OAuth、合规及运维帮助或LICENSE。

- 2026-09-28游客官网：`/` 仍重定向 `/home`，HomeView在自定义内容为空、正常站点模式下复用BrandHomeView；`/brand` 同组件，另有 `/plans`／`/faq`。此条优先于此前“独立品牌页不改变首页”的历史约定。后台模式仍限制公开入口，详见 `GUEST_PORTAL.md`。
- 升级核对新增公共目录只读白名单、订阅／支付开关、游客登录取消、注册与邮箱回跳、`plan` 不覆盖支付恢复／`group`续费；不要公开原结账或订单接口。上游接缝逐处登记在根CUSTOMIZATIONS的2026-09-28游客章节。
- `brand/` 管理中英文案；`components/` 管理品牌面板和主动主题切换；`assets/mofa-mark.webp` 是原批准透明 M 位图。
- 不复制登录业务、控制台页面、API、权限或数据模型。全部视觉规则在 `theme.css`，不重染上游语义 teal 分类色。
- 主题按钮挂载只读取当前主题，用户点击才保存；默认深色仍由上游 main.ts 现有定制初始化。

## 接缝表

| 文件 | 合并后必须保留／验证 |
| --- | --- |
| `src/main.ts` | style.css 后导入 custom/theme.css；默认主题策略 |
| `tailwind.config.js` | primary/dark 色阶引用主题变量 |
| `vite.config.ts` | 原 custom alias 扩展入口；本轮无新增影子替换 |
| `src/router/index.ts` | 404 前展开 customRoutes；保留认证／后台模式守卫 |
| `src/components/layout/AuthLayout.vue` | BrandPanel、主题按钮、mofa-auth-content 双栏容器；保留 default/footer 插槽、设置加载和 URL 消毒 |
| `src/components/layout/AppLayout.vue` | WorkspaceHeading、page-actions 及折叠状态钩子；保留 sidebar/header/default slot 与侧栏折叠宽度 |
| `src/components/layout/AppHeader.vue` | 路径栏 mofa-topbar-context；标题计算移至 custom/brand/useWorkspaceHeading，保留导航和移动菜单事件 |
| `src/components/layout/TablePageLayout.vue` | mofa-data-panel 包裹筛选、表格、分页，原 actions 插槽和移动模式不变 |
| `src/views/user/KeysView.vue` | 原操作移至 AppLayout page-actions；事件、refs、字段与 API 不变 |
| `src/views/user/__tests__/KeysView.spec.ts` | AppLayoutStub 渲染 page-actions，保留原24项业务测试 |

## 魔法家族统一首页（2026-09-30）

- `BrandHomeView` 现在装配 `FamilyHomeView`，默认 `/home`（无后台自定义 home_content 时）及 `/brand` 共用；原 API 介绍迁至 `ApiLandingView` 和 `/api`，新增明确入口 `/family`。保留游客、套餐、教程与原登录成功回跳；不修改全局默认登录落点。
- 最新用户纠正：首页恢复原 BrandPanel 的标志、文案及原三个 CTA，产品介绍放下面，移除账户摘要。`PublicLayout` 普通登录／注册恢复原默认 API 控制台，不回跳 `/family`；子产品匿名入口只携带 `/dashboard?product=id`，由控制台明确打开，不自动跨域跳转。
- 官网导航增加家族首页／魔法 API；`AppSidebar.vue` 的个人区 `[CUSTOM]` 家族首页入口保留。`router/index.ts` 的公开模式等待列表含 `/family`、`/api`，不能绕过后台模式。
- `AppHeader.vue` 新增 `[CUSTOM]` 的 FamilyProductSwitcher 装配与 import；只在有用户时展示，保留原账户、公告、移动菜单。具体入口／意图白名单与安全 URL 在 custom/family/products.ts，新增样式仍仅 theme.css。
- 产品区已改为四标签互动展台，保留 FamilyProductMark、Visual、Action 与 roving-tabindex／tabpanel关联；预览切换不得加入登录／请求／自动授权副作用。用户再次否决的旧卡片不作为视觉基线，当前局部契约在family/SURFACE.md；升级须验默认配方／已知锚点／左右HomeEnd焦点及手机两列，控制台右侧功能不可被产品坞挤压。
- 方法单浮签与图注保持正常流，工坊旋转图区与真实性图注分离，不恢复绝对定位遮字。页头完整工具区保留所需宽度，空间不足时产品坞排第二行；不能只用公告／语言文本替身验收。family-header夹具装配真实公告、语言和订阅组件，账户／数据／国际化为无凭据模拟；不把夹具点击或余额视为真实业务。
- Dockerfile 前端 stage 的 `[CUSTOM]` 三个 ARG／构建 RUN 与 `custom-image.yml` 的三个 Variables 接线共同保留：`VITE_MAGIC_RECIPES_URL`／`VITE_MAGIC_RECIPES_ORIGIN`／`VITE_MAGIC_STUDIO_URL`。生产空地址禁用；本机 DEV 才默认4178／4179，Docker必须显式配置且主站／目标都为回环HTTP。
- 配方静态页仍需另行发布。独立页底部返回家族首页为新窗口，避免丢失材料；本机 `api_site` 仅公开来源提示，必须继续走官方／本机白名单，不导出登录令牌、不自动授权密钥或调用模型。
- SynaRoute 仅走原密钥导入，工坊独立公开内容首版不运行模型／定时任务，无通用第三方账号 SSO 或单点退出。来源、会话与部署验收见 `FAMILY_PORTAL.md`。新增 `family-products.spec.ts`／`family-home.spec.ts`／`family-switcher.spec.ts`／`family-header.spec.ts` 与原升级、游客、路由守卫回归都须检查。

## 每次同步

1. 先检查工作树，备份／保留未提交文档和配置；按根 SYNC.md 演算后再合并。不要为了消除冲突覆盖整个上游文件。
2. 逐条检查根 CUSTOMIZATIONS.md（包括已有 SynaRoute 定制），逐个复核上表接缝。
3. 在 frontend 目录执行：

```powershell
pnpm exec vitest run src/custom/__tests__ src/views/user/__tests__/KeysView.spec.ts src/components/layout/__tests__/TablePageLayout.spec.ts src/router/__tests__/title.spec.ts src/utils/__tests__/synaRouteImport.spec.ts
pnpm exec eslint "src/custom/components/*.vue" "src/custom/brand/*.ts" src/custom/views/BrandHomeView.vue "src/custom/__tests__/*.ts" src/custom/routes.ts src/components/layout/AuthLayout.vue src/components/layout/AppLayout.vue src/components/layout/AppHeader.vue
pnpm run build
git diff --check
```

4. 契约测试只检查结构约定，不证明未来升级兼容。实际打开 `/brand`、登录／注册／找回密码及 keys/usage/dashboard，检查深浅色、360/768/1440、侧栏折叠、长名称、错误／加载状态。
5. 用户授权后用真实后端复核登录、权限、数据表格和图表，再做实际 API 冒烟；后端测试与构建按根交接约定执行。禁止把前端 fixture 或空态当成完整业务验收。

## 接入教程（2026-09-30）

- `/guide` 为公开页、`/help/guide` 为认证后 AppLayout 内嵌页、`/preview/guide` 为只读预览；公开路由仍服从后台模式。官网与控制台入口分开，不把个人路由放开。
- 示例使用 `app.apiBaseUrl` 公共设置（空值回退当前 origin），Anthropic 使用根地址、Codex 与 OpenAI 兼容工具使用 `/v1`；升级后若上游修改公开设置字段、响应协议或模型权限，请重新核对所有示例。
- 外部客户端文档链接会随版本变化，不能把本页步骤当实际调用成功证明；勿在示例、测试或用户反馈中提交真实 API Key。回归 `src/custom/__tests__/guide.spec.ts`、`guest-preview.spec.ts` 与 `src/router/__tests__/feature-access.spec.ts`。
- 样式只在 theme.css；`AppSidebar.vue` 菜单以及 `router/index.ts` 公开模式检查／限定教程锚点滚动共三处 `[CUSTOM]` 逻辑已登记根 CUSTOMIZATIONS.md。不要将目录锚点改为被原置顶逻辑覆盖的跳转；测试应覆盖浏览器返回位置优先。
- 教程scope与FAQ的guideTarget保持公开／控制台／预览各自入口；官网点密钥入口使用GuestAction，取消可继续阅读。安装和第三方配置以页面链接的官方版本为准，静态示例不读取真实密钥。
- 本轮80项定向、3项国际化、typecheck／ESLint／build通过；浏览器公开与预览验收证据见HANDOFF，当前只在4175预览，8080与生产未更新。

## 智力效率改接雷达Bench v1（2026-10-12）

本节优先于下方 2026-09-27 节中的数据源、30 分钟轮询、IQ、harness、低样本 30 等描述。

- 浏览器不再直连第三方，只读本站 `GET /api/v1/intelligence`。后端 `custom_intelligence_handler.go` 代读 codex-reset-radar.pages.dev 的雷达Bench公开汇总，并缓存 10 分钟；前端每 10 分钟轮询一次，超时 60 秒。CSP 的 connect-src 不需要放行源站。
- 合并上游时保留 `backend/internal/server/router.go` 的 `[CUSTOM] RegisterIntelligenceRoutes` 一行。前后端同版发布：只更新前端会因接口不存在而显示“数据源暂时不可用”。
- 服务器必须能访问 codex-reset-radar.pages.dev。需要代理时配置 `update.proxy_url`，与更新检查共用。
- 字段与精度和源站页面一致：分数 1 位、费用 2 位、耗时 1 位；覆盖少于 40 题标“数据不足”。缺失显示 —，不当作 0 分或免费；0 分是有效成绩。
- 回归：`go test -tags=unit -run TestIntelligence ./internal/handler/`；前端 `pnpm exec vitest run src/custom/__tests__/intelligence.spec.ts src/custom/__tests__/intelligence-view.spec.ts`，然后 build。

## 智力效率与订阅核对（2026-09-27）

### 本轮用户反馈后的差异

- 「查看源站」、JSON外链和底部来源文字已按用户要求移除；底部仅保留IQ可比性及低样本／缺失数据两段。界面不再提示刷新频率，但useIntelligence仍按原30分钟轮询。合并不要误删刷新行为。
- 初版IntelligencePlot已由IntelligenceComparison替换，不再使用无标签散点。模型／环境／档位逐条显示分数、耗时、费用及样本；可筛选档位、选择单指标排序，每页12条、各指标缩放范围跨分页保持一致，缺失显示破折号、零值不隐藏，不生成综合效率排名。
- 矩阵默认按JSON中模型＋环境首次出现的顺序，新增原始／名称／指定档位降序选项；同分保留原序、缺失最后，原档位列顺序不变。检查排序和筛选后数据仍正确。
- 支付空态宽度修正在theme.css的 `.mofa-workspace-main > .mx-auto.max-w-4xl`；这是定制flex布局与上游mx-auto容器的适配，不改支付业务。upgrade-contract新增此root class契约；若上游改变容器需人工核对，真实浏览器效果待验收。
- 本机订阅分组「测试分组1」已存在但售卖套餐0条：按后台「订单管理 → 订阅套餐」设置关联分组、售价、有效期并上架；分组额度不等于商品定价。用户没有授权具体售价，不能自行创建售卖套餐。
- 76项回归与定向ESLint／前端构建通过；浏览器连接仍报apikey认证不支持，不宣称已完成真实手机／双主题视觉验收。

- `/intelligence` 由 custom/routes.ts 注册，保留 requiresAuth；AppSidebar.vue 的 `[CUSTOM]` 入口属于新增上游接缝，用户主菜单及管理员个人区共用。WorkspaceHeading 按语言显示标题，主题仅改 theme.css。
- 依赖上游 `ModelIcon`、`AppLayout`；新对比组件已不用Chart.js。合并后检查模型图标黑色fill与深色主题可见性、矩阵固定首列、指定档位排序、对比分页和手机堆叠。
- 固定第三方数据源 `https://codexradar.com/data/intelligence-efficiency.json`。本轮实测 schema=2、CORS `*`，源更新时间 `2026-09-26T18:07:30+08:00`。源站改接口／CORS或部署者收紧CSP后可能读取失败；不得伪造新时间或把缺失字段补0。来源在维护文档与代码中保留，页面说明以用户最新指定的两段为准。当前上游默认 connect-src 允许 HTTPS，不改安全配置。
- 这是**浏览器打开页面期间**的30分钟轮询，不是无人访问时也跑的服务端采集；刷新失败保留同一挂载内的旧数据并报错，刷新浏览器后需重新拉取。网络超时15秒，退出页面清理计时器和请求；隐藏页暂停，回前台过期时补取。
- 使用源站模型名称及IQ，不等同官方型号认证／本站模型目录／人类智商。少于30样本标记，只是展示阈值；耗时和费用保持源值，费用不是本站售价，不混合不同harness。
- 回归命令：`pnpm exec vitest run src/custom/__tests__ src/views/user/__tests__/PaymentView.spec.ts src/components/payment/__tests__/SubscriptionPlanCard.spec.ts`，然后 `pnpm run build`；用户需在真实浏览器验收移动端／双主题／键盘／对比排序与分页／第三方网络受限路径。
- 充值与订阅继续使用 `/purchase?tab=subscription` 原生页面。没有套餐时显示空态属预期；本机0套餐，不得为了截图伪造售卖数据。配置入口为后台订单管理下套餐管理，实际分组、售价和配额待用户决定。
- 后续用户授权本机重建后，8080已运行新镜像`1a7a31d0e424`；HTTP入口`index-CYZhTaMW.js`引用智力页面`IntelligenceView-Cob7bqKS.js`，CSS新规则及数据来源地址已验收。数据库／Redis未重建、设置摘要未变；生产未动。以后本机重建仍先读 LOCAL_DOCKER_RUNBOOK，不能把HTTP检查当浏览器视觉和完整业务验收。

## 品牌局部回退

先备份当前 diff。仅移除本次品牌组件引用、mofa 展示钩子与 /brand 路由，回退对应新增视觉规则；保留原插槽、上游组件与既有错误态修复。不要 reset 整仓、覆盖整个布局文件、删除数据库卷或回滚其他人的未提交文档。

## 当前部署限制（本轮Docker发布后）

当前本机8080运行`cfa0a12c8bf6`，含本轮矩阵排序、三指标对比、隐藏频率／外链、支付宽度修正；HTTP入口`index-BW6NgO3M.js`／智力chunk`IntelligenceView-C0f1Aymd.js`。品牌配置和用户新建分组保持，76项回归及Docker前端／Go embed构建通过；真实视觉仍待。改动尚未提交推送，生产未动。最近回退镜像`sub2apicust:before-intelligence-ux-20260927-022446`，dump583628字节，证据output/brand/intelligence-ux-deploy-record.json；网络预检仍遵循deploy/LOCAL_DOCKER_RUNBOOK.md。

## 包号页面（2026-09-28）

- 2026-10-01最新未发布计费实现覆盖原首版扣余额说明：有效包号实扣0，普通组仍计费；服务端独立准入与迁移243，完整上游接缝见CUSTOMIZATIONS最新节，不能仅更新前端文案。
- 2026-10-10最新要求覆盖历史共享限额：包号只负责归属、有效期和免扣。DedicatedBillingPolicy组件、GET／PUT billing-policy接口和独立限额已删除，成员100人上限也已移除；通用请求/生图权限由sub2api原配置执行。同步不要恢复旧表单、API调用或旧默认限额文案。
- 同步保留用量表类型2标识／筛选、用户CSV与管理员Excel的计费类型；实扣0与原始参考金额分列。用户页复用管理员UsageTable，不应漏掉用户标签。前端关键回归保留dedicated-view、dedicated-config与UsageTable，已删除旧dedicated-billing-policy测试；2026-10-10包号43项、类型与相关lint通过，真实浏览器与业务客户端验收待。

- 新路由：用户 /dedicated-accounts，管理 /admin/dedicated-accounts；侧栏入口、requiresAuth／requiresAdmin、WorkspaceHeading中英文接线必须共同保留。两菜单仅标准模式可见，后端也拒绝简易模式创建／使用包号。
- 页面、选择器、API和文案位于custom；所有新增样式仅在theme.css的mofa-dedicated作用域，不覆盖上游账号管理界面。
- 用户数据只能来自专用只读API，不能改成调用/admin/accounts或在前端过滤管理员响应。不得把unknown/stale转为零额度或100%。
- 前端回归加入src/custom/__tests__/dedicated*.spec.ts（27项）；数据样例仅用于测试，不代表真实账号验收。
- 后端迁移、归属和调度接缝逐处见根CUSTOMIZATIONS.md，操作／回滚约束见DEDICATED_ACCOUNTS.md。

## 2026-10-07 订阅开通周期与官方直码

后端244迁移与服务代码同版：周／月窗口从开通初始化，列表和progress投影当前周期，前端仍只用服务器resets_at；不伪造等待首次使用的时间。支付宝新订单qr_code是校验后的官方签名WAP地址，pay_url与之相同，payment_mode仍qrcode；保留扫码等待、状态轮询和恢复原订单。旧短码只兼容存量，不能恢复为默认新码；新增2300字节编码容量回归与签名实付金额回归。
