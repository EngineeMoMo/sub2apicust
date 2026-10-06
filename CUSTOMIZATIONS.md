# CUSTOMIZATIONS — 本 fork 相对上游的所有改动登记

## 2026-10-06 发布门禁补修

- `deploy/update.sh` 指定标签替换：CI macOS 的 BSD sed 不支持 GNU `-i -E` 组合，改为先捕获 `sed -E` 输出再 printf 写回，保留原备份与文件权限，标注[CUSTOM]；Linux假Docker 22场景通过，macOS首次复验进一步发现Bash 3把相邻中文括号读入变量名，回退镜像提示改用显式花括号边界；由新SHA CI复验。
- `frontend/package.json` / `pnpm-lock.yaml`：Vue最低版本升至3.5.42，统一带入修复后的server-renderer；增加source-map-js<1.2.2的补丁override。依据本次pnpm audit的GHSA-g2v6-rqmx-r4w6／GHSA-68fv-2mgg-jv7q修复，不增加安全扫描例外。JSON不能写注释，以本节登记接缝。

## 2026-10-06 同步上游 v0.2.13（本地合并，未发布）

- 基于稳定标签 `3040209f2` 合并；`backend/cmd/server/VERSION` 对齐上游随后 `b8dece900` 的0.2.13元数据。原未提交的支付、订阅、更新脚本等48文件先备份到output/upstream-merge-20261006/before并保存本地快照e2adf6005。
- `backend/internal/server/routes/payment.go` 唯一文本冲突：同时保留标注[CUSTOM]的WAP扫码Guard／IP限流与游客套餐目录，以及上游匿名订单verify独立20次/分钟限流；不丢任一分支。
- 新上游接缝 `backend/internal/handler/gateway_inflight_reservation.go:reserveInflightBalanceCtx` 标[CUSTOM]：有效私有包号凭证跳过余额估价与预占；普通Key、身份不匹配Key及订阅原逻辑保持。包号准入／共享限额仍在认证层独立执行。新增 `custom_dedicated_inflight_test.go` 通过真实资格解析构造凭证，验证HTTP／WS共用入口、零余额及已有预占、未定价fail-closed、普通与不匹配Key不豁免。
- custom支付宝测试适配createOrderInTx新增bonusAmount及RegisterPaymentRoutes新增Redis参数；增加赠金／折扣订单的签名实付金额校验，避免以到账额替代实付。
- 上游测试接缝 `frontend/src/api/__tests__/settings.authSourceDefaults.spec.ts` 按实际新增TypeSafe保留六个平台；`frontend/src/views/user/__tests__/UsageView.spec.ts` 保留本仓用户CSV的Billing Type列，两处均标[CUSTOM]。Makefile关键测试加入这两个文件，避免只跑旧子集漏检。
- 迁移仍按完整文件名记录，不重命名已发布的241/242/243包号迁移；新增241_add_payment_order_bonus_amount与241_add_typesafe_platform已在独立PG的新库、旧定制库升级及重复执行中验证。主题、SynaRoute与custom目录既有代码保留，无新的业务配置／数据库写入。
- 验证与未完成门禁见HANDOFF最新状态，证据output/upstream-merge-20261006；完整integration已在用户明确授权socket后通过，52包／6,997顶层＋6,586子项；安全审计专用PG／Redis补测214项通过。仍有6项既有／外部依赖跳过，清单见integration-summary.json；不等同真实付款／模型请求成功。

## 2026-10-06 支付宝支付安全复查补强（未发布）

报告与证据边界见 [deploy/ALIPAY_SECURITY_REVIEW.md](deploy/ALIPAY_SECURITY_REVIEW.md)。新增 `backend/internal/payment/provider/custom_alipay_security.go` 统一严格签名字段／金额格式及固定公钥模式；新增 `backend/internal/service/custom_alipay_notification_security.go` 精确比较支付宝网关金额。对应真实RSA签名／SDK查询测试为 `custom_alipay_security_test.go`；实际仓储并发余额测试为 `backend/internal/server/routes/custom_alipay_fulfillment_security_test.go`，service中 `custom_alipay_desktop_wap_security_test.go` 覆盖URL歧义、可信站点、HTTPS与并发租约／Host fuzz。

上游接缝均标 `[CUSTOM]`：

- `backend/internal/payment/provider/alipay.go` 的 VerifyNotification 要求完整签名字段／app_id匹配及严格total_amount，SDK继续验签，禁止用配置补商户、用实收或买家金额补总额；剔除SDK不签名的alipay_cert_sn，裸公钥接入不能被请求触发下载证书。QueryOrder已付结果须原out_trade_no、非空trade_no及合法总额。
- `backend/internal/payment/provider/alipay_test.go`：原金额回退测试改验严格总额解析，旧未使用的parseAlipayAmount随生产回退逻辑一同移除；provider整包unit复跑通过。
- `backend/internal/service/payment_fulfillment.go:confirmPayment`：官方支付宝精确匹配两位网关金额，拒绝一分钱差异；其他通道原容差保持。
- `backend/internal/service/payment_order.go`：创建新扫码订单前，以及invokeProvider准备地址时，使用新canonical helper，仅信任管理员frontend_url／未配置时请求Host，不再用Referer扩大允许站点；HTTPS强制，本机回环例外。
- `backend/internal/server/routes/payment.go`：新增的公共扫码路由在原IP限流前挂AlipayDesktopWapGuard，保证错误和429也禁止缓存。

已新增的custom文件内部补强：handler严格解析单一token、限制query长度／规范订单ID，正常302只写Location不输出付款URL正文；service拒绝WAP参数及biz_content重复键。配置、迁移不新增；手机发起分支不变，官方通知安全加固覆盖各支付模式。同步不可恢复金额回退／Referer信任／一分钱容差；须运行TestCustomAlipay和既有支付／恢复回归。既有更新脚本专项与并行订阅UI不属于本次安全修改。

## 2026-10-06 我的订阅时间提示与样式优化（源码完成，未发布）

审查后修复：`SubscriptionsView.vue` 列表独立结束 loading，辅助 progress 异步补齐；通过请求版本号丢弃重试前及卸载后的旧响应，不再用 Promise.allSettled 阻塞列表展示。`subscription-timing.spec.ts` 增加慢响应、延迟失败及旧请求成功／失败回归，共25项；加既有quota共30项及类型／相关lint通过。未更改样式或后端，未发布。

用户要求核对并补齐截图中的到期／额度重置提示、优化现有样式。原页面已显示到期和重置文本，但只在渲染时计算，日窗口按首次使用加24小时、周／月按本地固定时长，过期窗口还会拼成“等待首次使用 后重置”。本轮仅调整订阅展示，服务端额度、续期、认证和数据库不变。

- 新增 `frontend/src/custom/components/SubscriptionCard.vue`：提取原套餐名称／平台／描述／倍率／高峰费率／用量内容；到期日期始终保留，有效订阅显示剩余时间，3／7天提示、过期／暂停／撤销状态区分；用量条提供原生progress语义。有效及过期可沿原 `/purchase?tab=subscription&group=…` 续费，暂停／撤销不提供入口。非有效订阅不展示未来重置提示；到期早于下一重置时显示额度在到期结束，已到重置时刻但尚无新窗口时明确“使用后更新”，不假定额度已发放。
- 新增 `frontend/src/custom/subscriptions/timing.ts`：通过原认证apiClient读取已有 `/subscriptions/progress` 的真实 `{subscription, progress}` 结构，只取服务端 `resets_at`；不复制服务端日历日、旧周／月锚点和最后不完整周期算法。共享15秒本地时钟、标签页可见变化时刷新及卸载清理，只更新时间显示，不自动发网络请求或清用量。窗口未启用显示等待首次使用，缺失／非法时间明确未读取。
- 新增 `frontend/src/custom/__tests__/subscription-timing.spec.ts`：21项包括真实响应结构、权威日时间、跨到期／过期／日卡／暂停／撤销／无效时间、分钟更新／后台标签返回／卸载、失败重试、原续费路由与英语。新增文案在custom层按现有locale显示，原中英文键仍复用。

上游接缝（均有 `[CUSTOM]` 标记）：`frontend/src/views/user/SubscriptionsView.vue` 提取原展示到custom组件，保留原列表API、空态、AppLayout和续费路由；并行读取已有时间接口，时间失败保留套餐用量并允许重试、列表失败明确错误而非空订阅。`Makefile` 关键回归新增一项，不删除原支付宝／包号等测试。视觉规则仅新增到 `frontend/src/custom/theme.css` 的 `mofa-subscription*` 作用域，复用既有雾钛青深浅tokens，平台徽标沿原语义色；桌面双栏／窄屏单栏、长名称换行、数字对齐及手机44px续费触控。

验证：21项新回归＋5项旧订阅quota回归、27文件399项关键回归、vue-tsc、相关ESLint和最终Vite构建通过。真实IAB浏览器组件夹具验证1440桌面深浅、390手机深浅／英语、900中间宽度、长标题及不同状态，无横溢出；手机按钮44px，测试页warn／error为空。证据 `output/subscription-timing-20261006/`，截图明确标模拟数据，不能当真实登录／付款验收。未改后端、配置、依赖、迁移或业务库，未重建8080、提交推送或生产部署；此前支付宝和更新脚本专项保留。同步核对见custom/UPGRADE.md。

## 2026-10-06 电脑端支付宝手机网站支付扫码（源码完成，未发布）

用户授权实施电脑扫码计划：默认关闭开关，开启后仅选中的官方支付宝桌面订单使用 `alipay.trade.wap.pay`；手机浏览器与其他通道仍走原路由。扫码短入口使用本站已校验回跳 origin 与 32 字节随机令牌；校验通过直接 302 到订单保存的官方 WAP 收银台，无本站登录或中间操作页。实际 App 扫码兼容性及实付尚未验证。部署验收见 [deploy/ALIPAY_DESKTOP_WAP.md](deploy/ALIPAY_DESKTOP_WAP.md)。

新增文件：

- `backend/internal/service/custom_alipay_desktop_wap.go`：模式判定、随机短链接、官方网关／商户／金额／订单／分钟期限校验、令牌及状态授权。
- `backend/internal/handler/custom_alipay_desktop_wap.go`：免登录 302／明确错误提示，no-store、no-referrer，不记录完整付款地址。
- `frontend/src/custom/components/AlipayDesktopWapSetting.vue`：复用现有 Toggle、可访问标签及中英帮助文案，无主题样式变更。
- 后端 `service`、`payment/provider`、`handler/admin`、`server/routes` 下四个 `custom_alipay_desktop_wap_test.go`：配置／SDK 路由、真实本地签名订单、免认证真实路由与攻击用例；前端 `src/custom/__tests__/alipay-desktop-wap.spec.ts`：开关、桌面余额／订阅扫码、恢复、手机旧跳转。
- `deploy/ALIPAY_DESKTOP_WAP.md`：启用、真机验收和证据边界。

逐处上游接缝（均标 `[CUSTOM]`）：

| 文件 | 本轮接缝与同步检查 |
| --- | --- |
| `backend/internal/service/payment_config_service.go` | 新存储键 `ALIPAY_DESKTOP_WAP_QRCODE`、支付配置 bool／更新指针、批量读取／默认 false 解析／非 nil 保存；遗漏更新保留原值 |
| `backend/internal/handler/dto/settings.go` | 管理员响应字段 `payment_alipay_desktop_wap_qrcode` |
| `backend/internal/handler/admin/setting_handler.go` | 管理员读取响应映射 |
| `backend/internal/handler/admin/setting_handler_update.go` | 指针字段、支付保存请求、更新响应及 hasPaymentFields；单独更新开关也必须生效 |
| `backend/internal/payment/types.go` | 内部请求增加桌面 WAP 标记及 `ExpiresAt`，只能由订单服务赋值 |
| `backend/internal/payment/provider/alipay.go` | 原手机分支之后、桌面分支之前选 WAP；桌面开关不受实例 redirect 模式影响；仅新模式设置东八区分钟 `time_expire`，不改变手机参数 |
| `backend/internal/service/payment_order.go` | 创建订单截断分钟并冻结 snapshot 标记／实际 `payment_mode=qrcode`；调用 provider 传模式／期限；保存前构造本站短 QR；响应模式覆盖，原 WAP URL 与 QR 用现有订单字段保存 |
| `backend/internal/server/routes/payment.go` | 在 payment/public 注册唯一 GET `/alipay/wap/:id`、PublicIP 限流；不得给个人订单接口解除认证 |
| `backend/internal/server/api_contract_test.go` | 两个管理员设置响应 fixture 增加新 bool，原 API 契约保持 |
| `frontend/src/api/admin/settings.ts` | 管理员读／写接口类型增加 optional bool，旧客户端遗漏兼容 |
| `frontend/src/types/payment.ts` | 支付配置 `alipay_desktop_wap_qrcode` 类型；结账继续由服务端响应／订单 snapshot 决定显示模式 |
| `frontend/src/views/admin/SettingsView.vue` | 新组件 import／支付区装配／form 默认 false／save payload；不改两个手机选项 |
| `frontend/src/i18n/locales/zh/admin/settings.ts`、`en/admin/settings.ts` | 对齐新开关标题和免本站登录说明 |
| `frontend/src/views/admin/__tests__/SettingsView.spec.ts` | 实际页面读取 true／保存 false／保留手机两项配置回归 |
| `Makefile` | FRONTEND_CRITICAL_VITEST 增加新扫码回归，保留包号等原关键项；既有 CI 调用自动覆盖 |

维护边界：不新增迁移、不修改 theme.css、依赖或现有支付通知入账逻辑；关闭开关只影响新订单。扫码令牌只授权读取该订单已保存的付款地址，不创建订单、不入账、不返回用户资料。白名单限定 HTTPS `openapi.alipay.com/gateway.do` 与 WAP 方法，禁止用户信息、端口及 fragment；保存地址与订单不匹配拒绝跳转。新功能不可降级成仅将完整 WAP 长地址直接塞入二维码。

验证：本轮前端 26 文件／378 关键项、定向 128 项、类型／ESLint／Vite 构建，后端支付 195 顶层＋87 子项、API 契约 1 顶层＋11 子项、embed 构建及 Go lint 0 issues 通过；本地 SDK 仅生成签名，不调用商户网络。Windows 全量 unit 已执行但图片路径及 Ollama 时间精度两项仍失败，Linux CI／integration／真机扫码／用户确认的小额实付待验证，不能沿用旧提交门禁。证据 `output/alipay-desktop-wap-20261006/`；更新脚本、UPDATE_GUIDE、两个假 Docker 测试及 backend-ci.yml 字节与开工基准一致。未提交推送、重建 8080 或部署生产。

## 2026-10-01 配方打开原图修复

