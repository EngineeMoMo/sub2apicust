# 日常更新操作手册（定制版 sub2api）

> 2026-10-10安全补修进行中：用户已授权修复发布阻塞。Go最低版本及根/部署Docker构建镜像统一1.27.2，CI、安全和release版本断言同步；x/net升级0.60.0，Go模块解析同步升级crypto/mod/sync/sys/term/text/tools。保留安全扫描，不新增忽略项。隔离快照32个顶层测试及34个子场景通过，5项PG因缺DSN跳过；工作树与发布快照govulncheck均0可达漏洞（模块中仍有2项未调用的漏洞，不宣称所有依赖全0），最终发布仍须新SHA的CI/安全/GHCR成功；生产未操作。

> 2026-10-10发布检查：包号功能提交80191c2afa4a16af6e8d6c3197d2ffff46835052已推送origin/main。当前安全扫描38023308260失败，govulncheck报告12项可达漏洞，涉及现有Go1.27.0与golang.org/x/net v0.58.0；日志给出的修复版本为Go1.27.2与x/net v0.60.0。本轮未修改这些依赖，不把安全失败误记为通过，也不交付上线命令。CI38023308250与GHCR38023308257仍在运行，真实PG结果尚未确认，生产未操作。证据位于output/dedicated-release-20261010；本条覆盖下方“未提交推送”历史状态。

> 2026-10-10包号职责收敛仅源码完成，尚无本轮已核验镜像。发布后同版更新前后端，旧billing-policy接口移除、旧策略值不再生效；数据库不删表不清零，回退旧镜像会恢复旧限额。上线验收见[包号当前规则](DEDICATED_ENTITLEMENT_ONLY.md)，不沿用旧SHA的CI结论。

## 2026-10-08 收款限制与人工开通（最新可更新版本）

功能`575467a5f4d194ebd01cf771edacef9f0075d395`已推送；同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37749962687)、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37749962637)、[镜像](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37749962677)全部成功。

固定镜像`ghcr.io/engineemomo/sub2apicust:sha-575467a`（linux/amd64），摘要`sha256:5539364a4f724569a4f0ef22bf12db74357fce5b6e3669327be619137d5e408e`。备份数据库、配置和app_data后，在部署目录执行：

```bash
./update.sh sha-575467a
```

新增247迁移及共享收款限额、快捷充值配置、套餐人工开通模式。限额默认关闭；先按[收款说明](PAYMENT_COLLECTION_LIMITS.md)核对旧付款链接及线下到账，再配置单笔50（严格少于50设49.99）／每日1000并启用。支付宝交易不存在不释放预占，人工续费在关闭在线支付时仍可联系管理员。真实商户和生产交互待用户验收，未代部署。下方待发布提示为历史，不影响此终态；纯文档提交不生成新镜像。

2026-10-08审查补修未发布：支付宝交易不存在或SDK业务失败继续保留额度，人工续费使用不受在线支付开关限制的公开套餐联系入口。待新版本发布后验收，不将既有镜像当作包含修复。

2026-10-08最新源码：收款限制与人工开通已完成但未发布，新增247迁移及默认关闭的50／1000人民币共享限额、可配置快捷充值／自定义开关、套餐contact_admin与联系说明，人工发放继续扣库存。迁移、接缝、证据与待办以HANDOFF最新状态、CUSTOMIZATIONS及deploy/PAYMENT_COLLECTION_LIMITS.md为准；旧sha-70e66cf不含此功能，不能沿用旧CI结论。发布前须核对历史付款链接及线下到账，前后端同版，生产由用户部署。

## 2026-10-08 库存与帮助改进（最新可更新版本）

最终功能`70e66cf7305e0cb4610671b5120e3612f3478fb6`已推送；同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37720127131)、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37720127128)、[镜像](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37720127116)均成功。