仅修改自有product-samples/magic-recipes/site/{model-ui.js,template.html,theme.css,ui.test.cjs}，无新增上游接缝、依赖或CSP例外。旧生成图片链接直接target=_blank打开data地址；现在每张图有页内预览按钮，以原生dialog和img展示原图，PNG／JPEG／WebP Base64链接仅用于download。远程HTTPS图沿原安全筛选、referrerpolicy=no-referrer，在dialog预览，提示长按或右键保存；不伪造跨域下载或追加收费请求。Esc／关闭回焦点、Tab循环，多图序号和MIME扩展名对应，材料／配方／连接／等待重置及pagehide清源。同步后不得恢复data新窗口跳转；母站frontend/src/custom/theme.css未改，产品新增预览布局复用既有配色。

61配方＋3打包、类型和生产构建通过；旧构建在新回归失败，本机已有Logo夹具在独立及外置CSP的同源iframe实际解码512×512，桌面浅色／390深色无横溢出。下载事件超时，文件落地及真实线上生成未知；证据output/recipes-image-open-20261001。2026-10-02功能4bf1436fb同SHA CI／安全／GHCR全部成功，固定镜像sha-4bf1436及摘要见HANDOFF终态，含上一版包号免扣。没有重建本机Docker或部署生产。

## 2026-10-01 包号独立免扣与共享限制（已推送，门禁全绿）

用户已授权“开始处理开发”及“提交并推送代码，我发布更新”。功能acd73d716已推origin/main，65文件；其CI（unit／integration、lint、前端类型与372项）、安全和GHCR均成功，实际镜像sha-acd73d7及摘要见HANDOFF最新终态／ci-final.json。本节覆盖首版包号仍扣余额的历史边界，发布同时包含既有同批成员多专属组、移除成员、历史隔离、Key恢复、旧表单保护及所需三条主题样式；运行细则见deploy/DEDICATED_BILLING_PLAN.md第六节。前后端和迁移243同版，生产由用户备份更新，真实业务仍待验收。更新脚本专项继续保留未提交。

新增文件集中在custom：service/custom_dedicated_billing.go与custom_dedicated_admission.go、repository/custom_dedicated_billing.go、middleware/custom_dedicated_billing.go、迁移243；测试包含私有凭证、独立计量、缓存不扣、两平台／WS记录、真实PG准入／结算／认证以及fixture。前端新增custom/components/DedicatedBillingPolicy.vue及其回归，原自有管理页／API／文案接独立策略表单，无新依赖，本轮未改theme.css。策略GET／PUT沿用管理员认证／审计，updated_at防冲突；并发2、RPM30、日上限0、2MiB、生图关闭是首版默认，并非按Token或美元的硬预算。

所有新增上游接缝均标[CUSTOM]，逐处如下；同步后必须保留相应行为并运行包号／WS和普通计费回归：

| 上游文件 | 原行为与本轮改动 | 同步后的验证 |
| --- | --- | --- |
| backend/internal/service/api_key.go | 增加不序列化的请求私有准入凭证，只有服务端认证副本带权益，不能写共享缓存 | 无绑定／伪造身份不免费；原缓存Key字节身份不变 |
| backend/internal/server/middleware/api_key_auth.go | 正向解析权益、复制Key，只对有效包号跳余额条件；认证完进入共享准入 | 零余额包号进入handler；公共组仍需余额，Key状态／到期／配额仍拦截 |
| backend/internal/service/billing_cache_service.go | 包号跳余额及普通user×platform消费预检，保留Key窗口／RPM等 | 有效包号零余额通过；普通请求余额／平台额度仍原行为 |
| backend/internal/service/gateway_usage_billing.go | 两平台共用免扣结算与原事务；Key增量改用独立参考量；包号拒旧兜底、不减余额缓存／不发低余额通知，Claude执行ID来自准入 | DB与缓存实扣0，Key窗口增量非0；错误不转余额；实际账号与准入一致；跨期不追扣 |
| backend/internal/service/openai_gateway_usage.go | OpenAI执行ID由服务端准入生成，WS轮次哈希不再被原上游ID逻辑覆盖 | 同次回调幂等；不同HTTP执行／WS轮次分别计量，客户重复编号不吞真实用量 |
| backend/internal/service/usage_billing.go | 命令增加权益归属／租约／参考量，包号独立指纹与金额量化，普通指纹保持原算法 | 普通历史请求去重不变，包号不同归属冲突，参考金额规范8位 |
| backend/internal/repository/usage_billing_repo.go | 原去重事务内验证准入租约，写独立包号账本，再执行Key增量；包号BalanceCost／SubscriptionCost必须0 | 伪造／错误账号拒绝；账本失败全部回滚；12并发同回调只写一次 |
| backend/internal/handler/openai_gateway_handler.go | 包号WS每个response.create含首轮共享准入；控制帧钩子只允许已适配生成／取消 | 到期撤销／限额阻止新轮；session.update与重复type键不能藏生图或实时音频 |
| backend/internal/service/openai_ws_forwarder.go | Hooks新增包号专用控制帧校验 | 普通连接钩子为空／校验nil不改变原行为 |
| backend/internal/service/openai_ws_v2_passthrough_adapter.go | 后续透传帧在发送前调用包号控制帧钩子 | 中途session.update不能绕逐轮生图／次数保护，普通WS仍能使用 |
| backend/internal/service/openai_ws_forwarder_ingress.go | 原生Ingress读取后续帧也接同一控制帧钩子 | 两种Ingress包号边界一致，原普通连接回归通过 |
| backend/internal/handler/gateway_key_billing.go | 倍率自省增加可选dedicated_prepaid，客户有效倍率0；不改原配置参考倍率 | 包号／普通倍率显示准确，原字段保持兼容 |
| backend/internal/server/routes/admin.go | 专用GET／PUT billing-policy接原管理员路由 | 普通用户无权写，缺失绑定404，冲突409，非法数值400 |
| frontend/src/components/admin/usage/UsageTable.vue | 成本列标包号实扣0，Token／原始金额保持，用户页复用此表；其.spec.ts增加金额回归 | 类型2显示实扣0／参考金额非0；普通成本保持 |
| frontend/src/components/admin/usage/UsageFilters.vue | 新增类型2筛选 | 可筛选包号，不混普通余额／订阅 |
| frontend/src/views/user/UsageView.vue | 类型2筛选，CSV新增计费类型 | 包号CSV实扣0，参考金额另列，普通字段同原口径 |
| frontend/src/views/admin/UsageView.vue | Excel新增计费类型 | 包号类型与实扣0可辨认，原成本明细不丢 |
| frontend/src/i18n/locales/zh/admin/resources.ts、en/admin/resources.ts | 包号计费筛选／标识中英文 | i18n键完整，金额与标签一致 |
| .github/workflows/backend-ci.yml | Unit job增加独立Postgres服务与DEDICATED_TEST_POSTGRES_DSN，使真实包号测试在CI实际执行 | PG专项不能只skip，服务为空库；shell更新脚本接线仅留工作区，本次不提交 |
| Makefile | 前端关键Vitest列表加入包号策略／管理／配置及UsageTable | 本轮25文件372项通过，新CI包含这些回归 |

验证：便携官方Go1.27.0与现有Linux Go测试镜像；真实空白PG随机schema，包号／WS定向81顶层测试＋82子场景成功；Linux全量unit通过，最终控制帧补丁后重跑定向、Go lint0 issues及embed构建；前端372关键回归、类型／相关lint／Vite构建通过。Windows全量第一次遇到路径／sh和Ollama时间精度失败，Linux初试只读ent与SDK代理问题已修正测试环境后重跑成功，未改这些无关业务源码。随后实际acd73d716的CI全量unit／integration、PG及372项回归、安全与镜像全部成功，证据output/dedicated-billing-20261001。真实浏览器及收费账号仍待；开发前后22个原并行文件SHA256一致，发布时仅将必要包号依赖纳入，更新脚本保持。

## 2026-10-01 用户授权发布门禁修复（已发布）

- 功能293c091b3ab0c8f187c031bed1955bb843939d40已推origin/main，限定10文件；同一SHA的CI36842744777（Go lint、unit／integration、前端／shell／release）、Security36842744680、GHCR36842744656均success。镜像实际embed编译／推送成功，固定sha-293c091、摘要c517187b769f8b49f1fb29b1667277b745bb42deb5be05eabbf3f39e54274a06；完整链接、证据与手动更新说明见HANDOFF及deploy/UPDATE_GUIDE。源码包含上一轮47图／滚动／导航与目录诊断，真实账号的线上模型列表仍待验收，生产未操作。
- 用户已明确“修复好再发布”，解除前轮依赖安装授权阻塞。上游依赖文件`backend/go.mod`仅x/image0.41.0→0.45.0并加[CUSTOM]注释，`backend/go.sum`两条校验来自官方sumdb；其x/sys／x/text要求与现有锁定一致。`frontend/package.json`与`pnpm-lock.yaml`仅Axios1.18.1→1.20.0，JSON／锁文件不支持注释，接缝在此登记；七项Axios high已从生产审计消失，原例外清单与扫描规则未改。
- 自有`custom_studio_handler.go`八处资源清理错误使用studioCleanup检查；正常重命名后的临时路径不存在不记失败，其他失败写服务端日志，保留原业务错误／响应，不记录投稿材料。新增`custom_studio_cleanup_test.go`及来源明确的WebP测试副本覆盖正常解码、截断格式、最小尺寸、2400万像素限制、拒绝无落盘和重命名失败清理。没有修改上传权限／发布确认、额度／大小／媒体白名单或视频暂停状态。
- 本机Go编译器仍缺、Docker未启动，未把本机格式化当编译通过；只使用官方Go1.27.0归档核验SHA256后提取的便携gofmt处理本轮两文件。限定暂存树c3423c3导出源码，Node106＋Vue538共644项、类型／源码lint／Vite构建及原审计例外检查通过；新SHA Go CI／安全与GHCR终态均成功，证据output/family-security-fix-20261001。29个并行文件SHA256未变，共享theme未改；UPDATE_GUIDE只补本发布说明，原并行章节部分暂存保留。包号及更新脚本工作继续保留，不代部署生产。

- **截图修正版安全扫描最新终态**：d84d60e的Security36836909442已failure；新任务两个job日志确认Axios七项high与x/image三个WebP／VP8L漏洞仍在，CI／GHCR仍运行时取证，不把已推送或镜像存在当安全通过。安全依赖安装被自动审批拒绝，未执行，等待用户明确授权；本轮代码推送完成，生产未改，见HANDOFF与output/studio-refinement-release-20261001。

## 2026-10-01 截图修正版已推送（最新状态）

- 功能`d84d60e07d9abb776f8d2dd8b0e8bdef60dc9f7e`已推origin/main，65文件；实际源码644项、类型／源码lint／Vite构建通过。29个并行文件SHA256未变，共享theme原并行规则按还原本轮新增规则后的精确SHA256确认保留。范围与证据见output/studio-refinement-release-20261001。
- 本SHA的CI36836909715、Security36836909442、GHCR36836909685已启动，当前in_progress；镜像标签／摘要尚未确认，不能提供更新通过结论。安全依赖升级被自动审批拒绝，安装未执行，等待用户明确授权；不关闭扫描、不增加例外，生产未改。新修复与正式更新仍待，旧章节的未提交状态属历史。

## 2026-10-01 截图修正版提交授权与范围

- 用户要求提交并推送以便手动更新，限定65个本轮产品／素材文件；共享theme只包含产品两处hunk，包号专项、更新脚本／测试和CI接线仍留工作区。实际暂存导出源码验证Node106＋Vue538、类型及Vite构建，lint生成副本与源码的边界见HANDOFF最新记录。证据output/studio-refinement-release-20261001。
- 安全依赖升级安装被自动审批拒绝，依赖／锁文件仍未改；已向用户请求Axios1.20.0、x/image0.45.0及投稿清理检查修复的明确授权。未扩大安全例外、关闭扫描或处理未授权依赖，未部署生产。旧765e5a9的单测／集成已成功，Go lint与安全仍失败，不能据旧镜像给通过结论。

## 2026-10-01 用户截图修正与素材收口（最新源码，未发布）

- 所有本轮代码位于自有`frontend/src/custom/`及`product-samples/`，没有新增上游接缝。`family/products.ts`为SynaRoute配置入口加入`product=synaroute`上下文；`components/FamilyProductSwitcher.vue`按该上下文显示SynaRoute当前态，普通密钥页仍选API，工坊子路由保持工坊当前态。紫色选中态只改`custom/theme.css`，保留并行包号规则；未改SynaRoute客户端／导入。
- `FamilyProductVisual.vue`移除用户指定提示文案与无用hint规则，三图预览和点按交互保留。`magic-recipes/site/theme.css`修复内嵌吸顶栏全宽、不透明背景、目录顶部间距和设置锚点；`magic-studio/theme.css`为内嵌栏补不透明背景与层级。沿用现有主题变量，没有另做品牌换肤。
- `family/workspace.ts`在有效密钥全部因生图权限被筛除时返回明确原因；`magic-recipes/site/model-ui.js`刷新／用途变化清空旧目录，空列表和请求失败禁用无选项下拉并显示诊断。不绕过服务端权限、自动开生图权限或创建密钥；生产账号响应尚未取得，不能据此确定用户现场根因。
- `magic-studio/new-gallery.mjs`与`catalog.mjs`接入五类各两张独立生成图及首页同一窗边人像；删除用户指定两项公开条目与四个WebP，旧PNG／provenance留作历史。当前47图、39份准确原prompt、8份旧缺失记录；10分镜继续封存，4工具保留。新增`assets/PROVENANCE-CURATED-20261001.json`记录实际工具输入、三种资产尺寸／SHA256，十张新图无参考输入，首页人像沿用既有参考来源与未核实权属边界。打包测试确认新WebP字节相同、被删WebP与来源JSON／PNG均不公开。
- 定向测试覆盖SynaRoute上下文与普通密钥页、权限／空目录与用途切换、47图尺寸／来源／提示词／被删旧链接、真实公开打包；Node106＋Vue80、类型／限定lint／构建通过。浏览器滚动／窄屏／主题／详情／选择流程见`output/studio-refinement-20261001`；密钥与模型仅夹具模拟，未调用生成模型。交接与自有产品说明同步，未提交推送、重建8080或部署生产；前节765e5a9旧发布门禁仍有效。

## 2026-10-01 家族产品限定发布验证

- 功能765e5a9516eaecf23d94202b4afc48f9bf723339已推送，GHCR sha-765e5a9成功、摘要341886529ae61b817a24da3a2a40dbb7062aabbc2148ecc9b7956153cb422e9d；安全门禁失败与Go errcheck八项失败已取证，不表示可安全部署。x/image的WebP／VP8L可达漏洞含本轮投稿路径，Axios七个high尚未处理；没有关闭扫描或增加安全例外。后续修复待办／授权与完整链接见HANDOFF。
- 用户授权提交并推送本线程产品代码。按已有各节登记包含一体化静态白名单／Docker打包、正文工作区／默认浅色与宿主主题、配方同页配置、工坊38图／对应prompt／瀑布流、人工投稿审核及视频／分镜待开放；包号专项和更新脚本／CI接线保持未暂存，共享theme与UPDATE_GUIDE只提交产品段落。
- 实际暂存快照通过Node104＋原站前端229项、类型／定向lint／Vite构建；证据output/family-release-20261001。不沿用混合工作树编译或旧SHA的远端结果冒充本次发布验证；新SHA的CI／GHCR需推送后核对，生产由用户更新。