固定镜像`ghcr.io/engineemomo/sub2apicust:sha-70e66cf`（linux/amd64），摘要`sha256:bf94b9dcfca53dd394660904566bef67adc979f8abc721d607d54fc77964440e`。先备份数据库、配置和app_data，再在部署目录执行：

```bash
./update.sh sha-70e66cf
```

包含套餐库存与人工分配扣减、帮助图标区分、六节教程分类、17题FAQ，以及“本站支持生图，需联系管理员开通权限”说明。245／246迁移会建立库存与归属账本，旧套餐默认不限量；设限额前核对历史占用，无法归属的人工记录按[库存说明](SUBSCRIPTION_PLAN_STOCK.md)处理。升级后验管理员分配／购买／售罄／取消释放及帮助页面；真实支付与客户端接入由用户验证。未代用户部署。下方未发布提示为历史记录，纯文档提交不生成新镜像。

2026-10-08帮助页改进尚未发布：更新后应看到常见问题用问号、接入教程用书本；教程六类目录与FAQ分类筛选在控制台／公开页／游客预览可用。本地56项、类型／lint／构建和四种公开夹具布局通过，不代表真实客户端接入成功。当前无本轮新镜像标签，未代用户部署。

2026-10-08同轮库存补充包含246人工库存归属迁移：管理员分配、默认赠送／兑换、延长期限同样受库存约束，我的订阅展示库存并禁用售罄续订。上线前核对无套餐／多套餐分组及历史混合来源，不能假定旧人工分配全部自动还原；操作见[库存说明](SUBSCRIPTION_PLAN_STOCK.md)。本轮尚无发布标签，仍未部署。

2026-10-08套餐库存仅源码完成，当前没有本轮可用镜像标签。待授权发布并核验对应SHA CI／安全／GHCR后，备份数据库再更新同版前后端；245迁移保留旧套餐不限量并回填现存已售／预占。库存总限额包括历史占用，补货方法及晚付人工处理见[SUBSCRIPTION_PLAN_STOCK.md](SUBSCRIPTION_PLAN_STOCK.md)。不把sha-07f0ee6当作已有库存功能，不代用户部署。

## 2026-10-07 版本权限与依赖安全修复（最新可更新版本）

功能`07f0ee69e0b9318edbdea0694e2d5ecd0ae2e24a`已推送；同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37582687152)（含Go unit／integration）、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37582687062)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37582687015)全部成功。包含版本仅管理员展示、公开版本值清空、依赖17告警清零；前端本地2928项通过。

实际镜像`ghcr.io/engineemomo/sub2apicust:sha-07f0ee6`（linux/amd64），摘要`sha256:638ef47fc3b3d7f185849f90ef0bd8cd48b38732032c20f6176962b87e14821e`，成功manifest推送与containerimage.digest已核验。

用户先备份数据库、部署配置和app_data并记录旧镜像，再在部署目录执行：

```bash
./update.sh sha-07f0ee6
```

更新后验证健康检查、管理员版本可见、普通用户与游客无徽标，公开`/api/v1/settings/public`及首页注入的version为空，原有图标／公告展示正常。本版包含此前244迁移与支付宝官方直码，仍需按下节新建订单真机验收。未代用户生产部署；下方本轮未发布提示为历史记录，纯文档skip-ci提交不替代本功能镜像。

> [CUSTOM] 2026-10-07版本权限修改尚未发布，下方已发布镜像不含该改动。发布后验证管理员版本可见、普通用户与游客无徽标，公开settings/public及首页注入version为空。详见[安全报告](SECURITY_REVIEW_20261007.md)与HANDOFF最新状态。

> 同轮依赖安全修复也尚未发布：移除未使用图标依赖链、DOMPurify3.4.16，生产依赖复扫全0、前端2928项及类型／构建通过。需以新SHA完成CI／安全／镜像门禁后再更新，不将旧sha-80c30c8视为包含本次修复。

## 2026-10-07 开通周期与支付宝直码（最新可更新版本）