## 2026-10-01 工坊视频与分镜待开放（未部署）

- 自有product-samples/magic-studio/core.mjs新增VIDEO_CONTENT_ENABLED=false与isItemAvailable，统一隐藏视频作品／分镜，原目录不删；app.mjs关闭筛选、搜索／收藏／历史、玩法分镜步骤、旧hash及公开投稿选择，保留图片和4份官方工具指引。收藏ID仍存储，仅可见数量排除暂关闭项。
- index.html增加常驻“视频与分镜待开放”及使用说明，theme.css使用真实disabled、muted与不允许光标；局部opacity=1避免叠加全局忙碌样式导致提示过淡。保持瀑布流、完整画幅、对应prompt及宿主主题，不新增依赖或上游接缝。
- 自有StudioSubmissionsView.vue禁用视频新投稿选项，提交／旧视频重投前端阻断并提示；保留图片投稿和旧记录私有查看／撤回／管理员管理，api.ts及服务端MP4协议不变。不能将UI关闭描述为后端访问控制或已上线。
- 新增core／UI与Vue边界回归；Node104＋Vue11、类型／限定lint／Vite构建通过。4185真实验证待开放、搜索、旧链接、玩法、图片详情与深浅手机；24个并行包号文件SHA256保持。证据output/studio-video-paused-20261001；恢复成片／提示词／授权／播放条件已记HANDOFF五，未更新8080、提交推送或部署生产。

## 2026-10-01 更新脚本旧镜像保留（未发布）

- 自有 `deploy/update.sh` 新增仅本仓 `sha-*`／`latest` 与来源标签悬空镜像的定向清理；当前镜像与实际更新前镜像按完整ID保护，同版重复运行保留回退版，无历史记录时取创建时间较早的最近sha版本作候选。增加本地update-before／rollback-previous标签以保护latest旧版，所有运行／停止容器引用镜像均跳过，不强制删除、不做全局prune、卷或备份清理。
- 健康检查超时改为失败退出且不清理；检查错误保留镜像并显示原因，增加同部署目录并发锁。宿主机脚本需单独替换，应用镜像pull不更新该文件；本地保护标签不属于GHCR发布标签。
- 新增 `deploy/tests/update-test.sh`、`deploy/tests/fixtures/update-docker.sh` 的22场景假Docker回归；上游接缝 `.github/workflows/backend-ci.yml` 的shell job新增带[CUSTOM]的语法和假Docker测试步骤。UPDATE_GUIDE同步保留／失败／锁／服务器替换规则。
- Git Bash语法检查与22场景回归通过，不操作真实Docker／生产、不编译后端、不混入并行未提交工作；本轮未提交推送。

## 2026-10-01 工坊瀑布流消除齐行空白（未部署）

- 仅工坊自有app/core/catalog/theme与定向测试：新增纯masonryPositions最短列算法及1px Grid位置；保留DOM与顶部阅读顺序、完整contain原画幅、已有雾青主题、提示词／收藏／复用行为。38张源PNG尺寸写入artSizes预留，逐一测试核对；不改图片像素或素材内容。
- CSS变量负责2／3／4／5列，降列先归位再测高，避免旧位置产生隐式空列。ResizeObserver合并重排、过滤容器仅高度变化；媒体加载／错误／元数据与resize回退，筛选断开旧观察、玩法隐藏／pagehide清理、pageshow恢复。无观察器或测高不可用仍保留可读Grid，不加依赖、不改宿主／后端白名单或认证。
- Node101（工坊41＋配方56＋宿主1＋打包3）、语法／限定diff检查、Vite构建通过；真实320／390／768／1280／2200、详情焦点／分镜／空态检查，桌面同列间距18–19px、手机12–13px、横溢出0、warn/error空。初轮缩窄产生隐式列问题已修，并增加回归。技能扫描器exit127，不宣称扫描通过。证据output/studio-masonry-20261001；24个包号并行文件SHA256保持。仅4185预览，未更新8080、提交推送、Docker／Go或生产部署。
- README／PRODUCT同步OpenAI只读官方查证：Videos API及Sora 2于2026-09-24关闭，没有一对一替代API；历史文档不代表可调用。本轮没有接视频服务或收费调用，本站第三方视频能力未知，10套分镜仍0段对应成片。

## 2026-10-01 工坊风格多样化与真实投稿审核（未部署）

- 工坊new-gallery新增六种明显不同媒材的真图（像素／水墨／美漫／水彩／剪纸／复古未来）与实际prompt；catalog题材／风格独立，core保持原始记录与反推区别、严格同源公开媒体适配和规范ID收藏，app/index/theme加入风格筛选、投稿入口、匿名公开目录、成片播放器与prompt详情。38图／10套无成片分镜／4工具，准确原始记录30图；PROVENANCE-STYLES记录资产与工具，未输入参考图，完整图无裁切。打包脚本仍只发白名单运行资源，不发PNG／来源JSON，无新依赖。
- 新增后端handler/custom_studio_handler.go与测试、middleware/custom_studio_audit_test.go，单实例私有JSON／media存储：原API用户上传、作者／管理员读取、管理员确认发布、拒绝／下架／撤回、期望状态409、Range、文件签名／尺寸／体积／额度。公开目录只返回published的作品／署名／prompt等公开字段。没有新数据库迁移、自动审核、收费模型、采集或定时任务。
- **上游接缝** backend/internal/server/router.go：RegisterPageRoutes之后新增[CUSTOM] RegisterStudioRoutes，复用原JWT／管理员认证、面板限流、合规与审计，dataDir仍来自Pricing配置。同步上游时保留单次注册，不能把私有素材映射成匿名静态目录。
- **上游接缝** backend/internal/server/middleware/audit_log.go：auditBodyOmittedRoutes精确加入投稿及admin审核两条路由[CUSTOM]，不记录私有文件／提示词正文，仍记录操作者、路径和结果。同步后运行TestStudioAuditOmitsPrivateBody，勿为省略正文关闭整个审计。
- 定制web/custom_family_products.go只把工坊connect-src从none改为self（[CUSTOM]），用于匿名审核后公开目录；继续禁止外部连接、脚本／图片／媒体限定同源、静态白名单不扩为私有目录。static回归保留CSP和范围读取；不是跨域SSO或模型API放行。
- 新增custom/studio/api.ts、StudioSubmissionsView.vue及两份定向测试，customRoutes增加用户／管理员两路由；FamilyWorkspaceView宿主只新增origin/source/nonce保护的studio-submit当前路由动作，不传JWT。theme.css新增mofa-studio-contribute命名空间，默认浅色／跟随主站深浅、完整媒体contain、手机单列；普通预览仍可后期拆分模块。
- 投稿需要授权、完整prompt及模型说明，original／reference明确区分；初始pending不会公开。管理员看片确认及原因、作者撤回、修改重新审核；文件选择只本地预览，点击提交才上传，账号切换清理私人输入／blob／迟到响应。无永久清理UI，撤回仍占额度；只支持单实例，备份含pricing.data_dir/studio-submissions，见工坊COMMUNITY及FAMILY_INTEGRATED。
- 证据：Node96／Vue64定向、类型／lint／Vite构建，隔离Docker投稿／审计5顶层与静态回归／Go embed成功；web全量两favicon旧用例失败未修，不标全量通过。4185六图／风格／桌面浅色／手机深色／完整prompt／匿名登录回跳实测，私有投稿／审核真实账号仍待。24并行包号文件保持；未更新8080或生产、提交推送、调用视频或付费模型，准确未验和部署门槛见HANDOFF。

## 2026-10-01 工坊再增十张原创图（最新目录）

- 内容增量，不重设计界面：new-gallery.mjs新增10份独立完整实际prompt与变量、共用recordedImage映射；catalog.mjs优先展示最新十图，32张图＋10套视频分镜＋4份Skills，共46条。index.html使用说明数量、README／PRODUCT／DESIGN同步，旧22图不撤换。视频仍0段新成片，明确分镜与影片的区别。
- 内置image_gen实际生成，未传用户参考截图；成年人物25岁以上、不复刻现有角色。PNG原图留档、无裁切WebP和缩略图进入运行包；assets/PROVENANCE-EXPANDED-20261001.json记录逐字实际prompt、工具、生成文件名、尺寸／字节数／SHA256。共24张准确原始记录，8张旧摘要图仍明确缺失，不以另写模板伪造原始prompt。
- core／ui新增十图资源摘要、尺寸比例、对应prompt、替换和两种复制回归；custom/family/build-products.test.mjs逐字节核对十对网页资源、不发布原PNG或JSON。工坊30＋配方56＋宿主1＋打包3共90项与Vite生产构建通过，4185十图／原始prompt、1280浅色／390深色contain及无横溢出通过、warn/error为空，详见HANDOFF与output/studio-images-more-20261001。
- 不新增运行时module或依赖，不改后端白名单／CSP、主站主题、宿主认证、配方或SynaRoute；24个包号并行文件SHA256保持。未收费视频调用、Go编译、Docker重建、更新8080、提交推送或生产发布，准确边界和视频待办在HANDOFF。

## 2026-10-01 工坊鲜活图像、作品与提示词详情（视频成片待补）

- 仅工坊源码／素材与家族打包回归：撤下四段Blender视频及其封面；new-gallery.mjs新增六张原创图的实际完整prompt和十套原创分镜，catalog.mjs同步目录／玩法。22张图、10套分镜、4份Skills；十套分镜没有可播放成片，不冒充完成视频需求。
- index.html／app.mjs／theme.css加入作品与提示词并排详情、卡片复制／复用、原始提示词记录与可替换模板区分。14张原始记录与两个PROVENANCE JSON逐字对应，8张旧图缺失原始记录明确标注；右侧材料只留内存，图片索引原画幅及完整contain，44px操作、手机堆叠、关闭焦点、重开归零与弹窗内手动复制。
- core.mjs导出保留当前材料模板与可用原始记录、条件和状态；core／ui回归更新，build-products.test.mjs确认0个mp4和无film-*、新图打包。不改上游运行时文件、安全白名单／CSP、宿主认证／主题、配方或母品牌首屏，不新增依赖或付费生成接口。
- 六张PNG留档、无裁切网页WebP及缩略图，精确prompt／工具／无参考图记录在assets/PROVENANCE-VIBRANT-20261001.json；旧影片来源文档只保留撤下历史。没有视频生成工具，第三方目标题材媒体授权未核实，待用户自己的接口／预算或授权素材，不下载抖音或复刻剧集人物。
- 本轮88项Node／jsdom／宿主／打包及Vite生产构建通过，实际4185桌面／手机、原始与复用复制提示／材料隔离／完整图像／分镜／Esc焦点通过。准确证据与未验见HANDOFF最新条目；未Go编译、Docker重建、提交推送或更新8080／生产，包号并行工作保留。

## 2026-10-01 控制台内子产品工作区、主题与作品扩充（最新源码）

- 用户新要求覆盖历史默认深色：未保存主题时默认浅色，已保存深色仍保留；内置子产品跟随宿主根节点主题。新增 `FamilyWorkspaceView.vue`，受原认证守卫保护的 `/tools/recipes`、`/tools/studio` 在 `AppLayout` 右侧正文加载同源静态产品。`familyProducts` 内置入口使用 RouterLink，显式独立外部地址仍有安全新窗口兼容；SynaRoute 不改。
- 新增 `family/workspace.ts` 与 `product-samples/family-runtime/host-client.js`，以来源、确切 iframe 窗口与每次加载的随机 nonce 配对。宿主读取本人既有密钥，只回传目录元数据，用户选择模型并点击应用后仅回传所选 Key；复查权限、轮换、模型目录时效及用途。退出、切换与卸载取消未完成读取并清空连接；不自动创建密钥、调用生成模型或把 JWT 放进 URL／消息。它是受信同源模块，不是隔离不可信脚本的安全沙箱。
- 上游接缝逐处登记并有 `[CUSTOM]`：`main.ts`、`AppSidebar.vue`、`HomeView.vue`、`KeyUsageView.vue` 的无偏好主题初始化统一浅色；`vite.config.ts` 新增开发入口适配器 import／插件；`Dockerfile` 新增共享 family-runtime COPY。`custom/theme.css` 仅新增 `.mofa-product-workspace` 专属布局；原主题颜色与并行包号样式不回滚。
- 新增 `dev-products.mjs`／`d.mts` 与6项回归：开发服务器仅将 GET／HEAD 的两产品根目录映射到其静态首页，补齐尾斜杠并保留宿主参数；不改 API、文件路径或其他路由。解决目录入口误落主站登录页；本机预览显式 `--config vite.config.ts`，避免旧的 gitignored 编译配置优先加载。
- 配方新增同页 Key／模型／协议选择与手填兼容，只有内嵌宿主模式启用直接目录选择；独立静态模式保留原登录选择窗口。材料／已选连接只留页面内存，应用配置本身不运行模型。子产品主题开关内嵌时隐藏，由主站统一控制。
- 工坊扩为16张原创图片＋4段真实授权原片＋4份 Skills／工作流，新增8张原创图包含年轻成年男女、穿搭、Cosplay、动漫、原创机甲、动物、电商。所有新人物明确成年人，未把用户截图作为生成输入。左侧点击直接大图／原生影片播放，右侧 `contain` 完整预览，统一网格与横向用途筛选；视频不再使用图片二维运镜，不将影片冒充本站生成或模板实测。影片出处／署名／CC BY 授权及处理记录在 `assets/PROVENANCE-20261001.md` 和每个视频条目。
- Go静态处理仅两产品安全头从 DENY 改 SAMEORIGIN、`frame-ancestors 'self'`，允许同源工作区，未放宽主站全局策略。工坊增加两个运行模块和严格 mp4 白名单／MIME，支持原生 Range 请求；原始 PNG、来源文档、测试和服务脚本仍不公开。打包器只复制运行时资源，不新增服务、SDK或调度。
- 169项自动测试（Node／打包83、前端及国际化86）、类型／相关 lint／生产 Vite 构建通过；Docker内 TestFamily 两函数回归与 Go embed 编译通过。本轮未提交推送／生产发布；本机8080仍旧版，真实登录选择、用户模型／计费及三条抖音完整观看未验。源码／镜像与截图准确状态见 HANDOFF 最新条目，不用历史浏览器或 CI 结果代替。

## 2026-10-01 配方与工坊同站点一体化（本机8080已部署，未发布生产）