功能提交 `80c30c8250d7de59da55b964251f1c437bcf6ddd` 已推origin/main，包含订阅开通周期、支付宝官方直码及SheetJS安全依赖补修。同SHA的 [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37575276507)、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37575276518)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37575276525) 全部completed/success，CI含unit／integration、Go静态检查、前端类型／关键回归、脚本及发布辅助检查。

成功构建日志确认镜像 `ghcr.io/engineemomo/sub2apicust:sha-80c30c8`（linux/amd64），摘要 `sha256:f3ed8f3fa622bc3ed629472c5c2f0d6e343723646ec030a5d50ee99ca59989f3`。前一候选ba405f1a8因SheetJS例外到期导致安全失败，不作为本轮推荐版本；现已升级官方0.20.3并移除两条例外。

先备份Postgres、部署配置及app_data并记录旧镜像，在部署目录执行：

```bash
./update.sh sha-80c30c8
```

本版新增244迁移：周／月按开通时间的7／30天周期计算，旧未对齐窗口回填当前周期，保留已用额度与到期时间；无法按新边界精确拆分的旧聚合用量保守结转到下个周期，不在迁移时赠送新额度。

支付宝必须重新创建支付订单后扫码，新码直接编码官方签名WAP地址；恢复旧订单仍显示旧码。更新后核对实际运行镜像、订阅重置时间、扫码付款页的商户／金额、小额支付后电脑自动成功与仅入账一次。真实App直达和商户权限仍待用户验收，CI成功不能替代真机验证；生产由用户部署。


## 2026-10-07 v0.2.13定制合并版（历史版本）

功能提交 `4bc5abbb53544fe89f9fc3bc001f5a69b7ca07e5` 已推origin/main；同SHA的 [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37492464883)、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37492464870)、[GHCR构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37492465165) 全部成功。固定镜像 `ghcr.io/engineemomo/sub2apicust:sha-4bc5abb`（linux/amd64），摘要 `sha256:df08bf887996ab87efe2a5f8a4c9958a925723ff9b0fd4764b30e4512dadde52`，由成功构建的manifest推送及digest输出确认；不要用中间失败候选2c282e409或a8baeb515作为更新目标。

本版含上游v0.2.13、订阅到期与重置时间提示、支付宝桌面WAP扫码与安全修复、更新脚本镜像保留，以及已认证包号跳过余额预占的兼容修复。Vue／source-map-js新增高危依赖已更新，前端安全仍按原SheetJS例外通过。更新脚本的BSD sed与Bash 3兼容性已修，macOS CI及Linux22场景通过。

先备份Postgres、部署配置、app_data，记录旧镜像；在实际部署目录执行：

```bash
./update.sh sha-4bc5abb
```

如需同步本轮更新脚本，可从仓库该功能提交的 `deploy/update.sh` 获取并替换部署目录中的脚本，保留原脚本备份。脚本代码推送不等于服务器脚本自动更新。

两项上游241迁移与本仓241/242/243按完整文件名并存，勿重命名已应用文件。升级前未使用的重置密码链接需重新申请；Key创建默认每用户200个有效Key／每小时60次，余额并发预占默认启用，有效包号仍不预占客户余额。更新后核验health、运行镜像版本、订阅时间提示、包号零余额准入，并按ALIPAY_DESKTOP_WAP.md验收实际扫码支付；本地／CI成功不等于真实业务已验收。生产由用户部署。

## 2026-10-06 v0.2.13本地合并候选（历史记录）

已合并稳定上游源码及本轮支付宝／订阅／更新脚本定制，本地检查结果见HANDOFF。尚未推送、没有本次提交对应的CI或新镜像，下面已发布版本记录保持有效。发布前须核验本次SHA；保留包号预付跳过余额预占、支付严格金额与赠金／折扣的实付和到账区分。两项上游241迁移与本仓241/242/243并存，勿重命名已应用文件；升级前备份数据库及配置。上游重置密码令牌改哈希，升级前未使用链接需要重新申请；Key创建默认每用户200个有效Key／每小时60次，余额并发预占默认启用。生产仍由用户部署。