- 新增 frontend/src/custom/family/build-products.mjs／d.mts／test，将权威配方生成外置脚本／样式版本、工坊仅复制运行模块／WebP／品牌PNG到两个带标记的 public 目录。清理前检查绝对目录为指定根的直接子目录且带本产品标记，拒绝覆盖用户文件；生成目录 gitignore，原独立应用仍可使用。
- 上游接缝逐处登记，均有 [CUSTOM]：frontend/vite.config.ts 的打包器 import 和插件 config 钩子；Dockerfile 的两项 product-samples COPY 和嵌入路由测试 RUN；.dockerignore 的配方 Markdown／web embed 测试例外与生成目录、原始工坊PNG、本机 output／配置编译产物排除；.gitignore 的两个生成目录；backend/internal/web/embed_on.go 的有设置／旧版两条 serveFamilyProduct 分流，不经过 SPA 兜底或本机 data/public 覆盖。
- 新增 Go custom_family_products.go／test：两个专属命名空间，GET／HEAD、无斜杠308、明确JS MIME，未知／穿越／源文件404、其他方法405、无目录列表、no-store／nosniff／DENY／no-referrer及不允许内联执行的产品 CSP；不改 API 路由、认证／计费或全站 CSP，不新增依赖。Docker 构建执行 go test -tags embed -run '^TestFamily' -v ./internal/web 限定本轮回归后才生成镜像，测试源码不复制到最终运行镜像；首次完整 web 测试因旧用例引用已不存在的 logo.png 失败，不修改其用例或宣称全量通过，原失败日志保留。
- 定制 products.ts 默认同源 /recipes/、/studio/ 并显示本站产品，显式独立覆盖与错误地址禁用保留；配方 builder 内置模式外置JS／CSS和构建标记，model-ui固定当前站点选择窗口／家族首页、不信 api_site；connector 只在受信内置模式允许精确当前来源，其余白名单／窗口／nonce／期限／用途配对不变。工坊只增加安全新窗口家族链接，维持原设计与模型边界。
- 同步 README／产品部署说明、FAMILY_PORTAL 与 FAMILY_LOCAL_DOCKER，新增 FAMILY_INTEGRATED 和不含凭据的 Compose 构建覆盖样例。177项自动测试及类型／lint／Vite通过，Docker内TestFamily两函数／14路径子用例及两种embed入口通过、Go embed编译通过；原完整web测试的旧logo.png夹具失败另存，不标全量通过。
- 用户选择更新8080后仅应用替换，镜像c599729e28df与容器一致，三服务healthy、25项HTTP检查通过、原DB／Redis ID／数据卷／业务计数及280项设置摘要保持。浏览器实际配方整理／同源配置字段、工坊内容／图片／复制提示、1280桌面及390手机无横溢出、双主题通过；原站登录回跳参数可见，IAB配置弹窗及真实账号授权／付费模型仍未验。准确证据与备份见HANDOFF及output/family-integrated-20261001，不提交推送或部署生产，24个包号并行文件SHA256保留。

## 2026-09-30 家族视觉与公开链接发布收口（最新授权）

- **最新发布状态**：本节功能已推送 origin/main，提交 25071d4d547c69d6c933bbcb0c276585b23b4b68，push 与远端 SHA 一致。GHCR 36736764218 和 CI 36736764231 收工时仍运行；安全扫描 36736764322、CI 前端／Go lint／shell／release-helpers 已成功，Go Unit 与镜像 Build and push 仍进行，不标全部通过或镜像已发布。准确终态见 HANDOFF 和本轮取证目录，生产未操作；下方“未提交／未发布”均为实施时历史。

- 用户再次授权提交推送，范围仅为下方登记的工坊三图／最新参考人像、家族皇冠图标、SynaRoute 公开链接及配置说明。AppSidebar 的新增图标 import 与原家族入口标记均保留 [CUSTOM]；视觉仍只在 theme.css。包号运行时及三条专用样式不混入本轮暂存版本，不新增依赖、模型调用、认证共享方案或一体化托管。
- 本轮验证、远端代码与镜像终态以 HANDOFF 最新记录和 output/family-polish-release-20260930 为准；下方“未发布”文字保留为实施时历史，不沿用旧 SHA 的 CI 结果，不部署生产。
- 实际限定暂存快照 72 项／8 文件及类型、相关 lint、Vite 构建通过；仅 16 个家族相关文件，24 个并行文件内容未变，三条包号样式未暂存。远端 CI／GHCR 尚须按功能提交核验，不把本地前端构建当后端或真实模型验收。

## 2026-09-30 SynaRoute 公开官网与下载链接（本地未发布）

- 新增custom/family/synarouteLinks.ts，官方地址来源为SynaRoute真实工程的README、site/config及download路由，集中https://synaroute.mofamilys.com与/zh/download；FamilyHomeView仅在SynaRoute面板的原配置动作下加入两个公开a链接，匿名也可点击，新窗口／noopener noreferrer／不带凭据，下载入口仅打开官网下载页，不自动下载或安装。KeysView、导入生成器、认证与其他产品行为保持，无新增上游接缝或依赖；视觉只在theme.css，继承紫色、44px操作高度及原焦点规则。
- family-home.spec.ts新增匿名／登录态、精确href／新窗口安全属性、无download属性、无凭据访问及无自动fetch／open回归。69项／7文件及相关lint／类型／Vite通过，日志output/synaroute-links-20260930；4184桌面／手机无横向溢出、两链接44px、Tab可见焦点及截图留证。官网在线访问超时，不能宣称实时可达、安装包或真实客户端导入通过；本轮不新增安装检测，既有100ms焦点启发式并非可靠安装判断。未提交推送、重建8080或部署生产，保留此前视觉及包号并行修改。

## 2026-09-30 工坊三图、深发参考人像与等宽放大（本地未发布，覆盖旧两图人像）

- FamilyProductVisual.vue的工坊分支保留原玻璃瓶／狐狸副本并新增左人像，三张均为原生button，单选aria-pressed及再次点按恢复，不跳转／运行模型。新增custom/assets/studio-reference-portrait.png／webp／provenance.json，内置image_gen参考用户最新深发色室内图片生成虚构成年女性，使用分体吊带／细条纹衬衫／牛仔裤，不使用裙装、摸头或截图边框；精确prompt与来源边界保存，参考权属未核实，不代表真人代言或模板实测。WebP113416字节，原PNG保留；被否决候选移入ignored output，原瓶／狐副本SHA256与工坊thumb逐一一致。caption改为“AI 视觉示例 · 非模板实测”，alt不再虚构欧美国籍；SURFACE记录最新要求。
- 所有视觉只在custom/theme.css：工坊文案／图区40:60，三图等宽4:5正常流Grid、轻微±2°／±8px错落、手机去除额外最小高度；悬停抬升／回正／1.1倍且z20，点按z10、键盘焦点z30，reduced-motion关闭过渡。没有新上游运行时接缝／依赖／业务配置，不触碰原首屏／入口／登录或并行包号样式。同步上游后验证三图默认无遮挡、等宽、键盘与手机交互，不把旧叠图效果当最新状态。
- family-home新增三图素材及按钮回归，upgrade-contract新增等宽正常流／工坊列比例／焦点层级／减少动画断言；57项／6文件、相关ESLint、类型及最新Vite构建通过。4184实际浏览器1280桌面／390手机三图加载、默认相交面积0、无横向溢出，三张实际悬停z20、键盘z30及手机尺寸点按true→false，非触摸实机。日志与browser-evidence-final.json在output/studio-reference-20260930，页面截图studio-balanced-{desktop,mobile}-20260930.png；源码预览公开设置回退不当品牌或生产验收。未提交推送、重建8080或部署生产。

## 2026-09-30 家族首页侧栏皇冠图标（未发布）

- 新增custom/components/FamilyHomeIcon.vue：简洁皇冠SVG，1.5px描边、currentColor、aria-hidden，继承原菜单20px／16px尺寸；不新增主题规则或依赖。
- 上游接缝AppSidebar.vue逐处登记：新增带[CUSTOM]的FamilyHomeIcon import；既有[CUSTOM]家族入口仅由DashboardIcon改为FamilyHomeIcon。共享buildSelfNavItems同时作用用户主菜单和管理员个人区，保留两类仪表盘四宫格、导航行为及认证。同步核对见custom/UPGRADE.md。
- upgrade-contract.spec.ts新增家族／用户仪表盘／管理员仪表盘图标区分及SVG语义断言，55项／6文件、相关lint和类型通过，日志output/family-icon-20260930；未真实账号侧栏视觉验收、提交推送、重建8080或部署生产，保留既有人像及包号并行修改。

## 2026-09-30 工坊女性人像与地址配置说明（未发布）

- FamilyProductVisual.vue仅替换工坊左侧图例，新增custom/assets/studio-rain-portrait.webp为已有rain-portrait-thumb.webp原样副本，虚构成年女性；右侧狐狸、非模板实测caption、所有入口／认证／原首屏保持不变，无主题改动或新上游接缝。来源见magic-studio/assets/PROVENANCE.json、family/SURFACE.md；未新调用模型或修改图片像素。
- family-home.spec.ts补人像src、alt、尺寸与懒加载断言；44项／5文件回归、相关lint、类型与Vite通过。4184源码预览1280桌面／390手机两图已加载且无横溢出，证据output/studio-portrait-20260930及本轮截图；公开设置请求失败导致首屏回退，不能作为原品牌配置或生产验收。
- FAMILY_PORTAL.md补充GitHub仓库Settings → Secrets and variables → Actions → Variables具体配置、三项MAGIC_*公开变量示例与重建流程，明确不是后台或旧容器运行时变量；不含密钥、不扩大CORS、不自动托管子产品。未修改实际变量／DNS、提交推送、重建8080或部署生产，保留包号并行改动。

- **本轮镜像发布补充**：GHCR36712632118现success，实际标签sha-321fe99／摘要c8dd5b34edacf29eec9e30f53f35f5de8c5898706b86715b81d70b8a73db4147；Go Unit已成功，CI36712632123集成仍运行，不标全绿。此条覆盖下方“镜像运行中”，准确取证在HANDOFF与output/family-release-20260930；后续skip-ci文档不生成新镜像，生产未操作。

- **2026-09-30 产品已推送321fe99**：用户授权的家族首页、配方／工坊与接入教程已提交origin/main，141文件，独立暂存快照285项与类型／lint／build通过。包号专项前后端及三个专用主题规则仍在工作区，不混入产品发布。GHCR36712632118与CI36712632123仍运行，安全扫描及frontend／Go lint／shell／release-helpers成功，不标镜像可用或全绿；终态以HANDOFF及本轮记录为准。三项MAGIC_*仓库Variables未设置，独立产品另行托管／配置并重建，当前生产入口禁用；不自动部署生产。

## 2026-09-30 用户再次否决旧卡片后优化产品入口

- 保留原BrandPanel及账号／授权边界；FamilyHomeView改为tablist＋四个tabpanel的互动展台，默认配方／已知锚点，支持左右键／Home／End／循环焦点和手机两列。每个面板仍只有真实登录或打开动作，不在预览时调用API或弹新窗口，说明详情可展开。
- 新增custom/components/FamilyProductMark.vue，复用原M以及专属几何图标；FamilyProductVisual重排配方结构示意／既有原创图例／接入与路由关系，caption及alt保留。FamilyProductAction加箭头，FamilyProductSwitcher换产品标识，保留原站API当前态与独立地址禁用规则；所有新增视觉仍只在custom/theme.css，无本轮新增上游业务接缝。
- family/SURFACE.md记录被否决的旧版与本轮局部方向；family-home增加4项展台与键盘回归，family-header增加可选夹具目录／原M数据嵌入及完整实际公告、语言、订阅组件装配回归（数据／账户／国际化模拟，无凭据）。三项收尾修复：配方浮签／caption正常流，工坊独立图片区与caption分离，页头右侧max-content保宽／不足第二行。252项回归、类型／lint／构建及最终Docker94dba406证据见HANDOFF与output/family-deck-20260930。独立Verdict三项resolved、ship仅限修复评分，不等于用户视觉或真实登录批准；继承检查No changes，未重写母品牌、未宣称检测器通过。未提交推送或生产部署。

## 2026-09-30 用户纠正品牌首页与控制台产品切换（最新方向）

- 最终本机8080运行d02c6ac5689c，四容器healthy、数据／设置不变；317项原站＋54项配方定向通过。独立审查五项P2修复复核均resolved，ship只限修复清单；真实登录授权与用户视觉待验，未提交推送／生产部署。准确镜像及备份见HANDOFF，不能沿用下方初版状态。
- 文档继承检查为No changes，既有母品牌设计系统未重写；只读范围及结论归档output/family-layout-20260930/documentation-check.md。检测引擎不可用，不登记检测通过。

- 用户否决账户摘要首页。FamilyHomeView恢复原BrandPanel狮冠M、文案与原三条CTA，下方通栏API＋配方／工坊／SynaRoute展示；PublicLayout普通登录／注册恢复原默认API控制台，匿名产品仅携带站内/dashboard?product=id，不自动跳外部。新增FamilyProductAction／Visual／Switcher、family/PRODUCT.md／SURFACE.md、header／switcher定向测试；所有视觉仍只改custom/theme.css。
- 上游新接缝逐处：AppHeader.vue模板仅在user存在时装配FamilyProductSwitcher，script新增import，均标[CUSTOM]；原账户、公告、移动菜单保留。Dockerfile在既有[CUSTOM]ARG与RUN追加VITE_MAGIC_STUDIO_URL；custom-image.yml在既有[CUSTOM]build-args追加MAGIC_STUDIO_URL Variables公开配置，不含凭据。保留原router、AppSidebar与登录处理，不增加后端接口或账号系统。
- products.ts为工坊增加与配方一致的URL安全校验；生产没有地址则禁用，本机DEV默认4179，非DEV只接受显式双方回环HTTP。工坊首版已完成，入口不等于通用SSO／统一退出。新增studio-jade-bottle.webp／studio-paper-fox.webp是原工坊thumb资产的原样副本，来源见其PROVENANCE.json及SURFACE.md，没有新增模型请求。
- 本机Docker补充：新增配方site/Dockerfile、deploy/docker-compose.family-local.yml.example／FAMILY_LOCAL_DOCKER.md，serve.mjs仅增加容器HOST／PORT读取。配方容器非root／只读／无能力／仅回环4178，精确4178两来源CORS；原库／Redis／app_data保留。叠加层显式配置4178配方及4179工坊预览，工坊不是本轮Docker容器。实际镜像、备份、验证与未验项以HANDOFF最新四／五为准，不将下方初版“未部署／工坊仅规划”当现状。没有提交推送或生产部署。

## 2026-09-30 同批成员多专属分组（未发布，覆盖严格单组历史）

- 后续现场反馈：用户明确两位成员已选全仍失败，不能把漏选作为本次结论。只读线上公开资源证实管理页尚无本轮config_issue／expected_updated_at；已发布HEAD中的SQL确实拒绝多组账号，不依赖成员数量。具体线上后端SHA与数据未取得，修正仍未发布；证据及待办见HANDOFF最新状态，本次只补文档，不新增运行时接缝。

- 用户明确允许账号多个专属组并确认同批成员使用；一账号一条有效包号保留，241／242唯一性及历史隔离标记不改。新增backend/internal/service/custom_dedicated_groups.go：关联组范围、实际Key组授权、认证阶段关联绑定／历史标记查询；custom_dedicated_accounts.go的结构查询不再拒绝所有其他组，而验证关联组为同平台标准专属、单账号、无备用路由及名单外授权／订阅，保存检查关联组非成员未停用Key。Check在原始用户／Key组身份下核验实际组，认证、选号／转发、WS原接缝复用；不会只让保存成功却令其他组请求403。
- custom_dedicated_members.go移除成员在原Save事务清理原账号关联专属组授权并停用Key，保留无关组／余额；相关mock及PG回归同步。具体诊断将合法多组改为通过，非法关联组报告account_unsafe_groups或related_group_other_accounts，并提示同批共用者加入同一条包号。config.ts、copy.ts及dedicated-view.spec.ts同步中英文规则与失败文案。均属custom文件／新增文件，无新增上游运行时接缝、样式或迁移。
- 46项前端（含原授权弹窗4项）／相关lint／类型，Docker包号service／handler定向及embed编译通过；隔离PG30个保存场景及原多人回归通过，覆盖同账号多组的认证／选号、拒绝非成员和错组、移除／撤销／删除／重分配与历史隔离。日志output/dedicated-config-20260930/{multi-build,multi-postgres}.log，生产未取证，未提交推送部署。截图两位同组只选一位的失败已隔离复现，不能据此认定线上本次payload；还需确定用户指授权弹窗还是包号绑定保存。

## 2026-09-30 包号配置400具体诊断（未发布）

- 新增backend/internal/service/custom_dedicated_config.go及config_test／config_postgres_test：保留原完整性查询，失败后仅在保存事务内诊断首个失败项，白名单19类检查＋最多10项编号，不读取缓存，不更改权限。custom_dedicated_accounts.go的Save返回具体诊断；custom_dedicated_handler.go错误转换保留原错误及元数据，不再以通用常量覆盖。仅管理员保存路径返回内部资源编号，普通用户原不可用响应保持。两处均为既有custom新增文件，不新增上游业务接缝。
- 新增frontend/src/custom/dedicated/config.ts及dedicated-config.spec.ts；AdminDedicatedAccountsView.showError按已知诊断键显示中英文原因／编号，非法原始字段不展示，旧版响应回退原提示。dedicated-view增加保存失败回归，后端原mock及handler测试同步。没有新增样式、迁移或外部依赖。
- 42项前端定向测试、相关ESLint、vue-tsc，Docker包号service／handler测试及embed编译通过；临时PostgreSQL真实23个新增保存场景＋原多人测试通过。日志output/dedicated-config-20260930/{build,postgres}.log。生产未取证，不能认定用户本次根因；需要请求账号／组／成员编号，用户只确认发生在线上。未提交推送部署，不包含其他线程未提交内容的发布授权。

## 2026-09-30 魔法家族统一首页（源码完成，未发布）

- 新增custom/family/products.ts、views/FamilyHomeView.vue与ApiLandingView.vue、FAMILY_PORTAL.md、family-products.spec.ts及family-home.spec.ts。BrandHomeView改为家族装配壳，原API内容保留/api；custom/routes注册/family与/api，PublicLayout导航增加家族与API入口，family登录／注册带站内回跳，不改原登录核心。仅custom/theme.css增加mofa-family作用域，无新依赖或后端业务修改。
- 上游接缝逐处登记：router/index.ts公开页等待名单新增/family、/api（保留后台模式限制）；AppSidebar.vue个人区新增家族首页（用户与管理员共用，不代替原菜单）；Dockerfile前端stage新增两个ARG和RUN注入；.github/workflows/custom-image.yml的build-args新增仓库公开Variables MAGIC_RECIPES_URL／MAGIC_RECIPES_ORIGIN。均有[CUSTOM]标记。实际变量／域名未设置、Docker／CI未运行，不等于镜像已发布。
- 产品地址仅已配置HTTPS或同源相对路径，不允许凭据／参数／片段／协议相对／通配；生产空地址禁用配方，本机DEV才明确4178预览。配方model-ui.js接受经过原白名单验证的api_site来源提示，template.html新增底部新窗口家族入口、独立theme.css仅补返回链接可见性；不导出JWT／Key、不自动打开产品／授权或运行模型，回到首页不丢当前材料。ui.test.cjs增加允许／恶意来源与安全链接断言。
- SynaRoute走已有/keys导入，不假冒桌面SSO；工坊禁用规划入口，无通用跨产品账号或单点退出。UPGRADE、site/README与HANDOFF四／五同步最新接线事实，覆盖旧咨询“无构建ARG”的结论，但保留历史记录与其他任务改动。
- 237项前端定向＋54项Node（合计291）及类型／相关lint／Vite生产构建、workflow YAML解析通过，额外覆盖公开来源不得自动预填本机登录地址，本机hint仅限本机HTTP预览。浏览器首页1280深色／390浅色、匿名回跳、API介绍与配方预填／取消及390登录控件无外层溢出、两页错误日志空；证据output/family-20260930-*.log及本线程family-home／family-recipes截图。配方最终136797字节，4178 session86473；4175原Vite保持，8080与生产未更新、未提交推送。真实账号／Key导入闭环、CORS、独立产品SSO及收费运行未验，不用模拟登录或静态构建替代。

## 2026-09-30 魔法工坊独立本地产品（定时任务关闭）

- 用户后续授权先做产品、定时先不启用；新增product-samples/magic-studio内的PRODUCT／SURFACE／README／DESIGN与.impeccable/design.json、package／静态页面／独立theme.css、catalog／core／icons／app模块、Node静态服务与core／ui测试。16条资源与4个玩法覆盖图片／视频、Skills·工作流；复用变量、中英模板、复制／TXT／脱敏ID分享、收藏／历史／主题、大图与二维运镜。设计归档仅实际token／断点／组件及品牌继承，不重定母品牌。没有上游运行时接缝、生成API或自动搜索／调度，现有产品不重设计。
- assets包含8张内置image_gen原创PNG、对应full／thumb WebP及既有狮冠M；PROVENANCE.json保存精确prompt／来源，prepare-assets.py嵌入元数据并压缩，verify-assets.py只读检查25张位图来源通过。外部方法保留Google／Runway／ComfyUI／Remotion官方出处；提示词未逐条目标模型实测，Skill未安装执行、视频不是生成视频。
- 21项Node／jsdom／HTTP定向通过，浏览器多尺寸与真实复制／TXT下载／ID分享／收藏重开等证据在该目录.impeccable/review。设计扫描器不可用，不冒充扫描通过；独立审查首轮三项P2已修（资源类型跨区匹配／手机返回保留材料／dialog名称），同审查者Verdict Pass为ship、三项均resolved，结论仅覆盖这三项修复，记录finish-review.md／finish-verdict.md。更新HANDOFF四／五及CLAUDE最新授权，PLAN保留历史规划并标注首版已实施。仅本机4179，未提交推送／Docker／生产部署，保留并行修改。

## 2026-09-30 魔法配方部署说明（文档）

- 导航核查补充：源码官网导航／首页与控制台侧栏尚未添加到配方首页的产品入口，/connect/recipes仅为反向登录选择页；建议待用户确认最终地址与新增入口后再实施。本轮仅更新HANDOFF现状／待办，没有导航或运行时改动；线上自定义菜单未读取，不以源码检查替代线上配置取证。

- 更新product-samples/magic-recipes/site/README.md发布方案：单文件静态托管与原站登录选择页分别发布，独立来源构建许可＋API CORS两层配置，备份／验收与同源路径备选边界。核实当前Docker／CI没有VITE_MAGIC_RECIPES_ORIGIN构建接线且.env.*排除，记录待实施；没有修改Dockerfile、workflow、原API配置或服务器，也没有部署／提交推送。
- HANDOFF四／五记录咨询与待确认域名、站点管理工具；官方Nginx／Vite／MDN资料用于部署取证，不代表实际服务器配置已验证。没有新增运行时定制或上游接缝，既有模型测试结果未重跑也未改写。

## 2026-09-30 魔法配方复用魔法 API 登录选择（源码完成）

- 新增frontend/src/custom/recipes/connect.ts、custom/views/RecipeConnectView.vue、custom/__tests__/recipe-connect.spec.ts及recipe-connect-view.spec.ts；custom/routes.ts注册/connect/recipes并要求登录。只用原keysAPI获取本人现有密钥，所选Key只读GET /models；原登录、账号密码／JWT、所有后端权限及计费逻辑保持。不自动创建／改绑Key或扩大权限，无新增上游接缝。原站新增样式只在custom/theme.css的mofa-recipe-connect作用域；并行教程、包号及其他修改未覆盖。
- 独立product-samples/magic-recipes/site新增magic-connect.cjs／magic-connect.test.cjs，build.mjs嵌入；template.html新增登录入口、model-ui.js配对接收并导入私有连接Map，独立theme.css补样式，package.json与ui.test.cjs纳入回归。登录入口仅官方域名或本机HTTP；精确来源／窗口／随机nonce／用途／五分钟期限及一次性接受，API Key仅内存、无账号凭据回传、不自动发生成请求，手动配置仍在。
- 生产只允许同源或构建环境VITE_MAGIC_RECIPES_ORIGIN的精确来源，本机站额外允许4178；实际env仍gitignore，不添加通配、不放松COOP等安全头。两端需一起发布；目录存在不等于模型能力实测。教程、PRODUCT、DESIGN、两级README、UPGRADE与HANDOFF四／五同步更新。
- 51项Node＋99项原站定向（含新增授权／选择25项和国际化3项）测试、类型、相关lint、生产构建通过。浏览器4178入口／1280布局、4175匿名登录目标保留及无配方控制台错误已验证；本轮视口覆盖未生效，新增手机控件与真实登录后选择／跨窗口闭环未实测。证据output/recipes-login-20260930-*.log及本线程recipes-api-login-desktop-20260930.png。本机4178 session18607已重启，用户旧页不刷新；未调用真实付费模型、未更新8080镜像／生产、未提交推送。

## 2026-09-30 第四产品规划草案

- 新增 `product-samples/magic-studio/PLAN.md`：按用户「先规划」要求记录暂名魔法工坊、图片／视频分类、Skills·工作流、新玩法、定期发现与审核、分阶段模型接入。仅文档，无运行时／上游接缝；母品牌雾钛青／狮冠M／默认深色与SynaRoute不改。
- 外部官方来源只做网页取证，不安装执行或调用模型；16／40条为未来内容目标，每日／每周频率尚未启用。更新HANDOFF四／五及CLAUDE入口，保留其他未提交工作，未提交推送或部署。

## 2026-09-30 魔法配方自带模型设置（独立原型）

- 新增product-samples/magic-recipes/site/model.cjs、model-ui.js、model.test.cjs与fixtures/model-mock.mjs；template.html增加页内模型配置及运行／追问／图片结果区，build.mjs嵌入模块，app.js仅发配方失效与整理完成事件，样式留在独立site/theme.css。文字与生图分别连接，视频不误发文字请求。无上游业务接缝、无新依赖或真实配置文件。
- 兼容Chat Completions／Responses文字和Images API生图；仅运行或追问直连用户地址，所有配置和Key均仅当前页内存。远程HTTPS、本机指定HTTP；不带Cookie／不跟随API重定向／不自动重试，纯文本显示模型回答，错误不回显服务商原始内容。中止、超时及配方切换隔离已覆盖；停止等待不等于服务端取消。
- 语法、构建、44项定向测试通过；浏览器本机模拟接口验证文字两格式、追问与图片显示，1280／390／320布局无外层溢出，实际服务权限／CORS与模型效果未验证。证据output/recipes-model-20260930-tests.log及本线程截图，最终4178 session64858。模拟生图只返回既有Logo，未用真实模型或扣费；未提交推送／生产部署，其他并行业务改动保留。更新上层README、07／08范围说明、site产品／设计／使用文档以覆盖旧无模型边界。

## 2026-09-30 包号专项修复（待发布，覆盖下方审查待修状态）

- 上游新接缝：`backend/internal/service/api_key_service.go`的APIKeyService.Update，在repo.Update前以[CUSTOM]标记统一核验专属资格，覆盖显式启用与扩额／重置额度／清除或延长到期的隐式激活。非包号原路径保持，停用仍可执行。
- 定制service：结构、保存与当前成员请求校验分离；运行时忽略非成员旧Key状态，保留独占及非成员授权检查。移除Key使用inactive，保存兼容历史disabled。Save要求expected_updated_at，撤销恢复需reactivate确认；custom handler映射409／400。AdminList增config_status，先关闭列表rows再检查，兼容单数据库连接。
- 定制前端：列表单列同时显示绑定状态及配置诊断；保存传原始版本，冲突提示刷新重开，撤销恢复checkbox明确确认。恢复控件与诊断文案样式仅theme.css。新增custom_dedicated_key_update_test.go测试适配器并扩展既有Go／Vue回归；无新迁移。
- 验证终态及证据见HANDOFF四与output/dedicated-fix-20260930。未提交推送部署，未触碰product-samples；生产历史请求根因仍未确定。
- 最终验证：40项前端定向＋3项国际化、typecheck／相关lint／build通过；Docker定向Go service／handler、embed编译及隔离PostgreSQL65项service测试通过。未全量Go／本轮CI／真实管理员浏览器验收，8080未更新。

## 2026-09-30 包号专项审查（未修复／未发布）

- 新增deploy/DEDICATED_AUDIT.md，记录三类隔离数据库复现缺陷及管理页诊断缺口；审计探针／日志在gitignored的output/dedicated-audit-20260930，不新增运行时接缝。19个既有service测试＋1个四子场景审计测试执行完成，审计断言坏行为，不能称缺陷已修复。当前成员移除补丁不宜单独发布；本轮不修改业务逻辑、不提交推送部署，详见HANDOFF最新状态。

## 2026-09-30 包号成员移除修复（待发布）

- 仅改现有定制service `custom_dedicated_accounts.go`／`custom_dedicated_members.go`：保存时事务内撤销被移除成员的原组授权并停用原组Key；校验不计入disabled的历史Key，仍拒绝非成员授权和其他未停用Key。成功后刷新用户／组缓存，失败回滚；无新迁移、不放宽账号独占、不改变余额／其他组，不删除Key记录。
- 更新定制copy中英文提示与排查说明；扩展现有sqlmock及真实PostgreSQL用例，覆盖原失败路径、自动清理、失败回滚、其他组不变与不自动重新启用Key。31前端测试／typecheck／相关lint、Docker Go定向测试及embed编译、19个service测试含真实SQL通过，证据output/member-removal-20260930。无上游新接缝，未提交推送或部署。

## 2026-09-30 控制台与多人包号（未发布）

- **最新发布状态（覆盖本节历史未发布文字）**：43b715a15已推origin/main，GHCR36657101655及安全扫描36657101610成功，镜像sha-43b715a已发布、摘要dc36b117e524262faddf0c13f0dad5c9ff510e385e69ddc2860145a7ea6566af。CI36657101544前端／Go lint／shell／release-helpers／Go Unit通过，集成测试尚未结束，不标全绿；发布前核对终态。证据output/console-release-43b715a，生产未操作。

- **收口决定**：用户确认保留账号不可加入其他分组的独占约束，授权提交推送；本轮不继续改业务规则。生产待用户发布，本轮CI／镜像状态按代码提交重新验证，未跟踪product-samples不提交。