## 2026-10-02 配方打开原图修复（最新可更新版本）

功能4bf1436fb8ecae0c0780c3b741e26cefba9dd170已推origin/main，[CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36887283643)、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36887283348)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36887283175)全部success；Go unit／integration、lint0、前端类型／lint和372项回归通过，镜像embed编译完成。实际固定镜像ghcr.io/engineemomo/sub2apicust:sha-4bf1436（linux/amd64），摘要sha256:038e95a020dc71a5615c02cec490a9a92cf8172242686977051b9686cdeae87e，经manifest推送与containerimage.digest核实。

生成后的“打开图片”改页内原图弹窗，带关闭／Esc和手机完整预览；Base64提供下载入口，远程图可长按或右键保存。仅自有配方静态资源与交接，不新增迁移；含上一版包号免扣和迁移243，旧acd73d7不含本次图片修复。61项配方＋3项打包、类型／生产构建与独立／同源iframe夹具通过；下载落地、真实线上根因与结果仍待验收，内嵌测试的一条MutationObserver错误来源未知。更新脚本专项未纳入，代理未操作8080或生产。

用户先备份Postgres、部署配置和app_data（含studio-submissions），记录旧镜像，再执行：

```bash
cd /sub2api-deploy
./update.sh sha-4bf1436
```

更新后刷新配方，用真实已生成图片检查打开图片1、多图序号、手机关闭与原图保存；无需为了验证打开功能重新请求模型。包号策略与计费验收仍按下节执行。

## 2026-10-01 包号免扣与共享限制（最新可更新版本）

功能acd73d716daa6a8e728a36fca5dce5fd523ae3f6已推origin/main，[CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36873854705)、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36873854536)、[GHCR构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36873854627)全部成功。Go unit／integration、PG、lint0 issues及前端类型／372项回归通过，镜像embed编译成功；govulncheck未发现漏洞，前端审计按既有例外通过。本次包括有效包号实扣0／余额0可用、独立参考计量、后台使用限制，以及此前同批成员多专属组、移除成员、Key恢复与旧表单保护；公共组按原规则计费，前后端与迁移243同版。293c091不包含此次包号改动。

实际固定镜像ghcr.io/engineemomo/sub2apicust:sha-acd73d7，linux/amd64，摘要sha256:b41acdce7446e1b7bca618e16e6f6dbdab5b13fd30732760247063deb576bfa7，经成功任务manifest推送及containerimage.digest核实。先完成下面的备份并记录旧镜像，再由用户执行：

```bash
cd /sub2api-deploy
./update.sh sha-acd73d7
```

更新前备份Postgres、部署配置与app_data（含studio-submissions），记录当前镜像。确认新SHA门禁全部成功后在服务器用固定sha标签更新。迁移启动时执行；更新后在包号管理配置使用限制，默认2并发／30次每UTC固定分钟／日上限0不限／2MiB／生图关闭，不是Token或美元预算。验证0余额专属组实扣0、普通组原计费、到期撤销拒新请求、共享限额和实际客户端。需要WS会话更新或实时音频的客户端须先验证兼容性；本版包号WS仅支持生成与取消。回退旧计费版本前暂停相关包号Key／账号。

宿主更新脚本专项仍未提交，本次应用镜像不会替换服务器上的update.sh。本轮未代部署生产或调用收费模型，线上历史扣款仍需流水核实；后续纯文档skip-ci提交不产生替代镜像。下方293c091及更旧版本记录为历史。

## 2026-10-01 截图与发布门禁修复（最新可更新版本）

功能提交`293c091b3ab0c8f187c031bed1955bb843939d40`已推origin/main：[CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36842744777)、[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36842744680)、[GHCR构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36842744656)全部成功。Go lint 0 issues、unit／integration与镜像embed编译成功；govulncheck无漏洞，前端审计按既有例外通过（两项SheetJS high例外仍存在，未扩例外）。本地实际提交快照644项、类型／源码lint／Vite构建通过。

固定镜像`ghcr.io/engineemomo/sub2apicust:sha-293c091`，linux/amd64，摘要`sha256:c517187b769f8b49f1fb29b1667277b745bb42deb5be05eabbf3f39e54274a06`，由该SHA成功任务的manifest推送及containerimage.digest核实。包含滚动吸顶栏、SynaRoute选中态、首页文案、47图目录及空模型列表诊断，还有Axios1.20.0、x/image0.45.0与投稿资源清理检查；视频／分镜仍待开放。下方765e5a9和d84d60e的暂缓状态是历史，更新应使用本固定标签；收尾文档skip-ci不生成替代应用镜像。

先备份Postgres、部署配置与app_data（含studio-submissions），记录当前镜像作为回退版本，再在服务器执行：

```bash
cd /sub2api-deploy
./update.sh sha-293c091
```

更新后确认健康检查，登录验收滚动栏、SynaRoute、五类新图与保留的单张时尚肖像、配方密钥／生图模型选择、投稿待审和管理员审核。线上空模型框没有取得真实账号的Key／分组／models响应，不能据模拟选择测试宣称现场根因或生图效果已确认；若仍无选项，按新版具体提示检查配置并提供脱敏状态码。包号与新版更新脚本及CI接线不包含在本次应用提交；宿主update.sh不会随应用镜像自动替换。本轮未代部署生产或调用收费模型。

## 2026-10-01 配方与工坊一体化（已推送，暂缓部署）

功能提交765e5a9516eaecf23d94202b4afc48f9bf723339已推送，[GHCR构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36825037358)成功，固定标签sha-765e5a9、摘要sha256:341886529ae61b817a24da3a2a40dbb7062aabbc2148ecc9b7956153cb422e9d。但是[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36825037421)发现x/image的WebP／VP8L可达漏洞与Axios七个high，[CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36825037384)的Go lint也有八处资源清理错误返回未检查。**镜像存在不等于发布门禁通过：暂不建议运行update.sh或拉latest更新公开服务器。** 修复后需新SHA、新安全／CI及镜像成功，再由用户备份更新；本轮未操作生产。

本机8080已使用同一应用镜像提供`/recipes/`与`/studio/`，无需两个额外服务或产品地址变量。原数据库／Redis／数据卷／业务计数及全部设置摘要保持，镜像／HTTP／实际页面证据见HANDOFF最新终态；迁移和可选独立托管详见[FAMILY_INTEGRATED.md](FAMILY_INTEGRATED.md)。

用户已授权本轮限定提交推送，实际暂存源码快照333项测试、类型／定向lint／Vite构建通过；包含最新主题／同页配置、38图提示词、投稿人工审核及视频／分镜待开放，包号和新版更新脚本不包含在本次应用发布中。**新功能SHA与GHCR发布状态以HANDOFF本轮最终记录为准，生产未操作。** 下方`sha-25071d4`和更旧标签不包含一体化；须核对新SHA的CI和镜像构建成功，再由用户备份Postgres、部署配置及app_data（含studio-submissions），更新固定sha标签。不能直接拉旧标签或latest认为已更新；真实账号配置授权／投稿审核及收费调用仍需用户测试，静态与模拟账号回归不等于真实模型效果。

## 2026-09-30 家族视觉与下载入口更新（最新）

功能提交 `25071d4d547c69d6c933bbcb0c276585b23b4b68` 已推送 origin/main：工坊三张等宽图片与鼠标／点按／键盘置顶、最新深发人像、家族首页独立皇冠图标、SynaRoute 官网与下载链接，以及发布地址配置说明。只包含 16 个家族相关文件，不包含工作区包号专项，未新增跨站 SSO 或一体化产品托管。实际暂存快照 72 项／8 文件、类型／相关 lint／Vite 构建通过；真实生产登录与模型授权仍需用户验收。