- **2026-09-30 已部署本机Docker（优先于本轮未部署记录）**：用户授权本机测试。原三容器处于停止状态，保留原卷启动Postgres／Redis，先完成597905字节数据库备份与pg_restore目录校验、旧镜像标签sub2apicust:before-console-20260930，再代理预检200并构建。前端类型／国际化3测／构建及Go embed成功；仅应用容器重建，运行镜像与标签一致be2f26f98e8c3dee62d765a8b78e6cd33021dc3b6f618c76085e310180b234b8，三服务healthy，health=ok。迁移242两列和三个索引确认存在；数据库／Redis容器ID及app_data卷不变，用户2／密钥0／组3／套餐1／订单0／包号0／设置280计数保持。备份逐项比较只有Claude版本2.1.283→2.1.285与Codex版本0.158.0→0.159.2两键自动同步，09:44启动日志证实，其余278键不变。HTTP入口index-Ckx_MqIV.js及三个新页面chunk验证通过，账户与管理员接口匿名401；浏览器8080官网正常打开并留给用户测试，尚未做带数据管理员交互验收。证据output/console-deploy-20260930，备份SHA256=1D900EC14FB9915B5E912F0A84778271CEB149178B7CC526E3A9F55411EE037E。未提交推送或操作生产；迁移后的回退需兼顾数据库，不能直接跑旧单用户版本。

- 定制前端：PublicLayout／BrandHomeView按登录态隐藏游客入口；新增ConsoleFaqView与 `/help/faq`，PublicFaqView支持隐藏重复标题。DedicatedPicker提供已选名称回退，AdminDedicatedAccountsView显示所有用户／账号／组名称并支持1–100用户、改绑与撤销后删除；DedicatedAccountsView仅公开本人组名。IntelligenceView默认矩阵展示平均任务费用、缺失与聚合口径，中英文copy补充费用定义；新增样式只在theme.css。
- 定制后端：新增 `custom_dedicated_display.go` 名称白名单、`custom_dedicated_members.go` 成员与删除逻辑；更新已有custom service/handler，迁移242添加JSON成员及逻辑删除，改绑保留旧隔离记录。sqlmock与真实PostgreSQL测试覆盖迁移、多人、非成员拒绝、名称隐私、改绑、删除和重新分配。不改计费、不自动授权、不公开管理员账号信息给用户。
- 上游接缝：`frontend/src/components/layout/AppSidebar.vue` 的 `[CUSTOM]` FAQ路径改为 `/help/faq`；`backend/internal/server/routes/admin.go` 新增带 `[CUSTOM]` 的管理员DELETE路由，沿用认证、合规、审计中间件。
- 验证：139前端定向＋3国际化、类型／lint／build通过；Docker Go定向service/handler单测及embed编译通过，18个service测试在临时PostgreSQL容器运行通过（含真实SQL）。证据output/console-fixes-20260930；未全量Go／CI／管理页浏览器视觉验收、未提交推送部署，生产403/503原因尚未确定。维护与回退见deploy/DEDICATED_TROUBLESHOOTING.md。

### 魔法配方创作扩展（2026-09-28）
- 新增独立样品07-image-prompt.md、08-video-prompt.md；site目录接入creative分类、媒体字段默认值与对应工具引导，按钮明确仅整理模板。无上游业务文件修改、无新依赖或模型接口。
- 2026-09-28创作配方更新：新增07-image-prompt.md与08-video-prompt.md，现为工作3份、大学3份、创作2份。主按钮改为“整理并生成提示词”，明确本地模板组装、不调用模型；结果与教程按文字／生图／视频分别引导。构建与语法检查完成，30项core／DOM测试全通过；真实浏览器验证创作筛选2项、两份虚构示例及正确目标工具提示，桌面和390px无外层横向溢出。本机4178服务已重启（session75587），用户旧页未刷新。未调用模型、未验证媒体效果、未提交推送或部署生产；真实系统剪贴板与下载落地仍待验证。

## 2026-09-28 游客控制台预览入口

- **发布收口**：功能9a31787de＋SQLite测试初始化修正fa45b2023已推送；修正SHA的CI36424788919、GHCR36424788938、安全扫描36424789053全部成功（含Go单测及集成）。发布标签sha-fa45b20，镜像digest=sha256:69083ada3ce6432eaee2357e57fa384aa0804d125444cc33998e742c601d1d44；替代历史「后端测试待验证／未提交」状态。证据output/guest-release-fa45b20，生产未部署，product-samples未提交。

- **响应式修正**：只改theme.css游客选择器：去掉main的1280px上限、弹性对称内边距、字段等宽网格和480px双列、手机顶栏避免操作另起一行；不改认证／账户数据逻辑。七档320–2560px几何检查、65项定向回归通过，已部署8080镜像ea6e3a5ecc3c。证据output/preview-responsive-20260928；无新增上游接缝。

- **本机部署补充**：用户后续授权后，8080已运行ecb3961845e8，前端及Go embed构建通过、三服务healthy。新预览入口及真实套餐／登录提示取消经浏览器验证；业务计数、设置摘要和原数据卷不变。备份／构建／验证记录见output/preview-20260928-175007。此条替代下方本轮未部署状态；未提交推送或部署生产。

- 新增 `frontend/src/custom/guest/preview.ts`、`custom/views/GuestPreviewView.vue`、`custom/__tests__/guest-preview.spec.ts`；首页主CTA和PublicLayout导航新增 `/preview`，custom/routes注册可选功能子路由。独立只读布局不导入真实账户页面；公共套餐和FAQ支持embedded以保留预览侧栏。
- 上游接缝 `frontend/src/router/index.ts` 的 `[CUSTOM]` 公共入口后台模式检查加入isPreviewPath；`router/__tests__/feature-access.spec.ts`补充预览子页访问与后台模式拦截。未开放真实账户路由。所有新增样式仅在custom/theme.css。
- 本轮179项定向测试、3项国际化测试、类型检查、相关ESLint和生产构建通过；4175浏览器实测入口／登录取消／真实套餐／FAQ搜索及响应式。新预览尚未同步8080 Docker镜像，未提交推送／生产部署。

## 2026-09-28 游客官网、公开套餐与FAQ

- **本机部署更新**：2026-09-28用户授权Docker部署，运行镜像`e7a582285cbd`，前端及Go embed编译通过，三服务健康。公开目录读取既有1个在售套餐，4个个人接口匿名401；浏览器套餐→登录携带目标及FAQ搜索通过。下方“后端未编译、8080尚未更新”为源码实现阶段历史状态；Go单测仍未运行，完整登录购买待验收。证据output/guest-20260928-171257，旧镜像和594458字节数据库备份保留。仅内置自动同步的Codex版本键变化，其余279项设置逐项一致。未提交推送或生产部署。

- 新增后端 `backend/internal/handler/custom_public_plans.go`／`custom_public_plans_test.go`：只读字段白名单接口、在售过滤、支付／订阅／后台模式边界；后端本机未编译未运行测试，等待本轮未来提交CI。
- 新增前端 `custom/guest/{api,navigation,faq}.ts`、`custom/components/{PublicLayout,GuestAction}.vue`、`custom/views/{PublicPlansView,PublicFaqView}.vue`、`custom/__tests__/guest*.spec.ts` 和 `custom/GUEST_PORTAL.md`；修改已有BrandHomeView、custom/routes.ts、theme.css及升级契约。FAQ四类十二题，不增加数据库或管理后台。
- 上游接缝逐处登记（均有 `[CUSTOM]` 标记）：
  - `backend/internal/server/routes/payment.go`：仅注册 `GET /payment/public/plans`，原认证接口不变。
  - `frontend/src/router/index.ts`：四个公共页面冷启动等待设置，并保留后台模式限制。
  - `frontend/src/views/HomeView.vue`：自定义内容优先，正常站点默认游客官网；紧凑首页不再覆盖此入口。
  - `frontend/src/components/layout/AppSidebar.vue`：个人区新增公开FAQ入口。
  - `frontend/src/views/user/PaymentView.vue`：按 `plan` 恢复套餐，支付恢复和原 `group` 续费优先；不自动下单。
  - `frontend/src/views/auth/LoginView.vue`：三种登录成功回跳使用站内校验，注册入口携带返回目标。
  - `frontend/src/views/auth/RegisterView.vue`：注册／返回登录／邮箱验证保存选购目标；无目标时保持原行为。
  - `frontend/src/views/auth/EmailVerifyView.vue`：邮箱验证完成及返回注册采用站内目标。
  - 相关上游测试：`views/__tests__/HomeView.compact.spec.ts`、`router/__tests__/feature-access.spec.ts`、`views/user/__tests__/PaymentView.spec.ts`、`views/auth/__tests__/RegisterView.spec.ts`，分别覆盖首页优先级、真实守卫、套餐回跳、注册目标保留。
- 验证：163项定向前端测试、类型检查、相关ESLint及前端构建通过；后端新增测试未运行，真实套餐和支付联调待更新后端后验收。Vite 4175预览不等于8080新版镜像；未提交／推送／部署，未修改品牌配置及其他任务样品。

## 2026-09-28 魔法配方内容样品

- 使用教程：site/template.html、app.js、theme.css、ui.test.cjs及README／DESIGN新增四步折叠教程、顶部／结果入口与常见问题；页面不调用模型，教程往返不清除材料。27项回归通过，真实浏览器确认往返和390px布局，无上游业务文件改动。

- 页面内确认框：site/app.js、template.html、theme.css、ui.test.cjs及README／DESIGN更新；覆盖／清空不再使用原生confirm，支持模态焦点、取消及跨配方保护。24项核心／DOM测试通过，真实浏览器验证取消与键盘、390px浅色布局。只作用于独立原型，无上游业务接缝；未触碰并行的中转站业务修改。

- 独立交互原型product-samples/magic-recipes/site/：catalog.mjs字段与示例，build.mjs从上层Markdown生成单页并嵌入既有Logo，core.cjs／app.js提供纯本地生成和交互，template.html／theme.css仅影响独立页面，serve.mjs仅绑定127.0.0.1:4178；package.json无新增依赖。index.html是gitignored构建产物。core.test.cjs／ui.test.cjs共19项通过，后者复用frontend的jsdom。没有修改上游业务文件或原站主题；复制系统落地、TXT真实下载及模型实测仍待，不宣称已上线。PRODUCT.md／DESIGN.md记录原型边界，不替代母品牌决策或用户最终技术选型。

- 大学扩展（用户已排除其他学段）：product-samples/magic-recipes/{04-student-understand,05-student-mistake-review,06-student-revision-plan}.md现为0.2大学论文精读、编程实验复盘、期末复习；README同步。原低龄内容已替换，工作3份不变。只读内容草稿，无业务接缝；六份结构／链接和大学例子数值检查通过，未执行Python或外部模型／学生实测，不宣称提分或所有专业适配。

- 新增product-samples/magic-recipes/{README,01-meeting-actions,02-requirements-acceptance,03-error-triage}.md：母品牌第三产品的内容评审样品，无运行时接缝、无依赖、无配置或上游业务代码改动。所有参考答案为人工编写，未外部模型实测；结构与链接检查不等于效果验证。不自动发布、不执行用户任务、不托管密钥。

## 2026-09-28 上游推广清理

- 上游接缝 frontend/src/components/layout/AppHeader.vue：移除管理员GitHub菜单项，保留客服配置与其他操作，添加中文 [CUSTOM] 标记。
- 上游接缝 frontend/src/views/HomeView.vue：移除页脚仓库链接及githubUrl常量，保留站点文档与版权，添加 [CUSTOM] 标记。
- 上游接缝 frontend/src/views/KeyUsageView.vue：移除页脚仓库链接及githubUrl常量，保留站点文档与版权，添加 [CUSTOM] 标记。
- 上游接缝 frontend/src/components/common/ProxyAdBanner.vue：移除广告模板，改为含 [CUSTOM] 注释且render返回null的兼容组件（空模板不符合vue/valid-template-root）；兼容ProxiesView、CreateAccountModal、EditAccountModal三处调用，不再展示sub2api.io/proxyip广告。
- 新增 frontend/src/custom/__tests__/promotion-removal.spec.ts：5项测试覆盖空广告组件、三个入口不含仓库推广、客服／站点文档／版权保留。Vitest及vue-tsc -b通过。
- 审计范围为frontend/src的Vue／TS外链及推广关键词；frontend/backend源码未命中截图QQ号码，客服来自appStore.contactInfo。保留管理员合规文档、支付帮助、TLS采集工具、GitHub开发者设置和OAuth功能链接，不删LICENSE。未读取运行站点数据库，无法判断自定义首页／菜单／客服配置是否另含推广。
- 2026-09-28已部署本机验证（详见HANDOFF），用户随后授权与包号一起提交推送；同步后检查四处接缝。生产由用户手动升级。

> 这是本 fork 的**唯一权威清单**。每一处偏离上游的改动都必须登记在这里。
> 同步上游（`git merge upstream/main`）解冲突时，逐条对照本表核对：
> **每一处定制是否仍然存在、是否被上游改动覆盖**。合并里定制被静默吞掉，是这类
> 维护最常见、也最难发现的失效——本表是唯一防线。
>
> 维护约定：
> - 优先**新增文件**，降低文本冲突面；仍须核对上游依赖，不能保证永不冲突。
> - 不得不改上游文件时，改动**尽量小、尽量集中**，并在下方「接线改动」逐处登记。
> - 配置类定制（`.env` / `config.yaml` / `docker-compose.override.yml`）走 gitignore，
>   不进本 fork，不登记在这里；它们记在你的私有 ops 仓。

---

## 一、新增文件（低冲突，仍需验证依赖契约）

### 客户端接入教程（2026-09-30，本地未发布）
- `frontend/src/custom/views/PublicGuideView.vue` — 公开教程；按站点公开 API 基础地址展示 Anthropic 根地址及 OpenAI `/v1`，CLI／桌面端与常见第三方客户端步骤、官方文档、错误排查，不接触真实密钥。
- `frontend/src/custom/views/ConsoleGuideView.vue` — AppLayout 内嵌相同教程，保留控制台侧栏。
- `frontend/src/custom/__tests__/guide.spec.ts` — 路由边界、协议配置和地址占位符；预览与守卫测试增加教程路径。
- 已有定制文件 `custom/routes.ts`、`custom/guest/preview.ts`、`custom/views/GuestPreviewView.vue`、`custom/components/PublicLayout.vue`、`custom/views/BrandHomeView.vue`、`custom/views/PublicFaqView.vue`、`custom/views/ConsoleFaqView.vue`、`custom/theme.css` 接入导航、FAQ往返、只读预览与响应式文档样式；所有样式只在 theme.css。教程scope确保控制台／预览中的套餐和FAQ链接不跳出对应浏览入口，GuestAction保护游客密钥操作。
- **上游接缝**：`frontend/src/components/layout/AppSidebar.vue` 新增带 `[CUSTOM]` 的 `/help/guide` 入口；`frontend/src/router/index.ts` 将 `/guide` 纳入已有 `[CUSTOM]` 公开页后台模式检查，另新增限定三个教程路径及五个已知锚点的滚动逻辑，保留浏览器返回恢复与其他页面置顶。同步上游时核对两文件中的三处逻辑。
- 80项定向前端、国际化3项、typecheck、相关ESLint与生产构建通过；浏览器验证1280深色／390浅色无外层溢出，游客登录取消、键盘折叠、章节滚动及预览FAQ往返。证据output/client-guide-20260930，截图在本线程visualizations目录；4175可预览。模型权限、客户端版本、桌面本地／云端与真实调用仍需单独验收，此轮不涉及后端与业务数据，未提交／部署。