**记录时发布尚未完成**：[镜像构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36736764218)在 Build and push、[CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36736764231)的 Go Unit 仍运行；[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36736764322)以及 CI 的前端／Go lint／shell／release-helpers 已成功，不标全部通过，不把旧提交的成功结果沿用到本轮。待三项成功并核对镜像任务日志中实际标签／摘要，再备份数据库和部署配置、记录旧标签后执行：

```bash
cd /sub2api-deploy
./update.sh sha-25071d4
```

上面是本轮功能提交的预期标签，不代表构建中已能拉取；后续交接文档 `[skip ci]` 提交不生成替代镜像。三项 `MAGIC_RECIPES_URL`／`MAGIC_RECIPES_ORIGIN`／`MAGIC_STUDIO_URL` 仓库变量名称只读查询仍为空，配方／工坊入口保持待配置，需按 `frontend/src/custom/FAMILY_PORTAL.md` 设置实际已托管地址并重建。上线后验证 `/health`、三图及交互、皇冠与仪表盘图标区分、SynaRoute 公开链接、真实登录后的家族首页／API 控制台往返；代理不代操作生产。

## 2026-09-30 家族产品与接入教程版本

功能提交 `321fe9943b700384b10c2fc1368ed24c6b90eb97` 已推送origin/main：保留品牌首屏，新增互动产品展台／控制台产品坞、配方模型设置与显式登录配置、工坊独立源码及接入教程。未混入工作区待确认的包号专项修复；本机8080此前整工作区镜像不等于本次发布范围。

**发布门槛**：[本轮镜像构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36712632118)已成功，实际标签 `ghcr.io/engineemomo/sub2apicust:sha-321fe99`、摘要 `sha256:c8dd5b34edacf29eec9e30f53f35f5de8c5898706b86715b81d70b8a73db4147`，从成功任务日志核验。[本轮CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36712632123)的Go Unit、frontend、Go lint、shell、release-helpers及安全扫描成功，集成仍运行；先确认集成成功，不把镜像发布等同全部CI通过。实际暂存源码已通过210项原站、54项配方及21项工坊测试与类型／lint／Vite构建。

集成成功后，备份数据库与部署配置，记录当前标签，在服务器执行（后续仅文档的skip-ci提交不生成新镜像，仍使用此功能标签）：

```bash
cd /sub2api-deploy
./update.sh sha-321fe99
```

独立配方／工坊不随原站镜像自动托管。仓库 `MAGIC_RECIPES_URL`、`MAGIC_RECIPES_ORIGIN`、`MAGIC_STUDIO_URL` 尚未配置（只读API核对变量名称），本次生产入口将显示待配置并禁用；需先单独托管、确认HTTPS或同源路径、设置公开构建变量后重建。独立配方还需精确API CORS，不放松认证或来源限制，不把本机4178／4179地址用于生产。具体见 `frontend/src/custom/FAMILY_PORTAL.md`。

更新后核查 `/health`、`/home`、`/family`、`/api`、`/guide`及真实登录后的产品切换。真实Key授权／计费及用户视觉批准仍待，不自动调用收费模型；后台自定义 `home_content` 仍可能覆盖默认首页。代理未操作生产。

## 2026-09-30 多人包号与控制台版本

代码 `43b715a15f599f5ae2125bc923849d57e4b3522c` 已推送；GHCR及安全扫描成功，镜像 `ghcr.io/engineemomo/sub2apicust:sha-43b715a` 已生成（摘要 `sha256:dc36b117e524262faddf0c13f0dad5c9ff510e385e69ddc2860145a7ea6566af`）。

**发布门槛**：记录时[本轮CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36657101544)的集成测试仍在运行，Go单测、Go lint、前端及其他任务通过；请先确认该CI全部成功，不把镜像成功当作全量测试通过。

通过后先备份数据库和部署配置，再在服务器执行：

```bash
cd /sub2api-deploy
./update.sh sha-43b715a
```

本次新增迁移242（多成员与删除隔离），保持一账号一专属分组的独占规则。上线后检查登录隐藏游客入口、FAQ右侧正文、包号名称／多人／改绑／撤销后删除及任务费用显示。不要直接回滚旧单用户镜像；详细规则与回退边界见[包号排查说明](DEDICATED_TROUBLESHOOTING.md)。生产尚未由代理操作。

## 2026-09-28 游客功能已验证版本

代码提交 `fa45b2023c01e5c84084dcbcb32d5a7c5bbe213a` 的 CI（含Go单测／集成）、安全扫描及GHCR构建全部通过，发布镜像 `ghcr.io/engineemomo/sub2apicust:sha-fa45b20`。本次包括公开官网／真实套餐／FAQ、游客控制台及宽屏手机适配；不开放个人数据接口。

服务器由用户操作：先备份数据库与部署配置，并记录当前镜像标签，再执行：

```bash
cd /sub2api-deploy
./update.sh sha-fa45b20
```

更新后检查 `/health`、首页游客入口、`/preview`、`/plans`、`/faq` 和登录回跳；如配置了自定义 `home_content`，其优先级仍保留。未执行生产更新或真实支付验收。后续仅文档提交不生成新镜像，不能替换此处已验证标签。

> [CUSTOM] **Windows本机源码构建**另见 [LOCAL_DOCKER_RUNBOOK.md](LOCAL_DOCKER_RUNBOOK.md)：已记录Docker Hub授权超时与代理预检。本文下方仍是服务器镜像更新流程；不要把本机127.0.0.1:7897代理配置带到生产。

> 速查卡：以后升级 / 回滚照这份敲即可。完整部署、首次平移、数据备份/恢复见 [DEPLOY_CUSTOM.md](DEPLOY_CUSTOM.md)。
>
> 你的环境（已确认）：部署目录 `/sub2api-deploy`，compose 用 `docker-compose.local.yml`，
> 定制镜像 `ghcr.io/engineemomo/sub2apicust`（owner 是 EngineeMoMo，中间**两个 e**），DB 用户/库均 `sub2api`。

---

## 更新分两步

### 第一步：出新镜像（代码侧，一般由维护者 / AI 协助）
上游出新版、或要改定制（换色 / 加功能）时 → 在 fork 仓库合并或改代码 → `git push` → CI 自动构建新镜像推到 GHCR。
（需要 Go 编译校验，本机若无 Go 就靠 CI。）构建完成后你会拿到新的镜像标签 `sha-xxxxxxx`。

### 第二步：服务器部署（你操作）
进部署目录：
```bash
cd /sub2api-deploy
```
按 override 里 image 钉的是 `latest` 还是 `sha-xxxx`，选一种模式：

**模式 A ｜ 用 latest（省心，日常一条命令）**
override 的 image 是 `...:latest` 时：
```bash
./update.sh
```
每次自动拉最新定制镜像并重建。

**模式 B ｜ 钉 sha（可控，当前就是这个）**
override 钉了固定 `sha-xxxx` 时，不带参数的 `./update.sh` 只会重拉同一版、**不升级**。
升级到新版本要带新 sha（维护者会给你）：
```bash
./update.sh sha-<新提交>
```
脚本会自动切 override 标签 + 备份原文件，再拉取重建。

**从模式 B 切到模式 A**（以后只敲 `./update.sh`）：把 override 重写为 latest：
```bash
cat > /sub2api-deploy/docker-compose.override.yml <<'EOF'
services:
  sub2api:
    image: ghcr.io/engineemomo/sub2apicust:latest
    environment:
      - DISABLE_ONLINE_UPDATE=true
EOF
```