### Claude / ChatGPT-Codex 包号（2026-09-28，源码，未部署）
- 使用与验收边界：`DEDICATED_ACCOUNTS.md`。仅标准模式；复用事先配置好的一账号／一标准专属分组／一授权用户，不自动搬号、不改变计费。
- `backend/migrations/241_custom_dedicated_accounts.sql` — 包号独立表、账号及分组唯一约束、撤销保留隔离记录。外键限制硬删除；同步上游时核对迁移排序及清理流程，不能回退旧镜像后假设独占仍生效。
- `backend/internal/service/custom_dedicated_accounts.go` — 绑定／续期／撤销、分页、服务端归属白名单、实时独占条件与到期校验；串行化写事务及失败关闭。
- `backend/internal/service/custom_dedicated_quota.go` — 只读上游保存的真实被动快照；缺失／无效／过期不推测余量；不调用Codex生成探针。
- `backend/internal/service/custom_dedicated_forward.go` — 两类Gateway共用转发前／WS每轮入口。
- `backend/internal/handler/custom_dedicated_handler.go` — 管理与用户独立API，错误脱敏、用户身份过滤、分页验证。
- `backend/internal/service/custom_dedicated_accounts_test.go`、`backend/internal/handler/custom_dedicated_handler_test.go` — SQLmock／归属／事务／额度／HTTP越权测试，本机未执行，待本轮CI。
- `frontend/src/custom/dedicated/{api,copy}.ts`、`custom/components/DedicatedPicker.vue`、`custom/views/{DedicatedAccountsView,AdminDedicatedAccountsView}.vue` — 用户只读快照与管理操作，中英文、分页、续期、撤销确认；继承原AppLayout，样式仅custom/theme.css。
- `frontend/src/custom/__tests__/dedicated*.spec.ts` — 27项新增前端回归；连同现有定制关联测试119项通过。Go编译／格式化／单测／真实数据库与浏览器验收仍待，不当成完整业务验收。

> 格式：`路径` — 用途 — 引入日期

### 智力效率页面（2026-09-27）
- 最新说明区精简：按用户指定仅保留IQ可比性与低样本／缺失数据两段，移除说明标题、来源归属、费用字段及直连提示；英文同步。数据获取及刷新逻辑不变。
- **修正版部署证据**：本机8080已运行`cfa0a12c8bf6`，三服务healthy，HTTP实际加载`IntelligenceView-C0f1Aymd.js`，新对比／排序控件、去外链／去频率元素和充值width规则已核验；76项回归与前端构建／Go embed编译通过。280项设置摘要及分组2／套餐0等条数不变；数据库备份583628字节，旧镜像保留，详见HANDOFF与output/brand/intelligence-ux-deploy-record.json。未跑Go单测，真实浏览器验收仍待。
- **本轮可读性修正（优先于下方初版描述）**：移除「查看源站」按钮、JSON外链和所有刷新间隔文案，保留纯文本来源归属、源更新时间、手动刷新及原30分钟轮询。矩阵默认仍为数据源首次出现顺序；新增模型名称／指定档位分数降序，缺失最后、同分保留源序，不计算跨档位综合分。
- `custom/components/IntelligenceComparison.vue` 替代并删除未提交初版`IntelligencePlot.vue`：不用无标签散点图，改为模型名＋三个独立指标条／精确数值，默认high档、可切全部档位与按分数／耗时／费用排序，每页12组合，缺失指标不丢整行。手机改纵向条目；三个指标单位不同，页面明确不跨列比较条长；图表不再依赖Chart.js。
- `custom/theme.css` 新增 `.mofa-workspace-main > .mx-auto.max-w-4xl { width: 100%; }`：适配原充值页在定制纵向flex正文里的居中面板，保留max-w-4xl最大宽度，避免空态随内容收窄。没有修改PaymentView、后端支付查询或订阅规则。`custom/__tests__/upgrade-contract.spec.ts` 登记该上游root class依赖；升级如改变结构需复核。
- 本轮只读核验：groups中「测试分组1」为active/subscription，日100／周500／月2000；subscription_plans仍0条。`PaymentConfigService.ListPlansForSale`只查for_sale=true套餐，checkout-info也用此方法，所以创建分组不自动上架。应在 `/admin/orders/plans` 新建关联分组的售卖套餐；不伪造价格、不自动创建订单。
- `frontend/src/custom/views/IntelligenceView.vue` — 认证后 `/intelligence` 页面：模型图标、分数条矩阵、三指标条形对比、模型搜索／评测环境／样本筛选、中英文及来源说明。
- `frontend/src/custom/intelligence/{data,copy,useIntelligence}.ts` — 固定读取用户指定 Codex Radar JSON；schema=2 校验、模型×harness×effort 去重，缺失不当零分；页面打开每30分钟读取、手动刷新、15秒超时、同页请求去重、隐藏页暂停、回前台过期补取、失败保留本次挂载上次成功数据。不是服务端定时任务，不持久化跨重载缓存，不携带本站凭据。
- `frontend/src/custom/components/IntelligenceIcon.vue` / `IntelligenceComparison.vue` — 线性雷达入口图标、原生HTML三指标条形对比；沿用已有ModelIcon，分页／精确数值／样本／聚合方式／缺失值，主题直接取CSS变量，无canvas和theme observer。
- `frontend/src/custom/__tests__/intelligence.spec.ts` / `intelligence-view.spec.ts` — 数据、轮询、超时、卸载、失败保留、零值／缺失、跨环境、筛选、图表明细与中英文回归。
- 既有定制文件接线：`custom/routes.ts` 注册认证路由，`custom/brand/useWorkspaceHeading.ts` 提供中英文标题；全部新增视觉规则在 `custom/theme.css`，含深色黑色品牌图标对比度修正，不改 SynaRoute 或原语义 teal。
- **上游接缝**：`frontend/src/components/layout/AppSidebar.vue` 新增 `[CUSTOM]` 图标／文案 import，`useI18n` 取 locale，`buildSelfNavItems` 添加 `/intelligence`，用户与管理员个人区共用。合并需保留入口、认证守卫、侧栏折叠和移动端行为。
- **订阅不重复造轮子**：原 `PaymentView.vue` 已有充值／订阅Tab、套餐购买与续费入口，本轮未改支付业务或数据库；本机只读SQL确认 `subscription_plans` 总数和上架数均0，须由用户提供套餐售价／额度／周期／分组后配置，不照搬参考截图。
- 验证：真实接口解析24模型／89组合／24矩阵行／15低样本组合；70项定制与订阅定向测试通过，定向ESLint与构建通过。后续用户授权本机Docker发布完成：8080镜像`1a7a31d0e424`，Docker内前端构建与Go embed编译成功，三服务healthy；HTTP实际入口引用`IntelligenceView-Cob7bqKS.js`并含源JSON地址，主题CSS含新矩阵规则；数据条数与全部280项设置摘要不变。备份与构建日志见HANDOFF及output/brand/intelligence-deploy-record.json。浏览器工具认证失败，真实浏览器视觉验收仍待；未跑Go单测、未提交推送、未部署生产。

### 品牌实施交接计划（2026-09-26）
- `BRAND_IMPLEMENTATION.md` — 已获用户确认的雾钛青方向、素材与色表、真实代码接线、实施阶段、验收及 fork 维护边界。实施及验收进度见文末第 9 节。
- `AGENTS.md` — 当前任务入口更新为读取该计划；SynaRoute 保留紫色，后续在其工程独立统一基础规范。


### 真实页面结构还原（2026-09-27）
- 新增 `custom/brand/useWorkspaceHeading.ts`：从原页头抽取同一套路由／计费模式／自定义菜单标题计算，保留管理员菜单隔离；`custom/components/WorkspaceHeading.vue` 在正文承载标题及操作插槽。
- `components/layout/AppHeader.vue`：顶部改为站点／页面路径；账户、公告、余额、移动菜单事件不变。`AppLayout.vue`：新增正文标题、`page-actions` 插槽、折叠状态钩子；原内容插槽保留。
- `components/layout/AuthLayout.vue`：现有品牌信息移至顶部品牌栏，保留原认证 default/footer、设置加载和 URL 消毒；登录业务未复制。
- `components/layout/TablePageLayout.vue`：新增统一面板包裹 filters/table/pagination；actions 插槽和移动检测不变。
- `views/user/KeysView.vue`：原刷新／列设置／创建按钮整块迁至 `page-actions`，仅增加筛选／搜索样式钩子；事件、ref、权限、API、字段、SynaRoute 导入不变。`views/user/__tests__/KeysView.spec.ts`：AppLayoutStub 渲染新插槽，保留24项原测试。
- `custom/theme.css`：原稿字体、224/72侧栏、独立正文标题、统一筛选表格分页面板、认证顶部品牌栏与大屏比例；覆盖 scoped 表头／sticky列背景时不改定位、语义色或错误态，无新增 !important。
- 新增 `custom/__tests__/workspace-heading.spec.ts`，扩展 upgrade-contract；定向61项测试、ESLint、类型检查和前端构建通过。详见 BRAND_IMPLEMENTATION 第15节。
- 维护成本：本轮扩大展示接缝，不是零冲突；同步上游逐项核对 `custom/UPGRADE.md`，不得整文件覆盖上游新功能。

### SynaRoute 深链接一键导入（2026-09-22）
按 SynaRoute 文档「深链接一键导入」（`synaroute://`）在 API Keys 页面加「导入到 SynaRoute」按钮，
**形态完全对齐上游已有的「导入到 CCS」（`ccswitch://`）**：并排放在密钥行操作里，并配一个管理员开关。
规格文档：`C:\Users\Administrator\Desktop\temp\demo\SynaRoute\docs\20-深链接一键导入.md`。
- `frontend/src/utils/synaRouteImport.ts` — 纯函数：URL 构造 + 由 platform 推导分类/协议/端点（镜像 `ccswitchImport.ts`）
- `frontend/src/utils/__tests__/synaRouteImport.spec.ts` — 单测（11 用例，已过）

### 部署文档（2026-09-22）
- `deploy/LOCAL_DOCKER_RUNBOOK.md` — 2026-09-26本机Docker Hub授权超时的证据、原因边界、代理预检／进程级设置及部署验收清单；不改变生产配置。
- `deploy/DEPLOY_CUSTOM.md` — 定制版 GHCR 镜像部署清单（登录/起服/升级/回滚/迁移）
- `deploy/update.sh` — Docker 下的一键更新脚本（`pull && up -d` + 健康自检 + 旧镜像清理；带标签参数可回滚）。用户明确选择「Docker 镜像更新」而非启用 App 内按钮（那按钮是 systemd 安装用的原地换二进制，Docker 下权限失败且会被 pull 覆盖）。

### 前端定制叠加层（换肤 + 新增/替换页面）（2026-09-24）
用户方向：「换肤 + 少数页面」。设计为 `frontend/src/custom/` 叠加层，把冲突面收敛到 4 个接缝文件（见下节）。
- `frontend/src/custom/theme.css` — 品牌变量层：`--color-primary-*`（雾钛青，500=#096B68）+ `--color-dark-*`（炭青，950=#162124）；移除 body 双径向光晕。新增 brand 语义变量及 btn-primary / input / card-glass / 深色 text-gradient 公共覆盖；夜间主按钮 #A1D9CE + #173B38，日间 #096B68 + 白字。认证页原有装饰仍保留。不改 gray/teal 分类色、Logo、默认深色和业务组件。公共类、图表与高亮需随上游升级复查；无后端时业务页未验收，见 BRAND_IMPLEMENTATION.md 第 9 节。换肤只改这里；logo/站点名走管理员后台设置（无需改码）。
- 2026-09-26 修复补充（仍仅 theme.css）：使用 html.dark .text-gradient 提高明确主题作用域的优先级，移除夜间文字渐变，避免 AuthLayout 懒加载 scoped 样式将标题变回透明深色渐变；普通输入边框/焦点排除 input-error，明确保留错误红边/焦点；btn-primary 内 animate-spin 继承按钮前景，避免浅青底白色加载图标。真实 Docker 页验证见 BRAND_IMPLEMENTATION.md 第 11 节。
- `frontend/src/custom/routes.ts` — 新增页面路由集中处（`customRoutes`）。
- `frontend/src/custom/views/CustomDemoView.vue` — 脚手架演示页（验证链路用，可删）。
- `frontend/src/custom/README.md` — 叠加层三种用法 + 影子替换代价说明。
- 新增页面/影子替换页放 `custom/views/`、`custom/components/`（降低文本冲突面，仍需复核上游依赖）。

## 二、接线改动（会冲突，重点核对）

### 包号接缝（2026-09-28）
| 文件 | 必须保留／核验的改动 |
| --- | --- |
| `backend/internal/service/api_key_service.go` | customDedicated依赖字段，不改变原构造器参数 |
| `backend/internal/service/gateway_service.go` | Claude Gateway的customDedicated字段 |
| `backend/internal/service/openai_gateway_service.go` | Codex Gateway的customDedicated字段 |
| `backend/internal/service/wire.go` | NewCustomDedicatedService提供者 |
| `backend/internal/handler/handler.go` | Handlers.CustomDedicated |
| `backend/internal/handler/wire.go` | 新处理器提供者与ProvideHandlers参数／字段 |
| `backend/cmd/server/wire_gen.go` | 与Wire提供者同步的显式初始化；重新生成后必须包含包号服务，不能漏装保护 |
| `backend/internal/server/middleware/api_key_auth.go` | 在SimpleMode早退前写入原始身份并进行实时包号认证 |
| `backend/internal/server/routes/admin.go` | 管理GET/POST/PUT/撤销，位于原管理员认证／合规／审计之后 |
| `backend/internal/server/routes/user.go` | 用户列表与单条额度，只在JWT及Heavy限流下注册 |
| `backend/internal/service/gateway_scheduling.go` | 两个公共选号入口defer终检；保留原始上下文，失败释放并发槽 |
| `backend/internal/service/openai_gateway_scheduling.go` | 传统选号及负载感知终检 |
| `backend/internal/service/openai_account_scheduler.go` | 高级调度统一包装层终检，覆盖粘性和图片回退 |
| `backend/internal/service/gateway_forward.go` | Claude主转发前检查 |
| `backend/internal/service/gateway_forward_as_chat_completions.go` | Claude Chat转换转发前检查 |
| `backend/internal/service/gateway_forward_as_responses.go` | Claude Responses转换转发前检查 |
| `backend/internal/service/gateway_count_tokens.go` | Claude计数转发前检查 |
| `backend/internal/service/openai_gateway_forward.go` | Codex Responses主转发前检查 |
| `backend/internal/service/openai_gateway_chat_completions.go` | Chat转换的共用内部转发前检查 |
| `backend/internal/service/openai_gateway_messages.go` | Anthropic协议转换前检查 |
| `backend/internal/service/openai_gateway_count_tokens.go` | 原生InputTokens和Anthropic计数两处检查 |
| `backend/internal/service/openai_images.go` | 图片转发前检查 |
| `backend/internal/handler/openai_gateway_handler.go` | WebSocket BeforeRequest每轮检查，拒绝沿用已撤销／到期账号 |
| `frontend/src/components/layout/AppSidebar.vue` | 用户和管理菜单及copy导入，两项均hideInSimpleMode |
| `frontend/src/custom/routes.ts` | 两个懒加载认证路由；管理页requiresAdmin |
| `frontend/src/custom/brand/useWorkspaceHeading.ts` | 两个页面的中英文标题和描述 |
| `frontend/src/custom/theme.css` | 仅追加mofa-dedicated作用域样式，保留现有主题与SynaRoute |