`update.sh` 每次都会：记录更新前镜像 → 拉镜像 → 重建容器 → **启动时自动跑数据库迁移** → 等 `/health` 就绪 → 定向清理本项目旧镜像。

### 旧镜像保留规则（2026-10-01）

- 仅处理 `ghcr.io/engineemomo/sub2apicust` 的 `sha-*`／`latest` 标签，以及带本仓来源标签的悬空镜像；保留当前镜像、实际更新前镜像和所有其他运行／停止容器引用的镜像。手工 `v*` 等标签保持，不清数据库、Redis、日志、备份、数据卷或其他项目镜像。
- 更新成功后，更新前镜像增加本地 `rollback-previous` 标签，支持只拉 `latest` 时保留旧版；重复更新同版时不覆盖已有回退镜像。首次使用新版脚本却没有发生版本切换时，保留按镜像创建时间排序、比当前更早的最近一个 `sha-*` 镜像作为回退候选，这不等于已验证该候选曾运行成功。
- `/health` 超时返回非零，跳过全部清理；拉取／重建失败同样不清理。检查失败时保留相关镜像并输出原因，删除不使用强制参数。`update-before` 是拉取前的保护标签，更新成功时移除；中断／失败时可能保留用于人工排查。
- 部署目录 `.sub2api-update.lock` 防止同目录并发更新，正常退出自动移除。被强制终止时可能遗留；确认没有更新进程后，使用 `rmdir /sub2api-deploy/.sub2api-update.lock` 移除空锁目录再重试。
- **服务器必须单独替换 `/sub2api-deploy/update.sh` 为本仓新版，并执行 `chmod +x /sub2api-deploy/update.sh`。仅拉取应用镜像不会更新宿主机脚本。** `rollback-previous`／`update-before` 为服务器本地标签，不是GHCR发布标签；指定版回退继续使用原 `sha-*` 标签，不把本地标签传给会执行pull的更新命令。
- 回归：`bash -n deploy/update.sh`、`bash deploy/tests/update-test.sh`。后者仅使用假Docker，不连接daemon；当前22个升级、latest（含无sha别名）、同版、回滚、仓库切换、容器占用、检查／删除失败及并发锁场景通过。真实服务器回收量需升级后用 `df -h` 和 `docker image ls ghcr.io/engineemomo/sub2apicust` 核验。

---

## 回滚
切回上一个正常的版本：
```bash
cd /sub2api-deploy && ./update.sh sha-<上一个正常的提交>
```
迁移是前向增量的，回滚镜像通常安全（旧代码忽略新列）；只有怀疑迁移本身出问题才用数据库备份恢复，见 DEPLOY_CUSTOM.md 第七节。

---

## 每次升级后验证
```bash
docker compose -f docker-compose.local.yml -f docker-compose.override.yml exec -T sub2api /app/sub2api --version
```
- 版本号是你要的 `0.2.8-custom.xxx`；
- 浏览器：主题正常、老数据都在、后台**无「立即更新」按钮**；
- 设过 logo / 站点名的，检查显示正常。

---

## 常见坑
- **`pull` 报 not found**：多半是镜像 owner 拼错——正确是 `engineemomo`（EngineeMoMo，中间**两个 e**），别写成 `enginemomo`。
- **`./update.sh` 跑了但版本没变**：override 钉了固定 sha（模式 B），要 `./update.sh sha-<新>` 或切 latest（模式 A）。
- **`pull` 报 unauthorized / denied**：GHCR 登录过期 → `docker login ghcr.io -u EngineeMoMo`（用有 `read:packages` 的 PAT；**别用 `echo <token> |` 的写法**，token 会进 shell history）。
- **升级前想备份数据库**：
```bash
docker compose -f docker-compose.local.yml exec -T postgres pg_dump -U sub2api sub2api | gzip > backup-$(date +%F).sql.gz
```