同步上游时若新增可直接转发Claude／Codex的入口，必须复核其认证、选号和转发前保护；源码存在检查不等于所有未来入口自动安全。

### 2026-09-27 展示接缝补充
- 本轮 AppHeader／AppLayout／AuthLayout 接缝已扩大，并新增 TablePageLayout、KeysView 和 KeysView 测试 stub 接缝；逐处内容见上方「真实页面结构还原」。旧条目中“未改业务页”“仅三个类名”的范围描述已由本轮更新取代；只移动展示节点，不改业务处理。
- 合并后按 frontend/src/custom/UPGRADE.md 保留新插槽并跑61项定向回归，不以CSS换色测试代替结构验收。

### 雾钛青品牌页面（2026-09-26，未提交／已本机部署，未发布生产）
- 新增 `frontend/src/custom/views/BrandHomeView.vue`；`custom/routes.ts` 增加 `/brand`，不抢占 `/`，沿用原守卫。
- 新增 `frontend/src/custom/components/BrandPanel.vue`、`BrandThemeToggle.vue`、`brand/copy.ts`：品牌展示、中英文案、主动切换主题；挂载不改变默认深色。
- 新增 `frontend/src/custom/assets/mofa-mark.webp`：原批准透明 M 位图，保留狮子／皇冠，非矢量重绘；站点名称与图标仍走公开设置。
- 新增 `frontend/src/custom/__tests__/brand.spec.ts`、`upgrade-contract.spec.ts`；补 `custom/UPGRADE.md` 并纠正 README 零冲突承诺。
- `frontend/src/components/layout/AuthLayout.vue`：品牌面板／主题按钮／品牌页链接及 mofa-auth 样式钩子；保留 default/footer 插槽、设置加载和 sanitizeUrl。
- `frontend/src/components/layout/AppLayout.vue`：mofa-workspace、mofa-workspace-backdrop、mofa-workspace-main 三类；保留 sidebar/header/slot 及侧栏折叠偏移。
- `frontend/src/components/layout/AppHeader.vue`：mofa-workspace-header、mofa-page-heading 两类；保留所有原事件和业务组件。
- 全部视觉规则在 `frontend/src/custom/theme.css`：认证双栏／手机堆叠、品牌首屏、控制台公共外观、长站名截断／换行；不改语义 teal 分类色，不改 SynaRoute。
- 本节更新此前“认证装饰保留／四接缝”历史描述：此轮另增三个布局接缝；装饰节点保留但由主题隐藏。未新增依赖或改业务页／后端。验证及尚未完成项见 BRAND_IMPLEMENTATION 第12节。

> 这些是为了把「新增文件」挂进系统而**必须编辑上游文件**的地方。
> 每次 merge 后逐条确认改动还在。所有改动都带 `[CUSTOM]` 注释便于 merge 时定位。

### SynaRoute 深链接一键导入（2026-09-22）
**前端**（已跑通 typecheck / check:i18n / eslint / KeysView 24 测）
- `frontend/src/views/user/KeysView.vue` — import util；CCS 按钮旁加「导入到 SynaRoute」按钮（`v-if="!publicSettings?.hide_synaroute_import_button"`）；加 `importToSynaRoute(row)` handler
- `frontend/src/i18n/locales/{zh,en}/dashboard.ts` — 加 `keys.importToSynaRoute` + `keys.synaRouteNotInstalled`
- `frontend/src/types/index.ts`、`stores/app.ts`、`api/admin/settings.ts`(两处接口) — 加 `hide_synaroute_import_button`
- `frontend/src/views/admin/SettingsView.vue` — 加 Toggle UI + form 默认值 + 提交 payload
- `frontend/src/i18n/locales/{zh,en}/admin/settings.ts` — 加 `hideSynarouteImportButton` 标签 + hint

**后端**（🔴 本机无 Go 工具链，未编译验证；commit 前必须 `cd backend && gofmt -w ./... && go build -tags embed ./cmd/server && go test -tags=unit ./...`）
设置键 `hide_synaroute_import_button`，全程镜像 `HideCcsImportButton`：
- `internal/service/domain_constants.go`（常量）、`settings_view.go`（两个 view 结构体）、`setting_parse.go`、
  `setting_public.go`（公共 key 列表 + 公共结构体 + 两处映射）、`setting_update.go`
- `internal/handler/dto/settings.go`（两个 DTO）、`setting_handler.go`
- `internal/handler/admin/setting_handler.go`、`setting_handler_update.go`（req 结构体 + 两处映射）、`setting_handler_audit.go`
- 🆕(0.2.8 合并引入) `internal/server/api_contract_test.go` —— 上游新增的 API 契约测试，含 `hide_ccs_import_button` 夹具 2 处（`require.JSONEq` 全量比对响应）。已镜像补 `hide_synaroute_import_button`，否则我们多出的键会让断言变红。判据：全后端 `hide_ccs` 与 `hide_synaroute` 命中数恒等（当前各 20）。

### 前端定制叠加层的 4 个接缝（2026-09-24）
> 全部打 `[CUSTOM]` 注释。**已跑通完整 `pnpm run build`**（含 check:i18n + vue-tsc + vite build），
> 产物 CSS 确含 `--color-primary-500: 20 184 166` 定义与 `rgb(var(--color-primary-500)/…)` 引用，
> 演示路由 `custom-demo` 已进产物 → 换肤与路由链路验证通过。
- `frontend/tailwind.config.js` — `primary` + `dark` 色阶、glow/glow-lg、gradient-primary、mesh-gradient、glow 动画均改为引用 theme.css 变量。⚠️ 上游若改 primary/dark 色阶或这些效果会冲突；解冲突时保留变量引用形式。
- `frontend/vite.config.ts` — `resolve.alias` 由对象改为**数组形式**，并留「影子替换」注释示例（覆盖项须在 `'@'` 之前）。
- `frontend/src/main.ts` — `import './style.css'` 后加 `import './custom/theme.css'`。
- `frontend/src/router/index.ts` — import `customRoutes` + 在 404 兜底前 `...customRoutes` 展开。

### 品牌换肤：上游文件里写死的品牌色 teal → 品牌变量（2026-09-26）
> 全站 `primary-*` 已走变量；但少数上游文件把品牌 teal **写死在 CSS / Tailwind arbitrary value 里**，
> 都是「首屏可见的品牌装饰」，逐处改为引用 `--color-primary-*`（带 `[CUSTOM]`）。已跑通 `pnpm run build`。
> ⚠️ 与**语义分类色**区分：`platformColors.ts` 的 deepseek、模型徽标、渠道监控状态色等 `teal-*` 是「按厂商/状态区分」的语义色，**刻意不改**（改了会撞色/丢语义）。
- `frontend/src/views/HomeView.vue` — 背景网格线、`.terminal-window` dark 光晕的 teal → `rgb(var(--color-primary-500)/…)`。（`.code-url` 的 teal 是终端 demo 语法高亮色，非品牌，保留。）
- `frontend/src/components/layout/AuthLayout.vue` — 登录/注册页背景网格线 teal → 变量。
- `frontend/src/styles/onboarding.css` — 引导高亮 outline + 引导「下一步」按钮底色/hover（teal / #14b8a6 / #0d9488）→ 变量。
- 残留（内部页、低优先，暂留并登记）：`KeyUsageView.vue`（图表线 `#14b8a6`、焦点环 `rgba(20,184,166)`）、`SubscriptionsView.vue`（emerald→teal 装饰渐变）、`platformColors.ts` `ACCENT_DEFAULT`。

### 交接文档进仓（.gitignore 例外）（2026-09-26）
> 交接协议要 `CLAUDE.md`（每会话自动加载 → 引导读 HANDOFF.md）随仓库共享，但上游 `.gitignore:122` 默认忽略它。
- `.gitignore` — 末尾加 `[CUSTOM]` 块 `!CLAUDE.md` + `!AGENTS.md`（last-match-wins 覆盖上游第 122/130 行的忽略，让交接入口进仓）。⚠️ `.claude/` 仍忽略、不受影响。上游若重排 .gitignore 需确认该例外仍在末尾且生效。
- 随之进仓的共享文档：`CLAUDE.md`（项目规则 + 交接协议）、`HANDOFF.md`（交接中枢：用户要求/准则/状态/待办）。

参考锚点：全仓搜 `HideCcsImportButton` / `hide_ccs_import_button` 就是本功能每一处的镜像位置。

## 三、上游文件的行为修改（高风险，A 方案应尽量为空）

> 直接改了上游核心逻辑的地方。每处都要写清：原行为、改后行为、为什么、
> 以及「如何验证这段行为在 merge 后仍正确」。

### 关闭 App 内在线更新 + 更新源可指向自己仓库（2026-09-22）
**为什么**：上游 App 内「立即更新」拉的是 `Wei-Shaw/sub2api` 的预编译产物，会覆盖定制镜像。
定制版只走「git merge + CI 构建镜像 + docker compose pull」升级，故从后端关掉这条路。
**改了什么**（全部集中在 `backend/internal/service/update_service.go`，带 `[CUSTOM]` 注释）：
- 原 `githubRepo` 常量（硬编码上游）→ 改为结构体字段，由环境变量 `UPDATE_GITHUB_REPO` 覆盖，缺省仍为上游常量（行为不变）。两处调用 `githubRepo` 改为 `s.githubRepo`。
- 新增字段 `disableOnlineUpdate`，由 `DISABLE_ONLINE_UPDATE=true` 开启：
  - `CheckUpdate`：直接返回 `HasUpdate=false, Disabled=true`，不联网 → 前端不出现「立即更新」按钮。
  - `PerformUpdate` / `Rollback` / `RollbackToVersion`：返回新错误 `ErrOnlineUpdateDisabled`（服务端硬拒，防直接调 API）。
  - `ListRollbackVersions`：返回空列表、不联网。
- `UpdateInfo` 加字段 `Disabled bool json:"disabled"`（前端未用，预留）。
**配置**（走 gitignore 的 override，不进仓库）：`deploy/docker-compose.override.yml` 的 `environment` 加 `DISABLE_ONLINE_UPDATE=true`（可选 `UPDATE_GITHUB_REPO`）。模板见 `deploy/docker-compose.override.yml.example`。
**merge 后如何验证**：① `DISABLE_ONLINE_UPDATE=true` 起容器后，管理员版本徽标只显示当前版本、无「立即更新」；调用 `POST /api/v1/admin/system/update` 返回 `ONLINE_UPDATE_DISABLED`。② 不设该变量时行为与上游一致（能查到更新）。🔴 本机无 Go，未编译，见下方待验证。

### 品牌默认深色主题（2026-09-26）
**为什么**：最初基于深蓝品牌设为默认深色；2026-09-26 换肤时用户明确要求保留该策略，已有用户选择仍优先。
**改了什么**（`frontend/src/main.ts` 的 `initThemeClass`，带 `[CUSTOM]` 注释）：
未显式选择过主题（`localStorage` 无 `theme`）时**默认深色**；原上游是「跟随系统 `prefers-color-scheme`」。
用户手动切换过（存了 `theme`）仍以其选择为准 —— 只改「没选过」时的兜底方向。
**merge 后如何验证**：① 清掉 `localStorage.theme` 首次访问 → 深色。② 手动切浅色后刷新 → 仍浅色。
**回退**：把该行改回 `savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)` 即可。

- 除上面两处外：无。

## 四、依赖 / 构建相关

- CI 构建镜像：`.github/workflows/custom-image.yml`（推 GHCR，见 [SYNC.md](SYNC.md)）
- 如新增前端依赖：改 `frontend/package.json` 后必须 `pnpm install` 并提交 `frontend/pnpm-lock.yaml`
  （上游 CI 用 `--frozen-lockfile`，不同步会失败）。
- 如改数据模型：改 `backend/ent/schema/*.go` 后必须 `cd backend && go generate ./ent`，
  并把**生成的代码一起提交**。

---

## 同步后自检清单（每次 merge upstream 后过一遍）

- [ ] 「二、接线改动」每一处都还在（`git log` / 直接看文件）
- [ ] 「三、行为修改」每处对应的行为仍符合预期
- [ ] `cd backend && go build -tags embed ./cmd/server` 通过
- [ ] `cd backend && go test -tags=unit ./...` 通过
- [ ] 若改过 schema：`go generate ./ent` 无未提交变更
- [ ] 若改过前端依赖：`pnpm-lock.yaml` 已同步
- [ ] CI 构建出的镜像能起来 + 冒烟测试（登录、发一个请求）

### 认证页第二次排版调整（2026-09-27，本机时间）
- 用户否决分散大屏布局后，AuthLayout.vue仅新增main.mofa-auth-content包住BrandPanel与原认证表单，保留default/footer插槽、设置读取与认证业务，带[CUSTOM]标记。
- theme.css重排为1120px统一双栏面板，左侧品牌区、右侧无嵌套卡片表单；标题42px上限、核心M180px、手机单栏。所有选择器限定mofa-auth，不更改业务页或独立品牌首页。
- upgrade-contract.spec.ts补mofa-auth-content契约；同步上游需保留此容器及原认证插槽。

### 多模型品牌标识（2026-09-27，本机时间）
- 新增 frontend/src/custom/assets/mofa-mark-flat.png：从原mofa-mark.webp透明轮廓派生，保留M／狮子／皇冠；扁平#096B68标志＋#E4F1EE底，512px PNG、16313字节。非矢量重绘，不新增依赖或上游代码接缝。
- 本机配置site_name=魔法家族、site_subtitle=多模型 API 服务，site_logo使用上述图片data URL。名称仍走上游管理员设置，不在代码里硬编码GPT或品牌名。生产未修改，部署到其他环境需配置这三项。
- 设置表单有无关字段规范化风险，保存请求被阻止，最终仅本机数据库三键事务更新并重启应用清缓存；其余277行精确比对不变，完整管理员设置响应前后也仅3字段不同。详见BRAND_IMPLEMENTATION第17节。
