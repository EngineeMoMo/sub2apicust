# 交接文档（HANDOFF）— sub2api 定制 fork

> **本文件是本项目的交接中枢。** 每个接手的智能体（含 Claude）：
> - **开工前必读**：本文件 + [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md) + [SYNC.md](SYNC.md)。
> - **收工前必更**：更新下方「四、当前状态」「五、待办」；任何新定制同时登记 CUSTOMIZATIONS.md。
> - 交接一律走本文件——**别把状态只留在会话里**（会话会丢，文件不会）。
>
> 最后更新：2026-10-19（Asia/Shanghai）。

---

## 一、用户是谁 / 怎么协作
- 语言：**全程中文**（回复、文档、注释）。
- 自部署运维者，目标：把开源 sub2api 定制成自有品牌的商用「GPT中转站」，并长期跟随上游。
- 用户**手动部署**；智能体负责改代码、验证、提交推送，**不代替用户上生产**。

## 二、用户要求（项目意图）
1. **核心**：定制 sub2api 的同时能持续同步上游更新、且定制不丢。
   路线 = **方案 A：配置 + 新增功能**，尽量不动上游核心 → 私有 fork + override 构建。
2. **镜像**：GHCR + GitHub Actions CI（push main 自动构建）。
3. **已交付定制**（细节见 CUSTOMIZATIONS.md）：
   - SynaRoute 深链接一键导入（`synaroute://`，镜像上游 CCS 导入，带管理员开关 `hide_synaroute_import_button`）。
   - 关闭 App 内「在线更新」（`DISABLE_ONLINE_UPDATE=true`；可选把更新源指向自己仓库 `UPDATE_GITHUB_REPO`）。
   - 前端定制叠加层 `frontend/src/custom/`（换肤 + 新增/影子替换页面，低冲突叠加层）。
   - **当前主题**：雾钛青（primary-500=#096B68），Logo/站点名仍走**管理员后台设置**。2026-10-01 用户新要求覆盖历史默认深色：无保存偏好默认浅色，已有偏好保留，子产品跟随 API 当前主题。源码与实际部署版本以第四节最新记录为准。
4. **品牌**：魔法家族 / GPT中转站；域名 https://ai.mofamilys.com ；保留核心「M」设计，允许标志重着色。**2026-09-26 最新确认：雾钛青 + 日夜配套；SynaRoute 保留紫色产品识别，统一基础规范，不全换青色。** 此前深蓝发光风为历史方向。
5. **升级方式**：**手动 Docker 更新**（`deploy/update.sh`），不用 App 内按钮。

## 三、用户准则（红线，务必遵守）
1. 🔴 **只说有证据的话；没证据就说「我不知道」**（2026-09-07 用户原话：「以后都拿证据事实说话，不要理论和以为，不确定的就不回答，找到实质性证据再回答」）。
   - 回答分栏：**已核实**（附取证方式：grep 命中数 / 读了哪个文件哪几行 / 命令完整输出）/ **推断**（附「要什么证据才能确认」）/ **不知道**。
   - 「代码里存在能产生该症状的路径」≠「用户遇到的就是它」；用户报的现场要在**那台机器**上取证。
   - 写「已核对」必须同时写核对方式，否则别写那三个字。
2. **定制维护**：优先新增文件；不得不改上游文件时改动尽量小并打 `[CUSTOM]` 注释，且**逐处登记 CUSTOMIZATIONS.md**。配置类定制（.env / override）走 gitignore，不进仓。
3. **本机无 Go 工具链** → 后端改动**未编译**，靠 CI 验证（`gofmt -w ./...` + `go build -tags embed ./cmd/server` + `go test -tags=unit ./...`）；提交时如实标注「未编译」。
4. **同步上游**：先 `git merge-tree --write-tree main upstream/main` 演算冲突，再真合并；rerere 已开；合并后过 CUSTOMIZATIONS.md 末尾自检清单 + 登录发一个请求冒烟。
5. **换肤**：品牌主色只改 `frontend/src/custom/theme.css`；上游组件里的 `teal-*` 若是**语义分类色**（deepseek、模型徽标、渠道状态）**刻意不改**（改了会撞色/丢语义）。
6. 会话上下文超 **200k** 先压缩再继续，压缩后先核对工作树再接着做。

## 四、当前状态（每次收工更新）

- **2026-10-19 智力效率发布准备（用户已授权）**：用户明确要求提交并推送本轮代码，范围为 v6 兼容、成功快照持久化、故障恢复／200 组合边界及本功能文档；其他会话产品／海报工作不混入。已 fetch 核实发布前 HEAD 与 origin/main 同为 7ee076e953506beaa754e066360372f08021bb3c，本地已完成 21 组／53 子用例、前端 38 项、类型／lint／Vite／Go embed。下一步记录本轮业务提交及其 CI／安全／GHCR 终态，生产不代部署；下方本地阶段“未授权／未提交”不再作为发布授权判断。

- **2026-10-19 智力效率可用性构建收口**：本轮最终 Go 1.27.2 后端 21 组／53 子用例、前端 38 项、类型／定向 ESLint、Vite 生产构建及随后包含新静态资源的 `go build -tags embed ./cmd/server` 全部成功。智力页面产物 `IntelligenceView-uaCUL_Nu.js`；Vite 仍有大于 500kB 的 chunk 提示，构建退出 0。证据 `output/intelligence-resilience-tests-20261019.jsonl`、`output/intelligence-resilience-{frontend,typecheck,eslint,vite-build,build}-20261019.log` 及 `output/intelligence-resilience-verification-20261019.json`。这是当前工作区本地验证，包含其他会话保留源码，不作为已隔离发布的镜像；未提交、推送或操作生产，本轮 CI／GHCR 与真实数据卷可写性、历史快照落盘及登录浏览器验收仍待。下条“生产构建收口”已由本条完成。

- **2026-10-19 智力效率可用性补强（接续 v6 修复，未发布）**：用户要求排查其他同类问题并避免再次整页不可用。本轮范围为智力效率读取链路：核实进程重启丢失全部内存快照、浏览器故障后固定等待十分钟、后端 64 组合上限与前端 200 不一致；已补自有 `custom_intelligence_storage.go`，经校验的成功快照写入现有 `pricing.data_dir/intelligence-cache/last-success.json`（默认 Docker 为 `/app/data/intelligence-cache/last-success.json`），临时文件同步后重命名替换。重启读回保留原取证时间并标记历史数据；未知契约／源站 503 不覆盖好文件，源站恢复后重新写入。损坏／超限／不同计分规则或费用口径的文件拒绝使用但不阻断新读取；写入失败只记日志，在线读取仍可成功。前端故障／历史或部分数据一分钟复查本站缓存，恢复后十分钟；隐藏页不请求，卸载清理。组合容量对齐 200，保留并发 3／单轮 40 秒及指标校验；日志现在带截断的实际汇总／计分／策略版本。唯一既有上游接缝是 router 的 `[CUSTOM] RegisterIntelligenceRoutes` 增传 `cfg.Pricing.DataDir`，已登记 CUSTOMIZATIONS 与 custom/UPGRADE。Go 1.27.2 后端 21 组／53 子用例、前端 38 项、类型／定向 ESLint 与后端 embed 编译通过，生产构建收口见同日下条补充；证据 `output/intelligence-resilience-tests-20261019.jsonl`、`output/intelligence-resilience-{frontend,typecheck,eslint,build}-20261019.log`。本轮未提交推送、未操作运行中 Docker／生产；正式 CI／GHCR 待本轮发布。缓存需至少成功保存一次且数据目录保留，不能承诺从未成功读取／磁盘不可写并遇源站故障时仍有历史数据，也不自动解释未知的新评分规则。

- **2026-10-19 智力效率不可用排查与本地修复**：实时匿名读取源站绑定和 `gpt-6.1-sol/low` 汇总均 HTTP 200，但汇总 schema 已变为 `radar-bench-public-summary-v6-150-details`，原后端只接受 v5。真实响应回放修复前稳定返回 503／`summary version`，修复后 200；已读取源站两个展示模块确认原计分规则、题库、150 满分、费用／耗时聚合未变。自有 handler 精确兼容 v5／v6，保留未知契约拒绝及原校验；新增真实响应 fixture 和七类异常回归，Go 1.27.2 本地后端 15 组／31 子用例、前端 33 项及 `go build -tags embed ./cmd/server` 均通过。证据 `output/intelligence-v6-{before,after,frontend,build}-20261019.log`、`output/intelligence-v6-final-tests-20261019.jsonl`、源站原文 `output/intelligence-summary-live-20261019.json`，变更见 CUSTOMIZATIONS 同日条目。本轮利用已有便携工具链完成真实编译，覆盖历史“无 Go 未编译”判断；embed 编译使用现有静态资源，只证明本轮后端可编译，不作为生产页面已更新的证据。保留其他会话未提交工作；本轮未提交推送或部署，生产运行版本、认证接口返回及日志未取证，不能宣称生产已恢复或唯一根因。

- **2026-10-19 上游v0.2.17同步发布终态（优先）**：2026-10-19上游v0.2.17同步发布终态：业务5233e9090698a9bd02f645340cd19e85e7e27743已推origin/main，包含upstream/main=b0a02a3304a0a78cdd003ddda1dd23ba02c03084及新增47条上游提交；同SHA CI 39196526010（Linux unit/integration、专用PG18/Redis8、Go lint 0 issues、前端53文件831项）、安全39196525919、GHCR 39196526008全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-5233e90（linux/amd64），index digest sha256:08fe2367c8241ca9b43fe741b1ccd03db65ba5a8a5579f8bbc4a8f4ce6d2c550；registry revision/source/摘要与正式embed构建日志交叉验真，正式镜像隔离28项HTTP检查含真实登录、角色权限及IP防护通过。无新增迁移或依赖；312个自有新增文件内容不变，支付两处冲突保留WAP期限/快捷金额定制，原未提交产品工作不混入。生产由用户备份后手动./update.sh sha-5233e90，未代部署或付费生图。PR8028仍未合并，本轮不包含，不能保证解决截图无图。证据output/upstream-sync-20261019/verified-release.json、release-source-proof.json、formal-smoke.json及job日志；Windows全量5项失败如实保留在backend-unit-summary.json，正式Linux CI通过覆盖本轮发布门禁。后续纯文档skip-ci提交不改变此业务镜像。

- **2026-10-19 同步过程与边界**：先完成merge-tree演算，再隔离合并；payment/types.go保留time导入原[CUSTOM]标记，payment_order.go在上游OrderType白名单balance分支保留快捷金额校验，其余自动合并接缝逐项核验。原42个未提交文件1218177字节备份在output/upstream-sync-20261019/before-main；其中40个业务/产品路径均未进入本轮提交，两份共享交接文件另保留原未提交内容。仅正式上游main，不含开放PR8028。Windows缺sh三项及两项协议/时序断言失败均保留原日志，未为此改无关源码；正式Linux同SHA完整CI已成功。

- **2026-10-18 登录400缺失人机凭证漏计修复（已发布sha-117e9a1）**：2026-10-18认证IP缺凭证计数发布终态：业务`117e9a1a08f1a225dbe5b041ae243be29251770b`已推origin/main；同SHA CI `39138612694`（unit／integration、专用PG18／Redis8、Go lint 0 issues、前端53文件830项）、安全`39138612602`、GHCR `39138612579`全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-117e9a1`（linux/amd64），index digest `sha256:166ede3f3a5aa77028b4479f754e353442285fa28b466d426a0194c5eb8e6896`；registry revision／source／摘要与正式embed日志交叉验真，IP管理chunk `AuthIPBlacklistView-b89c0411.js`。缺失必要人机凭证的400及包含captcha／TOTP字段名的格式错误现按规则累计，正常挑战失败／腾讯灾备票据／服务故障／成功操作仍不计数；原自动开关／阈值／范围／期限不变，无新迁移、不回灌历史。用户备份后手动`./update.sh sha-117e9a1`并确认自动规则。未代部署、未实际封禁189.24.123.4；生产私有规则／实际版本／历史响应原因仍未核实，复现路径不等于现场唯一根因。证据`output/auth-login-probe-20261018/verified-release.json`、`release-source-proof.json`、`local-verification.json`及job日志。本轮便携Go本地embed通过，覆盖历史无Go未编译判断；后续纯文档skip-ci不改变业务镜像，其他会话工作保留。 截图21:50:43–21:50:51至少5次登录400，详情尝试admin@sub2api.local且没有人机字段；公开settings/public确认Turnstile开启。本地真实Login handler验证前9次400、第10次403/IP_BLOCKED；相关后端631顶层＋321子项、IP页16项、本地类型／lint／embed通过，既有钉钉禁用项跳过。现有审计不保存响应reason，未发送生产登录或自动开启规则。
- **2026-10-18 订单核账菜单发布终态（优先于下方阶段记录）**：2026-10-18订单核账菜单发布终态优先：业务提交`57aad622fa4cdeea90df1d22866b6b514a92e94b`已推送origin/main；同SHA CI `39135534896`（unit／integration、Go lint 0 issues、前端53文件830项）、安全`39135534923`、GHCR `39135534858`全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-57aad62`，平台`linux/amd64`，index digest `sha256:f1e11005ea3fbd2f7beeeeff44ad94b2e8794568c8c71be64f6f2ee86515120c`；registry revision／source／平台／摘要与正式构建日志交叉核验，新页面chunk `AdminOrderReconciliationView-T4m7REbz.js`及同提交embed后端均已打包。新入口为管理员“订单管理 → 订单核账”，设置→支付保留限额／自动清理规则／联系方式及跳转；无新增迁移，原付款／等待期／审计保护不变。未代部署生产或处理真实订单，真实管理员／手机视觉／商户核账仍待用户更新验收。证据`output/order-reconciliation-release-20261018/verified-release.json`、`release-final-verification.json`、`release-source-proof.json`及job日志。后续纯文档`[skip ci]`提交不改变此业务镜像，下方未提交／待发布为历史阶段。

- **2026-10-17 独立订单核账菜单（本地完成，未提交发布）**：按用户“人工订单别藏在设置里”的要求，新增管理员“订单管理 → 订单核账”（`/admin/orders/reconciliation`），集中今日收款／占用、最近100条待确认单、立即批量复查、渠道关单、等待期后的人工释放及线下收款登记；设置→支付只保留限额／自动清理规则／联系方式和跳转入口。拆出定制 `PaymentCollectionOperations.vue` 与 `AdminOrderReconciliationView.vue`，实际守卫限制管理员与支付开关；未改后端释放时机、付款保护或审计，不处理真实订单。4文件82项相关回归、typecheck、本轮7个源码／测试文件ESLint及生产构建通过（i18n检查、vue-tsc -b、Vite；仅有既有大包提示），隔离产物与证据在`output/order-reconciliation-20261017/`（verification.json记录源码哈希及命令结果、build.log）。上游仅侧栏两行[CUSTOM]及实际守卫测试接缝，详见CUSTOMIZATIONS新节。尚未提交推送／发布新镜像／部署，浏览器真实登录／手机视觉／商户核账待验；其他会话未提交改动保留，下方sha-16ef917是上一次支付补丁，不含此菜单。

- **2026-10-17 支付修复发布终态（优先于下方阶段记录）**：最终业务提交`16ef9172cf666e6361f17161aa83a53178ee0f3f`已推送origin/main；同SHA CI `39029645317`（unit／integration、专用PG18、Go lint 0 issues、前端53文件824项）、安全`39029645211`、GHCR `39029645261`全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-16ef917`，平台`linux/amd64`，index digest `sha256:49b4285ec235151d36459e3cdd49320103fffa0864d1c552351378a43ef64d07`；registry revision／source／平台／摘要及正式embed构建日志已交叉核验。更新后在设置→支付开启“原支付通道已删除时，超期占用自动清理”并保存（默认关闭）；已取消／过期／失败、未记录付款且原实例确实不存在的历史占用，在期限＋配置等待（默认24小时）＋10分钟后按每60秒最多20笔处理，失败仍有30分钟退避，可点一次“立即批量复查”。只释放占用，不充值或发订阅，实际收款须商户账单核实。本地补修后104项支付叶子回归0失败0跳过、最终Windows embed编译通过，测试tmpfs数据库核验完整ID与挂载后已清理。证据`output/payment-auto-release-20261017/verified-release.json`、`release-final-verification.json`及同目录job日志。首轮2af3c432c的unused已补修并保留失败记录；不使用首轮候选交付。未代部署生产或处理真实订单，其他任务未提交工作保留。后续纯文档`[skip ci]`提交不改变该业务镜像。

- **2026-10-17 支付补丁首轮门禁补修**：2af3c432cc2fb51e6638e5d20f3ac49422492498已推送origin/main，安全39028834273与前端53文件824项通过；CI39028834186的Go lint任务117385058916发现定制复查中保留的 `customDueCollectionHolds` 包装函数仅在unit标签测试使用，无标签静态检查判为unused。已合并为唯一查询入口，显式传入复查时间，同步更新既有SQLite和PG测试调用，等待期／退避／批次数与付款保护不变。失败日志保留 `output/payment-auto-release-20261017/job-117385058916.log`；首轮候选不作为最终交付，补修后重新核验新业务SHA全部门禁。

- **2026-10-17 支付补丁发布授权（进行中）**：用户明确要求“提交并推送代码，我更新下”。本轮仅发布历史支付实例删除保护、缺通道超期占用自动清理及一次批量复查，保留其他聊天的官网／工坊／配方／海报等未提交工作。待发布14个支付源码文件SHA256与本轮104项后端叶子用例、9项前端及构建通过的记录一致；推送前fetch确认HEAD与origin/main均为964bdcd204d8e8f846cd481590fe4534f495e07b。提交后核验该业务SHA的CI／安全／GHCR与固定镜像来源，最终标签和门禁随后补记；不代替用户部署生产。

- **2026-10-17 积压收款占用可自动分批清理（最新支付状态，本地已验证，未发布）**：用户截图显示大量CANCELLED旧单仍占用，要求不用逐单点击。已在定制收款配置新增 `missing_provider_auto_release`，默认关闭，管理员保存开启后，原通道实例确实不存在、已取消／过期／失败且未记录付款的订单，在期限＋配置等待（默认24小时）＋10分钟后自动清理；通道仍存在而查询／配置失败、引用缺失／冲突及已付款均保留。事务内重新校验当前绑定／期限／状态／付款，释放与 `COLLECTION_HOLD_AUTO_RELEASED` 审计同事务，失败回滚；只释放占用，不充值或发权益。现有worker每60秒最多20笔，失败保留30分钟退避；新“立即批量复查”按已保存策略立即处理一批、只跳过退避。新增管理员鉴权组内 `/admin/payment/collection/recheck` 的 `[CUSTOM]` 路由接缝，前端仅定制组件和类型变更，登记CUSTOMIZATIONS最新节。最终后端104个叶子用例通过（service80、repository24），0失败0跳过，包含隔离PG18完整迁移、45笔20／20／5／0批次、审计失败回滚、迟到付款状态更新及10轮付款更新并发；前端9项、类型／ESLint／i18n3项／构建和Go1.27.2 Windows最终embed编译通过，证据 `output/payment-auto-release-20261017/verification.json`。首轮夹具未注入loadBalancer导致的失败日志保留，补齐后通过。临时PG按完整ID及挂载核验后清理；未提交推送、发布镜像、启动／更新业务站点或操作生产数据。**订单47真实收款和实例3删除经过仍未知；数据库更新测试不能当作已删通道真实webhook验签证据。下方仅人工释放是前阶段描述，新开关关闭时仍维持原行为；现场启用后可自动处理符合条件的积压，无需逐单。**保留其他聊天当前未提交工作。

- **2026-10-17 支付通道删除保护已修复（本地已验证，未发布）**：用户后续要求修复。新增 `custom_payment_provider_guard.go`，删除前按订单列或快照绑定检查账本未付款、未释放占用，覆盖已取消／过期／失败订单；有占用返回409及后台核账提示，查账失败拒绝删除。上游 `payment_config_providers.go` 的 `DeleteProviderInstance` 仅增加4行 `[CUSTOM]` 接缝，原进行中保护保留，无迁移／前端变更，登记于CUSTOMIZATIONS最新节。有效旧代码基线13个叶子用例失败（12种终态绑定＋查账失败仍删除），修复后服务专项45个叶子用例通过；该阶段唯一PG跳过已在隔离真实PostgreSQL18阶段补执行，真实PG完整迁移链及既有占用／人工释放回归共19个叶子用例通过、0跳过。Go1.27.2 Windows/amd64 embed编译成功，格式／diff检查通过；证据 `output/payment-provider-guard-20261017/verification.json`，首次SQLite共用夹具并行冲突的无效基线也单独留档。测试PG使用tmpfs、核实完整容器ID后清理，未启动或更新业务应用；未提交推送、发布镜像或部署。**本地删除保护不能恢复生产已删除的实例3；订单47真实付款仍未知，须商户核账后按既有等待期人工释放。**下方“仅源码取证”是前一轮状态，原其他未提交工作保留。

- **2026-10-17 订单47收款预占复查报错排查（仅源码取证）**：用户日志为 `resolve collection hold provider failed error=load order provider instance: order 47 provider snapshot instance 3 is missing`。本地 `payment_order_provider_snapshot.go:101` 按订单历史绑定直接 Get 支付通道实例，只有记录不存在分支才生成该 missing 文案（104行），不检查 enabled；`payment_order_lifecycle.go:410` 包装后由 `custom_payment_collection_release.go:247` 在后台预占复查中记录并跳过该单。筛选条件为已过期／取消／失败、未付款且预占未释放（261行），失败先记复查时间，正常按30分钟退避；因此该次不释放占用，不能从日志断定真实收付款。已确认代码存在删除保护覆盖缺口：`payment_config_providers.go:100,164,524` 仅检查 PENDING／PAID／RECHARGING 后硬删除，没有检查已结束但仍有预占的订单；这条路径可以造成历史绑定失效，但未连接报错服务器，不能认定实例3实际由谁／何时删除，也未核实报错部署SHA。已有后台设置→支付的 `PaymentCollectionSettings.vue:26`「检查待确认订单／人工释放」入口，须先在商户侧核实未收款、经过等待期并填写原因确认，后端 `CustomForceReleaseCollection` 写审计；当前配置默认等待24小时另加10分钟，实际以后台显示为准。读取了 missing 拒绝回退与人工释放既有测试，本轮未执行测试；未修改业务源码、数据库、配置，未提交／推送／部署，保留工作树其他未提交工作。

- **2026-10-15本轮分组补修发布终态（优先于下方阶段记录）**：最终业务提交`0e99748ba82575442aae95bda1f6358a43c46ddf`已推送origin/main；同SHA CI `38750224990`（unit／integration、专用PG18、Go lint 0 issues、前端53文件821项）、安全`38750225143`、GHCR `38750224942`全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-0e99748`，平台`linux/amd64`，index digest `sha256:9b3b82a8daeb04fa7d0816825f2401e9732c027746ec0ea62d25e2f876525251`；registry revision／source／平台／摘要与构建日志交叉验真。含R1–R4、管理员当前活动倍率及基础说明、分组到期事务平移后软删除与迁移258；福利时段结束仍只恢复基础倍率，沿用原计费逻辑。**新版启动会补处理已有到期规则，用户须先备份数据库／配置／app_data并审阅源组及平移目标，再执行`./update.sh sha-0e99748`。**生产由用户手动更新，真实收费与到期平移现场验收仍待；20:30历史故障根因未知。证据在`output/group-access-fix-20261015/verified-release.json`、`release-source-proof.json`及同目录job日志；首轮c4fa7fd的ineffassign失败已补修，旧候选不作交付。后续`[skip ci]`纯文档提交不改变镜像业务版本。

- **2026-10-15首轮发布门禁补修**：c4fa7fd809fac23773fa26b10ad08a2bed89ad83已推送，安全38749182272和前端53文件821项通过；CI38749182118的Go lint发现gateway_handler.go错误兜底先clone再覆盖的1处ineffassign。已移除无效赋值及失去调用的旧clone helper，沿用统一PrepareCustomGroupFallback返回的Key副本；目标资格与计费行为不变。失败日志保留output/group-access-fix-20261015/job-116470740717.log。c4fa7fd候选镜像不作为正式交付版本；补丁提交后重新核验同SHA全部门禁和镜像，最终固定标签随后补记。
- **2026-10-15用户已授权本轮提交推送及构建新镜像**：提交范围为R1–R4、管理员当前倍率显示、分组到期持久平移后软删除及迁移258，生产仍由用户自行部署。提交前已重新对比验证记录50个文件SHA256，源码未变；沿用已通过的本地门禁，不把旧794479760的CI／镜像当本轮结果。按本次实际提交核验CI（含真实PG／integration／lint）、安全与GHCR，全部成功且registry revision／source／平台／摘要一致后才交付固定更新标签。发布终态随后补记；下方未授权／未提交是此前阶段记录。
- **2026-10-15本轮最终本地验证**：后端5包41顶层／137含子项（含23项到期专项）全部通过，真实PG18实际执行、零失败零跳过；另6项handler模型／上下文回归通过。前端完整53文件821项通过，其中分组31项，最终类型检查通过；全量lint为0错误／26项已有警告，生产构建成功。Linux58包11970顶层／23371含子项通过、51个测试因外部服务等条件跳过（另51包无测试，不算测试跳过），embed编译退出0；这些PG场景已由前述真实PG专项覆盖。Linux构建期间仅一个已跳过PG测试的最小schema补齐deleted_at，业务源码未再变化，随后最终真实PG137项通过。证据output/group-access-fix-20261015/verification.json及linux-expiry。**未提交、未推送、未发布或替换本机／生产应用**；下方103／29为追加到期功能前的阶段数据。
- **2026-10-15用户追加语义：福利结束与分组到期分开处理（本地补修，未发布）**：已沿`APIKeyService.GetAvailableGroups → CustomGroupRate → rateAt`核对后台真实取值，正式边界回归验证时段开始前1、起点／时段内0.8、终点1，且数据库基础值不改；线上下拉观察不能替代真实收费账单。原分组到期只有请求内路由，现新增`custom_group_expiry_store.go`／`custom_group_expiry_worker.go`：后台按最近到期时间唤醒，启动补处理，保存和跨实例通知唤醒，最长30秒兜底；事务内将绑定Key的用户和显式授权用户平移到目标、持久修改未删除Key，再按既有语义软删除原组。保留Key字符串／状态／额度／有效期、用户余额和历史账单，目标现有专属倍率不覆盖，上游账号不搬。失败全部回滚、原组保留；同平台按量组适用，订阅／包号禁止自动合并权益。新增迁移258防并发Key／授权绑回已删组，规则保存亦防删组竞态；257不改。真实PG测试已覆盖事务、并发、回滚、到点／启动／通知唤醒与链式平移，最终全量结果见verification.json。**新版启动会处理现有已到期配置，升级前须审阅并备份；未改线上规则或部署。**本条及下条补修优先于20:26历史审查。
- **2026-10-15 21:17线上test活动与列表差异已核实，历史故障不能倒推**：用户配置test于北京时间21:06:19–21:17:27为0.8，管理列表基础值为1。使用独立临时浏览器页只读查看「API密钥→创建弹窗→OpenAI分组下拉」，活动期间实际显示test 0.8x；21:17:27结束后完整刷新并重新读取，显示test 1x。没有创建Key、改真实规则或发收费请求；test截图0账号，因此真实计费未知。管理列表源码只读取基础rate_multiplier，是确定的显示缺口，不能据此推断后端收费未切换。未抓到21:06:11–21:06:19空档；20:30原配置已被用户在失败后删除，当前开关关闭是后续处理，不能当作历史根因。证据output/group-access-fix-20261015/production-observation.json。
- **2026-10-15分组显示与R1–R4本地补修，未提交／发布／部署**：新增管理员批量rate-schedules只读接口和custom显示组件，按服务器对时的页面共用时钟更新当前倍率，活动下标保留基础值；保存后刷新，跨页面修改每30秒批量同步，隐藏页停止配置轮询；读取失败明确未知。只读显示不写回基础倍率、不新增定时改价任务，后端仍在新请求中以当前时间选择倍率并沿用原计费规则。既有降级统一校验目标资格／到期／活动／RPM，Messages识别客户端后同步目标订阅、上下文、模型白名单、分组策略及渠道映射；通知消费与心跳解耦并可关闭原连接解除等待；同组保留或按目标缓存RPM覆盖（含空值）；已拥有图片GET查询保留原鉴权并不受新金额门槛／到期切组拦截。迁移257未改；不是线上已修复。专项103项（32顶层、5包、真实PG18零跳过）与另6项handler目标模型／上下文通过，前端分组29项、类型及定向lint通过；最终全量结果见output/group-access-fix-20261015/verification.json。本机已核实Go1.27.2可用，旧「本机无Go」描述过期。

- **2026-10-15 20:26 分组规则审查发现4项未修复问题（本条优先于下方发布／效率结论）**：审查基线be769bf34，业务源码同sha-7944797。R1/P1既有Claude客户端降级可绕过目标余额／月充值门槛，实际调度到目标账号；R2/P1通知消费者延迟1.1秒且积压1000通知时，同步Ping与pq通知队列互等，4秒后健康租约过期，缓存持续回源，真实PG18及堆栈已复现（无延迟80万通知测试通过，不断言线上必现）；R3/P2PrepareKey替换User丢失RPM覆盖快照，100次同组热请求产生100次仓库查询，原快照为0；R4/P2余额跌破门槛后异步生图结果轮询被403拦截。现有专项含子项76项仍全过、零跳过，两个PG专项实际运行，不能替代新增失败复现。完整报告deploy/GROUP_ACCESS_REVIEW.md，overlay测试／日志／互等堆栈在output/group-access-review-20261015；临时PG已核验身份后清理。仅审查及交接，业务源码未改、问题尚未修复、未提交推送／发布／部署，生产触发情况与P95未知。

- **2026-10-15 用户要求保证效率，已补做性能与回源复核（无业务源码变更）**：在已发布794479760的实现上，用Go1.27.2 windows/amd64复跑缓存及活动专项，GOMAXPROCS=1／4／16各重复3次，10个顶层／35个含子项场景共315次通过、零失败零跳过；并发100个相同缓存项冷请求合并为一次回源、消费更新余额不重新汇总月充值、失效防旧值写回、月初／到期和福利边界均通过。9组热缓存微基准为897.4–1525ns/op（约0.9–1.5微秒），policy-loads/op与monthly-loads/op均为0；2687–2688B/op、19 allocs/op，保留用户／分组／Key副本以免污染认证缓存。CPU参数只设置GOMAXPROCS，基准循环本身串行；使用桩仓库，不代表生产QPS、P95或完整HTTP链路。证据output/group-access-performance-20261015/verification.json及cache-and-schedule.jsonl，记录4个相关源码／测试文件SHA256。未改业务代码、未重建镜像或部署，更新标签仍sha-7944797；真实生产吞吐与P95继续待验，不承诺所有环境绝无性能影响。

- **2026-10-15 分组门槛、缓存、到期与不定期福利已发布（最新终态）**：功能b5e3c217f、前端CI清单0339bb5b5及缓存类型补修794479760593961c08ea842095543ec1d0bd58c5已推origin/main；正式发布以最终794479760为准。同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38727724771)（全量unit／integration、独立PG18分组退款／双实例缓存、既有PG18／Redis专项、脚本、Go lint 0 issues、前端52文件815项及类型／lint）、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38727724898)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38727724764)全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-7944797`（linux/amd64），摘要`sha256:34cf7d44c87608d92964329571250f65127b20dc3ca5cc99f7bf582533e4b0c5`，registry的revision／source／平台、正式embed构建日志及26个生产源码哈希交叉验真。初次CI的3处errcheck已补齐五类缓存取值检查，19项缓存专项通过，失败记录保留；原69项／50项及Linux11955顶层unit属补修前本地基线，不能冒充最终CI计数。证据`output/group-access-cache-20261015/verified-release.json`及同目录各job日志。用户备份数据库／配置／app_data后，在实际部署目录执行`./update.sh sha-7944797`，前后端同版并自动执行最终257迁移。管理入口「分组管理 → 开放规则 → 分时福利倍率」，可填入三段示例并选择23点兜底；指定日期一次性生效，不每日循环。退款／余额失去资格仍拒绝，只有到期切兜底；正常缓存命中不聚合充值，通知异步不能承诺零传播延迟。未代部署本机8080或生产、未配置真实业务活动；真实商户退款、页面交互、生产P95和WS客户端重连待用户更新验收。下方未提交／待门禁均为历史；之后仅文档skip-ci提交不产生新镜像，仍使用7944797固定标签。

- **2026-10-15 本轮CI缓存类型检查补修，待新SHA门禁**：功能b5e3c217f及前端CI清单0339bb5b5已推origin/main。0339的安全／脚本／前端52文件815项成功，但Go静态检查发现自有资格缓存3处未经检查的类型断言；已对规则、用户、分组、月充值和订阅五类取值统一检查，错误类型／空指针返回错误而非panic，不放行旧资格。新增8种异常缓存回归；本机Go1.27.2缓存专项6顶层／含子项19项全部通过、零跳过，证据output/group-access-cache-20261015/cache-typecheck-fix.jsonl。原69项及Linux全量是补修前基线，最终以新SHA的全量CI／安全／GHCR为准，尚不交付旧失败SHA标签；未部署。

- **2026-10-15 用户要求更新测试，进入本轮发布门禁（最新状态）**：用户要求「好了吗你我更新测试下」，按既有协作约定提交并推送 origin/main，生产仍由用户手动更新。发布前 fetch 确认远端与本地基线同为879f3b2e85c5cd5105d3c2d8afcae15c4193959a；最终本地验证证据见output/group-access-cache-20261015/verification.json。新增CI独立PG18及CUSTOM_GROUP_ACCESS_TEST_DSN，让退款和双实例缓存专项不靠跳过通过；待本轮实际功能SHA的CI／安全／GHCR全部成功并核验固定镜像后交付更新命令。未部署本机8080或生产，未配置真实福利活动；下方未授权／未提交为阶段记录，最终发布状态须更新本条。

- **2026-10-15 分组缓存与不定期福利最终验证完成，未发布（最新终态）**：覆盖下方「待优化／最终验证进行中」历史。独立资格缓存及事务通知已落地，正常缓存命中不再查询全部规则或聚合订单；仅余额门槛不查月充值，消费事件更新资格余额而不重算月充值。缓存30秒TTL、月初切换、4096项上限、并发合并、防旧值回写；每实例一个LISTEN连接，每秒心跳／3秒健康租约，异常回源且失败拒绝；通知异步，不保证零延迟。最多32个一次性北京时间活动时段，可独立关闭；三段模板0.1至18:00、0.15至20:30、0.1至23:00，23点到期须选择兜底。用户专属／高峰／独立图片规则保留，倍率进入API与批量生图提交快照及列表展示，默认用户倍率缓存不锁住旧活动价。69项后端专项含真实PG18双实例／提交／回滚、活动边界与批量快照零失败零跳过，缓存／活动重复十轮通过；前端7文件50项、类型／i18n／lint／生产构建通过。最终Linux Go1.27.2全量58包11955顶层unit通过，含子项23316通过／36明确跳过；两个PG专项另在本机实际执行通过，不能把Linux跳过当数据库验证。首次全量server退出清理测试漏传新服务参数，补齐后最终unit与embed退出0，未改无关业务。最终GroupsView-CpXNiTby.js含福利配置；Linux/amd64二进制182566689字节，SHA256 265f34c3f869057e142daef0b5051a2cf177bd955077bb943643929838797e24。证据output/group-access-cache-20261015/verification.json，首次失败亦保留。临时PG容器精确ID核验后连匿名卷清理，业务镜像／容器／卷未操作；未提交、推送、发布镜像或部署。真实商户、生产负载/P95、管理员／用户页面与客户端WS重连仍待验收，不把隔离微基准当生产QPS；退款失去资格仍拒绝，仅到期触发兜底，未配置真实业务活动。

- **2026-10-15 分组资格缓存及不定期福利时段已实现，最终验证进行中**：用户追问优化并追加0.1至18:00→0.15至20:30→0.1至23:00的活动。已将规则／资格元数据／分组／月充值改为独立事件缓存，257增加事务NOTIFY，每实例LISTEN；30秒TTL、3秒健康租约、4096项上限、并发回源合并、加载令牌防失效后旧值回写、月初强制跨月。余额变化直接更新资格余额，不重算月充值；未改变现有认证／计费余额缓存。通知异步，断线回源且失败拒绝，不承诺零延迟。新增最多32个一次性北京时间活动时段与三段示例，独立开关；23点结束可联动原到期兜底。基础倍率进入请求及批量生图提交快照，可选组／模型广场显示当前倍率；用户专属、高峰和独立图片定价保留原规则，修正用户倍率缓存冻结旧默认倍率。69项后端专项（PG18双实例、提交／回滚、普通／Google、活动边界、批量快照）零失败零跳过、十轮缓存／活动重复通过；前端7文件50项、类型／i18n／lint／生产构建通过。热路径微基准新增policy及monthly回源0/op，仅为隔离缓存测试，不当作生产QPS。最终Linux全量与embed尚在运行，后续须用新证据覆盖初版；未提交推送、发布或部署，真实商户／页面交互仍待验收。最终证据目录output/group-access-cache-20261015。

- **2026-10-15 分组资格请求开销待优化**：用户质疑每请求实时查库的资源与效率。源码确认 PrepareKey 每次 List 查询全部规则，启用门槛时还读取用户／分组并由 Facts 汇总当月充值与余额；尚无压测证据，不能宣称具体延迟或吞吐影响。建议规则缓存、月充值按充值／退款事务提交后更新或失效、余额复用现有计费缓存、到期只比较缓存时间；须覆盖跨实例失效、并发旧值回写、月初边界及故障不放宽资格。当前仅核查和方案说明，性能优化尚未实现，原本地功能测试通过不代表性能验收完成；未提交部署。

- **2026-10-15 退款失去资格行为澄清**：用户询问实时检测还是切换兜底。源码取证 custom_group_access.go:171 的 PrepareKey 在新请求读取规则及数据库资格；230行门槛不满足直接返回 ErrGroupAccessRequirements，只有210行到期分支才使用 fallback_group_id。因此当前退款／余额降低失去资格会拒绝，尚未实现资格不足自动兜底。用户本次仅询问行为，未要求改变，未改业务或部署；退款须先同步入本站订单／余额，不能声称可实时识别站外退款。

- **2026-10-15 分组开放门槛与到期切换本地完成，未发布**：用户明确月充值／余额两个门槛可单独开关，同时开启为 AND；「不退款」指防止达标绑组后退款保留资格，故保留退款并在每次新请求重查。新增独立257迁移及服务、管理员「分组管理 → 开放规则」弹窗；北京时间自然月统计成功在线充值的 USD 账户本金，剔除赠送、已退比例，退款中订单暂不计入；兑换／管理员加余额不计月充值。整个分组可设北京时间到期时间和同平台目标，已有 Key 的下一请求自动采用目标权限、门槛和计费，不批改原绑定；不授权原本不可访问的目标，防循环。普通／Google／可选分组／模型广场与 Responses WebSocket 每轮均接入；长连接路由变化须重连，重连无需手改 Key，不能保证所有客户端自动重连。本机缓存 Go1.27.2 实测最终40项专项（含真实PG18、普通／Google HTTP、WS）零失败零跳过；Linux58包11946顶层unit、含子项23290通过／35明确跳过，真实PG专项另在本机实际通过；前端6文件42项、i18n、类型、lint、生产构建及Linux/amd64 embed编译通过。Windows首次全量有3项缺sh及1项未改动Ollama时序用例失败；补现有Git sh后三项通过，Ollama Windows仍失败，最终Linux全量退出0，不改无关代码也不认定其现场根因。最终包GroupsView-CsOAVPF9.js含新规则；编译产物182663821字节，SHA256 a2cdb91b93e819e40588e5461abddcaf08bc0efbab883a3f22984bea7f2e17de。临时PG容器及匿名卷已清理，旧生成依赖目录保存在忽略的output内。本轮未提交、未推送、未跑新CI、未部署8080或生产；真实商户退款及部署交互仍待验收。证据output/group-access-20261015/verification.json，说明deploy/GROUP_ACCESS_RULES.md；本条覆盖此前未实现／待澄清记录。

- **2026-10-15 分组充值／余额门槛需求取证**：用户要求指定定制分组按月充值金额、当前账户余额控制可见，两个门槛可单独或同时开启，并提到「不退款」。已读 APIKeyService.GetAvailableGroups／GetUserGroupVisibility、User.CanBindGroup、可用渠道／模型广场调用与 PaymentService.RequestRefund：现有公开／专属授权、有效订阅和退款流程均需保留兼容。拟按分组新增两个可配置门槛，同时开启时要求同时满足；月周期、金额口径和降额后的使用规则尚未定稿。尤其「不退款」可能改变统计或既有退款流程，须由用户明确具体含义，不能自行关闭退款或把充值全部锁定。本轮仅源码取证及需求记录，未改业务、配置或部署。

- **2026-10-15 充值页签开关位置取证**：管理员「系统设置 → 功能开关 → 站点类型 → 购买方式」控制充值／订阅页签；「仅订阅」关闭充值，「充值 & 订阅」同时保留，「仅充值」只保留充值。证据：SettingsView.vue 的 features 区块及 siteBillingMode 绑定、utils/siteBillingMode.ts 的两字段映射、PaymentView.vue 按 checkout.balance_disabled 生成充值 tab。「支付设置 → 启用支付」为另一总开关。本轮仅源码取证，未改配置或业务、未部署；线上当前选择未取证。

- **2026-10-15 搜索成功结果计数下限补修已发布**：功能2847434f96400713e81c72146d0202223a7c84c5已推送origin/main；同SHA CI38677336658（含unit/integration、前端及脚本）、安全38677336554、GHCR38677336670全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-2847434（linux/amd64），摘要sha256:4d9cd274e6c5fc711439c058cb34a418e6f5c79c8d392fcd9e5f2acbf412a42d，GHCR manifest/revision/source与源码SHA已验真。用户同时间127/587的本站0.001378仍少工具费；本轮在源码另复现成功结果被终态0覆盖、JSON的0/较小计量屏蔽结果，旧代码专项实际失败，补修后25项叶子专项（含父项38条）、58包/11935顶层unit（含子测试23257通过、34明确跳过）、gofmt及embed编译全部通过。共享观察器现取累计有效计量最大值与tool_use_id去重成功结果数的较大值，不相加；错误/普通工具/仅启动/无证据仍不收费，移除旧计量锁定状态，不改Token赋值、新增迁移或补扣旧账。条件搜索基础价0.05、有效倍率0.15、输入127/输出587及成功一次时总实扣0.00887790。分组单价/APIKey缓存传递源码及既有往返回归均已检查通过，但生产原始响应、实际版本/生效工具价仍未知，不能把本地补修认定为该笔现场根因或线上已验收；由用户备份后按固定标签手动更新并对新请求查看工具快照。本轮未发付费请求、未改生产配置或部署，本机8080未更新；临时验证镜像/导出容器已清理，原业务镜像及卷保留。证据output/newapi-ledger-20261014/result-floor-verified-release.json、result-floor-before.log及result-floor-unit.json；收尾纯文档skip-ci不生成新镜像。

- **2026-10-15 Token差异现场核验完成，搜索费仍待查**：用户补充确认使用相同请求后，继续只读检查已登录双方账单。本站页面同时保留07:53旧批与10:38新批；上游页面查询截止2026-10-15 08:52:49，排除了10:38新批。通过上游时间选择器“今天”及“搜索”将范围更新为当天00:00:00—23:59:59后，直接读到10:39:04为345/11、10:38:59为1268/618、10:38:47为127/587、10:38:45为19/245，与本站同时间四条输入/输出完全一致；上游缓存读取/创建分别180372/417、177569/1343、0/0、135255/43367，与本站缩写显示一致，未取得本站全部缓存精确值。此处用页面实际记录纠正此前仅凭截图配对产生的Token差异判断；不是修改Token代码，也不能从旧筛选截图证明Token统计有缺陷。10:38:47上游详情请求ID为202610150238460428009138268d9d69SCs9Nud，明确web_search一次、附加费0.000800、总0.001534；本站同时间127/587总0.001378，可由3/M、15/M、有效倍率0.15的纯Token费0.00137790复现，搜索费仍未体现在该条总额，不能宣称生产搜索漏计已解决。若基础搜索价0.05且有效倍率0.15实际生效，一次应加0.0075，总0.00887790；生产后端版本/同笔分组工具价/脱敏原始SSE仍未知，不把官方New API的BillingUsage候选机制当作这批现场根因。证据output/newapi-ledger-20261014/token-usage-live-reconciliation-20261015.json与token-usage-upstream-refreshed-20261015.jpg；只改查询时间与交接，无业务代码改动、无付费请求、无生产配置变更/部署。

- **2026-10-15 搜索流式初始0漏计修复已发布**：功能fb5916a4fe1853995b18f016d0806a2990a4c73b已推送origin/main，同SHA CI38665635231（unit/integration、前端与脚本）、安全38665635268、GHCR38665635087全部成功，golangci-lint日志0 issues。固定镜像ghcr.io/engineemomo/sub2apicust:sha-fb5916a（linux/amd64），摘要sha256:0c33d76be7e126af2e171d7d9086b1e84abcbe38cfde9c28507a1aa982154570；registry manifest、revision/source及正式构建日志交叉验真，二进制构建版本0.2.16-custom.fb5916a4f。旧代码在message_start搜索0→成功结果→终态无次数时漏计，新增28项叶子回归先失败后通过；共享观察器只忽略初始0，保留JSON/终态计量、去重、失败/普通工具边界。127/529、搜索基础0.05、有效倍率0.15的总实扣0.00874740，余额/订阅/Key额度/速率额度与工具快照一致。隔离Docker Go1.27.2全量58包11931个顶层通过（含子测试23219条pass、34明确skip），gofmt与embed编译通过，无新迁移。测试镜像/导出容器已清理，原本机业务镜像/卷保留，8080与生产未部署。线上该笔原始响应/版本仍未知，源码复现不等于现场根因已确认；由用户升级后用新请求对账，不重复索取已有双边费用截图。证据output/newapi-ledger-20261014/search-zero-verified-release.json；后续纯文档skip-ci不生成新镜像。

- **2026-10-15 搜索漏收双边费用详情已取得**：用户补全上游New API与本站同一127/529请求详情。上游/v1/messages、dynamic/base、输入3/M输出15/M、倍率0.08，明确记录web_search调用1次、调用花费0.000800，总扣0.001466；本站tooltip输入费0.00038100/输出0.00793500、倍率0.15、原始0.00831600、用户扣费0.00124740，没有工具费明细。现已证实线上该笔仅收Token、未体现该次搜索附加费；上游次数/金额不再只是反推。按双方额度1:1才有账面倒挂0.00021860，充值/币种换算仍未知。若用户所设本组基础0.05在该请求时生效且搜索计量传到本站，应额外收0.0075，总0.00874740；不能把0.0008采购工具费误当作本站应加费用。只读浏览器已进入现有Chrome的本站/usage，目标07:53:21行一致；导出CSV等待下载超时，没有取得原始usage/SSE或服务端版本，未改线上配置、未发付费请求。无法据账单断定具体是旧版部署、分组价未作用、协议字段未传或解析未覆盖；不对所有普通tool_use猜测收费、不以此截图宣称已修复。本轮仅更新诊断交接，业务代码/收费规则未改。证据output/newapi-ledger-20261014/search-fee-details-20261015.json及用户截图；下一步需运行版本与原始响应的计量字段，不能再索取同一账单费用明细。

- **2026-10-15 搜索单价0.05仍少收截图核查（仅诊断，未改收费代码/配置）**：截图4笔输入/输出345/13、1083/443、127/529、19/127，本站账面0.008366/0.010033/0.001247/0.030658，上游0.004462/0.005352/0.001466/0.016350；按显示数字仅127/529差-0.000219，其余为正。左$右¥，实际充值/显示币种换算未知，不称真实现金利润。沿用此前输入3/输出15/缓存读0.3/写3.75每百万、本站0.15/上游0.08的条件，用Decimal复算4笔本站金额均可由纯Token费重现；127/529基础0.008316×0.15=0.00124740，未体现搜索附加费。若该笔1次搜索且本组基础价0.05生效，应加0.0075、总0.00874740；上游0.001466与Token0.00066528+一次基础0.01×0.08=0.000800（再取整）数值吻合，次数/单价仍须原始日志证实。现行d3a2c3463代码只读usage.server_tool_use.web_search_requests或成功web_search_tool_result数组；明确0优先、普通tool_use不计。2026-10-15核对New API主线relay/common/tool_usage.go与relay/channel/claude/relay-claude.go：工具请求定义匹配和自定义ClaudeToolUseRules可配置按命名tool_use块计量，默认web_search规则的CountToolUseBlocks=false，不能断言所有普通WebSearch都收费或据UI文本直接补扣。截图仅Searched web/结果字符串无法确认原始协议；线上版本、该笔实际分组单价与tool_usage、上游附加费明细仍未知。需展开2026-10-15 07:53:21的两边账单费用详情/导出工具快照，再取原始usage与content_block决定兼容补丁；不继续靠调高价格试错。本地工作树起始clean、HEAD1eae1128e；证据output/newapi-ledger-20261014/search-fee-screenshot-20261015.json及本轮官方源码网页核查，未运行Go测试、未部署、未提交推送本条诊断。

- **2026-10-14 Claude网页搜索工具附加费发布终态**：最终功能d3a2c346352e7eca5aacc197785fd366f90f83b6已推送origin/main（原实现d83900dc21debd3ba21b1c926375d5dfed167729，后续只修复DTO测试两处类型断言）。同SHA CI38589112494、安全38589112074、GHCR38589112344全部成功；后端unit/integration及lint、前端49文件790项通过，正式镜像embed编译通过。固定镜像ghcr.io/engineemomo/sub2apicust:sha-d3a2c34（linux/amd64），摘要sha256:5baca94dfd6ad73ea03117406400253c083ff9886addca0e4b80528bb6149ba5，registry/revision/source与构建日志交叉验真。新增256迁移，Claude搜索工具费按分组基础单价与Token同一有效倍率叠加；用户/管理员账单及CSV/XLSX保留工具快照。隔离Docker原实现58包11928顶层条目、真实PG18工具金额/幂等及前端54项专项通过，修复测试由最终CI复核；工具PG专项CI未配置DSN，不能把其skip当通过。生产及本机8080未部署，配置/实际上游计数/单价/充值价值仍待现场核验，不自动同步New API价卡。证据output/newapi-ledger-20261014/verified-release.json；交接纯文档skip-ci不生成新镜像。

- **2026-10-14 Claude网页搜索工具附加费实现与验证（发布前记录）**：用户要求“增加相关收费”后新增custom_claude_tool_billing，读取原生JSON/SSE的usage.server_tool_use.web_search_requests，按最大累计值去重；无有效计数时仅从成功web_search_tool_result按tool_use_id回退，普通tool_use、声明/启动事件和错误结果不收费。复用web_search_price_per_call，默认0.01/次、0免费；Anthropic/composite分组创建/编辑入口可配置，按Token同一有效倍率叠加工具费（用户专属/分组及高峰），通用Claude、API Key直通、Claude→Responses缓冲/流式及OpenAI转发Claude原生Messages均接入。新增256 nullable tool_usage JSONB及完整Ent/SQL/DTO，用户/管理员共用账单提示和CSV/XLSX工具快照；账号统计模型价重算后补工具基础费，包号免扣/扣费失败的工具实扣0，旧账单不补扣。Docker隔离验证最终go test -tags=unit ./...退出0、58包/11928个顶层测试条目通过、34个条件项skip（真实PG专项另跑并通过）；Ent重生成哈希一致、go build -tags=embed通过；54项前端回归、i18n完整性3项、类型/改动文件lint/生产构建通过。首次全量3项阿里云验证码测试因构建代理导致本地HTTP超时，保留完整证据，测试进程取消HTTP(S)/ALL_PROXY后全量通过，未修改验证码业务或测试。本轮仅独立临时PG验收迁移重跑/同步/批量/最佳努力写入/历史NULL及余额/订阅/Key额度幂等；已清理3个临时容器、本轮网络和具名验证镜像，日志/生成/构建证据保留，无业务容器/配置更新。用户后续明确授权提交并推送origin/main，首个候选d83900dc2已推送，但CI38587939452的errcheck发现DTO回归测试连续类型断言未检查ok；仅修复该测试的两处类型检查；最终新SHA的CI/安全/GHCR核验结果见上方发布终态，生产未部署。配置/边界见deploy/CLAUDE_TOOL_BILLING.md，接缝已登记CUSTOMIZATIONS与frontend/src/custom/UPGRADE.md；证据output/newapi-ledger-20261014/{backend-final-check.log,unit-tests-final.jsonl,unit-exit-code-final.txt,final-validation.json,postgres-check.log,frontend-check.log,frontend-build.log,ent-before.sha256,ent-after.sha256}。真实用户上游工具种类/实际计数/单价/充值价值仍未知，不自动同步New API价卡，不承诺所有未知附加费和阶梯错价都已消除；下方“仅诊断未实施”是此前状态。

- **2026-10-14 Claude工具附加费新截图核查**：用户新增New API徽标tooltip明确“包含工具调用附加费”，127/637上游0.001594/本站0.001490、132/354上游0.001256/本站0.000856，按额度1:1合计负差0.000504；这些Token对不在前述原始CSV/206匹配样本，不改旧样本结论。当前HEAD5f2f08762的ClaudeUsage及原生SSE/JSON解析没有付费工具次数字段，搜索价表单仅OpenAI Codex/Grok；通用SearchCount叠加分支虽存在，Claude没有次数接入，不能把已有搜索收费代码当Claude已支持。推断按输入3/输出15、上游0.08/本站0.15、基础工具费0.01/次与上游500000额度单位取整，两笔金额全部重现，额外采购0.000800/笔；真实工具种类/次数/倍率/生产版本仍缺，不能直接认定web_search或该价格配置。需补Claude付费工具次数识别、单价配置、叠加扣费/日志，现有Token价卡和阶梯开关不足；仅诊断建议，未获实施新计费功能指令，未改业务配置/代码或提交部署。证据output/newapi-ledger-20261014/tool-surcharge-verification.md、gateway_service.go:596、gateway_anthropic_passthrough.go:627/764、gateway_usage_billing.go:1069、GroupsView.vue:3083/3150；Decimal复算通过，未运行Go测试。

- **2026-10-14 双账单逐笔对账完成**：读取用户下游xlsx206笔及上游csv394条，按模型/四项Token/≤5秒/双向唯一匹配，206笔全部匹配，最大差1.010875秒；两边额度按1:1，客户扣20.059911、对应采购14.153176、净正差5.906735，23笔Sonnet5合计负差0.015886，183笔Opus5.5合计正差5.922621。全部有效倍率0.4、账号1、上游kiro-aws。本站Sonnet价2/10/0.2/2.5，采购相对本站基础费等价约0.42，完整解释本批少扣；上游3/15/0.3/3.75×0.28及0.000002取整是全样本重现组合，不声称直接读到上游配置。全206笔固定价重算0不一致，取代前述仅截图尚无现场证据的阶段结论；不能认定本批源于阶梯或1h缓存。上游未匹配129笔：122笔早于范围12.778004、期间7笔2.385628，59笔错误均0；不得直接总额相减。建议只修该分组Sonnet基础价至3/15/0.3/3.75、保持0.4；历史模拟0亏损，总净差6.065635，Sonnet正差0.143014。原价全组0.45/0.50亦在样本0亏损，但影响其它模型且不能保证未来。真实现金利润仍缺双方实付/到账比；1h价、上游完整表达式及7笔归属未知。证据output/newapi-ledger-20261014/{analysis.json,report.md,对账与计费建议.xlsx,xlsx-validation.json}；Decimal加总及206笔费用/模拟全核对，成品导出0公式错误、已渲染。仅分析及更新交接，未改业务配置/代码或提交推送部署。

- **2026-10-14 阶梯开关与高倍率仍少扣核查**：用户明确账号倍率1、售价分组比上游高0.1、开启阶梯但无手填区间，要求验证为何部分上游实扣更高。源码已核实开关只启用本站已有价卡阶梯，不会同步New API表达式；无区间时用BasePricing，阈值/长上下文倍率缺失就不会加价。随仓Opus5/5.5价卡与兜底均无上下文阈值；不能据此认定生产动态价卡一致。截图含约9万缓存读/2.8万缓存写，选5356输入/463输出/94557读/27911写条件复算：按4/20/0.2/5基础合计0.1891504，缓存占83.78%；上游假设0.28并按500000额度单位取整可重现0.052962，本站同口径0.38应扣0.071877152。若上游写入1h按8、本站按5m按5，上游0.076407352反而更高；若上游整套价2x而本站基础档，上游0.105924224。两者均为验证条件路径的算例，不认定用户现场。已核实本站TTL细分缺失会全部按5m，覆盖设置可改计费分类，用户专属倍率优先于分组。New API官方日志源码确认+1是额外价格摘要片段数，缓存价会占片段，不能认定为第二档。证据与取证范围 output/newapi-billing-verification-20261014/verification.md，源码 billing_service.go:1567/1740/1985、model_pricing_resolver.go:305、gateway_usage_billing.go:803/919、resources价卡:175/770及官方固定提交4013729f的日志列代码:229/856。尚缺一笔确实上游高于本站的双方完整usage、有效倍率、TTL/档位/附加费、生产版本与上游billing_expr；不宣布实际差额根因。未改业务/费率配置或部署，未运行Go测试，仅Decimal复算与静态核查。

- **2026-10-14 New API 与本站日志对账（盈亏尚未证实）**：用户反馈本站倍率高于上游仍亏，并给出18:24:00的 Opus5.5 日志。截图双方均输入33/输出94；上游请求名claude-opus-5-5、发往上游claude-opus-5.5，命中standard，输入4/输出20每百万、倍率0.27，基础0.002012、未取整0.00054324，日志最终显示¥0.000544。本站绿色显示$0.000704，橙色A显示$0.000212。源码证据 UsageTable.vue:213 绿色来自 actual_cost；:580 橙色A为 (account_stats_cost ?? total_cost) × account_rate_multiplier，是本站计算的账号统计成本，不是读取New API本次实扣。account_stats_pricing.go:53 支持自定义成本规则/售价基准/默认模型价；不能只凭A认定采购成本。截图未给本站total_cost/实际倍率快照、部署版本及双方余额的实际充值兑换价值，币种符号也不能替代现金口径，因此目前不能判定这条实际盈亏。仅在每单位余额实际价值相同的假设下，0.000704大于0.000544；若客户支付1元获得2单位本站余额，实收折算为0.000352元，仍需上游实际额度成本才能比较。这是条件示例，不是用户配置。本站「余额充值倍率」明确每1 CNY得到多少USD余额（settings.ts:737），赠金/折扣另影响实际价值；源码 payment_amounts.go:28 按金额×倍率入账。优先取得双方实际充值到账比、本站该条费用tooltip快照和上游域名/价卡，继续核验，未改业务或费用配置/部署。

- **2026-10-14 New API 截图未显示上下文区间**：用户补充「按分组定价」截图，每个分组可见同一 base 档，cc-max1 输入5.75/输出28.75与1.15x、cc-max4输入4.5/输出22.5与0.9x对应基础输入5/输出25；这里只证明可见分组倍率，不能认定存在上下文阈值，也不能认定完整表达式无额外条件。前述动态计费不等于必须配置上下文区间。联网核实同一官方提交的 `pkg/billingexpr/expr.md:183` 同时有单档 tier(base, ...) 和 len 条件分档示例；`web/src/features/pricing/components/model-details.tsx:1226` 的分组档位单元格有 conditionText 才追加条件，否则显示 label；`dynamic-pricing-breakdown.tsx:655` 同样有条件摘要才显示。用户实际部署版本和完整 billing_expr 尚缺，不能据此解释其上游实际少扣/多扣。另核实本站 GroupsView.vue:1494/3144 明确 hide-token-intervals，手填上下文区间在渠道价卡（PricingEntryCard.vue:161），不在分组逐模型表单。下一步仍需上游地址或该模型 billing_expr，不凭截图编造200k等阈值。仅更新交接，未改代码/计费配置/部署。

- **2026-10-14 New API 定价查询接口已核实**：用户追问是否可直接查询使用。通过联网读取 QuantumNous/new-api 官方固定提交 `4013729f6a2a14b9b5988d3c1a529ea97d2b2d8d` 核实：`GET /api/pricing` 是官方模型广场读取的接口，返回 data 的 model_name/model_ratio/completion_ratio/cache_ratio/create_cache_ratio、可选 billing_mode/billing_expr，以及顶层 group_ratio/usable_group；动态模式 tiered_expr 会公开对应表达式（model/pricing.go:27/373）。登录查询会按用户所属分组覆盖部分 group_ratio（controller/pricing.go:38），匿名结果不能直接认定为当前 Key 的有效费率。middleware/header_nav.go 的 HeaderNavModuleAuth 允许站长配置公开、需登录或禁用；不能保证该用户上游可匿名访问。`GET /api/ratio_config` 另提供价格同步配置，仅在启用倍率暴露时开放（controller/ratio_config.go:13）。`GET /api/usage/token/` 使用 API Key Bearer 鉴权，返回累计额度，不能据此直接关联并发场景下单次请求实扣（controller/token.go:GetTokenUsage，router/api-router.go:298）。表达式系数使用 USD/百万 token，且可依赖上下文、请求字段/头、时间等条件；本站须适配规则与 token 归一化后才能自动计费。官方证据地址基址 `https://raw.githubusercontent.com/QuantumNous/new-api/4013729f6a2a14b9b5988d3c1a529ea97d2b2d8d/`，读取文件 router/api-router.go、controller/pricing.go、model/pricing.go、middleware/header_nav.go、controller/ratio_config.go、controller/token.go、web/src/features/pricing/api.ts、pkg/billingexpr/expr.md。网络浏览工具未取得正文，改用 Invoke-WebRequest 获 HTTP200；普通沙箱网络拒绝后按要求获审批读取公开源码。当前缺用户上游域名，尚未实测其部署接口、版本或实际价格。本轮仅更新交接，未改计费代码/配置或提交部署。

- **2026-10-14 New API 上游费用差异只读排查**：用户询问动态计费入口，反馈本站扣费有时低于、有时高于上游。当前源码已核实：渠道入口 `/admin/channels/pricing` 支持输入/输出、缓存读、5m/1h缓存写与上下文区间；分组入口 `/admin/groups` 的「分组逐模型定价」优先覆盖渠道/内置价，另有长上下文阶梯开关。证据 channels.ts:123/158、overview.ts:1108、model_pricing_resolver.go:72。价卡单位为 `$/MTok`（PricingEntryCard.vue:107），截图为人民币；用户扣费采用本站价卡和用户专属/分组有效倍率，账号倍率另用于账号统计/额度（gateway_usage_billing.go:165/189/921）。现有上游声明倍率探测访问 `/v1/sub2api/billing`，不是 New API 实扣同步。尚无生产版本、完整模型名、上游 Key 分组、同次请求双方 usage/扣费详情或币种换算配置，不能判定现场差额原因；截图可见仅 standard 档，不能推断其他条件。本轮仅更新交接，未改业务/计费配置、未部署或运行测试，未提交推送。

- **2026-10-14 全工作区发布终态**：用户明确要求“整个工作区提交”，此前保留的59项改动已全部提交推送origin/main，功能SHA为6406511c687e59b2afd37cf24da3db0d6011bea5；包含收款预占释放与用户封顶、248/249迁移、库存待付保护、工坊配额/清理/匿名限流、智力效率雷达Bench、前端工具链与更新脚本，并保留上游0.2.16及认证网段防护。同SHA CI38522649457（Go unit/integration、专用PG18/Redis8、Go lint0问题、Node20前端49文件790项及26组更新回归）、安全38522649572、GHCR38522649396全部成功。本地前端409文件3279项、类型/构建和26组更新回归通过，源码lint0错误26警告；宿主未重新编译Go，embed由本轮GHCR验证。正式镜像ghcr.io/engineemomo/sub2apicust:sha-6406511（linux/amd64），摘要sha256:9608581fa3c7e8c9b36e368e7815f58376c9c3ee9d2c48f449e0412f6b2244bb，registry/revision/source/构建日志与源码树6050fc24b4c47e4e8434c53fc75f7a8403ca163c已交叉核验。前端生产依赖审计0告警；Go0可达漏洞，仍有2项未调用模块漏洞，不宣称全部依赖零漏洞。源码提交后工作区干净，最终交接另以纯文档skip-ci提交；生产未部署、本机8080未更新本镜像，新版真实付款/素材/生产数据仍待用户验收。证据output/workspace-release-20261014/verified-release.json、preflight.json、远端回执与各job日志。本条覆盖下方“59项未提交/未发布”的历史阶段记录。

- **2026-10-14 上游0.2.16发布终态**：功能fe6323ca643af95cd779e8c4700b1bc2583eb29a已推origin/main，包含upstream/main=9c2cca9e627291bc719bada3d57de23b625a4180与正式v0.2.16；同SHA CI38517873853（unit/integration、专用PG18/Redis8、lint0、前端47文件774项含IP管理16项）、安全38517873829及GHCR38517873878全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-fe6323c（linux/amd64），摘要sha256:3740fc0c7b541500a4ba50f94a59696a4cb3dcf52335dca0578275547bb7660d；registry/revision/source、源码树605c0fbae241ffb0bda14d14798dfe9cd5465a0c及二进制0.2.16-custom.fe6323ca6交叉验真。Ent生成无差异、Go embed、定制专项与4189源码61条/正式镜像25条真实HTTP验收通过，覆盖登录、管理员权限、通配符/网段封禁、永久默认与在线更新拒绝。210个自有源码/迁移保持，主工作区main已快进；其他59个未提交文件已备份恢复，依赖仅合入Vue3.5.43并整理原锁文件，付款/248/249/智力效率等其他会话工作仍未发布。3个测试容器、测试卷/网络及工具镜像已清理，原8080镜像b311d7372950与原卷保留，8080/PG/Redis当前仍停止，未把原8080当0.2.16验收。生产由用户备份后执行`./update.sh sha-fe6323c`，生产真实IP及共享网段误拦率仍待；此前sha-bbee35d及本轮“待发布”条目是阶段记录。证据output/upstream-sync-20261014/verified-release.json、formal-smoke.json、original-work-restored.json与cleanup.json。
- **2026-10-14 上游同步本地收口**：原59文件共1258867字节逐字节备份在output/upstream-sync-20261014/before-main及manifest；恢复后48文件文本完全一致，11文件合入上游/交接变更，package.json除Vue3.5.43外原依赖升级完整保留。Git快照0a1d3f61b09be55d2c7a0c3ce8cab738189d9ade保留，不drop；未提交内容恢复为未暂存，本轮提交排除它们。原8080、PG、Redis及recipes容器保持停止、未删除，原应用镜像仍b311d7372950；正式sha-fe6323c镜像已拉取但只在4189验收。独立工作树已申请归档，测试/发布证据全部留在D:/ccfile/sub2apicust/output/upstream-sync-20261014。

- **2026-10-14 上游同步本轮**：用户要求更新开源fork，目标upstream/main=9c2cca9e6（正式v0.2.16=c3cb8588ce6a5aef4dc29fa4f71ac91f3c0bdf3d，主线随后仅版本号/赞助信息更新）；相对711bfc744新增235条上游提交。merge-tree演算后在独立工作树合并，14处文本冲突逐项解决，保留Go1.27.2、x/image0.45.0、SheetJS0.20.3、DOMPurify3.4.16及认证IP/包号CI存储；采用Vue3.5.43和上游完整13平台配额测试。210个自有源码/迁移文件哈希保持，原6平台fixture已由上游动态清单取代。前端47文件774项、lint/类型/生产构建；Docker内Ent生成无差异、Go embed、PG18/Redis8专项及4189隔离61条真实HTTP验收通过。原8080应用未替换、其他会话未提交改动保留，生产未部署；本轮功能SHA/CI/安全/GHCR终态及固定镜像以HANDOFF最新记录为准，不能沿用sha-bbee35d的门禁证明0.2.16。证据output/upstream-sync-20261014。

- **2026-10-14 网段防护发布终态**：功能bbee35dbcc30525bae354dc1d270f90b5591fab7已推origin/main；同SHA CI38508986548（unit/integration、专用PG18/Redis8、lint0、前端38文件504项含IP管理16项）、安全38508986630、GHCR38508986565全部成功。固定镜像ghcr.io/engineemomo/sub2apicust:sha-bbee35d（linux/amd64），摘要sha256:0dd2c3067bf73f3b5308ee72b7e27d22014ff71aa618fb7c02404bfda7ac9e22；28文件源码哈希、registry/revision/source及构建日志交叉验真。支持手动通配符/CIDR及自动同网段不同IP共享明确失败计数，默认/24和/64，单IP/前缀/期限可配置。8080于2026-10-14 11:56完成备份更新及三服务healthy验收、原数据/卷保持，254/255生效、自动仍false、permanent/version5、名单0；本机保留原智力效率，正式发布排除其他会话改动。独立测试环境已清理，生产由用户备份后`./update.sh sha-bbee35d`；同网段正常用户也可能受限，生产真实IP和误拦率仍待验证。证据output/auth-ip-network-20261014/verified-release.json、local-verified.json及HANDOFF。 新加入默认永久封禁，只有管理员可调整规则/期限及解封；已有临时名单保留原到期时间。复查旧自动单IP名单解封的网段计数边界，修复前真实PG回归失败、修复后8组网段回归及全量unit通过；最终45条真实HTTP断言通过。最终8080镜像sha256:b311d7372950873015c08c68567d36719ad37457693a7a1915b5ca4839b2a3c1、容器bd9fd0be85449daeb20028f49a3e29495d4337891e1ca9e18f6cf627f2e6d3e6、chunk assets/AuthIPBlacklistView-DwaftuX8.js；255前备份667706字节、目录1283行、SHA256cec4821389c3cf8bec26426184695a0e6922066f8be26d6c3dfd273c621e6ffb与before-auth-ip-permanent-20261014镜像保留。最新日志backend-permanent-final/frontend-permanent-final/build-permanent-release/build-permanent-local.log；旧aade433及下方timed/version4/44条为阶段记录，最终更新使用sha-bbee35d。 2026-10-14 13:08接续会话只读检查确认Docker引擎未运行，当前8080健康未复测；CI/GHCR及远端HEAD已重新核验，不沿用旧发布结果。

- **2026-10-14 网段防护本轮**：手动支持210.230.206.*及IPv4/IPv6 CIDR；自动按网段共享明确失败计数，默认IPv4 /24、IPv6 /64，可配置单IP/网段、前缀、临时/永久及时长。成功不计数，名单仍只保护认证；共享网段正常用户也受限。新增254保留旧开关和名单、规则version递增；250–253不改。本轮自审未发现阻塞问题；Docker真实PG18/Redis8、18项后端专项含7组网段回归、全量unit/Go embed、前端16项/类型/lint/Vite和8081的44条HTTP断言通过。正式提交/本轮CI/GHCR仍待，不能沿用旧发布结论。证据output/auth-ip-network-20261014。
- **2026-10-14 网段防护8080**：原应用更新到sub2apicust:auth-ip-network-local-20261014/b33cb472be2a，容器0fa7d82eda3d，三服务healthy；原PG/Redis/recipes容器及挂载保持，recipes仍停止。用户3/Key0/组3/套餐1/订单0/设置281与摘要686429b8e197b1d3d609158ba01f822c保持，名单0；254生效、enabled=false/timed900、version3→4、network/24/64。真实管理员登录/config/entries、新输入400、模型入口401和新chunk BznUWgMC核验。迁移前备份665933字节、pg_restore目录1282行、SHA256ff83db7ad955edffcc47e6cffc904c1eafa5a115d26a36592e6d085916e141af，旧before-auth-ip-network-20261014镜像保留。本机延续原智力效率源码；正式发布快照排除其他会话改动，生产未代部署。

- **2026-10-14 手动封禁发布终态**：功能065eee935a5eea3fb023c5faf469b988feb0fa73已推origin/main；同SHA CI38497226115（unit/integration，专用PG18/Redis8，lint0，前端38文件502项含IP管理14项）、安全38497226138、GHCR38497226114全部成功。镜像`ghcr.io/engineemomo/sub2apicust:sha-065eee9`，linux/amd64，摘要`sha256:8da3604012c2a2c121dd47a7fa22542c785b5c77c60335177498a40058807d38`，registry/revision/source/构建日志及23文件快照哈希交叉验真。用户备份后在部署目录执行`./update.sh sha-065eee9`；支持后台IP黑名单“手动新增 IP”、完整IP/原因、永久/临时，默认手动永久；自动开关关闭仍生效，只保护认证，重复有效封禁409。253迁移已在本机8080生效，原250–252不改，生产未代部署。
- **2026-10-14 手动封禁8080收口**：原应用已替换为镜像e9fd3b9c7ea2、容器112691d6fc2b，三服务healthy；实际镜像ID/运行容器一致，10条HTTP及真实管理员登录/新POST参数400/新版chunk AuthIPBlacklistView-CoRkhMZ9.js通过。原PG1f5c08bee7e1、Redis c531bb587a18及原挂载保留，用户3/Key0/组3/套餐1/订单0/设置281/名单0和防护配置enabled=false,timed900,version3升级前后完全一致（仅比较上述业务计数及防护配置，不宣称全部数据库字节不变）。基于原v3快照只叠加手动功能，保留之前本地智力效率源码；正式镜像不含其他会话改动。备份output/auth-ip-manual-20261014/database-before-manual.dump，664356字节，SHA2563998822DEF6124768CF5D35AD5ADEA42CF3A219FA3B45C66B3FF9B05384A69A0，pg_restore目录1282行检查通过；旧镜像before-auth-ip-manual-20261014保留。独立8081测试容器/网络/匿名卷已清理。证据verified-release.json、local-verified.json、smoke.json及本SHAjobs/logs；下方隔离阶段“未操作8080/待CI”已成为历史状态。

- **2026-10-14 手动封禁本地验收**：隔离Docker Go1.27.2全量unit及15项专项（真实PG18/Redis8）、前端14项/类型/lint/Vite、实际镜像embed和产品路由通过；8081独立auth_ip_smoke库32条HTTP断言及3条人工封禁审计通过，关闭自动开关、永久/45分钟/到期、重启/Redis清空、重复409/旧版本和管理员自解封已取证。模型入口仍401、管理员/auth/me仍200，新chunk AuthIPBlacklistView-D2EJCSaP.js。原8080服务/数据/配置未操作；本轮新SHA CI/安全/GHCR待核验，证据output/auth-ip-manual-20261014。

- **2026-10-14 管理员手动封禁（本轮）**：已补自有服务/存储、POST `/admin/ip-blacklist/entries`和后台“手动新增 IP”，支持完整IPv4/IPv6、永久/1–1440分钟临时期限、1–500字原因与确认；关闭自动开关仍有效，重复有效封禁409不覆盖期限。253迁移区分manual/auto、manual无失败统计；旧250–252未改。只保护现有敏感认证，管理员有效会话仍可自解封，模型调用不受名单影响。隔离源码基线06873209c及验证目录output/auth-ip-manual-20261014；测试及新SHA发布正在核验，不能沿用07bcef204结果。其他会话改动保留，生产未部署。

- **2026-10-14 IP防护最终发布（最新有效状态）**：功能提交`07bcef204d09349e0574b51a5b8cfdd6266288a3`已推送origin/main。同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38493252759)（unit／integration，专用PG18／Redis8，lint0 issues，前端38文件497项含IP管理9项及类型）、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38493252800)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38493252762)全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-07bcef2`，linux/amd64，摘要`sha256:81bcf46cd8e1008f058c296f09d32af059d0e7fadefa66cf355b522885d5bc13`；registry manifest／镜像revision／source标签与构建日志交叉验真，证据`output/auth-ip-release-20261014/verified-release.json`及jobs/logs。18个功能文件与10-13已测试v3哈希一致。用户备份数据库、配置和app_data后在部署目录执行`./update.sh sha-07bcef2`，启动应用250/251/252；首次自动拉黑关闭，10次可识别失败/120分钟、timed15分钟，可后台选永久或时长后开启。只保护公开认证：注册／登录／2FA／Passkey／验证码／密码找回和重置／OAuth；成功不计数，OAuth GET只查名单；模型／管理／当前用户接口、刷新Token／登出／微信支付OAuth不受此名单影响。发布未带入其他会话付款／库存248/249、投稿、依赖升级和智力效率改接。生产未代部署，真实IP链路、共享出口误拦和带记录浏览器验收仍待；10-14Docker引擎未运行，本轮远程门禁不冒充再次本地运行。

- **2026-10-14 IP防护发布准备**：用户授权提交并推送本功能，生产由用户更新。已从router/auth.go核实guard只挂公开认证组：注册／登录／两步验证／Passkey／验证码／找回和重置密码／OAuth认证受保护；成功不计数，OAuth GET只检查。模型调用、管理接口、用户信息／撤销会话、Token刷新／登出及微信支付OAuth不接入。发布仅包含本功能及自有文档、CI真实存储接线，不带其他会话付款／库存248/249、投稿、依赖升级或智力效率改接。CI新增专用PG18／Redis8及AUTH_IP_TEST_*变量，避免新存储回归仅skip。工作区源码与10-13隔离v3的本功能哈希一致；10-14Docker引擎未运行，不冒充本轮重测。发布SHA与CI／安全／镜像结果待推送后记录。

- **2026-10-13 23:31 IP 防护可选永久／临时封禁（最新有效状态）**：用户最新要求“加封禁，找管理员解除，或封禁时间可配置”，覆盖上一轮只新增临时封禁的决定。保留降低误拦规则：只计明确的密码／Passkey认证失败和请求格式异常；成功登录／注册／正常多步骤操作不计数，验证码／TOTP／业务校验／令牌失效／429／500／503排除。后台增加封禁方式：timed临时（1–1440分钟），permanent永久（管理员移除）；默认仍10次失败／120分钟、临时15分钟，开关关闭、名单0条。252迁移增加block_mode默认timed并递增version；250/251内容不改。规则切换仅影响新封禁，不改变已有记录期限，永久记录blocked_until=NULL。不能承诺共享IP零误拦或抓住所有攻击。
- **本轮审查与验证**：封禁模式、版本竞争／期限保存、管理员权限／审计及前端表单已审查，未发现阻塞性问题。Docker Go1.27.2的13项专项（真实PG18／Redis8）、全量`go test -tags=unit ./...`、运行镜像Go embed及嵌入产品路由回归通过；前端5文件55项（IP管理9项）、类型／定向ESLint／Vite构建通过。实际镜像专用guard_smoke库8081：阈值3下9次成功登录不计数，第3次失败403永久封禁；实际重启应用、Redis FLUSHALL后仍403，管理员有效JWT仍可解封。切换临时45分钟、解封后新失败生成正确期限；切回永久不延长旧临时记录，只在测试库移动到期时间后下一次正确登录200。遗漏模式400，旧规则及旧名单版本409；真实审计两条自动封禁、一条成功解封、两条永久规则更新。3只临时容器、匿名卷、专用网络已清理。
- **8080实际终态**：只替换应用容器`589ddc6d3ae7`，三服务healthy，运行镜像与`sub2apicust:auth-ip-mode-20261013`一致，inspect Id `sha256:7f423b3890c9a070657c0097adf7b486499ed042cc8e3a350a2266caf1b2bad5`，版本commit `local-auth-ip-mode-a6cdbd870`。真实原管理员登录及config/entries200，管理接口匿名401；入口`index-Cgh4f6-U.js`和页面`AuthIPBlacklistView-DotqpFXJ.js`已由HTTP验真。迁移250/251/252共存、总迁移301，248/249未带入；用户3／密钥0／分组3／套餐1／订单0／281项设置摘要及全部原卷保持，PG/Redis/recipes容器ID不变，recipes停止。构建源为HEAD a6cdbd870＋本功能＋已运行智力效率的docker-src-v3隔离快照，未带入其他未提交付款／库存／投稿／依赖升级。证据`output/auth-ip-local-20261013`的mode-*及backend-mode-tests.log、frontend-mode-tests.log、build-mode.log，见`deploy/AUTH_IP_BLACKLIST_REVIEW.md`。
- **备份与边界**：第一次部署前dump655269字节、SHA256 `BF50B77C63EC935F7565D55AFE28FEF9B6B06135344AF6F1D2B568FCFD9A958F`，旧镜像`sub2apicust:before-auth-ip-20261013-221424`；251前再次dump661963字节、SHA256 `464A3F7A1A3AA3EECC920006BDD172A58230E6C0D3E60B264DE40A3A2431CBBD`，两份均通过pg_restore目录校验。后一阶段旧镜像为最初永久封禁版本，已有临时名单后不能把它当作支持到期的安全回退版本。生产未部署、未提交推送、没有本功能新SHA CI/安全/GHCR或本轮golangci-lint；Cloudflare→Nginx真实IP及生产误拦率仍未知。仅HTTP实测原管理员，新页面带真实记录的浏览器交互仍待用户验收。截图400缺少reason/正文，新分类不能承诺截图10条全部计入。
- **252前新增回退证据**：database-before-mode.dump为662966字节，SHA256 `DFF741233A5D9FA6C649479F45594C27682EF014A25BE0DBD08B9B5E303A22CA`，pg_restore目录检查通过；旧镜像`sub2apicust:before-auth-ip-mode-20261013`为33e2ba735e66，支持临时到期及读取永久名单，但不能设置永久模式，旧前端PUT不含block_mode，新接口会400要求刷新同版页面。不得回退或删除252迁移。启动日志另有原定价更新源连接0.0.0.0失败；本轮未调整该独立网络问题，不宣称远程定价更新通过。

以下三条为首次全请求／永久封禁实现的历史证据，当前实现和本地运行状态以上述修订为准。

- **2026-10-13 敏感认证 IP 黑名单实施完成（源码，未发布）**：用户要求按截图慢速探查实施，默认周期改为120分钟。新增迁移250、Redis TIME滑动窗口、PostgreSQL持久名单、配置/状态版本、管理员接口与“IP黑名单”菜单。默认自动拉黑关闭，10次/120分钟，第10次立即403永久封禁；成功失败、跨敏感接口共同累计，管理员已登录会话及模型调用不接入名单。解封后从新窗口计数，旧请求不能重封。实施和上线说明`deploy/AUTH_IP_BLACKLIST.md`；接缝逐处登记CUSTOMIZATIONS。
- **本轮核验终态**：Docker Go1.27.2实际通过9项后端专项（真实PostgreSQL完整迁移链两次应用、真实Redis80路并发、截图第10次20:32:32、120分钟边界/同毫秒/时钟回拨、规则/解封竞争、服务重建和Redis清空、真实管理员JWT鉴权及封禁/配置/解封审计）、`go test -tags=unit ./...`、`go build -tags embed`。前端7项专项、40文件511项关键回归、typecheck、定向ESLint、生产构建通过，新spec已加入Makefile供CI执行。验证镜像`sub2apicust:auth-ip-verify`的inspect Id为`sha256:cb3b18ae84dd6089dad15d7625dca580a31d2cf570359040756aa9dd2a33f8db`，二进制SHA256`330dce73771db2658471eb6c61665ffd3181fe0cffaec38697c818de68b5070c`，仅本机工具验证镜像，不是发布镜像。证据`output/auth-ip-blacklist-20261013/VALIDATION.md`及backend-verify.log/frontend-critical.log。初次全量3项阿里云验证码测试超时，测试进程去掉HTTP代理后全量通过；前端关键回归有既存jsdom网络报错和Windows worker退出警告，最终退出0、全部测试通过。两只自建测试容器及自有匿名卷已清理，8080原三服务保持healthy。
- **发布边界/未知**：没有改生产、没有替换8080、未提交推送，新SHA CI/安全/GHCR及golangci-lint未核验。构建验证包含当前工作区源码，发布须先协调其他会话与248/249迁移，不能直接把工具镜像拿去生产。用户确认Cloudflare→Nginx，实际IP链路、头清洗和源站限制以及误拦率未知；默认开关关闭，开启前由用户按上线说明核验，历史审计不回灌永久名单。

- **2026-10-13 认证请求异常咨询（仅核查与建议）**：用户截图显示19:56:43–19:58:13，64.176.228.143先有3次login 401、2次send-verify-code 200，之后register 400及forgot-password 500；另有两个IP的login 400。截图没有成功登录记录，不证明全时段无入侵，也不能仅凭IP确认攻击者。20:29（Asia/Shanghai）匿名GET本站settings/public实取：registration_enabled/email_verify_enabled=true；turnstile_enabled/tencent_captcha_enabled/aliyun_captcha_enabled/totp_enabled/passkey_enabled/invitation_code_enabled/password_reset_enabled=false，脱敏证据`output/auth-abuse-20261013/public-auth-settings.json`。公开设置不是生产完整配置，生产镜像、代理拓扑和错误详情仍未知。本地auth.go确认按IP且各接口独立限流：login20/min，register/send-verify-code/forgot-password各5/min；截图未达到阈值，不据此判定限流失效。auth_service.go:345和email_queue_service.go:103确认验证码200仅代表入队，不保证邮件送达；auth_handler.go:617存在frontend_url缺失直接500的路径，但未证明本次500属于它。建议先启用现有服务端人机验证和管理员个人TOTP，再考虑认证接口共享IP限流、账号失败退避、入队前邮箱冷却/总发信预算及可追溯告警；源站访问和转发IP信任先取证。没有修改业务代码、生产配置，没有发起登录/发信探测或部署/提交推送；外部官方文档检索未取得可引用结果，本轮判断基于截图、本站公开接口和本地源码。



- **最终发布状态**：2026-10-13最终发布：官网与登录页动效、交互、首屏加载已推送origin/main，功能提交`f8928e20dac15b341fc8d657e421b6f8ed0a8d71`（从7cd18f823快进）。同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38361578948)（含真实PostgreSQL的unit/integration、golangci-lint、前端）、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38361578790)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38361578902)全部成功。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-f8928e2`（linux/amd64），摘要`sha256:cd8c1667f943c0a6706742dcc3c0e9a1c203e75ba53dd28ec528da9f920e3d6a`。镜像在026d9ab的基础上只增加本轮前端改动，没有后端代码和迁移；不含工作区里未提交的10-11安全修复，也不含并行“智力效率”会话的改动。证据在output/site-motion-20261012（release-evidence.json、job日志）。用户备份后执行`./update.sh sha-f8928e2`；生产未代部署。

- **上一版发布状态**：2026-10-10最终发布：包号职责收敛及安全补修已推送origin/main，功能提交`026d9ab2e78e244858e9a57fafe3297e193434ac`。同SHA [CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38025165503)、[安全](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38025165599)、[GHCR](https://github.com/EngineeMoMo/sub2apicust/actions/runs/38025165523)全部成功，含Go unit（独立PostgreSQL）、integration、lint、前端及脚本检查。固定镜像`ghcr.io/engineemomo/sub2apicust:sha-026d9ab`（linux/amd64），摘要`sha256:93d974774096864fcf60a67050586c01fbf7bab27f6f6eb7f349e0e6aa96adeb`。Go1.27.2、x/net0.60.0与x/tools0.50.0已生效，govulncheck零可达漏洞（模块中2项未调用漏洞，不宣称全部依赖零漏洞）；golangci-lint2.14.0通过，旧HTTP/2接口仅有五文件/明确符号限定的弃用兼容规则，迁移另列待办。证据output/security-release-20261010与output/dedicated-release-20261010。用户备份后执行`./update.sh sha-026d9ab`；生产未代部署，真实请求/到期/跨组计费待用户验收。本条覆盖下方历史未发布或失败候选记录，后续纯文档提交不产生新镜像。


- **2026-10-12 官网与登录页动效、交互、加载速度升级（已发布sha-f8928e2，生产未部署）**：用户选的范围是“公开页＋登录/注册页”，加载速度只动前端和构建分包，不改 Go。新增文件和上游接缝见 CUSTOMIZATIONS 顶部新节。
  - 发布（2026-10-13）：
    - 用临时索引构造“HEAD＋本轮改动”的提交，只取本轮内容，共23个文件。
    - 推送前，在干净工作树里用与远端一致的工具链（vite 5.4、vitest 2.1、eslint-plugin-vue 9）复验：lint 0 问题；typecheck 通过；关键回归37个文件、488项；全量372个文件、2974项；`pnpm run build` 成功，骨架和 zh 预加载都在，vendor-airwallex 独立成块且首页不预载。
    - 推送后，同 SHA 的 CI（真实 PostgreSQL 的 unit/integration、golangci-lint、frontend、shell、release-helpers）、Security Scan、GHCR 全部成功。
    - GHCR 日志里的 vendor-airwallex-BeRoWX1f.js、vendor-misc-COBd48M0.js 与复验构建一致。镜像和摘要见第四节首条。
  - 交互与动效：
    - 页头吸顶，滚动后变成磨砂效果；向下滚动收起、向上滚动出现，锚点跳转后章节标题不会被遮住。
    - 手机端导航收成菜单按钮，页头从 194px 降到 80px。
    - 首屏标题依次出现，品牌卡增加轨道与随指针移动的光斑。
    - 展台改为滑动指示条，四个产品各有专属动画。
    - 首页、套餐、API 页加了滚动出现；套餐页加载时显示骨架。
    - FAQ 展开箭头旋转，支持的浏览器上高度平滑展开。
    - 登录/注册页依次进场。
    - 主题切换从按钮位置圆形扩散。
    - 以上动画都受“减少动态效果”保护，motion.spec 会检查。
  - 加载速度（生产用 curl 实测，本地 dist 对比）：
    - 生产 HTML 有 gzip，JS/CSS 没有 `Content-Encoding`，未压缩的关键路径约 1.43MB。这是服务器 Nginx 配置问题，代码管不到，见第五节。
    - 生产首页 HTML 47KB，其中 43.5KB 是两份 base64 站点 Logo。
    - 构建时注入首屏骨架和早期脚本：提前设置深色，并让 140KB（gzip）的中文语言包与入口 JS 并行下载，原来要等入口 JS 执行完才开始。vite preview 中实测 zh 只请求一次。
    - 新发现：Airwallex SDK 原本并入每页都加载的 vendor-misc，首页因此会预载两个外域脚本。已仿照上游 Stripe 规则独立成块：预览首页的外部请求变为 0，支付页仍按需导入。
    - 体积对比（改动前 dist 对新构建）：首屏关键路径 raw −26.3KB、gzip −9.2KB。其中 vendor-misc 从 280.9KB 降到 238.6KB；CSS +16KB、HTML gzip +1.3KB 已计入。
  - 验证：
    - 前端全量在目录迁移前通过：372 个文件、2979 项，含新增 21 项。该日志后来被覆盖，这个数字来自会话记录。
    - 最终代码上：Makefile 关键回归 39 个文件、504 项通过；typecheck 通过；lint 0 错误（26 条警告都在未改动的上游文件里）；vite build 成功。
    - 浏览器（dev 代理生产公开只读接口）：1440/375 两种宽度、浅色与深色。控制台无警告。逐项核对了展台指示条位移与 SynaRoute 紫色、工坊三图回到原角度、教程锚点标题可见、磨砂页头、手机菜单 Esc 关闭后焦点回到按钮且关闭状态不可聚焦、主题扩散、套餐骨架、登录动画。
    - 证据：output/site-motion-20261012/（size-compare.json、各项日志、screens 下的前后截图、prod-headers）。
  - 未验证：真实网络和弱网下的首屏时间、Safari/Firefox、真机触屏、系统真正开启“减少动态效果”时的表现（只有单测覆盖）、Go 后端真实替换 nonce、生产部署。
  - 注意：
    - 骨架插件放在 `custom/shell/`，因为根 `.gitignore` 会忽略任何 `build/` 目录。最初放在 build/，自查时发现并已迁移。
    - 本机被忽略的 `frontend/vite.config.js` 是 vue-tsc 的产物，内容停在 10-01，会在 `pnpm dev` 时覆盖 .ts 配置。已重新生成。CI 是干净检出，不受影响。
    - 临时的 `.env.development.local` 和预览构建目录已删除。
  - 提交边界：本次提交只含本轮文件。不含 10-11 安全修复，它仍在工作区，未提交；也不含并行“智力效率”会话的改动（intelligence 前后端、router.go 及 theme.css 中两行 `.mofa-intel-*`）。中途一次全量里这两个 intelligence spec 曾有 11 项失败，之后复查 custom 目录加关键回归，共 60 个文件、708 项全部通过；它们的状态以该会话交接为准。

- **2026-10-12 智力效率改接雷达Bench v1（源码完成，未提交/推送/部署）**：用户给出 https://codex-reset-radar.pages.dev/#codex-community 和截图，要求更新数据。接缝、契约和缓存规则见 CUSTOMIZATIONS 顶部「2026-10-12 智力效率改接雷达Bench v1」。
  - 已核实的原因（curl/Node 实取）：
    - 旧 JSON 被源站标为“旧基准”，源时间停在 2026-10-01 23:00（+08），文件 9.7MB。
    - 新汇总接口 GET 响应没有 `Access-Control-Allow-Origin`；非默认档位不带 `X-Radar-Bench-Selection` 时返回 428。
    - 所以改为后端代读并缓存，前端只读本站 `GET /api/v1/intelligence`。
  - 数据核对：
    - 实取 34 个汇总，Go 解析结果与“按源站页面公式计算”的值逐项比对，0 处不一致。
    - 与用户截图相比，28 格完全一致；另外 6 格是源站在两次读取之间有新判分，数值已变化，证明是实时数据。
    - 实时一轮（Go 客户端直连源站）34/34 成功，用时 7.75 秒。
  - 验证：
    - Go：定向 13 个顶层、24 个子项通过，连续 20 次无偶发失败；handler、server、middleware、routes 四个包的全部 unit 通过；embed 构建、vet、gofmt 通过。golangci-lint 本机没有，以 CI 为准。
    - 前端：intelligence 两个 spec 共 33 项；关键回归 36 个文件、372 项；vue-tsc、定向 ESLint、生产构建通过。上面一条记录里“2 个 intelligence spec 共 11 项失败”是本会话改写进行到一半时的状态，现已全部通过。
    - 浏览器：用真实页面组件、实取数据、页内模拟接口的临时夹具检查。954 与 1440 浅色、1440 深色边界状态、375 手机：页面不横向溢出，手机端矩阵在自身容器内滚动、首列吸附，“覆盖／判分”组内不断行。检查后夹具已删除。
  - 证据：`output/intelligence-radar-bench-20261012/`（raw 原始响应、go-captured/go-live、compare.mjs、derive.mjs、make-fixture.mjs）。
  - 未验证：
    - 8080 和生产的真实登录调用（本机 8080 未运行）、golangci-lint 与新 SHA 的 CI。
    - 生产服务器能否访问 pages.dev，是否需要配置 `update.proxy_url`。
    - 源站以后是否收紧“显式选择”限制。后端使用源站页面同样的标记；一轮成功后 10 分钟内不再读取，失败后按 Retry-After 或 1 分钟起退避，而且只在有人打开页面时读取。
  - **2026-10-13 本机 8080 已部署供用户验证**（用户授权）：
    - 构建源是已发布的 HEAD `7cd18f823` 加本轮 10 个文件和 theme.css 三条规则（`output/intelligence-radar-bench-20261012/docker-src`）。不含其他会话的未提交改动，所以 248/249 迁移没有带入，数据库仍停在 247。
    - 先备份数据库：646652 字节，pg_restore 目录 1270 项，SHA256 `95c559e3…02d6b`；旧镜像打标签 `sub2apicust:before-intelligence-20261013-090718`。
    - 结果：build 退出 0；运行镜像与构建镜像一致（`5647debb2e93`）；postgres/redis 容器 ID 不变，recipes 保持停止。health 200，未登录访问 `/api/v1/intelligence` 返回 401。前端入口引用 `IntelligenceView-CEVNcjIO.js`，包里不再有旧 JSON 地址。各表计数和迁移不变，只有两个启动时自动同步的客户端版本键变化。
    - 应用容器创建于 10:06（+08）。13:48 执行 `up` 时它已在运行且是新镜像，没有重建。我没有找到 10:06 是谁启动的。
    - **阻塞：容器访问不了外网**。Docker Desktop 设置为手动代理 `http://127.0.0.1:7897`（Clash）。该端口在监听，但经它的请求全部 5 秒超时，包括 auth.docker.io 和 baidu；主机直连 pages.dev 和 baidu 都是 200。容器内 HTTPS 访问任何站点都被重置，HTTP 返回 502。所以本机页面曾会显示“数据源暂时不可用”。这是本机代理状态问题，不是代码问题；09:07 构建时同一代理还是通的。**14:07 已解除**：用户在 Clash 换了节点后，主机经 7897 访问 pages.dev 和 baidu 都返回 200；容器内读取绑定文件（24054 字节）和带选择标记读取 astra/high 汇总都成功。到 14:07 为止，后端还没有收到登录后的 `/api/v1/intelligence` 请求，缓存为空。**14:12 用户确认已在 8080 看到智力效率数据**。后端日志：登录后的首个请求返回 200，用时 10.9 秒（首次读取源站，冷缓存）；没有 `intelligence source` 失败或 panic 告警。构建用的 docker-src 副本（203MB）已删除，需要时可用 `prepare-docker-src.sh` 重新生成。
    - 记录：`output/intelligence-radar-bench-20261012/local-deploy.json`、`local-before.json`、`docker-build.log`。
  - 提交时分开：本轮文件是：
    - 新增 `backend/internal/handler/custom_intelligence_handler{,_test}.go`；
    - `router.go` 两行；
    - `frontend/src/custom/intelligence/{data,useIntelligence,copy}.ts`、`views/IntelligenceView.vue`、`components/IntelligenceComparison.vue`、`__tests__/intelligence{,-view}.spec.ts`；
    - theme.css 中三条 `.mofa-intel-*`（price/time 字号、`.mofa-intel-pair`、`.mofa-intel-sample` 右边距）；
    - CUSTOMIZATIONS、custom/UPGRADE 和本文件的对应新节。
    - theme.css 和文档里同时有官网动效会话的改动，要按块暂存。

- **2026-10-11 审查问题全部修复（源码完成，未提交/推送/部署）**：用户要求“全部修复”。接缝和规则见 CUSTOMIZATIONS 顶部新节与 deploy/PAYMENT_COLLECTION_LIMITS.md。
  - 收款预占：迁移248按用户封顶；已结束的订单在最晚可付款时间后由渠道复查释放，管理员也可写原因并经审计人工释放。启用限额后，手机 WAP 也带 time_expire。
  - 工坊：撤回删除记录和素材，退回删除素材；读正文前查额度；并发上传设上限；改为流式写盘；后端拒绝视频投稿；新增管理员删除；公开画廊按 IP 限流。
  - 其他：公开套餐目录限流；迁移249让限量套餐每人同时只能有一笔待付；update.sh 只改 sub2api 的 image；前端工具链升级到 vite6、vitest4、typescript-eslint8、eslint-plugin-vue10；清空过期的审计例外。
  - 验证：后端定向与包级 unit 共 10221 个顶层和 9471 个子项通过。唯一失败 TestOllamaProbeCallback_StaleLongDoesNotOverrideNewShort 是上游既有问题，在干净 HEAD 的临时 worktree 中 10/10 同样失败，与本轮无关。另有 15 项跳过，包括新增的 248/249 真实 PG 测试，原因是 Docker 未运行。gofmt、go vet、golangci-lint 2.14.0 均 0 问题；embed 构建和 TestFamily 通过；govulncheck 修复后仍是 0 个可达漏洞。前端 369 个文件、2958 项全部通过，Makefile 关键回归 35 个文件、479 项通过，vue-tsc、lint（0 错误）、生产构建通过。update-test 共 26 个场景通过。
  - 依赖：pnpm audit 生产依赖 0；全量从 55 条降到 2 条，剩下的都来自 tailwind 3：braces 上游无修复，postcss-selector-parser 6.x 为中危。
  - 复验：修复后用相同攻击步骤复验，用户1上传并撤回18份后磁盘剩0个文件，用户2投稿返回201（修复前是429）。
  - 旧数据：修复前已撤回或退回的投稿不再计入额度，但它们的素材文件仍留在磁盘上。作者可以在“我的投稿”里撤回删除，管理员可以在审核页删除。本轮没有自动清理，也没有改动线上数据。
  - 未验证：Linux CI（Node20、mawk）、完整 integration、支付宝真机、生产配置。证据在 output/security-review-20261011/（*-fixed.log、frontend-full-final.log、backend-unit.jsonl、pnpm-audit-all-step3.json 等）。改动前的 package.json 和锁文件备份在 deps-backup/。
  - 真实 PG（用户同意后启动 Docker；隔离容器 postgres:18-alpine，绑定 127.0.0.1:15432。55432 在 Windows 保留端口段内，绑定失败）：
    - PG 专项：12个顶层和79个子项全部通过，0 跳过。包括新增的 248/249 触发器、真实时间戳下的复查粗筛、同一用户并发10笔只成功3笔、限量套餐并发5笔只保留1笔待付，以及既有包号、收款、库存、订阅 PG 回归。
    - 既有收款与库存夹具改为在 248/249 最终函数上验证，并补 user_id 列。
    - 新增 `custom_collection_release_schema_postgres_test.go`：用完整迁移链建新库并重跑一次，在真实表上验证复查、待确认列表、人工释放审计和按用户封顶。
    - 设置 DSN 重跑改动包 unit：10232个顶层和9550个子项通过；跳过从15降到5，剩下的都是外部依赖占位；唯一失败仍是上游既有的 Ollama 用例。
  - 真实 PG 暴露的缺陷：247 时期的 `CustomCollectionUsage` 第二次查询失败时，会把第一个 defer 捕获的 rows 覆盖成 nil，导致崩溃（SQLite 夹具没有触发）。已改用独立变量，并补“查询失败返回错误而非崩溃”回归。
  - 收尾：测试容器和它产生的2个匿名卷已删除。10-06 的旧匿名卷保留；4个业务容器保持原来的停止状态。Docker Desktop 仍在运行，需要时由用户自行退出。
  - 10-12 补跑复核：
    - 09:18 的 lint 早于 nil-rows 修复（09:38）和新增 PG 测试（10:26），因此在最终代码上补跑 gofmt、go vet（含 -tags=unit）和 golangci-lint 2.14.0，结果均为 0 问题。日志：static-recheck-20261012.log。
    - 设置 DSN 的 unit 在 10:31 运行，晚于最后一次代码改动。新增的 TestCustomCollectionReleaseRealSchemaPostgres 不在 09:40 那批 PG 专项（12/79）里，是在这次 10:31 的运行中通过的（2.87 秒，真实执行，未跳过）。
    - update-test 的 26 个场景在 GNU Awk 下通过（update-test-20261012.log），在一次性 postgres:18-alpine 容器里用 busybox awk 也通过（update-test-busybox-20261012.log）。该容器断网、以只读方式挂载 deploy，运行后已自动删除。
    - 本机没有实测 mawk。

- **2026-10-11 最新代码漏洞审查（仅审查，未改业务）**：范围是HEAD=7cd18f823（=origin/main，已fetch核对），加上工作区既有的5个未提交收款文件。审查期间用过两个临时探针，均已删除；`git status`仍只有原来那5个文件。证据在output/security-review-20261011/（探针源码、日志、pnpm audit JSON、govulncheck）。按用户要求没有启用子代理。
  - 【已核实·P1，仍未修】普通用户可以主动占满收款预占。`is_mobile`由客户端声明（payment_handler.go:266、295–296）；只有电脑WAP会写time_expire（alipay.go:155–159）。用户在手机WAP下单后不打开付款页、直接取消，关单返回交易不存在，额度不会释放。默认每人最多3笔待付（payment_config_service.go:51），取消限流默认关闭（同文件:291），所以可以反复下单再取消；20笔¥50就占满默认¥1000。后果只在启用限额后出现，但账本平时也会累计这些占用（247:39–46）。工作区未提交的复查释放已经编译、vet、gofmt通过，探针结论如下：电脑WAP（已取消）和支付宝建单失败且无付款地址这两种情况可以释放；手机WAP、电脑网站支付不在范围内，易支付既没有CancelPayment也没有复查器，同样不释放。新增函数还没有任何测试。真实支付宝对“未打开的WAP订单”关单会返回什么，没有在商户或沙箱取证。
  - 【已核实·P2】单个用户可以占满工坊存储。撤回只改状态，不删文件（custom_studio_handler.go:424）；额度统计包含所有状态的记录（:236、:244）；没有删除接口。探针中，用户1上传并撤回18份，共536870912字节，正好是全站上限，媒体文件18/18都还在磁盘上；用户2随后上传一张32px PNG，返回429。提示文案写的是“每人最多…”，会误导用户。另外，额度检查发生在完整读取30MB正文之后，也没有并发上限，内存峰值风险未压测。
  - 【已核实·P2】工坊公开画廊和公开媒体接口对匿名访问实际不限流。它们挂的是`limiter.Global()`（:76），这个限流器按用户计数，没有登录主体时直接放行（panel_rate_limit.go:75–78），现有测试也把这一点作为预期行为。上游的匿名接口用的是`PublicIP()`（auth.go:244、model_plaza.go:24）。`/payment/public/plans`同样没有限流器，每次请求要查询多次DB（低危）。生产反代是否另有限流，我不知道。
  - 【已核实·低】有限库存可以被待付订单临时占住：PENDING计入库存（245迁移），每人最多3笔，30分钟超时。update.sh带标签运行时，sed会改写override里所有`image:`行（用副本验证过）。按DEPLOY_CUSTOM的生产override只有一行image，不受影响；本机8080的override有3行，postgres和redis的标签会被改坏。
  - 【已核实·依赖】pnpm audit --prod：所有严重度都是0。全量audit（含dev）共55条：critical 3、high 33、moderate 18、low 1，全部来自vite5／vitest2（tinypool）／rollup／eslint链等构建或测试工具，不进生产镜像，CI也不扫。govulncheck（DB 2026-10-08）：可达漏洞0，另有2个漏洞在依赖模块里但代码未调用。audit-exceptions里lodash／axios的例外已过期，nanoid的10-13到期；但检查脚本只在漏洞仍被扫出时才处理例外（check_pnpm_audit_exceptions.py:186–233），当前生产审计为0，所以不影响CI。
  - 【已核实·未发现问题】支付宝通知会拒绝重复参数、缺项和商户不符，删除cert_sn，按分精确比对金额；扫码令牌32字节，比较用常量时间，只跳转到openapi.alipay.com。包号的免扣凭证只挂在单次请求的Key副本上，结算时按实际账号核对lease；调度入口都有defer检查，WS逐帧逐轮复查。工坊有ID正则、nosniff和CSP sandbox，私有媒体只有本人或管理员能访问。fork没有新增v-html；postMessage会校验origin、source和nonce；JWT不读Cookie；在改动文件和已跟踪文件里做高特征密钥扫描，零命中。工作区代码build、vet、gofmt通过，定向190项unit通过，另有4项真实PG测试因缺DSN跳过。
  - 工具链：宿主默认GOPATH里有go1.27.2 toolchain模块和x/net 0.60.0，可以离线编译和测试（见记忆portable-go-toolchain）；output里的便携1.27.0已经不满足go.mod。
  - 未覆盖：真实PG和Docker、生产配置（限额是否启用、反代限流）、真实支付宝或易支付关单、上游最新变更、浏览器实测和压测。

- **2026-10-10 用户授权提交推送包号职责收敛**：本轮仅发布包号归属/有效期/免扣及通用限制回归相关修改，origin/main抓取后与本地基线一致；4个既有收款修复文件及先前审查记录保留在工作区、不混入功能提交。CI配置独立PostgreSQL，发布须以本轮实际SHA确认CI（含真实PG及integration）、安全和GHCR全部成功后再交付镜像更新命令，生产由用户自行更新。

- **2026-10-10 到期与跨组免扣专项验证**：新增custom_dedicated_expiry_test.go，真实Gin/APIKeyAuth链配SQL mock覆盖7场景：有效专属组零余额可用；到期专属组零/正余额均403；到期关联组403；普通组到期前后零余额均INSUFFICIENT_BALANCE；普通组正余额放行但IsCustomDedicatedPrepaid=false，缓存原Key始终无免扣凭证。连同到期边界、选中账号、跨Key/组结算及缓存隔离回归共7顶层+13子场景通过，无跳过；证据output/body-limit-fix/expiry-verification-final.jsonl。不是生产/真实PG验收。到期前已准入的HTTP请求可完成并按原权益结算，不表示到期后能发起新请求；普通组继续走原生余额/订阅规则。未修改业务逻辑或部署。

- **2026-10-10 包号职责收敛（用户最新要求，覆盖以下旧限额方案；未发布）**：用户明确“包号只负责账号归属、有效期和免扣，通用请求限制交给sub2api”。已删除包号独立并发/RPM/日计数、请求体解析与大小限制、生图开关、HTTP白名单、WS帧类型/JSON深度限制、成员100人上限、后台billing-policy接口及配置UI。普通Key限额恢复按原ActualCost含倍率口径计量，不再强制原始参考价；用户余额免扣、归属隔离、撤销/到期及WS逐帧/逐轮权益检查保留。历史策略/计数表不再读写；旧lease表仅保存已完成的免扣凭证，无活动占位、续期或失败取消。细节及上游接缝见CUSTOMIZATIONS、deploy/DEDICATED_ENTITLEMENT_ONLY.md。
  - 验证：包号后端37顶层+27子项通过，5项真实PG测试因无专用DSN跳过；原生计费/WS/异步图片等407顶层+361子项通过，无跳过。前端43项、vue-tsc、改动文件ESLint、便携Go embed构建通过；证据output/body-limit-fix/{scope-go-tests,native-go-tests}.jsonl。Docker未运行，未做真实PG/线上验收。未提交、推送或部署；保留既有收款修复改动。

- **2026-10-10 包号请求体修复与其他定制盘点（未发布）**：按用户要求移除HTTP/WS包号独立大小阈值及后台/API字段，历史数据库列保留但不读写；通用网关限制仍生效。前端27项与vue-tsc通过，便携Go定向service/middleware测试通过（真实PG专项未设DSN跳过；Docker未运行），便携Go embed构建通过。其他仍存在：共享并发默认2、RPM默认30（两者不能设0）、日请求默认0、默认关闭生图、HTTP入口白名单、WS只允许response.create/cancel、禁生图时嵌套深度大于64也拒绝；成员上限100、账号/标准专属组隔离、到期撤销、Key金额参考计量及租约DB依赖。仅盘点其他规则，未擅自删除。保留工作区既有收款修复，未提交推送部署。

- **2026-10-10 截图413核查（仅分析）**：DEDICATED_REQUEST_BODY命中custom_dedicated_billing.go:40，为本仓包号定制；引入提交acd73d716（2026-10-01，Git作者mohaiming）。默认2 MiB，可配置1024至33554432字节；未取证生产实际策略和请求大小，不能确认现场阈值。未改业务或部署，保留既有收款修复改动。

- **2026-10-09 设计／逻辑漏洞／页面美观专项审查（仅审查，未改业务）**：范围HEAD=16334e80c（=origin/main）。只读审查加一个临时探针测试，探针已删除，`git status --porcelain`为空。证据在output/design-review-20261009/（probe-collection-hold.log及探针源码副本）。按用户要求不启用子代理；首轮5个子代理因网关529失败，没有留下工作树或分支。
  - 【已核实·P1】收款预占可能永久占满全站日额度。只有`checkPaidWithOptions`里`CancelPayment`返回nil时才会置released（payment_order_lifecycle.go:204–213，全仓唯一调用点）。247触发器统计未付且未释放的占用时没有日期边界（247_custom_payment_collection.sql:39–43）。EasyPay没有实现CancelPayment。支付宝未创建交易时关单返回“交易不存在”，本仓custom_alipay_collection_test.go:23就是按这个语义编写的。网关失败产生的FAILED订单之后没有任何流程处理。ExpireTimedOutOrders只扫描PENDING订单。待付上限只统计PENDING（payment_order.go:285），用户可以用支付页“取消订单”（PaymentStatusPanel.vue:464）反复下单再取消。探针用仓库SQLite夹具复现了5种场景：支付宝未扫码后取消、过期、网关失败，以及不支持关单的渠道取消、过期。每种场景20笔¥50订单后，占用都是¥1000，管理员“查询并关闭”仍返回COLLECTION_CLOSE_UNCONFIRMED。“新订单会被拒”来自触发器的静态阅读，真实PG未运行。真实支付宝对未扫码订单关单返回什么，还没有在商户或沙箱取证。
  - 【已核实·P2】API Key鉴权热路径每个请求多出2–3次未缓存的DB查询：CustomDedicatedContext→Check、PrepareCustomDedicatedBilling，以及调度时的guardSelection。服务始终注入（wire_gen.go:329、custom_dedicated_accounts.go:85–88），所以所有带分组的Key都受影响。任一查询出错就返回403 DEDICATED_ACCOUNT_UNAVAILABLE（api_key_auth.go:168–179），普通用户也一样。延迟影响未压测。
  - 【已核实·P2】CI前端只跑Makefile关键清单里的34项；custom/__tests__共32个文件，其中23个不在清单内。真实PG测试缺DSN时会Skip；CI配置了DSN，但没有“必须实际执行”的守卫。
  - 【已核实·页面】支付宝按钮#00AEEF配白字，按WCAG相对亮度公式手算对比度约2.5:1（style.css:136，上游代码）。工作区加载层#f5f8f8没有深色样式（theme.css:1605）。工坊段使用硬编码hex颜色。多个custom页面直接写中文、没有走i18n，但页头仍有语言切换。PlanContactPanel与管理员说明重复，还向用户暴露了“库存”这个内部术语。充值页的提示放在“快捷金额”标签前面，快捷金额也没有币种符号。截图来自10-08模拟夹具，不是生产页面。
  - 【已核实·正面】后端68个被改的上游文件都有[CUSTOM]标记，也都已在CUSTOMIZATIONS登记。Ent生成文件没有手改痕迹。迁移执行器按完整文件名记录并校验checksum；同号前缀是上游惯例（共47组），241同号不会出错。已跟踪文件里没有真实密钥。PNG母图已被.dockerignore排除，不会进入镜像。
  - 【已核实·文档】CLAUDE.md:64和本文件三-3的“本机无Go”已过期。CUSTOMIZATIONS里21个“未发布”标题，多数实际已经发布。第五节有已完成但未勾选的项，例如update.sh已在4bc5abbb5提交。
  - 未覆盖：真实PG、商户和生产数据，全量前端浏览器实测，上游最新变更（未fetch），以及库存／订阅全部路径的逐行审计。

- **2026-10-08 收款限制与人工开通发布完成（最新终态）**：功能`575467a5f4d194ebd01cf771edacef9f0075d395`已推origin/main，同SHA [CI37749962687](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37749962687)、[安全37749962637](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37749962637)、[GHCR37749962677](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37749962677)全部completed/success，含Go unit／integration／lint及前端。实际发布镜像`ghcr.io/engineemomo/sub2apicust:sha-575467a`（linux/amd64），摘要`sha256:5539364a4f724569a4f0ef22bf12db74357fce5b6e3669327be619137d5e408e`。证据output/collection-release-20261008。包含247迁移、共享收款上限、快捷充值配置、人工开通模式及两项审查修复。用户备份后执行`./update.sh sha-575467a`，先核对旧付款链接与线下收款再启用限额；真实商户与生产交互待验，未代部署生产。本条覆盖此前待发布记录；后续文档skip-ci提交不生成新镜像。

- **2026-10-08 用户授权发布收款限制与人工开通**：本轮完整源码、两项审查修复及交接文档提交推送origin/main；本地84项补修回归、后端支付定向与Docker生产构建／迁移健康已通过。发布须检查实际功能SHA的CI、安全、GHCR，不沿用旧结果；生产由用户自行更新。

- **2026-10-08 本机Docker已更新供用户测试**：用户授权本地更新，代理预检200、首次构建成功；运行镜像与构建镜像一致`2a844d6e77237dd55986627d8fbb9609af29ab4ce2022b3437d0980234246bc0`。原Postgres／Redis启动并保持容器ID、卷，应用仅重建且挂载不变，三服务healthy；health、brand、login、plans及公开套餐接口200。库存／sales_mode字段与独立收款账本已确认存在，更新前后用户3／套餐1／订单0不变。旧镜像`sub2apicust:before-collection-20261008`、616870字节数据库备份已pg_restore目录验证，证据`output/local-collection-20261008`。地址http://127.0.0.1:8080；原recipes容器保持停止。未改付款设置、未下单或真实商户验收；未提交推送、未操作生产。此条覆盖本轮旧“未本地部署”状态。

- **2026-10-08 收款审查两项修复（未发布）**：支付宝 CancelPayment 不再吞掉交易不存在，且检查 SDK 响应成功码（业务失败可能 error=nil），只有明确关单成功才释放预占；人工续费改去不受在线支付开关限制的 `/plans` 目录查看联系说明。新增本地签名 SDK 关单回归，支付 provider／service 定向测试通过；前端4文件84项、vue-tsc及修改文件eslint通过。首次回归揭示 SDK 响应业务错误并补齐成功码校验，最终重跑成功。未测试真实商户、未重跑全量CI；未提交／推送／部署。

- **2026-10-08 收款限额与人工开通（源码完成，未发布）**：用户确认开始实现。新增247迁移、全站人民币共享单笔／日限额（默认关闭，预填50／1000与北京时间）、事务预占与独立收款账本；跨日未付保留，确认渠道关单才释放，退款／删除不恢复已收额度。后台独立配置快捷金额、自定义开关、联系说明、统计／待确认关单与线下收款幂等登记；套餐sales_mode默认online，可选contact_admin，公开目录／购买／深链接／续费同步，管理员实际发放仍扣库存。297项后端定向无失败／跳过（含真实PG32笔争抢仅20笔成功、完整迁移／库存／API契约），前端首轮357项＋新增配置与最终支付专项57项、类型／lint／前端生产和后端embed构建通过；375／1440浅深四组合模拟API浏览器校验通过，证据output/payment-collection-20261008。首次ORM生成遇Windows映射锁，已在隔离镜像目录完整生成并原子回写，非功能损坏文件还原。初始浏览器夹具缺邮件／备份字段导致错误提示，补齐后重验，不将夹具故障写为产品问题。历史／外部收款须核对，旧付款链接启用前先处理；真实商户关单与到账、完整Linux／新SHA远程CI仍待。未提交／推送／业务镜像重建／生产部署，详见deploy/PAYMENT_COLLECTION_LIMITS.md。

- **2026-10-08 库存与帮助改进发布完成（最新终态）**：最终功能70e66cf7305e0cb4610671b5120e3612f3478fb6已推origin/main；同SHA [CI37720127131](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37720127131)、[安全37720127128](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37720127128)、[GHCR37720127116](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37720127116)均completed/success，含Go unit／integration、lint、前端与脚本。镜像ghcr.io/engineemomo/sub2apicust:sha-70e66cf（linux/amd64），实际manifest及containerimage.digest为sha256:bf94b9dcfca53dd394660904566bef67adc979f8abc721d607d54fc77964440e。证据output/release-stock-help-20261008/runs.json、ci-jobs.json、image.log、tests.log。内容包含245／246库存与人工分配扣减、教程六节分类、FAQ17题、不同帮助图标，以及用户指定的“本站支持生图，需联系管理员开通权限”。首候选86ec63760仅测试关闭返回值漏处理触发errcheck，已补修，新SHA全绿。用户备份后自行执行./update.sh sha-70e66cf；先核对历史库存归属再设限额，真实客户端／商户／生产视觉待验，未代用户部署。此条覆盖以下未发布／待授权记录；后续纯文档skip-ci提交不产生新镜像。

- **2026-10-08 发布门禁补修**：候选86ec63760已推送，安全与前端检查通过；远端golangci-lint发现新增库存测试未处理client.Close返回值，已按本仓测试惯例显式忽略关闭结果。须用补修后的实际SHA重新核验三套检查，再交付镜像标签，不把候选状态当最终成功。

- **2026-10-08 用户授权本轮提交推送**：发布库存（含管理员分配）与帮助页改进；追加用户明确要求的“本站支持生图，需联系管理员开通权限”，同时放入接入前准备和API接入FAQ（现17题）。本机文案／回归确认后提交origin/main，必须检查实际功能SHA的CI／安全／GHCR，不能引用旧结果；生产仍由用户自行更新。

- **2026-10-08 帮助图标与教程分类补充（源码完成，未发布）**：用户指出常见问题／接入教程图标重复且教程不详细。控制台原两项均ChannelIcon，现分别复用questionCircle／book，游客预览同步；教程重排为准备、命令行、桌面应用、编辑器扩展、验证、排错六节，保留原客户端示例与旧锚点，新增字段含义、配置生效／验证和400／429／网络错误步骤。FAQ增加四题至16题及四类可组合关键词筛选。样式仅theme.css，官方来源及接缝见GUEST_PORTAL与CUSTOMIZATIONS。4文件56项回归、相关lint、类型及生产构建通过；Edge隔离夹具验证375／1440浅深色四组合，目录跳转／分类通过、无横向溢出或页面异常，截图output/help-20261008，日志output/help-{tests,build,lint}.log。只用了公开数据夹具，未登录真实账户或调用模型；库存未提交改动保留。未提交／推送／业务镜像重建／部署，真实客户端接入和生产视觉待验。

- **2026-10-08 库存补齐管理员与非购买途径（最新源码状态，未发布）**：用户明确管理员分配也须扣库存、订阅页展示库存、0 不可订阅。新增246归属账本，管理员单次／批量分配显式选择库存套餐，实际新增／过期重开／延长有效期与扣减同事务；有效订阅幂等复用不重扣，已付履约只登记来源。兑换码／默认赠送共用服务，无法确定套餐或售罄拒绝发放；默认赠送失败不回滚注册。历史唯一套餐且无已付订单者一次回填，多套餐／混合来源／旧多次续期不能准确重建，须人工核对。我的订阅展示同组可售套餐库存，售罄／读取失败／加载中不可续订；既有权益展示保持。真实PG验证购买与人工同时争抢最后一份仅1成功、回滚、批量部分成功、历史重跑与过期并发幂等；4包421项无失败／跳过，前端21文件216项、类型／lint／生产构建与便携Go embed均通过。证据output/plan-stock-20261008/manual-*，操作和边界deploy/SUBSCRIPTION_PLAN_STOCK.md。未提交／推送／重建业务镜像／部署，未运行全量Linux或新SHA远程CI，生产数据／商户／真机视觉未验。本条覆盖首轮仅购买库存范围及测试数。

- **2026-10-08 订阅套餐库存源码完成（未发布）**：用户要求限制售卖量。新增每套餐 stock_limit（-1 默认不限量、0 停售）及数据库维护 stock_used，历史现存订单回填；新购／续费各占一份，未付取消／超时／失败释放，已付／退款／到期不补货。订单事务内条件更新防超卖；延迟付款库存不足拒绝开通并写 PAYMENT_STOCK_UNAVAILABLE 审计，须人工核实商户到账与退款，原过期恢复期限保持。后台配置／列表、用户卡片、游客目录及深链接购买保护完成，样式仅theme.css。4包236项后端定向无跳过、新增CRUD、真实PG 24并发最后一份仅1成功及8组状态／完整迁移链重跑，前端18文件184项、类型／相关lint／构建、便携Go1.27 embed编译通过。证据output/plan-stock-20261008，维护与边界deploy/SUBSCRIPTION_PLAN_STOCK.md，接缝CUSTOMIZATIONS。Windows生成文件映射锁已以容器临时目录完整生成后原子写回解决，无依赖变更；原业务4容器保持停止，隔离PG测试容器已删除（临时数据使用tmpfs），原4容器仍停止，证据containers-after.txt。未提交／推送／业务镜像重建／生产部署；未跑本轮全量Linux测试或远程CI，真实商户与视觉待验。此条不改变上一版sha-07f0ee6的已发布状态。

- **2026-10-07 版本权限与依赖安全修复发布完成（最新终态）**：功能07f0ee69e0b9318edbdea0694e2d5ecd0ae2e24a已推origin/main；同SHA的[CI37582687152](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37582687152)、[安全37582687062](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37582687062)、[GHCR37582687015](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37582687015)全部completed/success，含Go unit／integration、lint、前端与脚本。实际镜像ghcr.io/engineemomo/sub2apicust:sha-07f0ee6（linux/amd64），摘要sha256:638ef47fc3b3d7f185849f90ef0bd8cd48b38732032c20f6176962b87e14821e，manifest推送与containerimage.digest已确认；证据output/security-review-20261007/release-runs.json、release-jobs-*、job-*。前端生产依赖全0告警，本地368文件／2928项通过，远端govulncheck无漏洞、Audit exceptions validated。用户备份后执行./update.sh sha-07f0ee6；验证仅管理员版本可见、公开JSON／首页注入version为空、图标和公告正常。生产未代部署，真实账号验收待用户完成；纯文档skip-ci提交不产生新镜像。此条覆盖下方本轮未提交／门禁待定状态。

- **2026-10-07 用户授权提交推送本轮安全修复**：发布范围为版本仅管理员可见、两个公开版本出口去除实际值、依赖告警清零及回归／交接文件。前端2928项、类型／构建／相关lint与后端embed已本地通过；推送后须核验本次功能SHA的CI、安全扫描与GHCR，不沿用80c30c825。生产由用户部署。

- **2026-10-07 用户要求修复依赖告警（本地已修，未发布）**：移除无源码导入的@lobehub/icons及其UI／Mermaid依赖链，DOMPurify最低版本和override升至3.4.16，锁文件同步，图标SVG保持。生产依赖在线复扫所有严重度均0，退出0，未增加豁免。全量前端368文件／2928项、含vue-tsc的完整build与相关lint通过，SVG安全回归加入Makefile；前轮版本仅管理员可见修复保留。本地重装先遇沙箱清理长时间无进展，终止后获自动审批在沙箱外完成；随后确认Vue入口缺失，按frozen-lockfile强制重导入后回归成功。证据output/security-review-20261007/*-fixed.log、audit-fixed.json及dependency-repair.log，报告deploy/SECURITY_REVIEW_20261007.md。未提交推送部署，线上仍未验证，不声称全项目无漏洞。此条覆盖下方17项待处理状态。

- **2026-10-07 安全抽查与版本仅管理员可见（源码完成，未发布）**：用户要求审查漏洞并隐藏普通用户版本。修复 VersionBadge 非管理员分支，同时公开 JSON 与 HTML 注入的 version 保留空串，不再暴露部署 SHA；管理接口原鉴权保持。13项前端回归、类型／相关lint／Vite构建、公开配置与schema 3包及权限／支付定向4包测试、便携Go1.27.0 embed编译通过；新增测试接入Makefile。在线生产依赖审计0 high／critical、13 moderate／4 low，未升级依赖，未证实可利用性；报告deploy/SECURITY_REVIEW_20261007.md、证据output/security-review-20261007。未全仓逐行审计、未重跑后端govulncheck或全量Linux／PG测试，不承诺无漏洞。未提交推送、未改容器或生产；实际已发布版本仍为下条80c30c825。

- **2026-10-07 开通周期与支付宝直码正式发布完成（当前终态）**：用户授权提交并推送供自行更新。最终功能80c30c8250d7de59da55b964251f1c437bcf6ddd已推origin/main；同SHA的[CI37575276507](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37575276507)、[Security37575276518](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37575276518)、[GHCR37575276525](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37575276525)均completed/success，CI含unit／integration、Go lint、前端类型及关键回归、macOS脚本和发布辅助检查。前一候选ba405f1a8安全扫描被原SheetJS两项过期例外拦截，已升级官方固定0.20.3并移除例外，未延长豁免；本地前端366文件／2923项、类型／构建通过，审计0 high／critical（4 low／13 moderate），远端Audit exceptions validated及后端No vulnerabilities found已保存。成功镜像日志核实ghcr.io/engineemomo/sub2apicust:sha-80c30c8（linux/amd64），摘要sha256:f3ed8f3fa622bc3ed629472c5c2f0d6e343723646ec030a5d50ee99ca59989f3；证据output/subscription-wap-20261007/release-runs.json、release-jobs-*及job-*。用户先备份，再在部署目录执行./update.sh sha-80c30c8；244迁移对齐周／月开通周期并保留已用量／到期时间，旧聚合用量保守结转。支付宝必须新建订单才使用官方直码；旧码兼容附加flow参数。真实App直达、商户权限及小额实付仍待用户更新验收，未代用户生产部署。此条覆盖以下历史未发布状态。

- **2026-10-07 发布门禁补修过程（已完成，见上条）**：功能ba405f1a8已推origin/main，安全扫描因原xlsx两条high例外于10月6日到期失败。改用SheetJS官方固定0.20.3包并锁完整性、移除对应旧例外，未延长豁免；真实导出回归加入关键测试。pnpm audit已无high／critical（仍4 low／13 moderate），全量前端366文件／2923项、vue-tsc与生产构建通过（sheetjs-tests.log／sheetjs-build.log）；新SHA远程工作流需核实后才交付更新标签。


- **2026-10-07 用户授权发布本轮修复**：提交并推送订阅开通周期／244迁移与支付宝官方直码，供用户更新真机测试。已完成的本地验证见下一条；本次实际SHA的CI、安全和镜像需另核验，不能沿用4bc5abb结果。生产由用户部署。


- **2026-10-07 按用户要求修复订阅开通周期与支付宝直码（源码完成，未发布）**：用户确认继续访问后显示“付款入口无效”，要求扫码直接付款、不经过本站。已用源码与路由回归定位旧入口拒绝支付宝追加flowT／flowSign／flow；旧码只放行这三种单值有界参数，token／订单／目标校验仍保留。新订单QRCode直接使用经过严格校验的支付宝官方签名WAP地址，不生成本站token链接，2300字节上限防二维码溢出；官方SDK四类实付金额与前端M级编码／恢复回归通过。支付宝App真正直接唤起及商户权限仍待真机，新订单必须重建，旧码不自动变新码。订阅开通即启动周／月窗口（7／30天），旧数据通过新增244迁移对齐当前开通周期并保留已用额度与到期时间，已对齐窗口保持。首次兼容激活后重读再推进，列表与progress同周期投影；日历日／有效期续费／管理员主动重置原语义保留。新PG测试独立schema验证回填、幂等及用量／到期保护；首轮测试夹具重名已修后重跑。Go lint 0 issues、前端56项及SDK定向通过；最终Linux unit58包／22,052项通过、17跳过，integration52包／13,583项通过、15跳过（9项安全审计PG/Redis未配置及6项既有／外部依赖，逐项见summary.json），两套均退出0、无失败；新增244真实PG回归实际执行，完整integration的新库迁移及embed构建通过。三个临时容器与专用网络／测试卷已清理，socket挂载已解除，原4个业务容器ID及停止状态保持。证据output/subscription-wap-20261007。未修改业务库或容器，未提交推送生产部署。

- **2026-10-07 用户真机问题核查（未改业务逻辑）**：用户截图显示支付宝对本站public/alipay/wap付款入口提示“您即将离开支付宝”，带继续访问按钮；源码prepareAlipayDesktopWapQRCode生成本站带随机令牌的URL，handler校验订单后302到官方WAP地址。因此先前真机“一扫直达”的预期在该设备未满足，不能再称零中间页。是否点击后能进入收银台、触发提示的具体支付宝规则／域名状态未知，已询问用户；未访问截图中的令牌链接或发起交易。订阅截图的到期时间已确定；createSubscription开通即设StartsAt/ExpiresAt，首次使用只激活周／月额度窗口（7天／30天），由后端CheckAndActivateWindow维护，v0.2.13上游同逻辑；前端window_start为空显示“等待首次使用”，不是延迟订阅有效期。文案可能混淆两种时间，计费规则未擅改。

- **2026-10-07 本轮正式发布完成（最新终态，覆盖下方未发布记录）**：用户授权提交推送以便自行更新。最终功能4bc5abbb53544fe89f9fc3bc001f5a69b7ca07e5已推origin/main；同SHA的[CI37492464883](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37492464883)、[Security37492464870](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37492464870)、[GHCR37492465165](https://github.com/EngineeMoMo/sub2apicust/actions/runs/37492465165)均completed/success。CI含完整unit／integration、专用PG包号测试、Go lint、前端类型与关键回归、macOS脚本与发布辅助检查。首次发布发现BSD sed和Bash 3中文变量边界问题已修；Vue升级3.5.43及source-map-js补丁消除新增两项high，govulncheck无漏洞，前端Audit exceptions validated（原SheetJS例外保留，未扩例外）；修复后本地全量2921项与类型／构建、Linux更新脚本22场景通过。实际固定镜像ghcr.io/engineemomo/sub2apicust:sha-4bc5abb（linux/amd64），摘要sha256:df08bf887996ab87efe2a5f8a4c9958a925723ff9b0fd4764b30e4512dadde52，由成功构建日志的manifest推送和containerimage.digest确认。证据output/upstream-merge-20261006/release-runs.json、release-jobs-*及job-*。用户先备份再在部署目录执行./update.sh sha-4bc5abb；更新后按支付宝／包号／订阅验收说明核实实际业务。未代用户部署，本机业务容器未动，真实付款／模型请求仍待用户验收。

- **2026-10-06 发布门禁补修**：2c282e409已推送，CI的macOS更新脚本因BSD sed参数不兼容失败，安全扫描发现source-map-js及Vue SSR新高危。已改跨平台sed写回并保留备份／权限，Vue锁定到3.5.43、source-map-js升级补丁；未扩安全例外。Linux更新脚本22场景与升级依赖后的前端365文件／2921项全量通过，类型／构建及新SHA远程发布门禁继续核验。2c282e409镜像不得当作本轮推荐更新版本。后续a8baeb515安全扫描已通过，macOS暴露Bash 3的UTF-8相邻变量边界问题，已以显式花括号补修；新发布版本结果待核验。

- **2026-10-06 用户授权发布准备**：已将上游合并8d47c4253及完整integration交接01abba70d推送origin/main。历史上游提交含skip ci指令，本轮另提交发布记录触发新SHA的CI／安全检查，并手动触发该SHA的GHCR构建；成功结果及固定镜像待核实，不使用旧镜像替代。用户自行更新，智能体未操作生产。

- **2026-10-06 完整 integration 补测完成**：用户明确授权临时测试容器挂载 Docker socket，覆盖此前审批阻塞。测试代码为5d155f9e0（业务合并8d47c4253）；Linux执行 `CI=true go test -tags=integration -count=1 -json ./...` 退出0，52包、6,997顶层＋6,586子项通过、无失败。使用测试指定postgres:18.1-alpine3.23／redis:8.4-alpine，仓储及认证限流实际运行。首轮15项跳过中9项安全审计因未配置专用PG／Redis，已创建隔离服务补跑整个securityaudit包：214项通过、0跳过、退出0。仍有6项既有跳过：钉钉占位、并发缓存TODO、TLS capture地址、TypeSafe实测凭据、OpenAI凭据及插件包缺失；未伪装全项执行。证据output/upstream-merge-20261006/integration-summary.json、integration.jsonl、integration-audit.jsonl及container-results。四个本轮临时容器、其测试匿名卷及专用网络已清理，testcontainers自动清理完成，Docker socket挂载已解除；before/after清单确认只剩原4个业务容器，ID与停止状态一致。无业务代码修改，未推送／部署；本地验证不替代新SHA CI／安全／镜像与真实付款／模型请求验收。

- **本轮合并收口**：功能合并提交 `8d47c42535bad245a9714248092493687a68eb1a` 已从隔离分支快进到本地main；双亲为快照e2adf6005与上游稳定v0.2.13=3040209f2，`git merge-base --is-ancestor v0.2.13 HEAD`退出0。合并后工作区干净（本条交接补记除外）；未推送，origin/main仍旧版本。6个本轮临时容器及codex-upstream-20261006网络已删除，证据与备份保留；container-results.txt确认unit／迁移生成构建／shell容器均退出0。完整integration已按下方补测记录通过；新SHA CI／镜像及真实业务验收仍待。

- **2026-10-06 用户授权同步上游v0.2.13（本地完成，未发布）**：原main=25b9071cf；开工48文件逐字节备份及manifest在output/upstream-merge-20261006/before，隔离分支codex/sync-upstream-v0.2.13先提交快照e2adf6005，然后merge稳定tag3040209f2（含164条上游提交），VERSION按上游b8dece900补为0.2.13。唯一文本冲突payment路由保留WAP Guard／公开IP限流、游客套餐目录，并纳入上游verify 20次/分钟限流。已复现并修复语义冲突：已认证包号私有凭证跳过新余额预占，否则余额0及普通请求占款会误拒包号；HTTP／WS共用入口，不取消普通Key预占或包号共享限额。新增回归、支付宝函数签名适配／赠金折扣扫码实付测试，以及两处前端旧断言修正，全部接缝登记CUSTOMIZATIONS；主题、SynaRoute、订阅卡片及更新脚本保持。
- **本轮验证证据**：output/upstream-merge-20261006。Linux全量go test -tags=unit ./...容器退出0，58测试包通过、11,536顶层＋10,510子项、17项环境／外部依赖跳过（skipped-tests.json）；专用PG包号认证、零余额准入、成员隔离、并发账本等实际运行。新库和已有包号241/242/243库升级至新增上游两项241迁移、重复执行均通过（migration-smoke.log）。Ent生成前后SHA256无差异；Linux go build -tags embed成功；golangci-lint 0 issues。前端全量365文件／2921项、vue-tsc、源码ESLint（排除gitignored生成的public/recipes与studio）、Vite构建通过；Windows依赖初次EPERM导致Axios缺文件已备份损坏目录后按锁文件离线恢复，失败日志保留。上游TypeSafe三条旧断言与本仓CSV Billing Type一条旧断言已修正并重跑全量。最后新增WAP赠金／折扣签名金额与充值规则定向通过。Linux假Docker更新脚本22场景及Compose安全检查通过；Windowsbash缺dirname的尝试不计成功。
- **隔离应用冒烟与边界**：新编译二进制连接专用空库及Redis，在127.0.0.1:4189真实登录成功，profile、订阅列表／progress、包号列表、支付config均200，匿名订阅401，前端嵌入资源200。新管理员settings因尚未人工确认合规条款返回预期423，未替用户接受条款。业务数据库／8080容器／生产均未修改；未做实际模型调用或商户交易。完整integration测试要求容器访问宿主Docker socket，自动审批以授予管理全部Docker资源的权限超出隔离验证为由拒绝；没有绕过，完整integration及新SHA CI／镜像仍待，不能称全部发布门禁通过。未推送或部署；本轮临时容器与测试网络在收尾清理，保留日志、备份及二进制。

- **2026-10-06 上游更新核查（只评估，未合并）**：用户询问开源更新是否值得同步。通过现有本地代理成功fetch upstream（直连未完成已中止），本仓HEAD=25b9071cf、上游共同基点a3eb7ef30／v0.2.8；官方最新稳定v0.2.13=3040209f2（10月2日），最新main=b8dece900（仅比tag多VERSION同步）。HEAD..main共165提交、共同基点到main共394文件／+18028/-1283；稳定tag差164提交。官方v0.2.12包含验证码原子尝试计数／重置密码token哈希单次消费、匿名订单查询限流、上游错误脱敏；v0.2.11余额在途预占与Key创建限制；v0.2.13含请求结束前Key被删除时仍结算用量的修复。建议保留现有fork，优先以稳定v0.2.13在隔离分支同步，先保全当前未提交工作。按规则执行git merge-tree --write-tree --name-only main upstream/main：退出1、唯一文本冲突backend/internal/server/routes/payment.go；这是已提交HEAD演算，不含未提交改动，与上游文件重叠的未提交文件共14个（主要支付／设置）。语义重点：上游余额预占与本仓包号免扣／余额0准入、充值赠金与支付宝安全补丁、数据库迁移回归；无文本冲突不等于业务兼容。上游994048e6a倒计时修复只涉及管理员账号UsageProgressBar，不替代本轮用户订阅卡片。仅fetch／演算与文档记录，无实际合并、业务源码变更、提交推送或部署；完整兼容性与合并后测试结果未知。

- **2026-10-06 订阅改动二次审查（仅本线程范围）**：复读custom卡片／时间辅助、SubscriptionsView异步装配、订阅CSS与Makefile订阅测试接线；未发现新的明确可操作缺陷，原辅助请求阻塞列表的P2已由回归确认修复。重跑25项展示＋5项quota，并以临时用例补查时间先返回／列表后返回及卸载后两个请求迟到失败，共32项通过，证据output/subscription-timing-20261006/rereview.log；临时测试已清理，未修改业务实现。未审查并行支付宝／部署脚本，未做本轮真实账号浏览器或生产验收，不表示线上无问题；未提交推送部署。

- **2026-10-06 订阅审查 P2 已修复（未发布）**：用户授权“修复”后，SubscriptionsView 的列表独立结束 loading，progress 通过已处理成功／失败的异步请求补齐；请求版本号防止重试前的迟到响应覆盖新结果，卸载后失效。新增慢响应补齐、延迟失败保留列表、重试后旧请求成功／失败四项回归；25项订阅展示＋5项旧quota共30项通过（output/subscription-timing-20261006/fix-targeted.log），vue-tsc --noEmit 与两个改动文件 ESLint 退出0。未修改后端或样式，未提交推送、重建8080或部署；本次未重新进行浏览器或生产验收，既有其他专项保持。

- **2026-10-06 支付宝严格安全复查（源码完成，未发布）**：复查并修复短入口重复／畸形参数、HTTP 令牌传输、Referer 扩展信任站点、官方通知补商户／金额回退、未验签证书提示触发 SDK 下载、已付查询订单绑定及一分钱金额容差。默认公钥模式强制配置公钥验签；扫码优先信任管理员 HTTPS frontend_url，未配置时仅请求 Host，反代应拒绝未知 Host。Guard 在限流前设置禁止缓存／referrer，302 正文为空。新 service/provider 辅助文件优先新增，上游接缝已登记 CUSTOMIZATIONS；原手机发起和其他通道保持。
- **安全证据与限制**：`output/alipay-security-review-20261006/` 保存修复前失败复现、最终111顶层＋103子项／7包定向unit、真实仓储SQLite的16并发通知＋重放仅入账一次（连续三轮）、Host模糊测试31,077次、embed构建及最终检查清单。详细发现、静态检查结果与发布边界见 [安全复查报告](deploy/ALIPAY_SECURITY_REVIEW.md) 和 verification.json；不是绝对安全承诺，也不是已证明生产盗款。Windows全量两项既有路径／时间精度失败未处理；本轮真实PostgreSQL并发、race、新SHA CI／漏洞扫描／镜像与支付宝App小额实付仍待。未提交／推送／部署，未请求商户接口或写业务库，保留既有更新脚本及并行订阅UI；生产由用户发布。

- **2026-10-06 订阅时间改动专项代码审查（仅审查）**：按用户“只审查你改的”限定本线程新增 SubscriptionCard、timing、SubscriptionsView 接线、theme.css 订阅段及测试／Makefile 接线，不审查并行支付宝或部署脚本。确认 1 项 P2：SubscriptionsView.vue 的 Promise.allSettled 同时等待列表和辅助 progress，后者慢响应时已成功列表仍被 loading 隐藏，可能等到客户端 30 秒超时；建议列表独立结束 loading，时间异步补齐。临时 Vitest 慢接口用例复现：原 21 项通过、新增“列表应可见”断言失败，日志 output/subscription-timing-20261006/review-probe.log；临时测试已删除。首次复现误用 vite.config.ts 缺 DOM 环境，已改用 vitest.config.ts 重跑，结论仅依据正确配置结果。未修复业务实现、未提交／推送／部署；生产影响未知。

- **2026-10-06 我的订阅时间提示与样式优化（源码完成，未发布）**：用户截图要求核查并补齐提示。原页面已有到期／额度重置文本，现提取custom/SubscriptionCard与subscriptions/timing.ts，用已有认证 `/subscriptions/progress` 的真实 `{subscription, progress}`／服务端resets_at取代本地窗口加小时；共享15秒本地显示时钟、可见性更新及卸载清理，不自动请求／清用量。显示具体到期／重置日期＋中文或英文剩余时间；过期／暂停／撤销不显示未来重置，过期保留实际日期；到期早于重置时提示本期额度在到期结束，已到窗口时刻只提示使用后更新。有效与过期续费仍沿原分组purchase路由，暂停／撤销无入口。卡片统一雾钛青深浅色、桌面双栏与手机单栏／长名称、数字对齐及44px续费按钮，样式仅theme.css，平台语义徽标保持；逐处接缝及同步事项已登记CUSTOMIZATIONS／custom/UPGRADE。
- **本轮订阅验证与边界**：21项新回归＋5项旧quota回归通过；27文件399关键项、vue-tsc、相关ESLint、最终Vite构建成功。IAB真实浏览器以实际组件和明确标注的模拟数据夹具验证1440深浅、390手机深浅／英语、900中间宽度及长标题／未启用／暂停／即将到期状态：无横溢出，手机按钮44px，测试页warn/error为空。证据output/subscription-timing-20261006/{frontend-critical.log,frontend-build.log,browser-proof.json}及截图；夹具不等于真实登录、服务器镜像或实际付款。未改后端、依赖、配置、迁移／业务DB，未重建8080、提交／推送／部署。便携pnpm在8.3临时路径权限及构建PATH首次失败后使用原已有Node／依赖入口、工作区TEMP与.bin PATH通过，无安装；设计技能context因引擎缺失／不可写未运行，直接读取项目规范后完成，不称自动检测通过。此前支付宝与更新脚本专项保留，实际生产版本未知。收尾逐文件SHA256核验开工29个其他专项文件：24个不变、5个支付宝后端文件在工作期间出现并行修改，另有安全专项新文件；本轮未编辑这些后端文件，未覆盖／暂存／提交并行工作，详见protected-before／after.json。临时4187预览、测试标签页已清理，夹具保留在output的preview.html供复核。

- **2026-10-06 重复购买订阅时长咨询（源码与定向测试核验）**：用户询问已有订阅再次订阅是否加时长。现有支付入账 `payment_fulfillment.go:584` 调用 `assignOrExtendSubscription`；按同一用户＋同一分组查原订阅（`subscription_service.go:232`），未到期从原到期时间加套餐天数，已到期从当前时间重新起算（同文件315–317），不同分组独立创建订阅。同一订单的完成状态／分配审计与订单备注保护避免重试重复加时长（`payment_fulfillment.go:514`、569–590）。后台“分配订阅”走 `AssignSubscription`，有效订阅语义一致时复用、不同则冲突，不自动累加（`subscription_service.go:506–542`）；延时需走 `ExtendSubscription`。使用现有已核验便携Go1.27.0、现有模块／编译缓存且GOPROXY=off，本轮执行 `go test -tags=unit ./internal/service -run '^Test(AssignOrExtendSubscription(UsesLockedCurrentRow|SerializedRenewalsAccumulateDays|_ExpiredDailyCardStartsNewOneTimeQuota|_ExpiredSubscriptionAppendsMatchingNotes)|AssignSubscriptionReuseWhenSemanticsMatch|ExecuteSubscriptionFulfillmentRecoversCommittedAssignmentWithoutExtendingAgain)$' -count=1 -v -timeout 2m`，6项全部PASS、退出0；另有Go telemetry目录权限提示，不影响测试结果，不标全量回归通过。仅更新本交接记录，未修改业务源码、配置、业务数据库，未提交／推送／部署或实际付款；生产运行版本及真实续购结果未知，既有支付宝与更新脚本专项保持。

- **2026-10-06 电脑端支付宝 WAP 扫码源码完成（未发布）**：用户明确要求实施扫码计划。支付设置新增默认关闭的“电脑端支付宝使用手机网站支付扫码”，存储 `ALIPAY_DESKTOP_WAP_QRCODE`，管理员字段 `payment_alipay_desktop_wap_qrcode`、支付配置字段 `alipay_desktop_wap_qrcode`；遗漏字段保留原值。仅桌面选中的官方支付宝实例调用 WAP，跳过当面付及电脑网站支付；手机两项旧配置和其他通道保持原逻辑。订单 `expires_at` 向下对齐分钟，以东八区传入 WAP `time_expire`，原 WAP 地址、本站二维码及实际 `qrcode` 模式复用现有字段与 snapshot，无迁移。关闭开关只影响新订单。
- **免登录扫码入口与取证**：新增 `GET /api/v1/payment/public/alipay/wap/:id?token=…`；本站 origin 取自原已校验回跳地址，32 字节随机令牌绑定订单，常量时间比较。校验通过直接 302 至保存的 `https://openapi.alipay.com/gateway.do` WAP 地址，不增加本站登录／中间操作页；方法、商户、金额、订单号和期限须与订单一致，错误令牌、非待支付状态、过期及非法目标被拦截。响应 no-store／no-referrer，应用现有请求日志只记 path，新代码不记录完整付款链接。真实 SDK 本地签名＋SQLite 订单／实际路由测试验证余额与订阅两种订单、无 Cookie/JWT 可 302、原个人订单仍需认证；没有向支付宝发网络请求或产生商户交易。启用与验收步骤见 [deploy/ALIPAY_DESKTOP_WAP.md](deploy/ALIPAY_DESKTOP_WAP.md)，逐处接缝见 CUSTOMIZATIONS 最新节。
- **验证与限制**：证据 `output/alipay-desktop-wap-20261006/`。便携 Go1.27.0 来自此前核验过的 output 工具链，宿主仍未安装 Go；支付相关 195 顶层＋87 子测试、管理员设置、公共路由及重复通知幂等回归通过，API 契约 1 顶层＋11 子项通过；前端定向 128 项与 Makefile 26 文件／378 关键项、类型／相关 ESLint／Vite 构建通过，后端 embed 构建及 golangci-lint 0 issues。已执行 Windows 全量 `go test -tags=unit ./...`，不能标全量通过：API 契约新增字段已修正并重跑，PgDumper 三项缺 sh 已补现有便携 sh 后通过；剩余 `TestResolvePageImagePath` 的 Windows 路径及 `TestOllamaProbeCallback_StaleLongDoesNotOverrideNewShort` 的时间精度失败，未改无关源码。Linux 全量、integration 和本次代码提交的 CI／安全／镜像仍待发布时核验，不沿用 4bf1436fb 旧结果。
- **范围与未知**：保留既有更新脚本专项；`.github/workflows/backend-ci.yml`、`deploy/UPDATE_GUIDE.md`、`deploy/update.sh` 与两个假 Docker 测试文件的字节 SHA256 与开工基准一致。未改 theme.css、依赖或迁移，未修改业务数据库，未重建本机 8080、提交／推送或部署生产。实际支付宝 App 扫码是否进入正确金额收银台、手机无本站登录及小额实付后电脑成功／仅入账一次尚未验证，不能把本地签名和 302 测试当实际付款成功。用户部署生产，真实小额支付由用户确认后验收。

- **2026-10-02 Google AI Pro 视频配置咨询（仅核验）**：用户确认会员是 Google AI Pro，询问能否配置 key 直接生成视频。只读官方 [Pro 权益](https://support.google.com/googleone/answer/14534406?hl=en)、[开发者权益 FAQ](https://developers.google.com/profile/help/benefits)、[Gemini API 计费](https://ai.google.dev/gemini-api/docs/billing)、[视频 API](https://ai.google.dev/gemini-api/docs/video)及[价格](https://ai.google.dev/gemini-api/docs/pricing)：Pro 包含 Gemini／Flow 视频相关权益及每月 10 美元 Google Cloud credits，需关联开发者资料并将抵扣应用到账单账号；API 仍走项目计费，Veo 免费 API 档不可用，不能把会员网页额度当作不限量 API。预付项目使用符合条件的 Cloud credits 前仍需有有效预付余额，具体账户领取／余额／地区／模型权限未取证，不保证其 key 已可生成。源码 rg 核实 magic-studio/core.mjs:2 的 VIDEO_CONTENT_ENABLED=false；magic-recipes/site/model-ui.js:262–272 隐藏视频配置并禁用运行，model.cjs:30 只接受 text／image，当前子产品不能仅填 Google key 就生成视频。咨询不视为开发或收费调用授权；未改运行时代码／配置、未启用视频、未提交推送或部署，既有更新脚本专项保持。

- **2026-10-02 配方“打开图片”修复正式发布（最新终态）**：功能4bf1436fb8ecae0c0780c3b741e26cefba9dd170已推origin/main，11个限定文件，push与ls-remote核实；同SHA的[CI36887283643](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36887283643)、[Security36887283348](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36887283348)、[GHCR36887283175](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36887283175)均completed/success。Go unit／integration、lint0 issues、前端类型／lint及25文件372项通过；镜像embed编译成功，govulncheck未发现漏洞，前端审计既有例外通过。旧data新窗口href改原生dialog页内原图，Base64 PNG／JPEG／WebP提供下载入口；远程图原地址预览、保存提示和错误提示。关闭／Esc回焦点、Tab循环，材料／配方／连接变化或退出清预览，无CSP放宽、依赖或收费请求。
- **本轮验证与范围**：61项配方及3项打包测试、类型／Vite生产构建成功，新回归对旧构建确实失败；本机真实浏览器独立与生产外置资源同源iframe均显示已有Logo512×512夹具，Esc／关闭／390深色无横溢出。独立控制台warn/error空；内嵌记录一条MutationObserver.observe参数错误且无来源栈，配方／共享运行时／宿主夹具无该调用，来源未知，不把控制台标全空。下载完成事件超时，实际文件落地未验证；未取得用户现场href／控制台，不知道截图的线上根因，不把本机模型mock／宿主夹具当真实账号或线上恢复。证据output/recipes-image-open-20261001/{ci-final,ghcr-image-proof,browser-standalone,browser-embedded,guide-preservation}.json、完整job日志和本线程recipes-image-open截图。沿此前“提交并推送，我发布更新”授权限定发布；更新脚本4文件字节SHA256等于基准，UPDATE_GUIDE去掉新增发布段后的SHA256也等于基准，原旧镜像清理节仍仅工作区。临时4178／4180服务与测试标签页已清理，8080和生产未改。
- **本次实际固定镜像**：ghcr.io/engineemomo/sub2apicust:sha-4bf1436，linux/amd64，摘要sha256:038e95a020dc71a5615c02cec490a9a92cf8172242686977051b9686cdeae87e，版本0.2.8-custom.4bf1436fb；成功job的manifest推送、containerimage.digest及revision逐项核实。包含上一版包号免扣与迁移243，acd73d7不含本次配方修复。用户备份Postgres、部署配置与app_data（含studio-submissions）后执行./update.sh sha-4bf1436；刷新配方检查真实已返回图片的打开／关闭／手机／原图保存，无需为验证打开功能额外发模型请求。后续仅交接文档用skip-ci，不能拿文档SHA猜新镜像。

- **2026-10-01 包号正式发布门禁全绿（本轮终态）**：功能acd73d716daa6a8e728a36fca5dce5fd523ae3f6已推origin/main，65个包号及依赖文件；直接读取同SHA的[CI36873854705](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36873854705)、[Security36873854536](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36873854536)、[GHCR36873854627](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36873854627)，均completed/success。Go unit／integration全量、独立PG准入／计量／认证、Go lint0 issues、前端类型与372项关键回归均成功；镜像日志确认embed构建成功，govulncheck No vulnerabilities found，前端Audit exceptions validated，既有两项SheetJS high例外未扩。
- **本次实际固定镜像**：ghcr.io/engineemomo/sub2apicust:sha-acd73d7，linux/amd64，摘要sha256:b41acdce7446e1b7bca618e16e6f6dbdab5b13fd30732760247063deb576bfa7；成功任务的manifest推送、containerimage.digest和revision=acd73d716逐项核实，版本0.2.8-custom.acd73d716。用户备份Postgres、部署配置、app_data（含studio-submissions）后，在/sub2api-deploy执行./update.sh sha-acd73d7，迁移243随启动执行；更新后验0余额专属组实扣0、公共组原计费、共享策略、到期撤销及真实客户端。默认日上限0不限、生图关闭、WS仅生成与取消；严格Token／美元预算、IP异常和归档仍未实现。
- **发布范围与留存**：既有包号多人多组、移除成员、Key恢复、旧表单与400诊断及三条关联主题样式随免扣同版；更新脚本／假Docker测试、其shell CI三行及UPDATE_GUIDE旧镜像清理段继续留工作区未提交，三个脚本字节SHA256仍等于开工基准。证据output/dedicated-billing-20261001/{release-initial,ci-final,ghcr-image-proof}.json及完整job日志。只补交接的后续文档提交用skip-ci，不生成替代镜像；未操作8080或生产、未改业务DB、未发收费请求，不宣称现场验收或历史扣款原因已确认。以下“未发布／门禁运行中”为此前进度。

- **2026-10-01 包号功能已推送（门禁运行中）**：功能提交acd73d716daa6a8e728a36fca5dce5fd523ae3f6已推origin/main，git push与ls-remote输出核实；65个包号及依赖文件，发布树6fc7cad1adec562ab069b04e5c06e653835badb9。该SHA的[CI36873854705](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36873854705)、[Security36873854536](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36873854536)、[GHCR36873854627](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36873854627)已启动，当前未完成，不能标全绿或提供镜像成功结论。更新脚本专项及其shell接线仍未提交，服务器未操作；结果以本轮后续终态为准，证据output/dedicated-billing-20261001/release-initial.json。

- **2026-10-01 包号发布授权已收到（提交准备）**：用户明确“提交并推送代码，我发布更新”。本次发布包含免扣／共享限制以及其依赖的既有同批成员多组、移除成员、Key恢复复查、旧表单409／具体400诊断和相关三条主题样式，形成同版包号功能；更新脚本、假Docker回归及其shell CI接线继续留在工作区，本次不混入。远端main经git ls-remote核实仍f846b970c4，无远端新提交；新SHA的CI／安全／镜像结果推送后逐项取证，不能拿293c091结果替代。生产由用户备份后更新，代理不操作服务器。

- **2026-10-01 包号免扣余额与防滥用开发完成（未发布）**：用户已明确“开始处理开发”，源码在原未提交包号专项基础上实现独立BillingType=2：服务端正向资格与请求私有准入凭证，余额0可用、客户实扣0、公共组原计费；两平台／WS保留Token和参考价，Key额度／窗口按参考量累加，余额DB／缓存／通知不扣不误报。迁移243新增策略、共享计数、可恢复租约和独立账本，参考账本与原去重／Key增量共事务；每次服务端执行独立计量，同次回调幂等，已准入请求跨期撤销不追扣，不把失败转成余额收费。后台“使用限制”默认共享并发2、RPM30、日0、2MiB、生图关，所有成员／关联专属组／Key共用；日0不设限，不是Token／美元预算，不自动IP封禁或退款。WS每轮重查权益／策略，控制帧仅生成与取消，session.update／重复type键不能旁路工具或实时音频；未适配异步／批量媒体付费入口拒绝。用户／管理员表格及导出区分包号实扣0和参考金额，详细边界见[计费实现](deploy/DEDICATED_BILLING_PLAN.md)第六节、接缝见CUSTOMIZATIONS最新节。
- **本轮验证证据**：output/dedicated-billing-20261001。完整便携官方Go1.27.0归档SHA256与前轮官方元数据相同，宿主未安装Go；模块／编译／lint缓存均在output。真实独立Postgres随机schema测试、包号／WS定向81顶层＋82子场景成功；Linux全量go test -tags=unit ./...通过，最后原生控制帧／handler校验补丁再跑定向、golangci-lint v2.13 0 issues和embed成功；前端65定向及Makefile25文件372关键项、类型／相关lint／Vite构建通过。Windows全量首次路径／缺sh及Ollama时间精度失败，Linux初试只读ent和SDK代理环境错误修正后全量成功，没有修改无关业务。临时PG及Linux测试容器已清理，未替换8080或生产服务、未写业务DB、未调用收费模型。
- **保留与未知**：开工30个并行文件中22个SHA256完全不变（含theme.css、包号service专项和更新脚本），另外8个仅增本轮接线／测试／说明，原功能保留。当前HEAD仍f846b970c4，本轮未提交／推送／部署；integration／新SHA CI／安全／GHCR、真实浏览器和业务客户端待验收，不能沿用293c091。没有生产请求／余额流水，不知道某笔线上扣款原因或是否已有滥用；历史补偿需另行逐笔核对。

- **2026-10-01 修复版正式发布门禁全部成功（最新终态）**：功能提交`293c091b3ab0c8f187c031bed1955bb843939d40`已推origin/main，10个限定文件。对该SHA直接读取GitHub任务及job日志：[CI36842744777](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36842744777)、[Security36842744680](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36842744680)、[GHCR36842744656](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36842744656)均completed/success。CI Go lint为0 issues；Go unit／integration全部成功（含投稿handler解码与清理回归），前端类型／lint及310项核心测试成功；镜像日志确认`go build -tags embed`成功。govulncheck报告No vulnerabilities found，pnpm审计既有例外检查成功，仍有两项既有SheetJS high例外，未新增例外或关闭扫描。
- **固定镜像已发布**：`ghcr.io/engineemomo/sub2apicust:sha-293c091`（linux/amd64），摘要`sha256:c517187b769f8b49f1fb29b1667277b745bb42deb5be05eabbf3f39e54274a06`；Build and push日志中manifest推送完成及containerimage.digest逐项确认。包含d84d60e截图修正／47图库及本轮Axios1.20.0、x/image0.45.0和清理检查，不用旧765e5a9／d84d60e失败门禁或更旧成功记录替代。备份Postgres、部署配置和app_data（含studio-submissions）后，用户在`/sub2api-deploy`执行`./update.sh sha-293c091`。本轮未代操作生产、8080、收费模型或真实账号配置；线上生图空框根因仍未取得真实Key／分组／models响应，源码诊断与模拟选择通过不等于现场恢复。
- **发布证据与隔离**：`output/family-security-fix-20261001/ci-final.json`、各job完整日志、ghcr-image-proof.json及本地实际提交快照记录；本地Node106＋Vue538共644项、类型／源码lint／Vite构建通过。本机仅经官方归档SHA256核验的便携gofmt，Go编译与测试由上述新SHA CI验证。29个并行文件字节SHA256与开工快照相同、共享theme本轮未改；UPDATE_GUIDE只补本发布说明，原更新脚本章节保留并单独部分暂存。包号／更新脚本及CI接线未混入应用提交；收尾文档用skip-ci，不产生替代应用镜像。

- **2026-10-01 用户已明确授权“修复好再发布”（覆盖等待答复）**：Axios已升1.20.0并更新pnpm锁；x/image升0.45.0、go.sum采用Go官方sumdb记录，其要求x/sys0.47.0／x/text0.41.0与本仓已有版本一致。投稿handler八处清理错误改为统一检查／日志，保留原业务错误或已发送响应，不记录投稿正文；新增有效／截断WebP、PNG最小／超限及保存失败清理测试，正常WebP使用本仓新图缩略副本并标来源。pnpm生产审计已不含Axios，现有例外检查通过；两项SheetJS high仍按既有未过期例外，未扩例外或关闭扫描。限定暂存树c3423c3导出源码，Node106＋Vue538共644项、类型／源码lint／Vite生产构建通过，包含请求与token刷新；lint仅排除gitignored产品生成副本，仍有既有测试Vue警告与构建500kB提示。29个并行文件SHA256未变，共享theme本轮未改。后端本机仍无Go编译器，Docker引擎未启动；从官方Go1.27.0归档核验SHA256后仅提取便携gofmt格式化本轮两文件，未把格式化当编译通过。编译、单测／集成、govulncheck与镜像等待新SHA的CI，提交／发布终态待下续；不代启动服务或部署生产。证据output/family-security-fix-20261001，未成功的初次下载是未跟随302的74字节HTML，校验阻止执行后按官方重定向下载并核验成功。

- **本轮安全扫描终态（覆盖下条刚启动状态）**：`d84d60e`的Security36836909442已completed/failure，直接读取frontend-security110286404922及backend-security110286405220日志确认：Axios七项high仍在，x/image0.41.0仍有GO-2026-6222／5061／4961可达漏洞，最高修复版0.45.0。本轮不是安全发布成功，镜像即使成功也先不更新公开站点；CI／GHCR取证时仍in_progress。代码已按用户授权推送，安全修复因自动审批拒绝正在等待用户明确答复。证据output/studio-refinement-release-20261001/security-confirmed.json与ci-status.json，后续文档skip-ci不生成替代应用镜像。

- **2026-10-01 截图修正版已推送（最新发布状态，仍待门禁）**：功能提交`d84d60e07d9abb776f8d2dd8b0e8bdef60dc9f7e`已推origin/main，git push日志与ls-remote确认一致；65文件，47图及本轮滚动／SynaRoute／下拉诊断修正。实际发布源码Node106＋Vue538共644项、类型／源码lint／Vite构建通过；源码lint排除仅由builder生成且gitignored的public/recipes、public/studio副本，保留首轮格式失败日志。29个并行文件SHA256全部不变；共享theme撤去本轮三条新规则、恢复hint后的字节SHA256仍等于开工快照，证明原并行三条规则保留。
- **本SHA的CI已启动，不能沿用旧SHA或宣称可更新**：[CI36836909715](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36836909715)、[Security36836909442](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36836909442)、[GHCR36836909685](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36836909685)当前均in_progress。`sha-d84d60e`只是预期固定标签，未确认构建发布成功或摘要，不给update.sh指令。已知依赖／Go lint阻塞仍未处理：安全依赖安装被自动审批拒绝，等待用户明确授权；收到答复后再处理并发新功能SHA，核对其CI／安全／镜像，不能直接更新当前公开站点。证据output/studio-refinement-release-20261001，未更新8080、调用收费模型或部署生产。

- **2026-10-01 用户授权本轮提交推送（发布准备）**：用户要求修好后推送以便手动更新。已限定暂存65个界面／素材文件，导出实际index源码快照；Node106＋Vue538共644项、类型／Vite构建通过。完整lint首次扫到构建后的gitignored运行副本7项旧静态格式错误，随后按干净CI源码范围排除public/recipes、public/studio重查，结果以`output/studio-refinement-release-20261001/lint-source.log`终态为准。包号与更新脚本继续排除，共享theme只暂存本轮产品段落。旧765e5a9的单测／集成现已由GitHub job API核实success，Go lint及安全扫描failure仍有效。
- **发布仍有明确阻塞**：官方公告／npm与Go模块目录确认Axios修复版1.20.0、x/image0.45.0；投稿handler八处未检查清理错误也已由失败日志定位。尝试安装Axios修复版时自动审批拒绝，理由是安全依赖升级超出当前授权；安装未执行、依赖及锁文件未改。已异步请求明确授权这两项依赖升级及清理检查修复，答复前不处理依赖，不把界面成功与可构建镜像当安全发布通过。取证在`output/studio-refinement-release-20261001`及旧SHA真实任务；本轮功能提交／推送及新CI状态待下续记录，不代部署生产。

- **2026-10-01 用户截图修正与图库收口（最新源码，未发布）**：修复配方／工坊内嵌吸顶栏透明背景及层级造成的滚动正文叠字；配方栏全宽，目录避开顶栏，模型设置锚点保留88px空间。SynaRoute入口使用`/keys?product=synaroute`，该上下文选中紫色SynaRoute，普通`/keys`仍为API；不改桌面客户端或导入功能。首页移除“悬停或点按，置顶欣赏”，三图交互保留。
- **本轮图库最新数量覆盖下方历史38图**：按用户要求删除“夜幕之后，一点绯红”，时尚肖像只保留“海风里的丝绸与牛仔”；删除“跑进珊瑚色的夏天”，两项全尺寸／缩略WebP从公开打包撤下，旧PNG及生成记录仅作历史。Cosplay／动物自然／海报社媒／动漫二次元／空间设计各新增2张真实独立生成图，另补首页窗边人像的同一素材，共47图＋4工具，10套分镜继续封存。39图保留实际prompt，8旧图缺原始记录。十张本轮新图未输入参考图；首页人像复用既有参考图生成结果，其参考权属未核实，不能标全新无参考原创。精确prompt／原PNG／WebP／缩略图尺寸与SHA256见`product-samples/magic-studio/assets/PROVENANCE-CURATED-20261001.json`。后续素材由用户自行补充与人工审核，本轮不再自动扩充。
- **线上模型空框的事实边界**：用户确认现场是`https://ai.mofamilys.com`，本轮浏览器只取得该站登录页，未取得真实账号密钥列表／分组／`/v1/models`响应，**不知道线上根因**。已在源码补充空列表／接口失败／有效密钥均无生图权限的具体提示并禁用无选项控件；切换用途或刷新先清旧密钥／目录，避免沿用文字配置。明确标记的宿主夹具验证“无生图权限”提示、“有权限”可选模拟密钥与模型／应用按钮启用；这不是生产认证或真实生图成功，未自动开权限、创建密钥或调用收费模型。
- **本轮取证与限制**：Node106＋Vue80共186项相关测试、类型／限定lint／Vite构建通过。真实浏览器检查内嵌滚动栏不透字、配方设置标题可见、SynaRoute路由与紫色当前态、五个分类新增标题、时尚仅保留一图；1280桌面与390窄屏深浅代表视图、完整竖图contain及无横溢出。模拟宿主／密钥／模型范围明确写入`output/studio-refinement-20261001/browser-evidence.json`与截图，夹具源仅留output，不进入产品源码。十图接触表完成视觉检查。设计技能扫描引擎不可用，没有自动设计扫描通过结论；公开源码预览投稿目录不可用，与旧后端一致。包号／更新脚本并行文件SHA256比对结果见该output目录，主题仅增加本轮产品规则并保留原并行规则。没有本轮提交推送、8080重建、Go／Docker验证或生产部署，下方765e5a9发布门禁仍未修复。

- **2026-10-01 家族版本已推送但发布门禁未全通过（最新终态）**：功能提交765e5a9516eaecf23d94202b4afc48f9bf723339已推送origin/main，165文件；actual暂存快照333项、类型／定向前端lint／Vite构建通过，包号24文件及更新脚本并行实现保留未提交。新SHA的GHCR任务36825037358已成功，真实日志确认ghcr.io/engineemomo/sub2apicust:sha-765e5a9与latest，摘要sha256:341886529ae61b817a24da3a2a40dbb7062aabbc2148ecc9b7956153cb422e9d；这是镜像构建成功，不代表安全／全部CI通过，**暂不建议更新公开服务器**。
- **该SHA实际失败／待完成项（不得沿用旧全绿记录）**：Security Scan36825037421失败。govulncheck发现golang.org/x/image v0.41.0的GO-2026-6222／5061／4961，修复版本至少v0.45.0；本轮StudioHandler.Submit的image.DecodeConfig→webp／vp8l为可达路径，用户头像旧路径也命中。前端audit发现Axios七个high缺少例外，不能自动补例外掩盖。CI36825037384的Go lint因custom_studio_handler.go八处Close／Remove／RemoveAll错误未检查失败；原站前端／shell／release-helpers成功，Go Unit完成、Integration仍运行时取证，不标全部成功。已异步询问是否继续修复依赖再发布，当前尚无本轮后续修复授权。证据output/family-release-20261001/ci-status.json、GitHub本SHA任务与日志；8080及生产未更新，收费调用未执行。

- **2026-10-01 用户授权家族产品提交推送（发布验证）**：本次限定配方／工坊同镜像、控制台同页与主题跟随、配方同页配置、原创素材／完整prompt、投稿审核和视频／分镜待开放；不提交包号专项、更新脚本及其CI接线，部分共享theme／UPDATE_GUIDE只暂存本产品段落。已从实际暂存区导出独立源码快照（不是整个混合工作树），Node104与前端229项共333项通过，类型／定向lint／Vite构建通过；首次Node因漏构建gitignored配方HTML而失败，先执行权威builder后全部通过，不以旧预览产物补进提交。两忽略文件混合Windows换行曾造成全文件diff，已仅在暂存区归一化，限定diff检查通过。证据output/family-release-20261001。推送和该SHA的新CI／GHCR待本轮后续取证，当前不使用旧25071d4结果；未更新8080、运行收费模型或部署生产。

- **2026-10-01 用户要求暂停视频与分镜（最新有效状态）**：工坊前台现仅展示38张原创图与4份工具，视频筛选显示“视频与分镜 待开放”并真实disabled；列表友好提示“你可以先探索图片，或使用 Skills 和工作流”。原10套分镜及提示词／收藏ID留档不删除，发现／搜索／收藏／历史、玩法分镜步骤、旧链接及审核后的公开视频均不能绕过前台关闭状态；旧链接提示并保留当前图片材料。图片完整预览／对应原始提示词、复用与原画幅瀑布流保持。
- **投稿与验证边界**：视频新投稿选项待开放，提交与旧视频重新投稿路径前端阻断；图片可投、旧视频私有查看／撤回及管理员管理保留，服务端MP4协议未停用。Node104（工坊44＋配方56＋宿主1＋打包3）与Vue11（投稿9＋API2）通过，类型／限定lint／Vite构建通过，已有500kB分块及Browserslist旧数据警告未处理。4185新建tab25验证38图、4工具、4个分镜步骤disabled、搜索及旧链接、完整图与原始prompt、1280与390深浅横溢出0、warn/error空；已恢复浅色与默认视口，未刷新用户材料页。证据output/studio-video-paused-20261001与studio-video-paused-*截图；24个包号并行文件SHA256保持。仅源码／4185预览，未更新8080、Docker／Go构建、提交推送、CI或生产部署；真实登录投稿／视频任务／收费调用未验。

- **更新脚本服务器安装命令交付（2026-10-01）**：用户要求完整操作命令，已生成gitignored的output/install-update-cleanup-20261001.sh并通过bash -n。流程读取服务器当前唯一sub2api容器的固定sha引用，备份旧update.sh，写临时文件／语法检查／chmod后替换，使用同一个当前sha运行更新与镜像清理，最后查Compose状态／镜像列表／df；不是猜latest或切到本机未发布版本。最终提供Windows PowerShell输入SSH地址／端口的scp上传命令和服务器bash执行命令，不依赖尚未发布的私有仓下载链接。尚未执行服务器上传、安装或清理、尚未提交推送；结果待用户输出核验。

- **本轮镜像清理最终验证补充**：增加无sha别名的latest升级及重复运行、更新中切换到其他仓库场景后，假Docker回归为22场景全部通过；三文件bash -n及限定git diff --check通过。CI接线已登记，远端CI尚未运行，服务器脚本尚未替换。下条20场景为中途记录，以本条22为准。

- **2026-10-01 用户授权新增镜像清理（本轮已实施，未发布）**：deploy/update.sh改为健康检查成功后仅本仓sha-*／latest和本仓来源标签悬空镜像清理，保留当前／实际更新前镜像的完整ID和其他运行或停止容器引用；latest通过本地update-before／rollback-previous标签保留旧版，同版更新保留已有回退版，首次同版无记录时选创建时间更早的最近sha候选。健康超时退出非零，拉取／重建失败不清理，检查失败保留镜像并显示原因；同部署目录锁防并发。不自动清理数据库、备份、日志或卷。新增20场景假Docker回归及CI shell接线；Git Bash三文件语法检查与20场景全部通过。服务器需单独替换宿主update.sh；未运行真实Docker删除、更新8080、提交推送或生产部署。并行工坊／包号工作全部保留。

- **2026-10-01 工坊空白修复与视频接口查证（最新布局）**：按impeccable布局与craft-floor保留既有品牌／功能，原齐行Grid造成1280同列最大空白165.69px，改为1px Grid最短列瀑布流；完整原画幅与contain、不改DOM精选顺序／键盘阅读顺序。38图源PNG宽高预留值逐一核对，媒体加载／错误／元数据、ResizeObserver与resize重排；筛选断开旧观察、玩法隐藏及pagehide清理、pageshow恢复。首轮4列缩到2列出现隐式空列已修为CSS列数变量、降列先归位，并加入回归；没有新增依赖、后端／认证／宿主修改。
- **本轮证据／发布边界**：Node101（工坊41＋配方56＋宿主1＋打包3）、语法／限定diff检查、Vite生产构建退出0，原500kB分块警告保留。4185新建tab24，320／390／768／1280／2200真实布局纵向间距手机12–13px、桌面18–19px，顶部DOM顺序不递减、外层横溢出0；完整图与原始提示词详情、关闭焦点、分镜三段／无伪视频、空态恢复实测，warn/error空。截图studio-masonry-*在本线程visualizations，数据output/studio-masonry-20261001/browser-geometry.json；视口已恢复、浅色、新预览保留，没刷新用户草稿页。24个包号并行文件SHA256与原基准一致。技能扫描引擎exit127不可用，不声称设计扫描／独立审查通过；未更新8080、Go／Docker构建、提交推送、CI或生产部署。
- **OpenAI视频事实（2026-10-01官方只读查证）**：当前 https://developers.openai.com/api/reference/resources/videos 顶部与 https://developers.openai.com/api/docs/deprecations 的Sora段明确：Videos API与Sora 2系列2026-09-24关闭，暂无一对一替代API；保留接口页是历史参考。不能说OpenAI从未有视频API，也不能把历史接口当现可用服务；本会话无视频工具是另一事实。本站是否另有第三方视频供应商未知，未检查真实账号权限或收费调用；当前10套分镜仍0段对应成片，投稿人工审核链与上一轮未部署边界保持。

- **2026-10-01 生产磁盘截图取证（覆盖下方待取证状态）**：用户四张服务器终端截图确认/dev/vda1 20G、已用5.5G、可用14G、33%；当前sub2apicust sha-25071d4唯一被容器引用，另六个sha版本均CONTAINERS=0，独占大小合计约1012.8MB。上一镜像sha-321fe99可作为保留回退候选；拟仅删除更旧sha-43b715a、sha-fa45b20、sha-a1b9137、sha-d592f94、sha-58b154f，不用force，五版独占空间合计约843.7MB，实际回收需df复核。部署目录278M，数据库备份backup-0.2.7-2026-09-26.sql.gz为5.0M、应用日志48M、Postgres数据168M、Redis数据58M；无构建缓存，不能把这些目录视为可删备份。旧镜像积累已证实，但无更新前后对比，不断言全部磁盘增长都由镜像造成。仅给用户定向清理命令，尚未执行服务器清理；原上游镜像与chatgpt-next-web虽0容器，用途未确认不列入删除。

- **2026-10-01 生产升级后磁盘减少（取证待办）**：已读deploy/update.sh，指定标签时创建docker-compose.override.yml.bak.<时间戳>但无轮转；不自动执行pg_dump，手动数据库备份不会由该脚本删除。末尾仅docker image prune -f且忽略失败，按Docker文档只清悬空镜像，带sha标签的旧镜像仍可能积累。仓内Compose未声明日志轮转，服务器Docker日志默认值／实际配置未知。尚无生产df／docker system df／镜像清单／部署目录大小证据，不能断言现场主因是备份或旧镜像。提供只读检查，不修改更新脚本、不清理服务器或本机Docker；并行未提交工作保留。

- **2026-10-01 工坊风格／投稿审核版（最新，覆盖下方旧状态）**：已明确本会话只有生图工具，没有可调用的视频生成工具；用户未提供视频接口／模型／预算，不调用收费服务或把分镜冒充成片。新增六张真原创素材：像素港口、水墨巨鲤、美漫信使、水彩海湾、剪纸小虎、复古未来火星站，不输入用户截图、不复制人物。当前38图＋10套待样片分镜＋4工具，共52条、发现48条；30图有准确prompt，8张旧图仍缺失。PROVENANCE-STYLES-20261001.json记录工具、完整prompt、无参考图、三种资产尺寸／字节数／SHA256；WebP等比编码无裁切。新增题材／视觉风格两套独立筛选及可辨识的投稿CTA，不改母品牌首屏或SynaRoute。
- **投稿是真后端，不只是占位按钮**：新增handler/custom_studio_handler.go及custom/studio/api.ts、StudioSubmissionsView.vue。工坊进入同源 `/tools/studio/submit`，复用API登录／正文与主题；管理员 `/admin/studio/submissions`。PNG／JPEG／WebP最大12MiB、32px以上／2400万像素以内，MP4最大30MiB；签名／体积检查不等于完整解码、病毒或自动内容审核。上传必填模型／工具、完整prompt、题材／风格及两项授权，标明actual或reference。初始pending仅作者／管理员可读；明确人工确认才能published进入匿名公开目录／Range媒体接口；支持带原因退回／下架、作者撤回及旧状态409。修改重新投稿，不能直接恢复发布；图片和成片可点开并查看对应prompt，反推模板不冒充原始记录。审核用现有认证、限流、后台模式与合规保护，不给静态子页传JWT；上游router新增[CUSTOM]接线，audit_log精确省略投稿／审核请求正文但保留元审计，工坊CSP只放同源目录读取，未放开外部模型连接。
- **存储／隐私边界**：`pricing.data_dir/studio-submissions`，默认Docker `/app/data/studio-submissions`，不在静态目录，JSON＋media私有持久化；每人5待审核／20记录，全站1000记录／512MiB媒体，撤回等仍占额度、无永久清理UI。只支持单应用实例，未来多副本需DB元数据／对象存储。备份要含整个目录，Postgres备份不含文件；反代体积上限需容纳30MiB加表单。当前页替换材料仍仅内存，用户点击投稿后文件与prompt会上传本站并持久保存，不能继续笼统说所有材料不上传。原始来源、许可、人工审核与模型验证分开；未接自动NSFW／版权服务，不保证完全防滥用。细节product-samples/magic-studio/COMMUNITY.md及deploy/FAMILY_INTEGRATED.md。
- **验证与发布边界**：Node96项（工坊36＋配方56＋宿主1＋打包3）与Vue64项定向回归通过，类型／限定lint／Vite生产构建退出0；日志output/studio-community-20261001。Docker代理先验200，在隔离镜像 `sub2apicust:studio-check-20261001` 中5个投稿／审计顶层测试、相关静态产品回归与Go embed编译通过；MP4测试仅签名／Range夹具，不是播放／生成样片。首次编译响应函数名错误已修复；尝试web全量时两个既有favicon.png用例返回HTML失败，未修无关测试，不标全量Go通过。4185另开tab22，六图真实加载、12种风格及组合筛选、完整原prompt、水墨桌面详情、390深色剪纸详情均验；无外层或详情横溢出、warn/error空。匿名投稿入口真实回跳 `/login?redirect=/tools/studio/submit`，未借模拟账号宣称真实投稿／审核闭环。截图studio-style-*在本线程visualizations，恢复浅色／正常视口，没刷新用户草稿页。24个包号并行文件SHA256保持，preserved-check.json；未替换8080、提交推送或部署生产，8080原服务及DB／Redis仍healthy。旧后端的4185会显示“投稿目录暂不可用”，这是未发布状态；不是已上线可用。真实账号、可播放视频、管理员视觉及后续CI／正式镜像更新仍待。

- **2026-10-01 用户再增十图（最新目录与证据）**：用户确认询问视频是否仅提示词，答复明确当前只有10套原创分镜、0段新成片。另按要求实际使用内置image_gen生成10张独立原创图：街角成年女性、唱片店成年男性、夜色时尚、赤红信使Cosplay、莲池剑修、海岸救援机甲、海獭、跑鞋、极光湖景、狸猫食堂；没有输入用户参考截图，人物明确25岁以上。当前32张图＋10套分镜＋4份Skills，共46条、发现区42条；最新十图前置，此前22图保留。
- **准确记录／资源**：十张PNG已复制入工坊assets，网页使用等比缩小的完整WebP和缩略图，无裁切或手工像素编辑；assets/PROVENANCE-EXPANDED-20261001.json保存精确实际prompt、工具、无参考图、生成文件名及三种资产尺寸／字节数／SHA256。共24张图有原始实际记录、8张旧图仍明确缺失；复用模板另写未重跑，不将原图生成成功扩大成任意模型模板实测。new-gallery.mjs共用记录映射、catalog前置新图，index数量和README／PRODUCT／DESIGN同步；没有新运行module、后端／安全白名单／宿主／认证／主题修改或新增依赖。
- **本轮验证**：工坊30＋配方56＋宿主1＋打包3，共90项通过，Vite生产构建退出0（既有500kB分块警告不在本轮修），日志output/studio-images-more-20261001/{node,build}.log。打包十张full／thumb与源WebP逐字节一致、不公开PNG和PROVENANCE JSON。4185另开tab21验证十张完整图全部真实加载／contain、实际prompt与来源逐字一致、1280桌面浅色和390手机深色无外层横溢出，手机横幅详情无内部横溢出，warn/error为空；证据browser-prompt-check.json、studio-more-browser-20261001.json及studio-more-{gallery-desktop,prompt-desktop,mobile}-20261001.png。截图在本线程visualizations，最终恢复浅色／正常视口并保留新预览，不刷新用户材料页。24个并行包号文件SHA256保持原基准，output/studio-images-more-20261001/preserved-check.json。未调用视频或付费API、编译Go、重建Docker、更新8080、提交推送或生产部署；真实剪贴板／TXT落地／账号模型授权仍未验，视频生成服务／预算或授权素材仍待确认。

- **本轮最终范围核对**：24个包号并行文件SHA256与原基准全部一致，证据output/studio-vibrant-20261001/preserved-check.json；限定工坊／打包测试／交接文档git diff --check退出0，app与new-gallery语法检查退出0。最终浏览器实证studio-vibrant-browser-20261001.json及最新prompt-desktop截图保存在本线程visualizations，视口已恢复正常、主题恢复浅色，另开交付tab20保留，未刷新用户原材料页。视频样片问题尚待回复，十套分镜不是已完成成片。

- **2026-10-01 工坊素材再次否决与作品／提示词改造（最新源码）**：用户要求撤下4段旧视频、鲜活图像与9–12段年轻人偏好的短剧／修仙视频。已物理撤下4段Blender原片及封面，目录／玩法／打包回归同步；新增6张原创图（音乐节成年女性、成年男性滑板、紫色Cosplay、赤金天龙、夏日动漫、软糖广告），未使用参考截图生成。当前22张图＋10套原创视频分镜＋4份Skills／工作流，共36条；十套分镜为美剧感4、韩剧感3、修仙动漫3，每套三段12秒计划及中英文完整模板。**没有这10套分镜的成片，本请求的视频部分未完成，不称10段视频已交付。** 可调用工具只有生图；用户视频服务／预算或授权现成素材待确认。搜索所得社区集合第三方媒体权利未核实，未下载或擅用抖音原片。
- **作品与提示词对应**：左侧查看完整媒体与提示词双栏详情，手机上下；卡片直接复制模板／改成我的，右侧材料、收藏、语言保留。14张图真实原始提示词（原8＋本轮6）与来源JSON逐字一致，此前8张只有摘要的图明确原始记录缺失；原始提示词、可替换模板与生成状态分开。TXT增加实际记录／当前模板标记；弹窗复制失败文本留在弹窗内，重开归零并聚焦关闭，关闭恢复焦点。默认浅色／宿主主题与控制台嵌入架构不改。
- **本轮证据与边界**：工坊28＋配方56＋宿主1＋打包3，共88项Node回归通过，日志output/studio-vibrant-20261001/node.log；Vite生产构建退出0，build.log。首次直接构建因node_modules/.bin未进PATH缺vue-tsc，补PATH后完成，没有安装依赖。4185真实浏览器六新图加载、对应实际prompt／当前材料、复制成功提示、Esc／焦点、390手机深色与1280桌面浅色无外层横溢出，warn/error日志为空。截图studio-vibrant-{gallery-desktop,prompt-desktop,image-mobile,video-script-mobile}-20261001.png在本线程visualizations；不把复制成功提示当系统剪贴板或真实TXT落地验收。未调用视频模型、收费API、Skills、定时任务；未编译Go、重建Docker、提交推送或更新8080／生产。8080仍按上轮镜像记录，不将4185源码预览当新版部署。

- **收工取证**：最终重跑仍为Node83＋前端86共169通过，限定范围`git diff --check`退出0；24个并行文件摘要一致，记录output/family-workspace-20261001/preserved-check.json。新测试镜像144c2190d8c0，运行容器仍c599729e28df且应用／原PostgreSQL／Redis均healthy（只读inspect／ps）。预览服务4185为本轮自建Vite session89719；实际最终工坊1280×720浅色、20份媒体索引、右侧contain与无溢出，工坊／配方warn和error日志均为空，证据studio-polish-browser-20261001.json及desktop-light截图在本线程visualizations。已清理本轮失败类型构建产生的两个多余dev-products.js／d.ts，仅保留源码mjs与声明d.mts；未动用户旧预览或8080。

- **最终测试镜像（未替换8080）**：2026-10-01本轮最终 `sub2apicust:family-workspace-20261001` 构建退出0，image inspect为sha256:144c2190d8c0096d9890d2c2401477770371b690a09eddab23c2e537a4ab1fee，日志output/family-workspace-20261001/docker-build-final.log；代理预检200且只给构建进程设置代理。包含最新工作区／媒体／样式与Go embed编译，TestFamily两函数通过。未tag覆盖local-theme、未up服务、未改DB或生产；用户确认本轮更新后才能按原runbook部署，前次c599729e28df仍是8080实际版本。

- **2026-10-01 子产品工作区与作品扩充（最新源码，覆盖下方旧新窗口／二维运镜／默认深色描述）**：登录后配方／工坊默认进入 `/tools/recipes`、`/tools/studio`，在 API 控制台右侧正文嵌入同源静态模块，不开新窗口；显式外部独立URL兼容保留。默认浅色、已有深色偏好保留，子页由宿主观察主站主题同步。原模块源码／共享 family-runtime 适配器预留独立模式，不新增账号或通用跨站SSO。配方直接同页下拉本人可用Key／模型／接口格式或手填，目录读取与权限复查由宿主处理，只有显式应用的连接进入子页内存；来源／窗口／nonce配对，退出取消并清空，不自动运行模型或下发JWT。
- **本轮内容与浏览器证据**：工坊16张原创图片＋4段真实授权电影节选＋4份Skills／工作流，共24条。8张新原创图涵盖年轻成年男女、时尚穿搭、Cosplay、动漫、原创机甲、动物、电商，未将用户参考截图作为生成输入；人物明确25岁以上。影片是Blender开放电影的CC BY节选，裁短／缩放／静音，逐条署名、链接与来源记录在assets/PROVENANCE-20261001.md，不是模型生成或模板实测。左侧点击直达大图／影片，右侧与弹窗完整contain预览、互斥暂停；统一网格／横向分类。4185真实浏览器已验完整大图、4个不同影片实际解码播放、1280／390无外层横溢出、浅深主题与手机弹窗。截图studio-polish-*在本线程visualizations目录，临时视口已恢复浅色／正常尺寸；未刷新用户8080材料页。
- **本轮自动验证与边界**：Node／打包83项＋前端／国际化86项，共169项通过；类型、相关ESLint、Vite生产构建通过，日志output/family-workspace-20261001/{node-final,frontend-final,types-final,lint-final,build-final}.log。Docker内TestFamily两函数定向回归和Go embed编译通过，最终镜像状态另记；不标全量Go单测或新CI通过。开发目录入口曾误落登录页，已新增精确dev-products映射，4185用`--config vite.config.ts`避免旧gitignored编译配置优先加载。24个包号并行文件SHA256逐项一致未改，不提交推送／生产发布。
- **参考与未验**：Chrome扩展连接连续返回失败，不能读取用户Chrome已打开的3条抖音；IAB仅确认首条标题／播放控件，登录遮挡，未完整观看三条／逐镜反推。用户截图只作为年轻甜美人物方向，不复制脸／衣服。真实账号下同页Key／模型应用与付费生成未验；模拟宿主／组件证明交互不能替代真实登录。此轮已异步询问是否继续更新本机8080，尚未收到回复，仍保留原8080镜像c599729e28df及数据／容器，不把4185新源码或新测试镜像当8080已更新。

- **2026-10-01 一体化已部署本机 8080（最新终态，优先于下方待构建记录）**：用户明确选择更新本机 8080，现由同一应用镜像提供主站、/recipes/、/studio/，首页标为“本站产品”，不要求另行托管或设置外部地址。运行镜像与构建镜像一致 sha256:c599729e28dfe8572d14c0577e93f4f5818904e489e918a085d6fd46f7bc994b，主站入口 /assets/index-CYEQGOb2.js；原应用单独替换，应用／PostgreSQL／Redis均healthy，health=ok，原DB／Redis容器ID及sub2apicust-local_app_data未变。25项真实HTTP产品边界／资源摘要检查通过、4个个人接口仍匿名401，原精确CORS许可与未知来源403保持；业务计数和全部280键设置摘要前后无变化。证据 output/family-integrated-20261001/{docker-deploy.log,verification.json,verify.ps1}。
- **本轮编译／测试及失败边界**：177项自动测试、类型／相关lint／Vite构建通过；Docker内新增2个Go测试函数（包含14个路径子用例及两种embed入口）通过，Go embed二进制编译完成，日志docker-build-targeted.log。首次尝试完整web测试失败，原因是旧用例请求已不存在的/logo.png（HEAD的public仅logo.svg），不是本轮新增路由断言；旧测试未修改，不恢复旧品牌图，构建仅执行新增TestFamily回归，测试源码不复制到最终运行镜像。原失败日志docker-build.log保留，不宣称全量后端测试或新CI通过。回退标签与606446字节数据库备份沿用下方本轮已核验记录，不删旧镜像／旧产品容器或恢复数据库。
- **实际页面取证与未验项**：8080浏览器验证配方8份、创作筛选2份、会议及生图示例整理、当前API来源只读、工坊12份创作灵感＋4份Skills／工作流、图片加载、复制成功提示和同源家族链接；1280桌面、390手机无横向溢出，浅／深主题可切换，最终恢复深色／临时视口。两页warn/error日志为空。截图family-integrated-{recipes,studio}-{desktop,mobile}-20261001.png及browser JSON在本线程visualizations目录。家族匿名入口和另开/connect/recipes均返回8080原登录页且保留返回参数；IAB点击配置按钮进入等待但未列出弹窗，未完成真实账号登录／已有Key授权／付费模型或系统剪贴板验收，不能说完整接入通过。仅本机部署，本轮未提交推送／生产发布，24个包号并行文件SHA256保持未动。

- **2026-10-01 配方／工坊一体化开发与本机部署授权（最新决定）**：用户要求两个产品开发后一起部署，并明确选择更新本机 8080；生产仍由用户手动更新。原静态源码保留，新增 Vite 打包器默认生成同源 /recipes/、/studio/，主站首页／控制台空地址直接使用真实内置产品，不再要求三个 MAGIC_* 变量或额外产品服务；显式独立 HTTPS 覆盖仍可用，错误地址仍禁用。配方外置 JS／CSS、内置模式固定当前站点配置选择并忽略 api_site 提示，工坊新增同站点家族首页链接；没有新生成模型、定时任务、自动安装 Skills、账号系统或 API Key 单点撤销。
- **实现与初步取证**：新增 custom_family_products.go／test 及 build-products.mjs／d.mts／test，Go embed 两条入口先处理独立命名空间，白名单 GET／HEAD、404／405、禁止目录和原始源图／文档，专属 CSP 不启用脚本 unsafe-inline，不改变全站安全策略或 API 认证。177 项自动测试通过（打包3、配方55、工坊21、原站98），类型／相关 lint／Vite 构建通过，日志 output/family-integrated-20261001/{bundle,recipes,studio,frontend,types,lint,build}.log；新 Go 文件由缓存 Go 容器仅定向 gofmt，镜像编译与后端测试仍等待本轮构建。不沿用旧 25071d4 CI；该远端旧功能不含一体化。
- **本机备份与范围**：Docker 最初未启动，启动本机引擎后发现原三服务停止，保留原容器／卷启动旧应用及数据库／Redis，代理授权200后完成606446字节备份与 pg_restore 目录／副本 SHA256 校验，摘要 0EEED4AF32A57EB5344D82321AD7660A1C75EE42E8C02ABC25AB133958DDBEBC；回退标签 sub2apicust:before-family-integrated-20261001-001108，旧镜像94dba4064ba6。原基线 users3／keys0／groups3／plans1／orders0／dedicated0／settings280，证据 before.json。本轮本机镜像将来自当前工作树，含此前存在的未提交包号代码；不修改、提交或推送其文件，不把本机完整工作树镜像当限定产品发布快照。只更新原应用、保留原精确 CORS 叠加层，不删除旧产品容器或数据；部署终态以后续记录为准。

- **2026-09-30 家族视觉功能已提交推送（最新发布状态）**：功能提交 25071d4d547c69d6c933bbcb0c276585b23b4b68 已推 origin/main，git push 与 ls-remote 核对远端 SHA 一致。16 个文件包含工坊等宽三图／置顶／最新人像、皇冠图标、SynaRoute 公开官网／下载链接及文档／回归；实际暂存快照 72 项及类型／lint／构建通过，秘密模式扫描无命中，git diff --cached --check 通过。24 个包号并行文件和三条专用主题规则保留未提交，不代用户部署生产或重建 8080；本机 4184 源码预览不等于新镜像已发布。
- **本轮远端构建取证（收工时未完成）**：仅核对功能 SHA 25071d4 的工作流，GHCR 36736764218、CI 36736764231 仍 in_progress；安全扫描 36736764322 已 success，CI 的 frontend、golangci-lint、shell、release-helpers 已 success，test 仍在 Unit tests，镜像处于 Build and push，不标全绿或镜像已可拉取。证据 output/family-polish-release-20260930/{runs,jobs-36736764218,jobs-36736764231,jobs-36736764322}.json；三项 MAGIC_* Repository Variables 名称只读查询仍为空，生产配方／工坊入口仍待地址配置和重新构建。预期功能镜像标签为 sha-25071d4，必须从本轮成功任务日志确认实际标签／摘要后再更新，不沿用 321fe99 发布结果；后续交接 skip-ci 文档提交不生成替代功能镜像。

- **2026-09-30 家族视觉更新再次获准提交推送（最新授权）**：用户要求“提交并推送代码，我更新下”。本次仅发布工坊三张等宽图片及置顶交互、最新深发参考人像、家族首页皇冠图标、SynaRoute 官网／下载链接和相关配置说明／回归；工作区包号专项前后端及三条专用主题规则继续隔离，不混入本轮提交。不新增跨站 SSO、统一退出或配方／工坊一体化托管；主站家族首页与 API 控制台共用当前来源登录，8080 与 4184 是不同来源的本地运行／源码预览。发布验证与远端 SHA／CI／镜像以本轮后续记录为准，不沿用 321fe99 的成功结果，不代用户部署生产。
- **实际暂存快照验证**：本次限定 16 个文件，使用 git checkout-index 导出独立暂存源码，复用现有前端依赖；72 项／8 文件（家族相关 69 项＋国际化 3 项）、vue-tsc、相关 ESLint、Vite 生产构建全部通过。日志 output/family-polish-release-20260930/{tests,types,lint,build}.log；24 个并行文件 SHA256 前后一致，专用包号配置新文件未进入快照，主题仅剩三条包号规则未暂存。未重建本机 8080、部署生产、登录或调用收费模型；原有 Browserslist／chunk 警告不在本轮处理范围，完整后端 CI 仍需核对本轮 SHA。

- **2026-09-30 SynaRoute 免登录官网／下载入口（本地源码，未发布）**：用户要求在家族展台增加下载或官网链接，也可提供未安装引导。本轮选择确定性的公开链接，不新增安装检测或改动原密钥导入：FamilyHomeView的SynaRoute面板在原配置动作下增加“下载客户端”／“访问官网”，匿名与登录用户均可点击，新窗口且noopener noreferrer、不携带Key／JWT、不自动下载或安装。新增family/synarouteLinks.ts集中官方站点https://synaroute.mofamilys.com与/zh/download；来源为真实工程D:/ccfile/SynaRoute的README、site/src/config/site.ts及App.tsx下载路由，不采用搜索结果中的同名第三方。KeysView原导入和100ms焦点提示保持未改，不能把其启发式当真实安装检测；此轮无新上游接缝。样式仅theme.css，原品牌紫色／配置登录／SynaRoute关系图和工坊三图保留。
- **SynaRoute 入口取证及未知边界**：69项／7文件回归通过（此前57项＋home新增1项＋synaRouteImport11项），相关ESLint、vue-tsc与Vite构建通过，证据output/synaroute-links-20260930/{tests,lint,types,build}.log。4184实际匿名1280桌面／390手机两条链接可见、均44px高、无横向溢出，键盘Tab到官网链接focus-visible及2pxoutline；截图synaroute-links-{desktop,mobile}-20260930.png与browser-evidence.json，临时viewport reset，交付tab9保留。网站在线访问本轮未成功：web工具不可访问／GitHub cache miss、隐藏官网页超时导致工具重置、两条HEAD各12秒超时；来源已确认不等于官网此刻可达，不能说安装包下载或真实客户端导入已验。本轮未登录、调用收费模型、修改SynaRoute工程、提交推送、重建8080或部署生产，保留包号及此前并行修改。

- **2026-09-30 工坊三图与参考人像终态（本地源码，未发布；覆盖下方雨夜两图及金发版本）**：用户要求保留原玻璃瓶和狐狸、新增第三张人物，继而否决金发欧美人物，最新以用户提供的深发色室内写真为视觉参考。内置 image_gen 生成虚构28岁成年女性：深棕长发、安静表情、米白吊带／淡蓝灰细条纹衬衫／牛仔裤、自然坐姿，手放膝上；不使用裙装、摸头、截图黑框或尺寸徽标。新增studio-reference-portrait.png／webp／provenance.json，原PNG保留，网页WebP1122×1402、113416字节，精确prompt与参考边界已登记；不代表真人身份、代言或模板实测。此前被否决资产移入ignored output/studio-floral-20260930/iterations，不混入发布；原瓶／狐副本与magic-studio原thumb的SHA256逐一一致。
- **最新尺寸纠正与交互**：用户指出另外两张过小、不协调，theme.css中仅工坊列宽改为40%文案／60%图区，三图统一4:5等宽正常流Grid，轻微±2°与±8px错落，默认互不遮挡；手机去除图区多余最小高度。三张button保留悬停轻抬／回正／1.1倍并置顶、点按单选／恢复，键盘焦点层级高于悬停，reduced-motion关闭过渡。FamilyProductVisual仅custom组件、无新上游业务接缝；caption为“AI 视觉示例 · 非模板实测”。原首屏、品牌、登录／入口、模型功能与包号并行修改保留，没有触发网站模型请求。
- **本轮取证**：57项／6文件通过（home11、products17、switcher8、header3、upgrade8、AppSidebar10），相关ESLint、vue-tsc与最新Vite生产构建通过，日志output/studio-reference-20260930/{tests-final,lint-final,types,build-final}.log。4184实际源码浏览器1280桌面三图约178.46×223.07px、390手机约84.66×105.82px，三对默认相交面积均0、三图均加载、页面无横向溢出；每张实际鼠标悬停z20且pressed=false，玻璃瓶Tab焦点focus-visible/z30，手机尺寸点按true→false。手机尺寸不是实机触摸验证，减少动画仅源码契约测试。证据browser-evidence-final.json及本线程studio-balanced-{desktop,mobile}-20260930.png；临时viewport已reset，交付预览tab7保留。此源码预览设置请求仍回退Sub2API，不当原品牌配置、真实登录或生产验收；未提交推送、重建8080或部署生产。

- **2026-09-30 家族首页侧栏图标（本地源码，未发布）**：用户指出家族首页与仪表盘四宫格重复，已新增custom/components/FamilyHomeIcon.vue的1.5px皇冠轮廓（currentColor、aria-hidden），AppSidebar仅新增带[CUSTOM]的import并将共享buildSelfNavItems内`/family`条目改为该组件，用户端与管理员个人区均复用；两类仪表盘仍DashboardIcon，菜单、路由、登录及主题未改。upgrade-contract新增图标区分回归，55项／6文件通过，相关ESLint与类型通过，证据output/family-icon-20260930/{tests,lint,types}.log；UPGRADE与CUSTOMIZATIONS登记接缝。未提交推送、重建8080或部署生产，没有真实登录侧栏视觉验收；此前工坊人像本地改动与包号并行工作保留。

- **2026-09-30 工坊人像与地址配置说明（本地源码，未发布）**：按用户要求仅替换首页工坊左侧玻璃瓶为既有虚构成年女性雨夜写真，右侧纸艺狐狸及非模板实测说明保留。新增custom/assets/studio-rain-portrait.webp为magic-studio/assets/rain-portrait-thumb.webp原样副本，SHA256均B030EEF9067933F921A8580053620610D7C5C46F31167487544D98DF826D5B40；原生成来源PROVENANCE.json，未新增模型调用、图片编辑或真人代言。FamilyProductVisual仅改import／src／alt，family-home加素材与懒加载断言，SURFACE登记来源，没有改主题或认证业务。FAMILY_PORTAL新增GitHub Settings → Secrets and variables → Actions → Variables的具体操作、三项Repository Variables示例、精确CORS与手动重建／拉取边界；未操作实际变量、DNS或服务器。44项／5文件定向回归、相关ESLint、类型与Vite构建通过，日志output/studio-portrait-20260930。4184为源码预览（自启session42274），1280桌面／390手机浅色无横溢出、人像完整可见、两图加载；截图studio-portrait-{desktop,mobile}-20260930.png与browser-evidence.json留证。该预览公开设置请求失败，首屏回退默认站名，不能当现有品牌配置、真实登录或生产验收；未重建8080或提交推送，保留包号并行工作。此前321fe99镜像不含本轮人像与说明。

- **本轮镜像已发布／集成待完成（覆盖下方构建中状态）**：GHCR36712632118现success，已从成功任务109877820940日志核验真实标签ghcr.io/engineemomo/sub2apicust:sha-321fe99及摘要sha256:c8dd5b34edacf29eec9e30f53f35f5de8c5898706b86715b81d70b8a73db4147，证据output/family-release-20260930/image-publication.json。本轮CI36712632123的Go Unit已success，Integration tests仍in_progress；frontend／Go lint／shell／release-helpers及安全扫描成功，不标全部CI通过。功能提交仍321fe99；随后仅交接文档skip-ci提交，不产生替代镜像。生产仍由用户备份后更新，不建议在集成未完成时直接更新latest。包号专项继续保留未提交，三个MAGIC_*变量未设、独立产品托管／入口仍待。

- **2026-09-30 产品代码已提交推送（发布状态优先）**：功能提交321fe9943b700384b10c2fc1368ed24c6b90eb97已推origin/main，git push与ls-remote核对远端SHA一致。内容为家族展台／控制台切换、配方模型设置／登录配置桥接、工坊独立源码与接入教程，共141文件；未确认合并的包号专项前后端、专用测试／文档及3条主题规则仍保留未提交。实际发布快照285项通过及类型／相关lint／Vite见下一段；本机94dba406镜像包含此前并行工作，不是本次限定范围的GHCR镜像。新提交GHCR36712632118在Build and push，CI36712632123在Go Unit；远端frontend／Go lint／shell／release-helpers及Security Scan36712632091已成功，未标全绿或镜像可用。记录output/family-release-20260930/{runs,jobs-36712632118,jobs-36712632123,jobs-36712632091}.json，不沿用旧43b715a结果。仓库MAGIC_RECIPES_URL／MAGIC_RECIPES_ORIGIN／MAGIC_STUDIO_URL均未设置（只读API核对名称），生产入口将禁用；独立页面仍需单独托管。未操作生产、域名／Variables或用户数据。

- **2026-09-30 本轮产品提交授权**：用户要求“提交并推送代码，我先更新一版看看”。按当前会话默认整理家族首页／产品切换、魔法配方、魔法工坊及接入教程的完整依赖；已询问是否同时合并包号专项，未确认前不混入其前后端运行时修改。原站镜像不自动托管独立产品，生产公开URL／精确origin仍须配置；不把本机4178／4179注入生产。推送与本轮CI状态以随后取证记录为准，不沿用旧镜像结果，生产仍由用户操作。
- **实际暂存版本验证**：141个产品相关文件，主题仅暂存产品／教程部分，包号3条专用规则及其前后端仍保留未提交。git checkout-index导出独立暂存快照，复用现有frontend依赖，无新安装；210项原站／19文件、54项配方、21项工坊通过，共285项，类型／相关lint／Vite生产构建通过。日志output/family-release-20260930/staged/{frontend-tests,frontend-types,frontend-lint,frontend-build,recipes-tests,studio-tests}.log，不能沿用整个工作树252项作为该发布版本计数；基础镜像验证仍等待本轮CI，未重新部署本机或生产。秘密模式扫描未命中，git diff --cached --check通过；3个配方源仅去除末尾空行。

- **互动展台最终修复终态（2026-09-30，覆盖紧随段落的初次a514镜像／251项／审查待完成状态）**：本轮用户再次否决旧卡片，已重做产品展台而非沿用旧ship。三项修复为配方浮签正常流不遮动作、工坊图区与caption分离、页头完整工具区保宽且空间不足产品坞第二行。family-header改为真实公告／语言／订阅组件，仅账户／数据／国际化模拟，不再以文本替身验收。252项／22文件回归通过（家族44项、i18n3项已包含，夹具另跑3项不重复计数），类型／相关lint／Vite通过；Docker内前端及Go embed成功，未全量Go单测／本轮CI。最终8080镜像与容器一致94dba4064ba6d382772e6cc90630cca85a56d315e253715f05142bf098a0c278，入口index-BSFHhB79.js；四容器healthy、health=ok。原数据库／Redis ID、app_data、users3／keys0／groups3／plans1／orders0／dedicated0／280项设置与摘要均不变；CORS及匿名保护通过，配方4178仍f74233cb565c，工坊4179原服务未动。证据output/family-deck-20260930/{regression-final,types-final,lint-final,build-final,header-fixture-final,docker-build-final,docker-deploy-final,verification-final}.log及verification.json；原604527字节备份／SHA256 A0A683397FF0DD821ECE58F6F0D65FFD623C91BEBC0A362A18C18E5F25804444与before-family-deck-20260930-175621回退标签保留。
- **最终视觉／审查范围**：同路径覆盖七张family-deck-*-20260930.png均打开检查。实际8080首页1440／1280深色、390浅色无外层溢出，方法单／浮签不遮字、叠图加载并与caption分离；完整组件夹具1600／1280／1024／390实测外层无横溢出，1600页头141px产品坞第二行、390坞client364／scroll496。实际点击配方CTA到8080/login?redirect=/dashboard?product=recipes后返回，未输入凭据；API与SynaRoute预览运行，紫色191/169/243，手机ArrowLeft／focus-visible与详情展开验证。browser-evidence-final.json含几何及空warn／error日志。首位审查长时间未完成已向用户披露、关闭并换人，中途记录review-first-interim.md不算完整报告；Mendel完整初审fix，两项与首位手机遮字合并三项，最终Verdict三项均resolved、disposition: ship，仅限该修复清单，非全产品／真实登录／用户视觉或生产批准。报告review-{initial,verdict}.md。Godel文档继承检查No changes，documentation-check.md明确仅源码与既有规则匹配，未读取DESIGN／sidecar，不扩大其范围；母品牌未重写，检测器仍不可用。最终用户页tab22保留8080/home配方／深色、默认1280尺寸；只清理本轮17–21测试页与4182／4183两自启进程，用户配方材料页不刷新。未提交推送、付费调用或生产部署，保留并行工作；真实账号与模型授权待用户验收。

- **2026-09-30 产品入口再次优化（优先于下方旧卡片版）**：用户明确“效果太丑了、入口也不新颖”，上一轮修复评分不等于用户批准。保留原BrandPanel首屏／狮冠M／雾钛青／默认深色，下面由长篇卡片改为四产品互动展台；默认配方，产品标签只预览、不登录或请求，当前面板才提供真实进入动作，已知锚点优先，左右／Home／End及循环焦点、手机两列。新增FamilyProductMark，控制台继续四个直接入口并用相同标识，SynaRoute保留紫色；登录先进入API、显式授权及独立产品边界不改。长篇登录说明移入可展开详情，结构示意与非实测图例说明仍可读屏。
- **本机迭代证据**：251项原站回归通过（含家族43项、i18n3项，导出夹具2项与Docker3项不重复计数），类型／相关ESLint／Vite生产构建通过。首次pnpm因短路径临时目录权限失败，改为进程内工作区TEMP与直接Node；初次直接Vite缺node_modules/.bin路径，修实际进程PATH后构建通过，没有安装依赖。Docker代理授权200、原库备份604527字节／SHA256 A0A683397FF0DD821ECE58F6F0D65FFD623C91BEBC0A362A18C18E5F25804444，回退sub2apicust:before-family-deck-20260930-175621；仅应用重建，8080镜像与容器一致a514f4628b29275bf761a9e8ab152e806dfe40a91427a91b18b5ed69f668caef，入口index-rQ0oFKaS.js。四容器healthy、health=ok；原依赖容器ID／app_data／业务计数／280项设置摘要不变，配方仍f74233cb565c，CORS与匿名边界通过。证据output/family-deck-20260930；新的完整独立审查及Docker最终视觉尚在收尾，不能沿用旧ship或宣称用户视觉批准。检测器本次独立尝试退出127，engine未安装且缓存不可写，没有安装或宣称通过；未真实账号／付费模型、提交推送或生产部署，保留并行业务改动。

- **首页修复复核终态（2026-09-30）**：独立审查者Kuhn对原五项P2逐项评分均resolved，disposition: ship，范围仅该修复清单，不是完整产品／真实登录闭环／用户视觉认可或生产批准。报告output/family-layout-20260930/review-{initial,verdict}.md；最终Docker与371项定向证据见紧随段落。Chandrasekhar文档继承检查为No changes，报告output/family-layout-20260930/documentation-check.md；既有母品牌规范未重写，检测引擎不可用，不宣称检测通过。最终用户页tab15保留8080/home深色，恢复默认浏览器尺寸；该页warn／error日志为空。只关闭本轮有效空白测试页10／12／13／14与停止4181布局夹具服务，没有刷新配方材料页；临时错误页11因工具URL限制未手动关闭，交由默认临时页清理。

- **2026-09-30 用户确认两位成员都选全仍保存失败（最新现场反馈）**：不再把本次问题归因为成员漏选，不要让用户重复勾选。只读取线上/login及其公开JS：HTTP200，入口index-BM8dZCAf.js引用AdminDedicatedAccountsView-gZZSIQhx.js；管理页包含user_ids，但没有expected_updated_at、config_issue或本轮多组诊断键，说明该前端尚未包含当前未发布修正。Git HEAD 70271f0aa所含包号SQL明确拒绝账号关联其他分组，即使全部成员选中也会拒绝多组账号；当前源码已修且真实PG双成员多组场景通过。保存校验直接读事务内PostgreSQL，不读Redis。没有管理员认证或生产数据库访问，不能凭前端资源确定线上后端SHA／实际失败项。公开资源证据output/dedicated-config-20260930/online-frontend.json；本轮未修改业务、未提交推送或操作生产，保留并行改动，需前后端同版本发布后再验。

- **2026-09-30 品牌首页纠正与控制台顶部产品切换（最新有效方向）**：用户否决账户摘要＋目录首页，要求原品牌首屏保留、下方介绍附属产品、登录默认先进入魔法API控制台并在顶部切换。FamilyHomeView现原样复用BrandPanel标志／文案／三CTA，新增API通栏及配方／工坊／SynaRoute介绍和真实图例；普通登录／注册恢复原默认落点，匿名产品仅携带站内/dashboard?product=id，控制台显式继续打开，不自动跨域或弹窗。AppHeader仅在user存在时装配FamilyProductSwitcher，原账户／公告／移动菜单保留；所有原站视觉仍仅custom/theme.css。工坊独立内容首版已完成，4179原Node预览可打开，不再说仅规划；SynaRoute仍只是密钥页主动配置，不等于桌面SSO。
- **本机部署与审查终态**：已保留备份output/family-layout-docker-20260930/database.dump（603403字节，SHA256 B082162B87EC2C7804B444186746EC475C49888CE711BC0B06BA1C0929A35657，pg_restore目录及副本哈希验证），回退sub2apicust:before-family-20260930-164415；两次代理授权200，实际构建进程临时代理、Docker内前端及Go embed成功，只更新应用。最终8080镜像／容器一致d02c6ac5689c263ca970768a7e461eb378d905c712b339251bb4e738617d2484，入口index-B2Aw5OM6.js；配方4178仍f74233cb565c，四容器healthy、health=ok。库／Redis ID与app_data保持，users3／keys0／groups3／plans1／orders0／dedicated0／settings280及设置摘要无变化；4178两来源CORS204、未知来源403、真实个人接口匿名401。独立完整UI审查判fix的五项P2已修：keys保持API当前态、原M按日夜accent作mask、交接登记、figcaption读屏说明、brand-focus浅色键盘焦点；修复已部署，最终复核五项resolved，ship仅限修复清单。源码镜像包含工作树已有并行包号／教程代码，未回滚其他任务；构建成功不等于这些业务已真实验收。
- **最终验证及边界**：317项原站定向＋54项配方回归通过，共371项（Docker另跑3项i18n已包含，不重复计数），vue-tsc／相关ESLint／Vite生产构建通过。最终8080实际浏览器1280深色／390浅色无外层横溢出，原创图例加载且真实性说明进入AX树；API节点日夜颜色实测161/217/206与9/107/104，浅色导航键盘focus-visible实测9/107/104、2pxoutline（登录按钮继承原深色outline）。顶部切换为真实组件静态夹具1280／1600／390，1600同排高度75，手机胶囊client364／scroll453、页面390；明确模拟账户，非真实登录。最终五张截图为本线程visualizations的family-restored-*-final-20260930.png及family-switcher-fixture-*-final-20260930.png。先前4175验证匿名产品入口登录目标，后续该独立Vite端口停止／拒绝连接，未重启或杀其未知进程，最终改用实际Docker验收。无真实账号／Key输入或收费模型调用、无通用跨产品SSO／单点退出、无全量Go单测或本轮CI；生产未动、未提交推送。旧备份保留，下方账户摘要首页、登录返回/family、工坊仅规划以及8080未更新均为已被本轮覆盖的历史状态。工程日志output/family-layout-20260930-{regression,types,lint,build,recipes}.log，备份／build-final／deploy-final／verification在output/family-layout-docker-20260930。

- **2026-09-30 多专属分组规则修正（最新用户要求，源码完成未发布）**：用户明确账号关联多个专属组正常，并确认用于同一批包号成员；此要求覆盖此前“一账号只能关联一个组”的历史规则。保持每账号一条有效包号及主要绑定组，允许关联多个同平台标准专属组；各组仅含此账号、无备用路由／订阅、无名单外授权，保存检查所有关联组的非成员未停用Key。成员需授权主要绑定组，其他关联组按需授权同批成员，用户仍可使用无关公开组，此包号账号不进入公开池。新增custom_dedicated_groups.go，保存／配置健康／额度视图统一新结构校验；Key认证、调度／转发／WS复用Check按实际Key组识别关联绑定并检查授权，不只移除保存条件。移除成员事务撤销原账号所有关联专属组授权并停用对应Key，保留无关组；撤销／到期／删除仍阻断全部关联组，旧隔离标记保留，无迁移。
- **截图与证据边界**：用户截图显示两位用户都勾选同一标准专属组且公开组默认可用，截图不能证明弹窗保存成功、包号实际提交user_ids或账号关联数据。已在隔离PG复现：两位已授权但包号只选一位时，另一位被判名单外授权；同一条包号同时选择两位可保存，无关公开组授权不影响。新失败文案提示“同批共用者请全部加入同一条包号”，仍不擅自自动增加成员。待用户确认所指是用户授权弹窗保存还是包号绑定保存，线上根因尚无现场数据，不宣称已修线上。
- **最终验证**：46项前端定向（含原用户授权弹窗4项）、相关ESLint、vue-tsc通过；独立Docker包号service／handler测试及Go embed编译通过。临时PG18真实执行30个配置场景及原多人回归全部通过；多组场景额外验证两成员两组Key认证／选号、错账号／非成员／未授权组拒绝、改公开组与名单外授权拒绝、移除成员清两组且无关组Key保持、换主要组历史标记、撤销／删除／重分配。日志output/dedicated-config-20260930/{multi-build,multi-postgres}.log，临时容器清理；未部署8080／生产、未提交推送，保留并行家族等改动。

- **2026-09-30 线上包号保存400诊断（本轮源码，未发布）**：用户报告已配置专属分组与授权仍返回DEDICATED_ACCOUNT_CONFIG，确认发生在线上，尚未提供account_id／group_id／完整user_ids或服务器只读访问。源码核实Save在串行化事务内直接查询PostgreSQL，完整性校验不读取Redis或调度缓存；不能仅凭通用错误断定线上具体失败项。新增custom_dedicated_config.go：失败后在同事务读取19类条件中的首个失败项，返回白名单config_issue及最多10个资源编号；保留原错误reason与独占判定。handler保留具体错误元数据，管理页按白名单诊断显示中英文失败原因／编号，旧响应回退原文案，不直接展示原始错误或凭据。未调整授权或放宽隔离规则，无迁移。
- **本轮验证**：42项包号前端回归、相关ESLint、vue-tsc通过；独立Docker镜像包号service／handler定向测试及Go embed编译通过。临时PostgreSQL18（无宿主端口、无业务卷）实际执行23个新增保存配置场景及原多人回归通过：完整配置可保存，分组名称不替代专属开关、关联其他组／账号、备用路由、非成员授权／未停用Key、缺成员授权等正确拒绝并给出具体编号，inactive／disabled旧Key场景正确通过。证据output/dedicated-config-20260930/{build,postgres}.log；测试容器已清理。未连接生产数据库、未部署、未提交推送，保留其他线程现有改动；线上根因仍待编号与现场配置证据。

- **2026-09-30 魔法家族统一首页（最新入口范围，源码与本机预览完成未发布）**：用户要求首页登录后跳转／打开家族产品且避免重复登录。按现有魔法 API 账号复用实现/family，原默认/home（无home_content时）与/brand装配新首页；原API介绍迁至/api。官网导航与控制台个人区新增家族首页。首页为账户摘要与四产品目录：API直接本站控制台，配方新窗口打开，SynaRoute进入原密钥页主动导入，工坊标为暂名规划、禁用入口。沿用狮冠M、雾钛青及默认深色，原站样式仅custom/theme.css，不改认证页、计费或任何后端权限。
- **本轮登录与发布边界**：站内沿用authStore有效会话，配方选择配置复用原站/connect/recipes；没有新账号系统，不向产品链接发送密码、JWT、刷新令牌或Key，不自动创建密钥／授权／调用模型。SynaRoute并未实现桌面账号SSO，全产品单点退出也未实现；原站退出后配方页已授权Key须清除或刷新。生产VITE_MAGIC_RECIPES_URL未配置时禁用入口，只有本机DEV默认明确4178预览；api_site仅公开origin提示，配方用原官方／本机白名单预填，本页底部新增新窗口返回家族首页，保留材料。Dockerfile已新增URL／ORIGIN两个ARG及前端构建注入，custom-image.yml从仓库Variables接线；覆盖下方咨询“尚无构建接线”的历史结论，但尚未设置实际变量、构建Docker镜像或运行本轮CI，配方静态产物仍需独立发布。
- **本轮验证证据**：237项原站定向回归（含家族20项、路由45项、原定制及国际化3项）与54项配方Node回归通过，共291项；vue-tsc、相关ESLint、Vite生产构建及workflow YAML解析通过，日志output/family-20260930-{focused,regression,recipes,types,lint,build,workflow}.log。额外来源测试确认公开配方链接不能自动预填本机登录目标，本机来源提示仅限本机HTTP预览。浏览器4175验证/family与默认/home、匿名登录回跳/family、/api原介绍保留；首页1280深色／390浅色实测无外层横向溢出。4178新空白测试页预填本机4175、返回链接安全属性、接入等待／取消、390新增登录设置布局通过，两页错误／警告日志为空。截图family-home-desktop-dark-20260930.png、family-home-mobile-light-20260930.png、family-recipes-connect-mobile-20260930.png在本线程visualizations，首页tab6已留给用户；临时视口复原、主题恢复深色。配方最终构建136797字节并重启自己服务session86473，未刷新用户旧材料页。登录态切换由模拟组件／守卫测试证明，未输入真实账号／Key，真实授权跨窗口闭环、独立产品账号SSO、CORS及收费调用仍未实测。Vite4175原服务保留，8080镜像／生产未更新，无提交推送；其他并行工作未回滚。

- **2026-09-30 魔法工坊本地首版（后续授权，定时关闭）**：用户从「先规划」推进为「先做产品，定时任务先不启用」。独立product-samples/magic-studio已实现8图片／4视频提示词／2 Remotion Skill／2 ComfyUI模板指引，共16条及4个组合玩法；搜索／媒介／用途／资源类型、排序、材料替换／中英模板／复制／TXT／ID分享、收藏／最近浏览／日夜主题／大图／视频运镜播放暂停。8张内置image_gen原创视觉素材及原狮冠M，25个PNG／WebP来源已只读核对通过，精确prompt见assets/PROVENANCE.json。视频是二维示意、模板未目标模型实测、Skill未安装运行，不以视觉图代替模板效果证据。无模型API、自动搜索或调度；页面CSP connect-src none、仅白名单静态文件。
  - 本地4179运行（serve.mjs，session77912），最终21项Node／jsdom／HTTP定向通过，实际浏览器复制替换材料／ID分享、TXT落地内容、收藏重开、空搜索恢复、大图Esc、视频播放暂停已验证；320／390／768／1280／1440／2560宽度无外层横向溢出，截图／测试／素材核对／下载证据在该目录.impeccable/review。部分全页导出及快速跨尺寸截图有工具失真，Skills／320／视频以补录实际视口及DOM几何为准；设计扫描引擎不可用exit127，不报告扫描通过。独立审查首轮fix：资源类型跨区匹配、手机返回列表保留材料、dialog名称三项P2已修；同审查者Verdict Pass为ship、三项均resolved，仅覆盖这三项，不扩大为全产品或生产批准，记录finish-review.md／finish-verdict.md。DESIGN.md与.impeccable/design.json已归档实际token、组件、断点及来源，不重定母品牌；最终用户窗口1280×720恢复深色首页、图片全加载、无横溢出，preview.png为实际视口。未改中转站／配方／SynaRoute运行时代码或配置，未Docker重建／提交推送／生产部署，保留所有并行未提交工作。

- **2026-09-30 魔法 API 到配方入口核查（只读源码，未新增入口）**：用户询问魔法 API 是否已有跳转到配方的入口。rg检查frontend/src及官网PublicLayout、BrandHomeView、控制台AppSidebar，配方相关命中只有登录选择页、辅助与测试，没有本轮新增的官网／侧栏产品跳转入口。现有/connect/recipes是从配方发起的登录后配置选择页，不是配方首页。线上管理员自定义菜单配置未读取，不能宣称线上配置一定不存在。建议部署地址确认后在官网导航与控制台个人区各加“魔法配方”入口、新窗口打开；本轮仅回答现状，没有修改导航、站点配置或部署。

- **2026-09-30 魔法配方部署咨询（仅方案，未执行）**：用户询问发布部署方式。源码确认配方由build.mjs生成单一index.html，serve.mjs仅本机预览；现有Dockerfile构建原站，不包含配方静态发布。建议独立静态入口复用现有魔法 API 登录与服务，可选子域名（recipes.mofamilys.com仅候选）或同源/recipes/路径，未替用户选择／配置域名。发布须分别托管静态文件及更新原站/connect/recipes；独立来源须构建时VITE_MAGIC_RECIPES_ORIGIN与运行时API CORS双许可。查证Dockerfile无该ARG、custom-image.yml未传该build-arg、.dockerignore排除.env.*，因此子域名方案正式发布前尚需接构建链路，不能给运行容器加变量冒充生效。已查官方Nginx、Vite与MDN来源并更新site/README部署说明；没有改构建链路、域名、服务器、安全配置或生产，也没有提交推送／模型请求。站点管理工具及反向代理配置未知，待用户提供。

- **2026-09-30 魔法配方登录魔法 API 选择配置（最新范围，源码完成未发布）**：用户要求在自带模型设置之外，登录魔法 API 直接选择配置。独立4178页新增登录入口和手动连接并列；原站custom新增受保护/connect/recipes，复用既有登录／2FA／passkey成功后的站内回跳，不在配方页收集账号密码。登录后按本人keysAPI分页读取已有密钥，过滤停用／过期／额度耗尽／无效分组，生图排除明确禁用分组；用所选Key只读GET模型目录，用户选模型与文字格式、核对接收来源并显式授权后回传所选API连接。账号密码、登录JWT不出原站，不自动创建／改绑密钥，不修改权限或计费。
- **本轮安全与发布边界**：生产接收来源默认同源，独立配方需原站构建环境VITE_MAGIC_RECIPES_ORIGIN精确允许；本机源额外支持4178。配方登录目标仅官方ai.mofamilys.com或本机HTTP，其他服务仍手动接。回传严格匹配窗口／origin／32字节随机nonce／用途／五分钟有效期，单次接受且有确认消息；不使用通配targetOrigin，密钥不入URL或持久化。登录和导入不自动调用生成模型；视频仍复制使用。窗口被COOP隔离时提示重试或手动配置，不绕过安全头。新增原站样式仅custom/theme.css专属作用域，无新增上游业务接缝或依赖。
- **本轮验证证据**：51项独立Node测试及99项原站定向测试通过（新增选择页／辅助25项，99项含国际化3项），类型／相关ESLint／Vite生产构建通过，日志output/recipes-login-20260930-{node,vitest,regression,types,lint,build}.log。自动测试覆盖允许来源、错误窗口／nonce／用途／过期／重放、显式发送且无登录令牌、文字／生图导入、分页过滤、迟到目录与接收确认；浏览器4178新增入口和1280桌面无外层溢出，4175匿名选择路径进入原登录并保留完整站内目标，配方页警告／错误日志为空。390视口设置工具没有改变新测试页实际DOM宽度（仍1280），因此新增手机控件未实测；不沿用上轮手机证据。未用真实账号密码／Key，登录后真实选择及跨窗口导入闭环、实际CORS／权限待用户验收；没有付费生成请求。本机4178重启到新构建（session18607），另开空白页保留给用户；用户旧页未刷新。只在Vite4175验证原站新源码，8080镜像与生产未更新，未提交推送，其他并行改动保留。

- **2026-09-30 第四产品规划（本轮仅方案）**：用户提供创作工坊截图，要求第四款产品增加图片／视频分类及定期发现提示词、Skills和新玩法，并明确「先规划」。新增product-samples/magic-studio/PLAN.md，暂名「魔法工坊」，建议以内容创作者优先，独立媒介与资源类型分类，效果预览→改变量→复制／复用；16条候选＋交互原型、约40条精选首版、采集审核、站内图片、异步视频分阶段。已网页读取核实Google图片／视频、Runway指南、ComfyUI模板及Remotion Skills来源存在；未收集完首批条目或实测。每日09:00发现／周五18:00精选仅为Asia/Shanghai频率建议，未创建自动化；优先人群与首版生成深度已询问，暂按建议假设，名称／托管／接口／预算待定。未改业务、主题、模型配置、Docker或生产，未安装执行外部Skill、调用模型或提交推送；保留既有未提交任务。

- **2026-09-30 魔法配方自带模型接入（覆盖历史“不调用模型”产品边界，独立本地原型）**：按用户要求新增顶部页内“模型设置”，文字／生图分别配置API地址、Key及模型名。文字兼容Chat Completions／Responses并支持多轮追问，生图兼容Images API返回base64／URL图片；视频仍复制使用。整理及应用配置不发送请求，运行／追问才直接向所填接口发送提示词与当前对话；密钥与配置仅当前页内存，不存储、不写URL、不读取中转站业务配置。CSP开放HTTPS及指定本机HTTP，API请求不带Cookie、不跟随重定向、不自动重试；修改材料／切换配方／更改当前连接清除对话并中止等待，迟到结果隔离。文字120秒／生图180秒，停止等待不能保证服务端不计费。
- **本轮验证**：语法检查、构建及44项core／model／DOM测试通过（日志output/recipes-model-20260930-tests.log，含真实本机HTTP往返）；浏览器用本机4180模拟接口验证Chat Completions、Responses、追问及512px图片显示，生图返回既有Logo，不是AI效果实测。1280深色、390浅色、320窄屏无外层横向溢出，浏览器错误／警告日志为空；截图在本线程visualizations的recipes-model-settings-20260930.png、recipes-model-settings-mobile-20260930.png、recipes-model-run-20260930.png。模拟服务已停止，4178服务运行最终128364字节页面（session64858）；新用户预览留空配置，未刷新用户旧页。未知用户真实服务CORS与兼容性，未调用外部真实或付费模型、未提交推送或生产部署；未修改并行控制台／包号／接入教程业务。

- **2026-09-30 接入教程（本地已完成，未发布）**：新增免登录 `/guide`、登录后右侧正文 `/help/guide`、游客预览 `/preview/guide`，官网导航、首页 CTA 与 FAQ 可直达；覆盖 Claude Code CLI／Local 桌面、Codex CLI／本机桌面、Cherry Studio、Cline、Continue、OpenCode v1、Aider，以及 Cursor BYO Key 限制。补安装入口、终端密钥变量、Codex 用户配置及桌面“使用密钥”／API Key Mode／模型目录步骤，Continue／OpenCode／Aider可折叠示例；地址基于公开 apiBaseUrl，示例仅占位符。上下文链接保留控制台／预览；游客密钥入口支持登录提示和取消。router/index.ts新增限定教程锚点滚动接缝，其他页面行为保留。官方配置文档已在页面列出，特定客户端版本及桌面云端会话需用户实测。与现有未提交包号修复及 product-samples 并存，未提交、推送、重建8080或发布生产。
- **接入教程验证**：80项定向前端测试、国际化3项、typecheck、相关ESLint及生产构建通过，证据output/client-guide-20260930。4175真实浏览器验证游客密钥提示／取消、键盘折叠示例、Codex锚点滚动到章节、预览教程↔FAQ保留一个main与侧栏；1280深色和390浅色无外层横向溢出。截图client-guide-desktop-dark.png／client-guide-mobile-light.png在本线程visualizations目录；Vite仍监听127.0.0.1:4175，可浏览 `/guide`。未使用真实密钥请求模型，未完成所有客户端及登录后控制台带数据验收；本轮不涉及后端，包号验证沿用该专项记录。

- **2026-09-30 包号专项修复已完成源码及定向验证（未发布，覆盖下方待修结论）**：用户授权修复四项。Save仍校验完整配置，但运行时只验证当前成员健康及授权，其他成员停用／软删除／撤权不再拖累全组；退出者旧Key状态不参与运行时全组拒绝，非成员自身仍拒绝。成员移除事务停Key统一inactive（保存兼容旧disabled），APIKeyService.Update新增[CUSTOM]接缝，显式启用及扩额／重置额度／清除或延长到期等隐式激活写入前统一验包号。编辑必传原始expected_updated_at，陈旧／缺失返回409；已撤销恢复必明确reactivate确认。管理页分开绑定状态与配置诊断，提供刷新重开和恢复checkbox；新增样式仅theme.css。严格一账号一组、无备用路由／影子及历史隔离标记均保留，无迁移。
- **本轮证据**：40项前端定向／定制契约＋3项国际化、typecheck、相关ESLint及生产构建通过；Docker最终Go定向service／handler测试及embed编译通过；最终隔离PostgreSQL真实执行65项service顶层测试（含版本冲突、确认恢复、移除事务回滚、同伴停用／删除／撤权、6类Key激活拦截、合法成员Key更新及改绑／删除隔离）通过。证据output/dedicated-fix-20260930（frontend-test.log、frontend-build.log、build-final.log、postgres-final.log），临时PG已清理。无宿主Go，使用Docker编译；未全量Go／本轮CI、未做带真实管理员数据浏览器验收。未提交推送、未更新8080或生产，product-samples不变；生产历史403/503仍不知实际根因。

- **2026-09-30 包号专项审查（优先于下方移除修复已验证状态）**：隔离PostgreSQL复现：退出成员仅改旧Key状态即可让剩余者拒绝访问（active及inactive均复现，未发布修复有此缺口）；停用一个成员连带阻断其他成员；旧编辑payload覆盖后来的撤销并恢复绑定。源码另确认管理页缺配置健康状态。19个既有service测试＋1个四子场景审计测试完成，审计通过指坏行为复现而非修好。报告deploy/DEDICATED_AUDIT.md，证据output/dedicated-audit-20260930；三份被测业务源码哈希与工作区一致。此次仅审查及文档，不新增业务修改，不提交推送部署，不碰product-samples，临时PG已清理。

- **2026-09-30 移除包号成员400修复（未提交／未部署）**：用户报告从两个成员中移除一个无法保存。源码及隔离PostgreSQL证实：旧授权／密钥仍存在时新名单完整性校验返回DEDICATED_ACCOUNT_CONFIG。现Save在原串行化事务内撤销退出成员原组授权、停用该组Key但保留记录和组归属，再校验及保存；失败全部回滚，提交后清组和退出用户认证缓存。已停用的非成员Key不再阻断独占检查，非成员授权及未停用Key仍拒绝，不放宽一账号一组约束；重新加入不会自动启用Key。不触碰余额或其他组，无新迁移。前端提示同步。
- **移除成员验证**：31项前端包号测试、相关ESLint及typecheck通过；Docker Go service/handler定向测试、embed编译通过；19项service顶层测试含真实PostgreSQL成员移除／旧校验复现／回滚／其他组保留／重新加入Key不启用均通过。证据output/member-removal-20260930；仅临时测试库，无业务库操作，8080和线上仍为旧版。未全量测试或本轮CI、未提交推送部署，product-samples不动。

- **2026-09-30 已推送43b715a15**：完整代码SHA `43b715a15f599f5ae2125bc923849d57e4b3522c`，push及ls-remote核实origin/main一致。GHCR36657101655与安全扫描36657101610成功；镜像 `ghcr.io/engineemomo/sub2apicust:sha-43b715a` 已发布，摘要 `sha256:dc36b117e524262faddf0c13f0dad5c9ff510e385e69ddc2860145a7ea6566af`（构建日志确认）。CI36657101544的前端、Go lint、shell、release-helpers和Go Unit通过；8分钟等待窗口结束时Integration仍在运行，不能宣称全部CI通过。证据output/console-release-43b715a；用户发布前应等待该run成功，备份数据库／配置，迁移242后不要直接回滚旧单用户版。生产未操作，product-samples保留未提交。后续交接仅文档[skip ci]提交不产生新镜像。

- **2026-09-30 提交发布收口**：用户确认不放宽包号独占规则，并授权提交推送以便自行发布。保留“一账号一专属组、多成员”以及改绑／删除的旧Key隔离保护。本次提交仅控制台、多人包号、费用显示、测试及维护文档；product-samples与本机配置／备份不纳入。既有139前端定向与Docker Go／PostgreSQL验证及本机部署证据见下文；远端CI和GHCR须查询本轮提交，不沿用旧记录。生产由用户备份后操作。

- **2026-09-30 已部署本机Docker（优先于本轮未部署记录）**：用户授权本机测试。原三容器处于停止状态，保留原卷启动Postgres／Redis，先完成597905字节数据库备份与pg_restore目录校验、旧镜像标签sub2apicust:before-console-20260930，再代理预检200并构建。前端类型／国际化3测／构建及Go embed成功；仅应用容器重建，运行镜像与标签一致be2f26f98e8c3dee62d765a8b78e6cd33021dc3b6f618c76085e310180b234b8，三服务healthy，health=ok。迁移242两列和三个索引确认存在；数据库／Redis容器ID及app_data卷不变，用户2／密钥0／组3／套餐1／订单0／包号0／设置280计数保持。备份逐项比较只有Claude版本2.1.283→2.1.285与Codex版本0.158.0→0.159.2两键自动同步，09:44启动日志证实，其余278键不变。HTTP入口index-Ckx_MqIV.js及三个新页面chunk验证通过，账户与管理员接口匿名401；浏览器8080官网正常打开并留给用户测试，尚未做带数据管理员交互验收。证据output/console-deploy-20260930，备份SHA256=1D900EC14FB9915B5E912F0A84778271CEB149178B7CC526E3A9F55411EE037E。未提交推送或操作生产；迁移后的回退需兼顾数据库，不能直接跑旧单用户版本。

- **2026-09-30 控制台修复与多人包号（未发布）**：完成登录后隐藏官网两处游客入口、控制台FAQ右侧内嵌 `/help/faq`（公共 `/faq` 保留）、管理员用户／账号／组名称与用户端组名；智力效率默认矩阵新增平均任务费用及聚合口径，明确不是本站／Token单价。用户确认多人共享一个账号和专属组、独立Key及余额；包号可选1–100成员、编辑账号／组、撤销后确认删除。242迁移兼容旧单用户，逻辑删除与改绑旧资源保留隔离标记，重新分配仍严格检查授权及非成员Key，不自动修改授权／计费。
- **本轮验证与边界**：139项前端定向测试、国际化3项、类型检查、lint及构建通过；Docker隔离Go定向service/handler测试与embed编译通过；额外在临时PostgreSQL18真实执行18个service测试（含旧数据迁移、多人、移除成员、改绑、删除及新分配SQL），全部通过。证据 `output/console-fixes-20260930/`。未跑全量Go测试／本轮CI，管理页浏览器带数据视觉验收仍待；本机8080业务容器未替换，生产未操作，未提交推送，保留product-samples。
- **线上403/503**：用户无法提供日志；已定位源码触发条件，但两个历史请求根因仍不知道。403由专属校验失败返回，旧版单用户限制与多人授权是可能条件；503既可能来自无可用账号兜底，也有Responses依赖缺失路径。不得宣称此次修改已经解决生产错误。配置规则／具体请求ID／迁移回退注意事项见 `deploy/DEDICATED_TROUBLESHOOTING.md`。

- **2026-09-28 游客发布最终核验（优先于下方待CI状态）**：功能9a31787de及测试初始化修正fa45b2023均已推送origin/main，ls-remote确认完整代码SHA为fa45b2023c01e5c84084dcbcb32d5a7c5bbe213a。该SHA的CI36424788919全部成功（含Go Unit与Integration、前端、Go lint、shell、release-helpers），安全扫描36424789053成功，GHCR36424788938成功；镜像日志确认latest及sha-fa45b20已发布，digest=sha256:69083ada3ce6432eaee2357e57fa384aa0804d125444cc33998e742c601d1d44。证据output/guest-release-fa45b20。新增SQLite测试按现有用例开启外键并限制单连接，已由本次CI验证；本机此前ea6e3a5ecc3c仍含相同业务源码，后续仅测试及文档变化，不为此重复重建本机。用户备份服务器后可在/sub2api-deploy运行./update.sh sha-fa45b20；生产未操作。179项本机定向回归和全量lint通过，product-samples未跟踪文件保留且未提交。本次收口文档使用[skip ci]，不将其当成另一个镜像版本。

- **游客发布CI修正**：代码9a31787de已推送origin/main，CI36423807298前端／Go lint／shell／release-helpers通过，Go单测仅新增TestPublicPlansAnonymousCatalog失败：SQLite外键开关未启用、ent建表报错；Integration skipped。已参照现有PaymentConfig测试开启PRAGMA并将测试连接池限定单连接，未改运行时代码。安全扫描36423807397成功，首轮镜像仍构建中；需要查修正提交对应CI后再给更新版本，不将首轮描述为全绿。

- **2026-09-28 游客功能提交收口**：用户授权提交推送并更新。本轮范围为游客官网／公开套餐／FAQ／游客控制台预览和响应式修复；product-samples保持未跟踪，不混入本次业务发布，共享交接历史保留。代理fetch确认提交前HEAD与origin/main一致；全量前端lint和179项定向回归通过，新增Go测试仅修正gofmt空格。本机8080已运行本轮响应式修复；Go embed此前Docker编译通过，Go单测与远端镜像发布结果必须查本次代码提交CI，当前不宣称通过。生产仍由用户手动更新。

- **2026-09-28 游客预览响应式修复已部署（最新）**：用户截图否决宽屏适配。2031px浏览器复现main限制1280px导致卡片右留白551px；仅theme.css去除工作区宽度上限、设置弹性对称边距、字段等宽列／手机双列、窄屏顶栏隐藏重复徽标以保持操作同排。320／390／768／1024／1440／2031／2560七档实测左右留白对称、无溢出、顶栏操作可见；65项定向回归通过。代理预检与数据库备份后重建本机Docker，前端类型／国际化3测／生产构建及Go embed通过；运行镜像ea6e3a5ecc3c021f3f128725320fcee293751c69dc795cc6ac8fb678b65e1b2d与标签一致、三服务healthy，数据库／Redis容器ID、用户／密钥／套餐／订单计数和设置摘要保持。8080浏览器2031px实测卡片左右边距均40.609px。证据output/preview-responsive-20260928，595135字节备份及回退sub2apicust:before-preview-responsive-20260928保留。未提交推送／生产部署，未下单；视觉待用户再次确认。

- 2026-09-28创作配方更新：新增07-image-prompt.md与08-video-prompt.md，现为工作3份、大学3份、创作2份。主按钮改为“整理并生成提示词”，明确本地模板组装、不调用模型；结果与教程按文字／生图／视频分别引导。构建与语法检查完成，30项core／DOM测试全通过；真实浏览器验证创作筛选2项、两份虚构示例及正确目标工具提示，桌面和390px无外层横向溢出。本机4178服务已重启（session75587），用户旧页未刷新。未调用模型、未验证媒体效果、未提交推送或部署生产；真实系统剪贴板与下载落地仍待验证。

- **2026-09-28 17:52 游客控制台预览已部署本机Docker（优先于下方未部署记录）**：用户明确授权部署查看。代理预检200，先备份数据库并保留回退镜像，再构建；前端类型／国际化3测／构建及Go embed编译成功，仅应用容器重建。运行镜像与本地标签一致`ecb3961845e8580395b31170a8a31bfa533e92710671beddf6c1249ed0594c2e`，三服务healthy、health=ok；数据库与Redis容器ID、app_data卷、2用户／0密钥／3分组／1套餐／0订单／280设置计数及设置摘要全部保持。HTTP入口index-C4UWqsIx.js引用GuestPreviewView-D-VurzQx.js；4个真实个人接口匿名401。浏览器8080实测首页游客按钮→预览、密钥操作登录提示／取消、真实套餐加载。备份和证据output/preview-20260928-175007，回退sub2apicust:before-preview-20260928-175007；数据库备份SHA256为6762F0F26B1ED26CA99867F9F67DCBB828730BC4DDB5D095393C4105742C0FA2。用户可访问http://127.0.0.1:8080自行验收；未下单扣款、未提交推送或部署生产，Go单测／完整登录支付仍待。

- **2026-09-28 新增游客控制台预览（源码完成，尚未进入Docker）**：用户确认首页增加「游客预览控制台」，现主CTA与公共导航均通往 `/preview`；独立只读侧栏包含概览、真实公开套餐、FAQ及五项账户功能说明，不挂载实际账户布局、不读取个人数据。账户操作弹可取消登录提示并保留目标；手机折叠菜单导航后收起。179项定向前端测试、3项国际化测试、类型检查、相关ESLint、生产构建通过；浏览器4175实测首页入口、密钥操作提示／取消、真实套餐、FAQ搜索空态、390手机菜单及1280桌面深浅主题，检查宽度无横向溢出。截图在本线程visualizations目录preview-desktop-dark.png／preview-mobile-light.png。8080仍为上一版e7a582285cbd，不含本轮新预览；本轮未重建Docker、未提交推送或部署生产，保留并行product-samples改动。
- **2026-09-28 配方使用教程**：用户要求补教程，独立原型新增顶部“使用教程”、首页折叠说明及结果区“复制后怎么用？”入口。四步涵盖选配方／填材料／生成复制／到自己的AI工具粘贴发送，另说明核对追问、手动复制、虚构示例、隐私与刷新丢失；明确本站不运行模型。返回按钮保留原配方、材料与结果。构建及27项core／DOM测试通过；浏览器验证大学论文草稿往返保留、焦点与hash不丢，390px无外层横向溢出且无错误日志；截图recipes-tutorial-desktop.png／recipes-tutorial-mobile.png在本线程visualizations目录。仅改site/{template.html,app.js,theme.css,ui.test.cjs,README.md,DESIGN.md}及交接／登记，未触碰其他并行业务改动，未提交推送或部署生产。127.0.0.1:4178已重启到本轮页面（会话54984）；用户原页未刷新，旧页刷新前需保留材料。
- **2026-09-28 游客官网已部署本机Docker（优先于下方源码阶段未编译／404记录）**：用户授权本机部署；desktop-linux本地npipe、代理预检200，实际构建进程配置临时代理，首次构建退出0，前端／类型／国际化3测及Go embed编译通过。仅替换应用容器，新镜像与容器Image一致`e7a582285cbdaeb78412a0414ac4280322c1654c05614fec869cced2c2e13b11`；三服务healthy，health=ok，Postgres／Redis容器ID及app_data卷不变。users=2、api_keys=0、groups=3、subscription_plans=1、payment_orders=0、settings=280计数未变；设置仅`openai_codex_client_version_synced`变化，逐项与备份比对及17:14:59自动同步日志确认，其余279键不变，不宣称全部设置摘要相同。公开目录返回既有1个在售套餐，字段白名单检查通过，4个账户接口匿名401；HTTP入口index-C6kd8wBQ.js引用PublicPlansView-DAcrv8XQ.js与PublicFaqView-Beye-j_3.js。浏览器在8080实测首页、真实套餐、FAQ搜索、购买→登录并保留plan=1；390px套餐无横向溢出。未登录账户执行购买、未下单／扣款，Go单测与完整支付联调仍待。旧镜像`sub2apicust:before-guest-20260928-171257`；备份output/guest-20260928-171257/database.dump共594458字节，pg_restore目录验证通过；同目录含before.json、build.log、settings-diff.json、verification.json、plans-mobile.png。未提交推送或操作生产；保留并行配方与交接改动。
- **2026-09-28 配方自定义确认框**：按用户截图要求，独立原型两处原生confirm已替换为主题内dialog，区分示例覆盖／清空；取消优先、关闭／Esc、Tab双向循环与焦点恢复、异步跨配方保护，旧浏览器不支持时保留数据。node build.mjs及24项core／DOM测试通过；真实浏览器验证模态、取消保留、Esc及焦点循环，桌面深色和390px浅色截图存recipes-confirm-dark.png／recipes-confirm-mobile-light.png，手机弹框边界19至371px，无页面错误日志。浏览器仅使用新测试页和虚构示例，未刷新用户原页或清除其材料；确认执行由DOM测试覆盖。127.0.0.1:4178已重启到新版（会话53940），旧页刷新前需保留输入。仅改site/{app.js,template.html,theme.css,ui.test.cjs,README.md,DESIGN.md}及交接／登记，其他并行的中转站业务改动不碰；未提交推送或生产部署。
- **2026-09-28 游客官网与FAQ实现**：按用户确认方案新增公开官网／套餐目录／FAQ（四类十二题），账户功能弹出可取消的登录提示；套餐跟随后台在售数据，登录／注册／邮箱验证保留选择，支付恢复及原续费优先。后端新增只读字段白名单 `/api/v1/payment/public/plans`，个人接口认证不变；首页自定义内容优先，正常模式默认游客官网。163项定向前端测试通过，类型检查、相关ESLint及前端构建通过；本机无Go，新增后端单测未运行、后端未编译，不沿用旧CI。浏览器实测官网、FAQ搜索、登录提示／取消与1280桌面／390手机深浅色，无外层横向溢出；截图为本线程visualizations的guest-home-desktop-dark.png和guest-faq-mobile-light.png。Vite4175预览进程本轮session73339；8080后端未更新，新目录实测404并就地提示，真实套餐／登录购买联调仍待。详情frontend/src/custom/GUEST_PORTAL.md，接缝见CUSTOMIZATIONS。保留原交接／product-samples改动；未提交推送、未改Docker或生产、未下单扣款。
- **2026-09-28 魔法配方独立原型**：用户要求开始做配方，本轮将已确认的工作3份＋大学3份落成独立本地静态原型product-samples/magic-recipes/site；提供分类搜索、材料表单、示例／方法、完整提示词生成、复制回退、TXT下载入口及双主题。通过构建从现有Markdown抽取提示词，嵌入既有狮冠M；无账号、模型调用、网络接口或持久化输入，无现有业务接缝。node build.mjs成功，node --test core.test.cjs ui.test.cjs共19项通过（DOM测试复用frontend已安装jsdom）；浏览器实测会议／论文生成、必填、大学筛选，1280px桌面与390px手机未见外层横向溢出，日志无错误，深色桌面／浅色手机截图存本线程visualizations的recipes-desktop-dark.png与recipes-mobile-light.png。系统剪贴板落地未由工具证实（虚拟剪贴板为空），TXT下载事件等待超时，保留用户真机验收；不说完整端到端已通过。生成index.html被本目录忽略，可直接打开；127.0.0.1:4178预览进程仍运行（本轮会话74036）。未改中转站、SynaRoute、Docker或生产，未提交推送、未付费调用或创建定时任务；保留并行海报及原有交接改动。最终栈与正式托管未定，本轮静态原型为可逆选择。
- **2026-09-28 大学配方定位修正（优先于此前学生扩展）**：用户明确学生系列只做大学，不做其他学段。04／05／06原文件已整体替换为大学论文精读、编程实验复盘、大学期末复习，版本0.2；README同步，不保留原小学示例作为现行样品。工作系列3份SHA256修改前后一致。Node检查六份章节／围栏、6个链接、大学定位及均值／代码表达式／时间等示例通过；没有执行Python或外部模型试跑，不冒称实测有效。未改业务、创建自动化、调用付费API、提交推送或部署。
- **2026-09-28 海报客服文案更新**：用户明确要求将「微信扫码」改为「客服咨询或相关名词」，双二维码日间／深色海报统一采用「客服咨询」，左侧说明同步为「访问官网 · 客服咨询」。输出仍在with-wechat目录，二维码图案、位置与其余设计不变；此文案决定优先于下方初版标注。不改网站、不发布。
- **2026-09-28 海报双二维码补充**：按用户要求将上传的微信二维码放在官网二维码右侧，仅标「微信扫码」，不推断身份。日间与深色新版均输出到本线程mofa-poster-20260928/with-wechat/{day,dark}/，各含1800×2400 PNG及1080×1440 JPG；原版保留，生成源为add_wechat_qr.py。原二维码字节副本一致、高清微信码区域与合成素材逐像素一致、y<1990区域与旧版一致（脚本断言）；已打开两版分享图目视检查并排标签及排版。尚未手机实扫或验证微信目标。仅更新会话海报及本交接，不改业务、配置、Docker，不提交推送或发布，保留其他任务工作区变更。
- **2026-09-28 魔法配方学生扩展**：用户认为首批内容有用并要求学生场景。新增04知识点理解、05错题复盘、06复习安排三份样品，具体示例为小学高年级数学；强调分轮提示、真实学生回答、来源／冲突、隐私及时间预算，不承诺提分。README更新工作／学习两组及实测门槛。Node检查六份各7项必要章节、围栏、6个本地链接通过；学生例子的平均数、分数及90分钟计划核算通过（分数使用浮点容差）。共24个设计样例，未做外部模型试跑或学生验证；未付费调用、自动发布、开发平台、提交推送或部署。保留其他任务的海报与交接记录。
- **2026-09-28 日间宣传海报**：按用户要求新增雾白底／深钛青M的日间版，保留夜间版。输出在本线程mofa-poster-20260928/day/（绝对父目录见下方海报初稿记录），含1800×2400 PNG、1080×1440 JPG及规格；生成源为同级上层build_day_poster.py。已目视检查、确认与深色版26组文案及坐标一致、两种尺寸二维码模块对比零差异。未手机实扫、未发布、未改业务/配置/Docker/其他任务文件；保留已有CUSTOMIZATIONS、HANDOFF及product-samples工作区变更。
- **2026-09-28 魔法配方首批样品（优先于实验室建议）**：用户选择将AI工作方法包装成模板，并同意先做3份样品。新增product-samples/magic-recipes/说明及会议转待办、需求转验收、报错转排查三份内容，含完整提示词、虚构输入、人工参考答案、边界样例、验收和修订提示词。Node文件检查确认三份各7项必要章节、围栏配对及3个目录链接有效；共12个设计样例，未做外部模型试跑，不能标已验证。未调用付费API、开发平台、创建定时任务、提交推送或部署；仅新增样品及更新交接／定制登记，保留原有HANDOFF修改。
- **2026-09-28 产品线咨询**：用户提出魔法家族母品牌、狮子＋皇冠M共同标识及两产品分工，询问第三款获客产品。建议魔法实验室（暂名）：模型实测对比与自愿分享入口；属于待验证假设，未确认立项。本轮仅补交接，未改业务、提交推送或部署。
- **2026-09-28 发布最终核验（优先于此前状态）**：功能9edc7a7bb及CI修正a1b91372e均已推送origin/main；ls-remote核实修正完整SHA a1b91372e55f23f4592ab1f95b9e043c57e4b498。本地全量前端lint、推广5测通过，此前124项关联回归通过。修正SHA的GHCR运行36375537015成功，日志确认sha-a1b9137及latest已发布，manifest digest 076905e07d5966cfc9c4f2cabdce4c1c0b3d8d4adee7cfed4368c78d0f2f10e2；安全扫描36375537029成功。CI36375537042前端、golangci-lint、shell和release-helpers成功，最后查询Go Unit tests仍in_progress、Integration tests pending，不能宣称全绿。建议待本次CI成功后，用户备份并在/sub2api-deploy执行 ./update.sh sha-a1b9137；生产未操作。
- **发布CI修正**：功能提交9edc7a7bb已推送并核实远端；镜像运行36375045453成功（sha-9edc7a7），安全扫描成功，Go lint成功，但前端CI因ProxyAdBanner空模板触发vue/valid-template-root失败。已改为render返回null的兼容组件并补推；升级应使用修正提交镜像，不能把首轮CI描述为全绿。
- **2026-09-28 包号与推广清理提交发布**：用户授权提交推送origin/main供手动升级；fetch确认提交前HEAD与origin/main一致。已用Go 1.27.0 Docker对本轮Go文件执行gofmt，124项前端定制关联回归全部通过，diff空白检查通过；此前本机Docker前端及Go embed构建通过。此次推送须核对对应SHA的CI／GHCR结果，不沿用旧CI；Go单测与lint以本次CI为准。生产不操作。
- **2026-09-28 本机Docker部署完成（优先于下方未部署记录）**：用户授权本机验证；启动原Docker Desktop，代理授权接口200，实际构建进程设置代理，前端构建／3项i18n／类型检查及Go embed编译成功。原项目仅替换应用容器，PostgreSQL和Redis复用，三服务healthy，health=ok；镜像与运行容器一致为d0c2c1d46026d0699ce3e153c4aba9c102e707275e84320990549512265bf5a8。数据库custom_dedicated_accounts表存在；HTTP入口index-B18iYC10.js引用两张包号页面，新首页及用量页资源不含上游仓库推广链接；两类包号接口未登录均401。地址http://127.0.0.1:8080，待用户真实登录／带数据验证，未执行Go单测、不制造调用费用。回退镜像sub2apicust:before-dedicated-20260928-112340，数据库备份output/brand/before-dedicated-20260928-112340.dump已pg_restore目录校验，构建日志output/brand/docker-dedicated-build.log。未提交推送、未操作生产。
- **2026-09-28 上游推广清理（源码，未部署）**：移除AppHeader管理员GitHub菜单及HomeView／KeyUsageView页脚仓库链接；ProxyAdBanner改为空兼容组件，禁用代理管理／创建账号／编辑账号三处代理购买广告。保留客服配置、OAuth、合规和运维帮助及LICENSE，不动包号未提交工作。新增5项推广回归测试和vue-tsc -b通过；未浏览器验收、未重建镜像、未提交推送或部署。审计边界见CUSTOMIZATIONS。
- **2026-09-28 Claude／ChatGPT-Codex包号首版（源码，未部署）**：用户确认首期平台后，新增管理员「包号管理」与用户「我的专属账号」、独立绑定迁移241、到期／撤销／续期、用户白名单响应及上游剩余百分比快照。首期仅标准模式，需先配置一账号／一标准专属分组／仅授权目标用户；不自动搬号、不改变余额计费。API Key原始身份、两平台选号／转发前及WS每轮均接入实时隔离检查；失败关闭，撤销保留记录不回公共池。额度只读被动快照，无模型探测或强制OAuth刷新；缺失、过期、重置后不冒充0%／100%。使用与回滚边界见DEDICATED_ACCOUNTS.md，所有上游接缝已登记CUSTOMIZATIONS及custom/UPGRADE。已执行119项前端定制关联回归、3项i18n检查、vue-tsc、定向ESLint、Vite生产构建；最终复测同样119项通过，构建退出0，完整构建日志在output/dedicated-frontend-build.log（Git忽略）；构建仅既有Browserslist／chunk提示。新增Go/HTTP单测已写但本机无Go，尚未gofmt／编译／执行后端测试；Docker仅只读探测，desktop-linux引擎管道不存在，未启动Docker或重建服务。数据库迁移、真实账号／双用户隔离、负载与浏览器视觉仍待，不宣称已上线或CI通过。未提交推送／修改生产，保留开工前HANDOFF未提交记录。
- **2026-09-28 包号需求初查**：用户希望单个上游账号指定给用户，并提供菜单查看状态及剩余额度。本轮仅源码初查，未实现功能。已找到User.CanBindGroup专属分组授权、Account.GroupIDs多分组归属、Group备用分组及管理员GetUsage的用量百分比与重置时间能力。专属分组不等于账号独占，本站消费不等于上游余额。待确认首期平台，再设计用户归属校验、脱敏展示及独占调度约束。未改Docker或生产，未提交推送，保留此前交接改动。
- **2026-09-28 宣传海报初稿**：用户要求制作海报，已沿用「魔法家族／多模型 API 服务」、核心M／狮冠与雾钛青，制作1800×2400 PNG及1080×1440 JPG。素材为Pictures/mofamily/mofa-mark-flat.png，保留轮廓后重着色排版；不重画M、不修改原图。输出位于C:/Users/Administrator/.codex/visualizations/2026/09/26/01a0dd78-fd99-7e82-bbcc-5d095ed76b1c/mofa-poster-20260928/，含build_poster.py与poster-spec.json可复现源。采用现有Logo合成与确定性文字排版，未调用AI生图API（内置工具不可用，未擅自调用需要密钥的替代接口）。文案不含价格、稳定率、延迟或官方授权等未经验证承诺；二维码编码为https://ai.mofamilys.com。已目视检查成品、验证尺寸/文字边界无重叠，两种输出的二维码模块采样与生成矩阵一致；尚未手机实扫或验证生产站点可用性。本轮仅海报及交接记录，无业务代码、Docker、配置、提交推送或生产变更，保留原有未提交交接内容。
- **2026-09-27 SynaRoute官网与客户端只读评估**：用户确认两者都看。定位真实工程D:/ccfile/SynaRoute（HEAD 317ce19，package 0.1.69，git status检查前后干净）；线上官网中文页桌面/窄屏及本地Vite纯mock客户端的Key列表、大脑聚合、设置日夜主题、用量页已查看，未启动真实代理、未读取真实密钥、未调用付费模型。已核实官网steps.s4.title/desc原样显示（site/src/data/features.ts:107与语言表缺项）；客户端深色说明色#52525B在实测卡片底色上对比度约2.30:1（src/styles.css:118），设置页9个switch缺可访问名称（ToggleRow.tsx:59）；Linux默认加密分支仍用编译期固定密钥（secret.rs:861、888），而官网Linux标available及统一DPAPI文案不匹配，需优先处理，但未验发布二进制或证明用户密钥泄露。官网卸载清理说明漏掉可能位于exe同级的logs（store.rs:290）。建议保留紫色，精简首屏/分层高级配置属设计判断。SynaRoute产品文件未改、未运行完整测试/构建、未提交推送部署；本轮仅本仓交接补记，不另建SynaRoute真机验收编号清单。
- **2026-09-27 SynaRoute调整咨询**：已读品牌计划及本站导入生成器与测试源码；既定方向保留紫色。未检查客户端工程或当前官网、未运行测试，不能断言产品缺陷。本轮仅补交接，保留原有未提交修改，未改业务或部署。
- **2026-09-27 安全初查**：用户反馈异常登录；静态审查本地c11990cb6密码登录/JWT/管理员鉴权/限流/IP解析/审计/会话绑定及部分定制外联，未在已读路径发现万能密码或无凭据管理员放行，不等于全仓无后门。auth_service.go:535验证密码；admin_auth.go:120存在独立管理员API Key入口，不经过密码/TOTP，泄露后仅改密码不能撤销；config.go:2083默认信任转发IP，结合pkg/ip/ip.go:55、rate_limiter.go:121，若源站直达或反代不清洗伪造头，可污染审计IP和分散限流桶（未证明生产满足条件）；auth_handler.go:264登录200可能只是进入2FA，需关联最终认证；session_binding.go:61的IP/UA不匹配不能单独证明入侵。git diff upstream/main...HEAD核对上述核心认证文件无fork差异，仅为本地已有基线。未取得生产日志/配置/镜像；本机Docker只读ps失败（引擎管道不存在），未编译测试、未改业务代码/配置、未部署/提交推送。在线公告查询未取得可引用结果，不断言无已知漏洞。
- **CI最终结果补充（优先于下方中途状态）**：GitHub API最后核验功能提交d592f9481的CI运行36264121775已completed/success；结合镜像36264121792和安全扫描36264121782，三项全部成功，集成测试不再待办。镜像标签仍为sha-d592f94。纯文档补记b631b82ae已收到git push成功回执，但其后ls-remote连续网络连接失败，不把失败查询描述为远端核验通过。
- **发布已推送／镜像已生成（2026-09-27，本轮最终记录）**：功能提交 `d592f9481c2006993b9825ef830766f81431cf18` 已推送origin/main，并用git ls-remote核实一致。GitHub Actions镜像运行36264121792成功，日志确认已推送 `ghcr.io/engineemomo/sub2apicust:sha-d592f94`（同时latest），manifest digest `sha256:627aebd01fbf7d403b66b45e3e533180d1269b87a5427d348ac9dc6c9e456735`，linux/amd64，版本 `0.2.8-custom.d592f9481`。对应安全扫描36264121782成功；CI 36264121775的前端、golangci-lint、shell、release-helpers和Go单测已通过，但最后查询集成测试仍运行，不能宣称CI全绿。该发布结果以随后纯文档提交补记，不另造镜像。生产未操作；建议待该CI完成后用户备份并执行 `cd /sub2api-deploy && ./update.sh sha-d592f94`。
- **提交前最终检查**：本轮76项定向测试、3项i18n完整性测试、定向ESLint、vue-tsc与Vite生产构建及git diff --check均通过。构建有既有Browserslist／chunk体积提示，未阻断；本机未运行Go单测，CI结果待本次推送后核验。
- **本轮发布收口（用户授权提交推送）**：提交范围为智力效率页面、可读效率对比、排序／两段说明精简、充值空态宽度及维护文档。已fetch确认提交前HEAD与origin/main一致；76项定向回归、3项i18n检查及定向ESLint通过。推送后须核验本次SHA对应CI与GHCR镜像，不能沿用旧版本成功结果。生产由用户按UPDATE_GUIDE备份并手动更新，代理不操作生产、不创建售卖套餐；本机8080尚未包含最新两段说明精简。
- **订单菜单开关取证**：用户新截图已显示展开的订单管理及订阅套餐。源码确认该组为上游原生，受adminSettingsStore.paymentEnabled控制（支付配置enabled），系统设置→支付设置→启用支付对应form.payment_enabled；简单模式隐藏，路由要求管理员且公开payment_enabled明确false时重定向。当前侧栏diff仅增加智力效率，不改订单开关。本轮未改配置，无法从截图确定此前不可见的实际原因。
- **说明区再次精简（最新）**：按用户指定只保留IQ可比性及低样本／缺失数据两段，中英文同步，移除标题和其余说明。intelligence-view.spec.ts 8项组件测试及git diff --check通过；本轮未重建Docker、未提交推送或部署生产，8080仍是上一版。
- **套餐管理入口与上游来源核验（本轮仅取证）**：用户指出订单管理下看不到套餐，所附截图当前选中的是「订阅管理」，截图中的「订单管理」仍折叠。不能据此断言用户展开后的界面一定正常。已用git show／blob哈希确认：本机upstream/main=`a3eb7ef302961cba716dc78b39b93b60c467db0e`，上游AppSidebar原本包含订单管理→`/admin/orders/plans`；AdminPaymentPlansView文件hash `93a5303a76218e6f34fec8f185fd5ccc17d7d35d`、payment_config_plans.go文件hash `095730c1e8a4350a77749687af72807a162d7be4` 均与当前工作树相同。实际按钮中文是「创建套餐」，之前答复称“新建”不够准确。运行8080的AdminPaymentPlansView-COR3nusR.js HTTP200，含payment.admin.createPlan动作；已请求Codex面板打开直达地址，工具仅返回queued，不代表用户已看见或实际点击已验证。本轮不改业务、不建套餐、不重建Docker，当前运行仍cfa0a12c8bf6。
- **当前本机部署：可读性修正版已运行8080**：应用镜像`cfa0a12c8bf6`，与本轮构建ID一致；三服务healthy，数据库／Redis容器ID未变。HTTP实取入口`index-BW6NgO3M.js`引用`IntelligenceView-C0f1Aymd.js`，新条形对比、矩阵排序控件、移除旧频率元素／外链以及支付容器width:100%规则均存在；/health、订阅页和套餐管理入口HTTP200。Docker内前端构建与Go embed编译通过，未跑Go单测；用户=1、分组=2、套餐=0以及280项设置摘要与更新前一致。本轮回退镜像`sub2apicust:before-intelligence-ux-20260927-022446`和583628字节dump已保留，部署证据`output/brand/intelligence-ux-deploy-record.json`，均Git忽略。真实浏览器连接仍不可用，视觉／交互验收仍待；未提交推送、未部署生产。
- **最新用户反馈：智力页面可读性与订阅释疑**：已移除查看源站按钮／JSON外链／刷新频率文字，保留来源归属和实际30分钟轮询；效率分布改为带模型名的分数／耗时／费用三列条形对比，默认high档、可筛选及排序、12条分页和手机堆叠；不合并环境，不造综合排名。矩阵默认源顺序，新增名称／指定档位分数降序，缺失最后、同分保留原序。支付空态窄卡对应定制flex正文与上游mx-auto缺宽度的适配，已仅在theme.css补width:100%，真实浏览器效果待用户验收。76项回归、ESLint和前端构建通过；浏览器工具仍因apikey认证失败不可用。部署结果以本轮后续记录为准。
- **订阅现场再次只读取证**：本机「测试分组1」确有记录（id=2，anthropic／subscription／active，日100、周500、月2000额度），但subscription_plans为0条。代码`ListPlansForSale`只读上架套餐，故当前空态与数据一致，不能将“只建分组不显示商品”称为上游bug。应到订单管理→订阅套餐创建商品、关联该分组、设置售价／周期并上架；未擅自定价、未创建套餐／订单、未改用户分组。
- **上轮初版Docker发布（历史）**：用户要求「重新发docker」后曾部署`1a7a31d0e424`，三服务healthy，构建与HTTP资源验证通过；证据见output/brand/intelligence-deploy-record.json，旧镜像和备份保留。该镜像已由上方可读性修正版替代，不再作为当前运行版本。
- **2026-09-27 智力效率新功能（本轮）**：用户提供 Codex Radar JSON，新增认证 `/intelligence` 和侧栏入口：雷达图标／模型品牌图标、分数条矩阵、耗时／费用散点图、搜索／环境／样本过滤、中英文、来源更新时间与本页获取时间、低样本／过期／失败提示。页面打开期间每30分钟拉取，支持手动刷新；不属于服务端常驻采集。真实源数据24模型／89组合解析通过；70项定制与订阅测试、ESLint及前端构建通过。浏览器检查工具返回 `unsupported Codex auth method: apikey`，真实浏览器视觉验收未完成，不以组件测试代替视觉验收。本轮未提交推送、未重建本机Docker、未部署生产，历史提交授权不自动扩展到本轮。
- **充值／订阅核验（本轮）**：原生 `PaymentView.vue` 已有两Tab与套餐购买／续费；本机公开设置订阅／支付开启且未禁用余额充值。本轮只读SQL `SELECT count(*), count(*) FILTER (WHERE for_sale) FROM subscription_plans` 返回0／0，尚未配置可售套餐；不自行套用参考截图的价格、额度或周期，未改数据库或支付业务。
- **本轮开发预览**：此前启动只监听127.0.0.1:3001的Vite开发服务，`/intelligence` 使用本机8080后端代理，需登录；曾向Codex浏览器面板请求打开（工具返回queued），不等同浏览器验证。开发会话22909；本轮Docker更新后应优先使用8080正式本机入口查看新页面。没有写入登录凭据或创建订单。
- **本轮用户收口决定（最高优先）**：用户表示「暂时先这样」，明确授权提交并推送代码、更新待办、复制新Logo。当前视觉与多模型名称按阶段性接受保留，不再作为必须立即重做的任务；下面关于“未确认／未提交”的条目是历史过程。生产仍未授权部署，CI结果必须按本轮提交单独核验。新Logo副本已校验，路径见第五节。
- **本机2026-09-27 00:58 最新品牌状态**：用户要求去除GPT限定并统一顶部Logo后，本机三项品牌配置已改为「魔法家族」／「多模型 API 服务」／扁平雾钛青M。镜像仍b4fa2d8e97a1，无需重建；只新增素材mofa-mark-flat.png。完整管理员设置前后仅这三项变化，其他277条原始设置事务核验未变，实际登录、侧栏、标题和双主题手机／桌面通过。数据库备份607430字节保留，操作风险与证据见BRAND_IMPLEMENTATION第17节。名称与视觉仍待用户认可，生产未修改。
- **2026-09-27 00:43 最新有效状态**：用户再次否决登录页排版后，已改为统一1120px双栏面板，缩小标题和M、表单左对齐、手机上下堆叠。只改认证展示容器与theme.css，业务页／认证逻辑／SynaRoute不改。镜像b4fa2d8e97a1已部署本机8080，三服务healthy；61项测试、构建、真实30组认证布局与实际登录通过。回退镜像和数据库备份已保留。详见BRAND_IMPLEMENTATION第16节；用户视觉确认仍待，上一版不是已认可设计。
- **2026-09-27 当前有效状态**：已针对上一轮视觉否决完成结构调整并重建本机Docker，不再只换色。认证顶部品牌栏＋大屏双栏；控制台路径顶栏＋独立正文标题／操作＋统一筛选表格分页面板；核心M、完整业务字段、默认深色及SynaRoute保持。61项测试、定向ESLint、前端构建通过；Docker内Go embed编译通过，未跑Go单测。镜像／容器一致30804c47efc0，三个服务healthy，入口127.0.0.1:8080。真实64组双主题／四宽布局检查通过；创建弹窗取消、列设置及移动菜单实测。证据和局限见 BRAND_IMPLEMENTATION 第15节，用户视觉确认及真实带数据验收仍待，不自称设计完全交付。旧镜像／数据库备份保留，未提交／推送／生产部署。
- 以下2026-09-26条目为历史过程，不代表当前运行状态；旧9c8c版本的用户否决仍保留作为改进依据。
- **用户最新视觉反馈：未通过验收**。用户提供2560×1317登录截图，指出与概念稿偏差过大、进入控制台后仍像原版换肤。已对照原稿mofa-aurora-system.html与源码：认证布局限制1160px内容区／440px表单，大屏留白明显；控制台主要新增类名和公共样式，尚未落地原稿的正文页标题、筛选与表格一体化布局。部署及功能冒烟成功不等于视觉达标，不得把当前版本表述为设计完整交付。保持雾钛青／核心M／上游业务契约／默认深色，SynaRoute不动；后续应先对齐登录及真实API密钥页的结构和比例，再扩展共用布局，不能再仅换色或只做独立品牌页。此次仅核查并记录反馈，未改产品代码或重新部署。
- **Docker网络故障防复发记录**：已补 deploy/LOCAL_DOCKER_RUNBOOK.md（直接原因／推断边界、代理预检、构建进程环境变量、失败即停、镜像一致性验收）。后续本机重建先读此文件，不盲目重复构建；本次记录操作未重启服务或修改代理。
- **2026-09-26 23:50 本机新版部署成功（以本条为准）**：用户授权代理调整并重启Docker后，给构建命令进程额外设置HTTP_PROXY/HTTPS_PROXY=http://127.0.0.1:7897，构建成功。镜像与运行容器ID一致，前缀9c8c1dfb87bc，三个容器均healthy；入口 http://127.0.0.1:8080/brand 与 /login。真实设置保留GPT中转站／M标志，健康接口200，正常登录成功。五页面双主题×360/1440共20组无横向溢出，业务API观察无>=400。仍为空库验收，未真实调用上游，不代表带数据全流程完成。
- 原数据卷和旧镜像回退标签保留；新增即时备份 output/brand/before-brand-final-20260926.dump（596221字节）。未提交、推送或部署生产。成功日志 output/brand/docker-brand-build-client-proxy.log，详见 BRAND_IMPLEMENTATION 第14节。此前“未部署／容器停止”描述均为历史状态。
- **2026-09-26 用户要求再次重试**：第三次本地Docker构建仍在 auth.docker.io/token 连接157.240.20.8:443时超时，无法取得 node:24-alpine 元数据。日志 output/brand/docker-brand-build-retry2.log；未进入编译、未启动容器、未改代理或DNS，新版仍未部署。下一步须先解决Docker Hub网络访问，单纯重试尚无改善。
- **2026-09-26 23:28 本机Docker部署尝试（最新状态）**：用户已重新授权本机Docker部署，取代此前暂不运行限制；不授权生产、提交或推送。核对 desktop-linux 本地命名管道与 sub2apicust-local 项目后，保留旧镜像标签 sub2apicust:before-brand-20260926，备份数据库到 output/brand/before-brand-20260926.dump（595798字节，Git忽略）。两次构建均在拉取基础镜像元数据时因 auth.docker.io:443 超时失败，新版未构建／未部署。构建未到编译步骤；未改网络配置、未改设置API、未删除数据卷。最终已停止本次临时启动的postgres/redis，三个本地容器均回到原停止状态。须恢复Docker Hub访问后再继续，日志 output/brand/docker-brand-build.log 与 docker-brand-build-retry.log。
- **2026-09-26 本轮收尾**：独立 /brand、认证品牌双栏和工作台展示钩子源码已落地；长站名手机溢出已修复。20项回归、定向ESLint、i18n 3项及前端构建通过；公开页面24组浏览器检查无溢出，详见 BRAND_IMPLEMENTATION 第12节。
- **部署状态以本条为准**：本轮没有运行Docker、提交、推送或部署；旧8080镜像不含本轮布局。前端503隔离预览不代表后端验收，带数据控制台／真实调用／用户视觉确认待继续。三个新接缝已登记，维护步骤见 frontend/src/custom/UPGRADE.md；不保证未来零冲突。
- **2026-09-26 用户最新限制：先别跑 Docker**。后续不启动 Docker Desktop、不执行容器启动/镜像构建/重建；仅继续代码及前端验证，恢复 Docker 操作须用户重新授权。中断前曾发出 Docker Desktop 后台启动命令；当时项目 compose up 因引擎未运行而失败，本轮尚未重建镜像。当前新增独立 /brand 页面与品牌展示组件，AuthLayout / AppLayout / AppHeader 仅加入展示接缝/样式钩子，视觉仍集中 theme.css；新增品牌/升级契约 9 项测试及 SynaRoute 11 项回归通过，前端 pnpm run build（含 i18n 3 项）通过。工作尚未完成：真实业务视觉复验、升级维护说明与完整变更登记待补；不能宣称本机 8080 已包含本轮改动。原有未提交内容保留，未提交/推送/生产部署。
- **2026-09-26 截图问题已修复并本地重建**：仅 theme.css 修复深色标题优先级、输入错误态红边与主按钮加载图标前景。容器实测标题 #A1D9CE 且 background-image=none；登录、密钥、用量、管理员页深浅主题及 360/768/1440px 检查无横向溢出。真实创建密钥弹窗可打开/取消；当前为空库，未验证带数据图表或上游请求。用户首次合规确认已不再出现，本轮未代确认。通过已有后台设置将本地实例名称/副标题设为 GPT中转站 / 魔法家族，并接入保留 M/狮冠的银色位图加炭青底板（非矢量重绘）。构建、验证证据、局限及一次已恢复的设置接口连带默认值变化见 BRAND_IMPLEMENTATION.md 第 11 节。未改 SynaRoute、业务结构或生产；未提交/推送。
- **2026-09-26 用户截图复核 / 完成口径修正**：用户质疑实际效果是否为改造版。直接读取本地容器登录页 computedStyle：primary-500=9 107 104，dark-950=22 33 36，主按钮底/字=161 217 206 / 23 59 56，证明基础主题已加载；但并非完整品牌改造成品。仍是默认 Sub2API + /logo.svg，核心 M 仅未改原资产，尚未接入本地实例。发现真实缺陷：登录 h1 颜色实际透明，仍用 #085E5B→#096B68 深色渐变，主题层预期浅青标题覆盖未生效，暗底标题偏暗。此前无后端截图未显示品牌区，漏检此项；不得再称主题验收完成。此次只核查和记录，未改业务代码/重建镜像。
- **2026-09-26 本地后端已就绪**：Docker 从当前源码完成前端 + Go embed 构建，sub2apicust-local 三容器 healthy，http://127.0.0.1:8080/health 和公开设置接口均 200；真实管理员登录成功。独立测试数据卷、仅本机端口、在线更新关闭，未碰生产。配置与随机凭据只在被忽略的 deploy/.env 和 deploy/docker-compose.override.yml；启动方法见 BRAND_IMPLEMENTATION.md 第 10 节。首次登录合规确认未代用户接受，业务页完整验收待用户自行阅读确认后继续。没有跑 Go 单测，没有提交/推送，SynaRoute 未改。
- **2026-09-26 阶段一实施**：产品代码只改 frontend/src/custom/theme.css，原有未提交文档保留并续写。雾白/炭青色阶、日夜成对主按钮、输入焦点和卡片已落地；不动核心 M、业务结构、默认深色、分类色或 SynaRoute。真实登录页双主题、360/768/1440px 无横向溢出；主按钮对比度 6.34:1 / 7.76:1。公开设置接口本地返回 500，品牌区未加载；密钥/用量/管理员页均跳登录，业务验收未完成。验证与截图见 BRAND_IMPLEMENTATION.md 第 9 节。未提交/推送/部署，未重新查询 CI。以下设计记录为历史。
- **最新定案 / 下个对话直接执行（2026-09-26）**：用户已确认「就按照你这个设计执行」，并要求先安排工作、换新对话实施。已新增 **BRAND_IMPLEMENTATION.md**，含最终稿绝对路径、色表、核心 M 保留原则、SynaRoute 保留紫色、阶段划分、代码证据、验证与维护边界。**不再重选风格，从阶段一真实业务页换肤开始。** 本轮仅修改交接文档与任务入口；主题、后端、默认深色策略均未改变，未构建/提交/推送/部署。下列「待确认」均为历史讨论，设计方向以本条和计划为准。
- **2026-09-26 品牌第四轮（保留核心 M，颜色可重做）**：用户明确允许更换 Logo 色彩，核心 M 设计必须保留，要求重新思考设计。新增线程预览 mofa-aurora-system.html「雾钛青」：以现有标志透明遮罩保留 M/狮冠轮廓，重着钛青/浅银绿；雾白日间与炭青夜间组成同一套主题；首页与控制台分屏预览，增大正文字号、去掉轨道/网格堆叠、减少装饰。已浏览器验证页面/深浅切换、搜索、空状态、1024px 桌面及 360px 视口外层无溢出，并修正搜索图标定位。仅概念验证，非真实 Vue 页面实装，非矢量重绘，舒适度/转化效果未做用户测试。产品默认深色和业务代码均未变。
- **2026-09-26 用户补充品牌验收要求**：设计必须同时考虑用户舒适度与吸引力，不能仅追求科技感。后续建议区分品牌首屏与高频操作区：首屏用克制的 B+ 品牌表达，控制台减少装饰、保证信息层级与可读性；深浅模式保留选择。舒适度和转化效果尚无用户测试证据，不能宣称已验证。当前默认深色策略不擅自更改，下一步应在真实页面检验，而非继续单纯追加概念稿。
- **2026-09-26 品牌第三轮（B 与 C 对比、增强科技感）**：用户询问 B/C 哪个更好并要求更有科技感。新增线程预览 mofa-tech-directions.html：B+「深蓝精密科技」与 C+「银白精密科技」，采用第二轮克制排版、细网格/轨道/局部冷光与可选接入路径示意，业务表格区不添加装饰。建议 B+ 作主视觉、C+ 作浅色配套，属于设计判断，尚未获用户确认。已检查浏览器方案切换、标志加载及桌面/360px 视口外层无横向溢出。未改产品主题、业务代码或生产环境。
- **2026-09-26 品牌第二轮（高级感 + fork 维护优先）**：用户认为 A/B 不够高级，强调能长期同步开源更新。新增线程预览 mofa-refined-directions.html：C「铂银纸白」、D「石墨冷银」，减少渐变/卡片/发光，控制台示意保留侧栏、筛选与表格结构。已通过浏览器检查外观、切换、搜索示例及 360px 视口外层无横向溢出（表格内部允许横滚）。代码取证：primary/dark 已接入 theme.css，但 AppLayout 仍含 bg-gray-50 与独立 bg-mesh-gradient 覆盖层；不能承诺仅改变量就完整复刻概念稿。建议品牌展示页与业务控制台分开：展示页独立新增，控制台只做受控主题覆盖，不影子替换业务页、不改后端。CSS 覆盖仍需上游升级后检查，不能承诺零维护。尚未实装、未部署。
- **2026-09-26 品牌视觉对照稿已制作（未实装）**：用户确认先看两套稿。在线程可视化目录保存 mofa-brand-directions.html，含 A「银白清透蓝」与 B「墨蓝冰银」，各展示品牌官网首屏和 API 密钥控制台，内容一致、示例数据明确标注。沿用用户提供的朋友设计标志，仅做去背景预览处理，未完成矢量重绘、未改原图。浏览器已检查两套外观、标志加载、方案切换及导入按钮提示；360px 视口下未发现横向溢出。未修改产品主题或业务代码、未部署。
- **2026-09-26 生态品牌讨论（方案待确认）**：用户提供 SynaRoute 官网及银色狮子/皇冠 M 标志参考，希望统筹生态品牌，允许调整色调。本次通过浏览器读取官网文字及首屏截图：当前首屏为浅底紫色强调，定位为本地 API 路由与多模型协同；仅验证页面展示，未验证软件实际行为。讨论方向为母品牌统一背书、SynaRoute 与 API 服务各自定位、银白与蓝色为主的视觉家族；尚未获用户确认，未修改主题或业务代码。
- **2026-09-26 Codex 接手核查（仅结构与待办）**：已读取交接文件、目录与依赖清单；核查开始时本地 main 与本地 origin/main 引用一致，工作树干净，HEAD 为 0d1e3e7b8，VERSION 为 0.2.8。本次未联网核验远端/CI、未运行构建、未访问生产环境；仅更新本交接文件，不改业务代码。下列 CI/部署历史记录不代表本次重新验证。
- 本地分支 `main`，HEAD=`0d1e3e7b8`，与本地 `origin/main` 引用相同；VERSION=0.2.8。**当前工作树有未提交交接文档改动，不是干净工作树**；本轮未 fetch，远端现状未重新验证。
- fork 接线：`origin`=https://github.com/EngineeMoMo/sub2apicust.git ；`upstream`=Wei-Shaw/sub2api（**push 已禁用**，只读）。
- 上游同步：**2026-09-26 已合并上游 0.2.8 快照（a3eb7ef30，207 提交）→ 零冲突、11 处关键定制全存活、VERSION 0.2.8**。合并后 `pnpm run build` 通过；后端 `api_contract_test.go` 新增的 CCS 夹具已镜像 SynaRoute（全后端 hide_ccs = hide_synaroute = 20）。
  - ⚠️ 该快照是**上次成功拉到的** 0.2.8；当日 github 从本机拉不动，快照之后若有更新的上游提交未纳入 —— 下次网络好时 `git fetch upstream` 增量再同步即可。
- 前端：`pnpm run build` 通过（typecheck + i18n + vite）。本机预览：`pnpm -C frontend run dev` → :3000（无后端时仅外观预览，数据页会跳登录）。
- ✅ **CI 已验证（2026-09-26）**：0.2.8 合并提交 `58b154f53` 的三条 workflow 全绿 —— Build custom image(GHCR) / CI(单测，含 `api_contract_test`) / Security Scan。即后端已编译通过、单测通过、**0.2.8-custom 镜像已推到 GHCR**。
- **部署/升级方式（2026-09-26 用户定案：Docker + `deploy/update.sh`）**：日常升级 = 部署目录跑 `./update.sh`（拉 GHCR 定制镜像 `ghcr.io/engineemomo/sub2apicust:latest` → 重建 → 启动自动迁移 → `/health` 自检 → 清旧镜像）；指定版/回滚 = `./update.sh sha-<提交>`（自动切 override 的 image 标签并备份）。首次从现有 0.2.7 平移见 `deploy/DEPLOY_CUSTOM.md` 第七节。**不用 App 内按钮**（原因见「五」）。

## 五、待办 / 下一步

- [ ] **2026-10-19 智力效率 v6 与可用性补强发布验收**：一并发布已复现的 v6 兼容、成功快照持久化、故障一分钟复查与 200 组合容量。用户已授权本轮提交／推送，待同 SHA CI／GHCR 核验，生产由用户更新；前后端同版，确认现有 `pricing.data_dir` 所在卷保留且可写（默认 `/app/data`），至少成功读取一次后核实 `intelligence-cache/last-success.json` 已落盘。登录核验 `/api/v1/intelligence` 200、历史数据提示／原读取时间、自动恢复及矩阵费用／耗时；生产故障或重启演练由用户授权后执行，不用模拟测试冒充现场。若仍失败，取实际镜像 SHA、接口响应和 `intelligence source refresh failed`／`intelligence snapshot load/save failed` 日志；不要用旧发布的 CI 代替本修复验证。

- [x] **2026-10-19 上游v0.2.17发布门禁**：业务5233e9090698a9bd02f645340cd19e85e7e27743已推送；同SHA CI/安全/GHCR全成功，固定sha-5233e90及registry/source/embed/隔离登录验收通过，见第四节及output/upstream-sync-20261019/verified-release.json。
- [ ] **2026-10-19 上游v0.2.17生产验收**：用户备份数据库、配置及app_data后，在服务器部署目录执行./update.sh sha-5233e90，检查真实登录、业务请求与支付/退款流程；生产未代部署。PR8028不在此版本，截图无图仍须实际请求取证。

- [x] **2026-10-18 连续登录400漏计源码修复与本地验证**：缺失必要人机凭证标记计数，含验证码字段名的格式错误不再误排除，覆盖三服务商及故障/正常挑战/关闭规则；详见第四节及deploy/AUTH_IP_BLACKLIST.md。
- [x] **2026-10-18 连续登录400修复发布**：业务117e9a1a08f1a225dbe5b041ae243be29251770b已推origin/main；同SHA CI／安全／GHCR成功，固定sha-117e9a1及摘要、来源、embed均已核验，见第四节与output/auth-login-probe-20261018/verified-release.json。
- [ ] **2026-10-18 连续登录400生产验收**：用户备份后手动`./update.sh sha-117e9a1`，确认IP黑名单自动开关／阈值／范围和真实客户端IP；截图5次不能证明达到默认10次，历史请求不回灌。未实际封禁189.24.123.4、不代部署；需立即处理历史IP可由管理员按现有手动名单功能操作。
- [x] **独立订单核账入口**：新增“订单管理 → 订单核账”，人工处理迁出设置；限额与自动清理配置保留设置→支付，权限及操作回归已覆盖，详见第四节。
- [x] **核账菜单发布门禁**：业务提交57aad622fa4cdeea90df1d22866b6b514a92e94b已推送origin/main；同SHA CI39135534896／安全39135534923／GHCR39135534858全部成功，固定sha-57aad62及registry来源已核验，详见第四节。
- [ ] **核账菜单真实验收**：用户备份后执行`./update.sh sha-57aad62`，验收“订单管理 → 订单核账”、手机操作与真实商户核账；生产由用户更新，未代部署或处理真实订单，其他任务工作保留。

- [x] **本轮支付补丁发布门禁**：业务提交16ef9172cf666e6361f17161aa83a53178ee0f3f已推送origin/main，同SHA CI39029645317／安全39029645211／GHCR39029645261全部通过，固定sha-16ef917和registry来源已验真，详见第四节发布终态。
- [x] **积压占用自动清理与一次批量复查**：2026-10-17已补默认关闭的历史通道缺失超期清理策略及管理员批量入口；后端104个叶子用例0失败0跳过、前端9项及类型／lint／构建、最终embed编译通过，详见第四节最新支付状态和 `output/payment-auto-release-20261017/verification.json`。
- [ ] **支付补丁发布和旧单现场处理**：代码与固定镜像sha-16ef917已发布；由用户备份后在实际部署目录运行`./update.sh sha-16ef917`，更新后保存开启缺通道自动清理，可点一次立即批量复查。现场核对实际等待期、审计和占用下降，已收未入账仍须商户核账；订单47及截图旧单尚未代处理。
- [x] **支付通道删除保护修复与回归**：2026-10-17用户要求修复后已完成源码，原有订单状态保护之外新增未结清收款占用保护；45项服务叶子用例＋19项真实PG叶子用例、Go1.27.2 embed编译通过，详见最新第四节和 `output/payment-provider-guard-20261017/verification.json`。未提交推送、发布镜像或部署。
- [x] **支付通道保护补丁发布**：删除保护与自动清理已一并进入16ef9172c；以本轮同SHA三项门禁及固定镜像为准，不借用旧sha-0e99748结果。生产部署和实际订单处理仍待用户执行。
- [x] **订单47通道实例3缺失的源码定位**：2026-10-17已核对完整报错链、删除保护范围、30分钟复查退避及现成人工释放入口，详情见第四节；仅证明代码触发条件，不将本地源码取证当生产数据库查询。
- [ ] **订单47现场核实与处理**：在报错服务器只读核对 `payment_orders.id=47` 的状态／provider_instance_id／provider_snapshot、`payment_provider_instances.id=3` 是否存在，以及 `custom_payment_collection_ledger.order_id=47` 的 paid_at／released／release_checked_at；结合管理审计／备份确认删除历史和商户账单核实收款。最新代码已支持管理员开启后自动分批清理缺通道的超期占用，关闭时仍走既有人工释放；不重建已删实例、不换绑历史订单，当前尚未代处理生产。删除保护无新增数据库级约束，不能将源码测试等同生产已处理。

- [x] **本轮新镜像发布（用户已授权）**：最终业务0e99748ba已推送；同SHA CI38750224990／安全38750225143／GHCR38750224942全部成功，固定sha-0e99748的revision、source、linux/amd64及digest和embed构建日志交叉验真。完整证据output/group-access-fix-20261015/verified-release.json；未代部署生产。
- [x] **分组到期持久平移后删除源码与真实PG专项**：原实现只在请求中切组的缺口已补齐；新增258及后台到期任务、旧认证快照兜底、目标自指兜底清理、事务回滚与并发保护。福利时段结束仍只恢复基础倍率。订阅／包号权益不自动合并；用户收到的最终验证以output/group-access-fix-20261015/verification.json为准。
- [ ] **到期平移现场验收**：用户更新sha-0e99748前审阅已有到期规则并备份，验证测试组到期后从列表移除、旧Key持久改绑／授权到指定组、历史账单及余额不变；同SHA门禁及新镜像已核验，生产真实收费仍未知。
- [x] **分组规则R1–R4源码补修与专项验证**：统一既有降级目标资格及上下文、拆分通知心跳、保留／缓存RPM覆盖、已付费图片GET不走新金额门槛；正式回归含真实PG通知积压、断线恢复及Stop通过，管理列表当前活动倍率也已补修。已随sha-0e99748发布，详情deploy/GROUP_ACCESS_REVIEW.md及output/group-access-fix-20261015/verified-release.json；旧sha-7944797不含补修。
- [ ] **分组补修真实收费验收（发布已完成）**：本轮sha-0e99748同SHA CI／安全／GHCR及镜像身份全部通过；用户备份后自行同版部署。验收活动期间列表0.8x及基础说明、到点恢复、原Key新请求账单、真实目标订阅／专属权限／退款；20:30历史账单仍未知，不自动补偿或调整真实倍率。

- [x] **用户效率要求专项复核**：35项缓存／活动场景在9组复测共315次通过，100并发同项冷请求仅一次回源；9组热缓存基准规则／月充值新增回源0/op，约0.9–1.5微秒。证据output/group-access-performance-20261015/verification.json；仅本地桩仓库基准，生产负载／P95仍待实际环境验收，业务源码和sha-7944797镜像不变。

- [x] **分组缓存及福利发布门禁（用户已授权）**：最终794479760已推送；同SHA CI38727724771（unit／integration、专用PG18、Go lint 0、前端52文件815项）／安全38727724898／GHCR38727724764全部成功，固定sha-7944797的revision、source、平台与digest和embed日志交叉验真；证据verified-release.json，未代部署。

- [x] **分组缓存与福利最终全量验证**：最终69项后端／50项前端、Linux58包11955顶层unit（含子项23316通过／36跳过）及embed退出0；真实PG双实例另实际通过，十轮专项通过。完整证据output/group-access-cache-20261015/verification.json，临时PG已清理。
- [ ] **分组缓存及活动部署后验收**：用户已要求发布用于更新测试；本轮CI／镜像门禁通过后，由用户备份并同版升级最终257；验证通知传播延迟与故障回源、真实退款撤销资格、活动日期／18:00／20:30／23:00及目标权限／模型／账单、批量生图快照和WS重连。生产负载与P95尚未压测，不以隔离微基准代替；不自动配置真实活动或部署生产。

- [x] **分组开放规则性能优化源码与专项**：规则／月充值及资格元数据缓存、事务提交后的跨实例通知、余额事件独立更新、并发回源合并与旧值防回写、月初／到期边界和断线回源已实现。真实PG双实例与十轮专项通过，热请求policy／monthly新增回源0/op；保留独立资格缓存，不修改计费预占缓存。生产吞吐／P95未压测，不能用隔离微基准代替。
- [x] **不定期福利时段源码与专项**：最多32段一次性北京时间活动及独立开关，三段示例0.1→0.15→0.1和23点到期兜底；普通请求、用户默认倍率缓存、展示与批量生图快照接入。69项后端／50项前端专项通过，未配置真实业务活动或部署。

- [x] **分组月充值／余额门槛与到期切换源码及本地验证**：两个门槛独立开关、同时开启AND，退款后既有Key重查；分组统一到期及目标自动路由，目标权限／门槛／计费生效。40项后端专项含PG18实际通过，Linux58包unit通过，前端42项及构建／lint／Linux embed编译通过；证据见第四节与verification.json，未提交发布。
- [ ] **分组开放规则真实业务验收**：用户已授权发布，门禁见上项；用户备份后部署同版前后端及257迁移，核验管理员保存／独立关闭／退款失去资格／到期原Key自动使用目标及实际账单；确认目标已授权且支持原模型，验证客户端WS重连行为。不能把本地数据库状态测试说成真实商户退款成功，不能用旧镜像或旧CI代替本轮结果；不自动关闭退款或部署生产。

- [x] **充值页签开关位置答疑**：已确认「系统设置 → 功能开关 → 站点类型 → 购买方式」，选「仅订阅」隐藏充值页签，选「充值 & 订阅」恢复；未代改配置。

- [x] **更新后Token差异现场核验**：已在用户登录的上游日志将08:52:49截止改为“今天”并搜索；10:38—10:39新批四条输入/输出与本站同时间记录全部一致，缓存与缩写显示一致。旧筛选导致截图右侧未展示新批；不再以截图箭头或未验实的BillingUsage路径认定Token有误。证据见第四节最新条目。
- [x] **搜索成功结果计数下限补修发布核验**：2847434f96400713e81c72146d0202223a7c84c5已推送；本机25项叶子专项、58包11935顶层unit、gofmt/embed，以及同SHA CI38677336658/安全38677336554/GHCR38677336670全部通过；固定sha-2847434及manifest/revision/source已验真，完整摘要见第四节与result-floor-verified-release.json。生产仍待用户手动更新与新请求工具费验收。
- [ ] **10:38:47同笔搜索费现场验收**：上游明确127/587、搜索一次、附加费0.000800、总0.001534；本站0.001378仍为纯Token费。源码已补终态0/较小计量不能覆盖成功结果的边界，仍需核实生产运行版本、请求当时搜索基础价及实际计量格式；用户手动部署后用新请求看web_search次数、基础单价和实扣快照，不要求重复费用截图，不凭截图推断最终usage，不改Token数量补差。
- [x] **初始0搜索漏计修复发布收口**：fb5916a4fe1853995b18f016d0806a2990a4c73b已推送origin/main；28项专项、全量unit（58包/11931顶层通过，34跳过）、gofmt/embed与同SHA CI/安全/GHCR全部成功，固定镜像sha-fb5916a及摘要/revision/source已核验。生产验收仍见下方待办，旧账单不补扣。
- [ ] **0.05搜索价未计入现场定位**：双边费用明细已齐，上游明确web_search×1/实扣0.0008，本站0.00124740仅Token，不重复索取相同费用截图。初始0抑制成功结果的源码漏计已修复并发布sha-fb5916a；现场仍需线上backend版本、该笔实际分组价及脱敏原始JSON/SSE，确认是否同一边界或存在计量未传/其他格式问题。0.05与0.15生效且搜索1次应加0.0075、总0.00874740，不按聊天文字/链接数量收费。
- [x] **Claude工具附加费代码发布**：d3a2c346352e7eca5aacc197785fd366f90f83b6已推送origin/main；同SHA CI38589112494/安全38589112074/GHCR38589112344全部成功，固定ghcr.io/engineemomo/sub2apicust:sha-d3a2c34及摘要/revision/source已验真，交接skip-ci仅更新文档。
- [x] **Claude网页搜索工具附加费本地实现与验证**：实际计数/去重/错误免费、分组单价、Token叠加、双原生网关、工具JSON快照及导出、账号统计、迁移256、真实PG余额/订阅幂等和前后端回归已完成；配置与覆盖边界见deploy/CLAUDE_TOOL_BILLING.md；代码发布与镜像核验已完成，生产未部署。
- [ ] **Claude工具费生产现场对账与部署验收**：完整双边收费明细已取得，现仍缺生产原始usage/SSE、实际版本/分组价和充值价值；用户备份后升级固定sha-fb5916a（含初始0修复），确认前后端同版和既有256迁移。在新请求中核验web_search调用快照、基础单价、实扣工具费及上下游同笔账单。生产由用户手动部署；上游删除所有计量/结果或其他工具格式须取得实际响应另做，不承诺所有价差已消失。

- [x] **高售价倍率仍少扣的双账单定位**：206笔全部匹配，23笔Sonnet合计负差0.015886，根因为本站基础价低、采购等价倍率约0.42超过售价0.4；全批净正差5.906735（仅额度1:1）。已交付逐笔对账及三种模拟，证据见第四节最新条目。

- [ ] **New API 实际资金与配置验收**：仍需双方实付/到账比（含赠金/折扣）及期间7笔未匹配上游2.385628的归属，才能判断完整真实现金利润。取得上游域名后实测/api/pricing、完整billing_expr、本人有效倍率和5m/1h价格；用户应用建议Sonnet逐模型价3/15/0.3/3.75、有效倍率0.4后，核验新请求生效价卡/用户专属倍率覆盖。需自定义阶梯时配置渠道区间并排除分组逐模型覆盖；本站开关不会自动同步上游。当前未授权更改实际业务费率/配置，未接入自动同步。

- [x] **2026-10-14 全工作区提交发布**：全部59项已提交推送6406511c6，同SHA CI/安全/GHCR全绿，正式镜像摘要及源码标签验真完成；交接以纯文档skip-ci提交，不生成新镜像。忽略配置、凭据、备份与构建输出不入库，证据见第四节最新记录。
- [ ] **全工作区版本生产验收**：本节最新统一更新目标为`./update.sh sha-6406511`，包含上游0.2.16、IP网段防护及原59项改动，替代下方历史候选镜像。用户先备份数据库/配置/app_data（含studio-submissions），前后端同版更新；核验248/249/250至255迁移、收款待确认与人工释放、限量套餐待付限制、投稿撤回/退回/清理、登录后的智力效率和真实IP。支付宝小额实付及过期入口、真实模型/计费/共享网段误拦率仍待，生产由用户部署；本机8080未更新。

- [x] **2026-10-14 上游0.2.16同步发布**：fe6323ca6已推origin/main，同SHA CI/安全/GHCR全绿，registry/revision/源码树/二进制及实际运行镜像ID已验真，固定sha-fe6323c。前端774项、Ent无差异/Go embed/专项、源码61条及正式镜像25条HTTP通过；原main已快进，59个其他会话改动已备份恢复且未混入发布，独立测试资源清理。
- [ ] **2026-10-14 上游0.2.16生产验收**：用户备份数据库/配置/app_data后执行`./update.sh sha-fe6323c`；核验13平台配额、渠道监控/模型目录、Gemini缓存等新增上游功能与原包号/支付/家族入口。核验实际客户端IP后选择网段/前缀/期限并开启自动；生产未代部署，真实模型/费用及共享出口误拦率仍待授权环境验证。本机8080尚未更新0.2.16，原容器当前停止。

- [x] **2026-10-14 网段防护发布**：功能bbee35dbcc30525bae354dc1d270f90b5591fab7/sha-bbee35d及同SHA CI/安全/GHCR/镜像摘要验真完成，默认永久、仅管理员调整/解封；11:56本机8080三服务healthy及独立45条HTTP断言通过，测试环境已清理。13:08本轮只读复查Docker未运行，当前8080状态未复测；生产未代部署。
- [ ] **2026-10-14 网段防护生产验收**：已由上游0.2.16同步镜像包含，统一使用本节最新sha-fe6323c；用户备份数据库/配置/app_data后`./update.sh sha-fe6323c`；核验Cloudflare→Nginx真实IP，刷新IP黑名单页并选择范围/前缀/期限，再开启自动。共享网段正常用户可能受限，误拦率待真实流量确认；不自动加入截图历史网段。

- [x] **2026-10-14 管理员手动封禁发布**：065eee935独立提交推送，新SHA CI/安全/GHCR成功且registry/标签/摘要交叉验真，固定镜像sha-065eee9；本机8080备份后更新，三服务健康，原业务计数/防护配置及PG/Redis/挂载保留。证据output/auth-ip-manual-20261014，其他会话未提交改动保留。
- [ ] **2026-10-14 手动黑名单生产验收**：用户备份后`./update.sh sha-065eee9`，刷新后台IP黑名单使用“手动新增 IP”；核对生产真实客户端IP、代理头清洗/源站限制、已知恶意IP证据及共享出口影响。本机已测永久/临时、重启/Redis清空、到期/解封/旧版本；生产是否已更新未知，未自动回灌截图IP。253后不直接用旧程序处理曾手动封禁的IP，见AUTH_IP_BLACKLIST回退边界。

- [x] **2026-10-14 IP防护发布收口**：07bcef204独立提交并推送origin/main；同SHA CI（含真实PG/Redis、unit／integration／lint）、安全、GHCR全部成功，固定镜像sha-07bcef2及摘要交叉核验，更新命令和备份要求已记录；其他会话改动保留，生产未代部署。
- [x] **IP防护可选永久／临时封禁、代码审查和本地Docker运行**：失败计数、防误拦分类、管理员解封、可配置期限及永久模式、迁移252完成；13项专项／全量unit／55项前端／实际镜像重启和Redis清空／原管理员接口及原数据保留通过，详见第四节最新条目。
- [ ] **IP防护生产开启前验证**：8080目前关闭自动拉黑、默认timed15分钟。用户登录IP黑名单页确认封禁方式和期限／解封交互；按审计观察共享出口失败分布，再决定阈值、永久或临时模式及开启。保留成功操作不计数；永久模式须人工解封，共享出口连续失败仍可能影响正常人。新SHA CI／安全／GHCR及Cloudflare→Nginx真实IP链路待验证，生产由用户部署。

- [x] **IP 黑名单实现与本机验证**：管理员菜单、版本解封、120分钟滑动计数及迁移250完成；9项后端专项、真实PG/Redis、全量unit、嵌入构建、前端511项关键回归通过，证据见第四节。
- [ ] **IP 黑名单生产部署/开启**：本功能已独立发布sha-07bcef2，248/249及其他会话内容未带入。用户备份后同版更新，核对250/251/252和管理员菜单。默认关闭，核验Cloudflare→Nginx真实IP、转发头清洗及源站限制，观察共享出口失败后选择永久或临时并开启。历史审计不回灌名单；验证码人机验证、邮件预算、forgot-password 500取证及真实记录浏览器交互仍待。

- [ ] **2026-10-13认证防护与500取证**：确认用户希望落地的防护范围后，先启用现有人机验证并实测登录/发验证码/注册流程，管理员须同时开启系统TOTP和个人绑定；有恶意持续请求时可对截图IP临时封禁并设到期，但先核对真实客户端IP。按生产请求量设计认证共享IP限流、账号维度失败退避（避免永久锁号）、入队前原子邮箱冷却和全局邮件预算；边缘规则仅覆盖指定认证入口，真实IP信任与源站封闭须核对代理拓扑再调整。取得截图forgot-password 500及send-verify-code 200的脱敏详情/request_id与对应应用/队列日志，查更宽时间窗的成功登录和管理员变更。公开password_reset_enabled=false不解释500；不直接开启密码重置或凭截图断言已入侵。本轮仅建议，未获新安全功能实施/生产变更指令。

- [x] **官网动效发布核验**：f8928e20d 同 SHA 的 CI（含真实 PG 的 unit/integration、lint、前端）、安全扫描、GHCR 全部成功，固定镜像 sha-f8928e2，详见第四节首条。
- [ ] **官网动效生产更新与验收**：用户备份后执行 `./update.sh sha-f8928e2`。确认首页、套餐、教程、登录页和手机菜单的视觉，并补测 Safari、真机触屏和系统“减少动态效果”。这一版没有迁移。生产未代部署。
- [ ] **智力效率雷达Bench生产验收**：
  - 本机 8080 已部署，用户 2026-10-13 已确认看到数据（后端首个请求 200，用时 10.9 秒）。以后如果再显示“数据源暂时不可用”，先在容器内测出网。回退方式：用标签 `sub2apicust:before-intelligence-20261013-090718` 重新打回 `local-theme` 并 `up --no-deps --no-build`。本轮没有迁移，不需要恢复数据库。
  - 用户2026-10-14已明确授权整个工作区提交，包含本功能；用户视觉确认及生产矩阵、对比、“数据不足”等标注和说明文字验收仍待。
  - 已随6406511c6发布，同SHA CI（含golangci-lint）、安全扫描和GHCR全绿，统一固定镜像sha-6406511；生产视觉与真实数据验收仍待。
  - 前后端必须同版发布。更新后在生产登录打开 `/intelligence`，首次读取最多约 45 秒。
  - 如果显示“数据源暂时不可用”，先在服务器上用 `curl -sI https://codex-reset-radar.pages.dev/data/radar-bench-binding.json` 确认能否出网，必要时配置 `update.proxy_url`。
- [ ] **生产 Nginx 压缩（用户在服务器操作，最高收益）**：curl 实测生产 JS/CSS 未压缩。在站点的 server 块加入 `gzip on; gzip_vary on; gzip_comp_level 5; gzip_min_length 1024; gzip_types text/plain text/css text/javascript application/javascript application/json application/xml image/svg+xml;`。
  - 必须包含 `text/javascript`：Go 后端发 JS 用的是 `text/javascript; charset=utf-8`。deploy/EDGE_SECURITY.md:138（上游文档）只列了 application/javascript，照抄的话 JS 仍不压缩。本行覆盖此前的列表。
  - 不要列 text/html（HTML 本来就压缩，写了 nginx -t 会报 duplicate 告警），不要用 `*`，也不要包含 text/event-stream。
  - 粘贴前先用 `nginx -T | grep -n gzip` 查同一 server 块里有没有已有的 gzip 行，同一块写两遍会报 emerg、配置无法加载。
  - 用 `nginx -t` 检查后 reload，再用 GET 请求页面实际引用的 css/js，确认响应带 `Content-Encoding: gzip`。
  - 本地 nginx 1.28.3 容器已复现并验证：只写 gzip on 时仅压缩 HTML，与生产现象一致；用文档列表时 JS 不压缩；用上面的列表时 HTML/CSS/JS 都压缩、SSE 不压缩。证据在 output/nginx-gzip-test/。
  - 5 级压缩下 CSS 从 377KB 降到 54KB，vendor-misc JS 从 239KB 降到 83KB。
- [ ] **可选：站点 Logo 改用图片地址**：后台 site_logo 现在是 base64，每次打开页面 HTML 都要带两份，共 43.5KB，而 HTML 是 no-cache。改成可缓存的图片 URL 前，先核实后端 safeImageURL 接受的地址形式和托管位置（我只核对了前端校验接受相对路径和 http(s)）。
- [ ] HTTP/2上游接口迁移：x/net0.60标记旧配置/GOAWAY类型弃用，目前保留五个既有文件中明确符号的SA1019兼容规则。后续验证H2C参数、流式PING和第三方错误匹配后迁移并删除该规则；不混同安全漏洞豁免。

- [x] 本次发布安全门禁完成：026d9ab2e的CI、安全、GHCR全部成功，含独立PG与integration；固定镜像sha-026d9ab。本地隔离66项、schema回归、embed及lint通过，govulncheck零可达漏洞。

- [x] 包号到期/跨组免扣源码及认证链专项验证通过；真实PG及生产验收仍归入职责收敛发布待办。

- [ ] 用户备份并更新sha-026d9ab，实测长请求、有效包号免扣、到期/撤销拒绝、普通组余额及原生限流；CI不能替代生产验收。

- [x] **2026-10-11 审查问题修复源码**：用户要求“全部修复”。已完成收款预占（按用户封顶、到期复查、人工释放）、工坊存储与匿名限流、库存待付、update.sh、前端工具链及审计例外；验证结果见第四节最新记录。
- [x] **真实 PostgreSQL 回归**：已在用户同意后启动 Docker，用隔离容器（postgres:18-alpine）运行，12个 PG 专项和79个子项全部通过，0 跳过，结果见第四节。完整 integration 本机缺 testcontainers 模块缓存（GOPROXY=off），没有运行，以提交后 CI 为准。
- [x] **本轮CI确认**：6406511c6的Node20类型/关键前端49文件790项、Go unit/integration/专用PG18/Redis8、lint、安全及GHCR均成功；macOS更新脚本26组回归通过。vite6/vitest4/typescript-eslint8/eslint-plugin-vue10及冻结锁文件安装已通过CI，本机Node22.9全量3279项另验；mawk单独专项未实测。
- [ ] **支付宝真机**：启用限额后，手机 WAP 新增 time_expire（格式与电脑扫码相同）。须小额实付确认能正常付款、过期后无法付款，且预占在期限后约10分钟释放；电脑网站支付和易支付按等待期释放。生产启用前先在后台查看“待确认占用”。
- [ ] **剩余低危**：tailwind 3 带来的 braces（上游无修复）和 postcss-selector-parser 6.x 中危只在开发构建中，去除需要升级 tailwind 4。多账号仍可合计占用收款预占和限量库存，这是预占制度本身的限制；可开启取消限流、缩短订单超时。
- [ ] **鉴权热路径DB依赖（P2）**：评估给包号判定加短TTL缓存并在变更时失效，或只在分组确实是包号组时查询；需要压测和故障注入验证。
- [ ] **页面低成本修复**：在theme.css覆盖.btn-alipay对比度，给加载层补深色，把工坊硬编码颜色改成令牌；调整充值页提示顺序并补币种；改PlanContactPanel文案；决定隐藏EN，或给custom页补i18n。
- [ ] **文档精简**：CLAUDE.md只保留规则和指向HANDOFF的指针；本文件第四节只保留当前状态，历史另行归档；更正“本机无Go”；勾掉已完成的待办。

- [x] 收款限制本轮功能已提交推送，实际功能SHA 575467a5f的CI／安全／GHCR全部通过，固定镜像sha-575467a；以下同轮“待提交／CI”条目为历史步骤，已完成。
- [ ] 用户备份并更新生产到sha-575467a，核对历史收款、启用配置，再验人工开通／库存及真实商户关单付款；生产未代部署。

- [ ] 用户在本机8080验收收款配置、人工开通与库存交互；本机已更新，新SHA远程发布仍待，真实商户待验。

- [ ] 发布时包含本轮两项审查修复；真机验支付宝交易不存在时仍保留额度、确认关单后释放，以及关闭在线支付后的人工续费联系入口。

- [ ] 本轮收款限制源码待提交／推送及新SHA CI、安全、镜像验证；发布后备份再更新，按 `deploy/PAYMENT_COLLECTION_LIMITS.md` 核对历史付款链接／外部收款，配置单笔50（严格小于则49.99）与每日1000后手动启用；需要人工开的套餐改为contact_admin，真实商户关单与小额付款由用户验收。

- [x] **本轮库存与帮助发布**：用户授权提交推送已完成，最终70e66cf73的CI／安全／镜像全部成功，固定标签sha-70e66cf。
- [ ] **用户生产更新与验收**：备份后./update.sh sha-70e66cf；确认管理员分配扣库存、售罄不可购买、历史人工归属核对、教程分类及生图开通说明。生产由用户操作，不把CI通过当实付或客户端接入成功。

- [x] **帮助入口与教程源码**：问号／书本区分图标、六类场景教程、FAQ分类与详细步骤，已完成本地组件及浏览器夹具验证。
- [ ] **帮助页生产验收**：已随sha-70e66cf发布；用户更新后确认真实控制台图标、教程可读性，并按自己的客户端版本验证实际接入。不能把公开夹具预览称为生产验收。

- [x] **套餐库存源码与隔离验证**：售卖限额／剩余／售罄购买与人工分配校验、订阅页库存、事务防超卖完成，测试证据见第四节和deploy/SUBSCRIPTION_PLAN_STOCK.md。
- [ ] **套餐库存生产验收**：已随sha-70e66cf发布且同SHA三套检查通过；用户备份后同版升级（245／246迁移）。核对历史购买及人工分配，特别是无法自动归属的多套餐／混合来源记录，按“已占用＋新增售卖量”设限额；验新购／续费／管理员单次与批量／默认赠送或兑换／售罄／未付取消释放。真实延迟付款和商户退款待验，不能宣称自动退款。未修改业务套餐、创建真实订单或代部署。

- [x] **版本仅管理员展示源码与安全抽查**：组件及两个公开版本出口已修，定向回归、类型、lint及前后端构建通过，详见第四节与安全报告。
- [x] **本轮版本权限与依赖修复发布**：功能07f0ee69e已推送，同SHA CI／安全／GHCR全部通过，固定sha-07f0ee6，见第四节。
- [ ] **本轮版本权限真实账号验收**：用户备份更新sha-07f0ee6后，以管理员／普通用户／游客核实版本显隐，公开JSON／首页注入version为空，图标与公告正常；未代部署生产。
- [x] **17条生产依赖告警处置**：删除未使用图标依赖链、DOMPurify升级3.4.16，生产审计全0；全量2928项、类型／构建与相关lint通过。修复前是否可在本站利用仍未证实；源码尚未发布，详见deploy/SECURITY_REVIEW_20261007.md。

- [x] **订阅开通周期与支付宝直码发布**：80c30c825已推origin/main，同SHA CI／安全／镜像全通过，固定sha-80c30c8，见第四节。
- [ ] **本轮真机与订阅迁移验收**：用户备份更新sha-80c30c8后新建订单，核实扫码直达、小额实付及电脑自动成功／仅入账一次；检查订阅按开通7／30天重置、迁移保留用量及到期时间。未代用户部署，未知商户与App实际接受情况。

- [x] **上游v0.2.13本地合并**：用户已授权，完成快照备份、文本／包号预占兼容、Linux全量unit、真实PG及迁移、前端2921项、类型／源码lint／构建和隔离登录冒烟；详见第四节。
- [x] **v0.2.13代码与镜像发布门禁**：4bc5abbb5已推送，同SHA CI／安全／GHCR全部成功，固定镜像sha-4bc5abb；真实业务验收仍待用户更新后完成。

- [x] **订阅改动二次代码审查**：限定本线程改动，未发现新的明确缺陷；32项含临时竞态／卸载验证通过，真实账号与发布验收仍待。

- 支付宝安全收尾：先读 [安全复查报告](deploy/ALIPAY_SECURITY_REVIEW.md)。发布时对实际新提交核验 Linux unit／integration、安全与镜像，补生产同型 PostgreSQL 并发验证；用户部署后按 [扫码验收](deploy/ALIPAY_DESKTOP_WAP.md) 核验 HTTPS／可信 Host／日志去敏、真实 App 无本站登录、正确金额、小额实付后电脑自动成功且仅入账一次。源码和本地测试不等于线上已完成验收。

- [x] **订阅审查 P2 已修复**：辅助 `/subscriptions/progress` 慢响应不再阻塞订阅列表；列表独立结束 loading、时间异步补齐，重试／卸载丢弃旧响应，四项新回归及类型／相关lint通过；尚未发布。

- [x] **订阅时间提示与样式优化源码**：服务端权威重置日期、显示倒计时／实际到期日期、过期／暂停等边界及深浅／响应式已完成；21新项、27文件399关键项、类型／相关lint／最终Vite和浏览器夹具通过，证据见第四节。
- [ ] **订阅提示真实账户验收**：发布已完成，当前固定版本sha-80c30c8。部署后用真实账号核对日／周／月重置日期、到期／续购更新和过期状态；模拟数据与CI不能当真实账户或付款验收。生产由用户部署。

- [x] **重复订阅是否累加时长咨询**：现有购买入账、同组／跨组、过期续购、同订单重试与后台分配入口已核验，6项现有定向测试本轮通过；结论和证据见第四节。本次不视为开发、发布或真实付款授权，原发布与现场验收待办保持。

- [x] **电脑端支付宝 WAP 扫码源码与本机回归**：后台默认关闭开关、仅桌面官方支付宝 WAP、32 字节令牌免登录 302、官方目标与订单状态校验、分钟期限、恢复和旧手机流程已实现；195 顶层＋87 子项支付回归、378 前端关键项、API 契约、类型／lint／前后端构建通过，详见第四节。
- [x] **本次支付宝代码发布门禁**：随4bc5abbb5同步上游版正式推送，实际SHA的Linux unit／integration、前端、安全及GHCR全部成功。生产HTTPS／真实App／小额实付与通知幂等仍由用户更新后验收，见第四节与更新手册。
- [ ] **支付宝真机扫码与实付验收**：用户部署本版后开启设置，用支付宝 App 扫新电脑订单二维码；确认没有本站登录／二次支付操作、商户和金额正确，余额与订阅、手机旧流程、刷新恢复及过期／取消分别验证。用户确认小额实付后检查电脑自动成功、余额或订阅只入账一次；失败保留去敏后的错误码与时间，按 [验收说明](deploy/ALIPAY_DESKTOP_WAP.md) 取证，不保存完整付款链接／令牌／商户密钥。真机兼容性未知，生产由用户部署。

- [x] **Google AI Pro 视频配置可行性咨询**：会员类型已确认，官方权益／API 计费与现有视频入口源码已核验，结论见第四节。若后续要求本站直接生成，先落实具体 API 项目、抵扣领取／预付余额、模型权限与费用预算，再按下方“视频与分镜重新开放条件”适配接口和验收；本次未授权开发、收费生成或恢复视频入口。

- [x] **配方打开原图源码及本机验证**：data新窗口链接改页内dialog，61项配方＋3项打包、类型／生产构建及独立／内嵌真实浏览器通过，细节见第四节。
- [x] **配方打开原图发布门禁**：4bf1436fb已推，实际同SHA CI／安全／GHCR全成功，固定镜像sha-4bf1436与摘要已由成功job核实，详见第四节终态。
- [ ] **配方打开原图现场验收**：用户备份后更新sha-4bf1436、刷新配方；用真实生图结果点打开图片1／多图、手机关闭和原图保存。现场响应／href及下载落地仍未取证；内嵌夹具MutationObserver错误来源未知，若现场出现再取栈，不代部署生产。

- [x] **包号免扣余额源码与隔离验证**：按用户开发授权完成迁移243、正向准入、DB／缓存实扣0、独立参考计量、共享并发／RPM／日请求上限、WS控制帧防旁路、策略后台和用量报表。Linux全量unit、真实PG专项、Go lint／embed、372前端关键项／类型／lint／构建成功。默认2并发／30rpm／日0／2MiB／生图关，规则与实际局限见DEDICATED_BILLING_PLAN第六节。
- [x] **包号新版本提交与发布门禁**：acd73d716已推origin/main，65文件；同SHA的CI（含unit／integration）、安全与GHCR全部success，固定镜像sha-acd73d7及摘要从成功job日志核实，详见第四节本轮终态。前后端＋243同版，旧293c091不含免扣。
- [ ] **包号真实业务验收**：用户备份后用sha-acd73d7部署；真实浏览器编辑策略、合法Claude／Codex／HTTP流式／WS客户端、0余额专属组及同用户公共组、同步生图启用与未适配入口拒绝仍待。没有合法账号不发收费请求，日限额按实际档位调整；回滚旧计费版先暂停相关Key／账号。未知生产现场和历史扣款不自动补偿。
- [ ] **包号进一步运营约束与对账**：本版日上限是硬请求数，尚无输出Token硬限／美元预留／自动IP异常处置；按实际档位与兼容性决定是否增强。租约／账本保留供异步结算，观察增长后制定归档策略。历史已扣款只据请求与余额流水逐笔形成补偿清单，不自动全额退款、不把参考量当真实上游成本。

- [x] **本轮安全修复授权已收到**：用户明确“修复好再发布”，前轮自动审批的授权阻塞已解除；Axios1.20.0安装执行完成、官方sumdb核对x/image0.45.0，投稿清理检查和解码回归已写。先前“等待授权”均属历史。
- [x] **修复版新SHA发布门禁**：293c091b3已推，实际提交快照644项与类型／源码lint／构建通过；同一SHA的Go lint、embed编译、单测／集成、安全审计与GHCR均成功，固定标签sha-293c091及摘要已核实。详见第四节最新终态与ci-final.json，生产由用户备份后手动更新。

- [x] **截图修正版提交推送**：`d84d60e`已在origin/main，限定65文件与实际发布源码644项验证；并行包号、更新脚本和CI接线未提交。
- [x] **前轮发布授权阻塞已解除**：用户明确“修复好再发布”，本轮Axios／x/image升级和八处清理检查已实施，本地644项／类型／源码lint／构建及审计例外检查通过；正式发布仍等待新功能SHA的CI／安全与镜像终态。

- [x] **本轮截图所指源码修正**：滚动吸顶栏、SynaRoute选中态、首页提示文案、两图撤下、五类各两张新图与首页人像入库已完成；186项回归、类型／lint／构建与浏览器证据在`output/studio-refinement-20261001`。最新47图优先于下方历史数量，剩余素材由用户自行添加与审核。
- [ ] **线上生图目录真实取证**：同一登录账号确认API密钥页是否有启用、未过期、额度可用且绑定启用分组的密钥；若文字有而生图无，核对分组生图权限；若有可用密钥却无模型，取所选密钥`/v1/models`响应／状态码（不发送密钥明文）。当前仅模拟目录流程通过，不能标线上问题已解决；下拉诊断源码已推d84d60e，用户现已授权安全修复后发布，最终固定镜像以新SHA门禁成功记录为准，不代部署生产。

- [x] **家族产品限定提交与推送**：765e5a9已在origin/main，GHCR固定标签与摘要由成功任务日志确认；排除包号与新版更新脚本。交接文档后续skip-ci提交不生成替代应用镜像。
- [x] **765e5a9／d84d60e失败门禁已由新修复版本解决**：工坊八处清理检查、x/image0.45.0与Axios1.20.0已在293c091b3；有效／截断图片、尺寸与保存失败清理回归随Go unit／integration成功。新SHA安全扫描、Go lint与GHCR全部成功，固定标签sha-293c091；旧失败镜像仍不作为本轮更新目标。

- [x] **本轮家族源码与镜像发布完成**：47图库／截图修正与安全修复均包含在sha-293c091，三项新SHA门禁全部成功，未混入并行包号与更新脚本。
- [ ] **用户备份更新与真实账号验收**：用户备份Postgres、部署配置及包含studio-submissions的app_data后手动更新sha-293c091；验证滚动栏、SynaRoute当前态、47图库和真实登录下主题同步、配方Key／生图模型选择、图片投稿待审／管理员发布／撤回。线上空框需实际账号响应取证，视频与分镜保持待开放，不代部署生产。

- [x] **暂不开放视频与分镜**：已关闭工坊展示、玩法／旧链接与新视频投稿前端入口，显示友好“待开放”，原始资料与历史记录保留。服务端协议和管理员历史管理不属于本次停用范围。
- [ ] **视频与分镜重新开放条件**：先核实实际可用视频供应商、任务提交／查询接口、账号权限与费用预算；准备9–12段题材／风格不重复的原创或明确获授权成片，逐条附对应完整实际提示词、参数及来源／授权。经人工审核和桌面／手机播放、完整画幅验收后，再同时恢复core.mjs的VIDEO_CONTENT_ENABLED、投稿选项与提交／重投路径，并重跑相关回归。不凭历史Sora文档接入、不自动收费、不搬运未授权作品；当前保持待开放。

- **瀑布流待用户视觉确认**：查看4185工坊与studio-masonry-desktop-final／mobile截图，确认原画幅紧凑布局；本轮仅预览与静态构建，8080仍旧版。后续提交／镜像替换需明确授权，包号并行修改继续隔离。
- **视频接口不能沿用旧Sora文档接入**：OpenAI官方Videos API2026-09-24已关闭且未提供一对一替代。若用户希望自动生成样片，先确认实际可用供应商／任务接口文档与预算，不发送密钥明文；本站第三方视频能力与实际账号权限尚未验证。也可沿已有投稿审核链展示有许可的成片与完整实际prompt；十套分镜仍不能说已生成。

- [x] **生产磁盘占用定位（截图）**：已取得df／Docker镜像独占大小／部署目录输出，证实带sha标签旧镜像约1GB，数据库备份5MB；已给保留当前及上一版、删除五个更旧镜像的具体命令。
- [x] **更新脚本镜像保留策略（源码）**：用户已要求增加，当前＋实际更新前版本＋其他容器引用镜像保护，失败不清理，22场景假Docker回归与语法通过；同步UPDATE_GUIDE与定制清单。
- [ ] **新版更新脚本交付与生产验收**：尚未提交推送，需把新版deploy/update.sh单独替换到服务器/sub2api-deploy/update.sh并chmod；镜像pull不更新宿主脚本。用户执行更新后复核df／镜像清单／回退版。备份／日志轮转未在此轮实现，不做全局prune或数据卷删除。

- **最新风格／投稿版验收与发布**：看4185风格预览；投稿／审核已实施并通过隔离测试，但旧8080尚无新接口。获得本轮明确更新授权后按runbook先备份DB＋app_data投稿目录，前后端同版本重建原应用，不改数据库／Redis、不删卷；用真实用户完成图片／真实MP4上传→pending私有→管理员确认发布→公开作品和对应prompt→下架／撤回不可公开，以及深浅／手机／反代上传体积验收。不可把静态预览、mock用户或验证镜像当部署；多副本上线前先更换存储层，清理额度与自动内容安全待另议。
- **视频能力仍待条件**：当前无视频生成工具，十套分镜仍0段对应样片。已向用户询问自己的魔法API视频模型／任务接口与单条预算（不用发送密钥明文），等待补充；未确认前不收费调用、不搬运未经授权的第三方原片或用静态动画冒充。用户可先投稿自己有展示授权的成片及真实prompt／明确反推模板，人工审核通过才公开。

- **最新十图验收**：本轮十张真图已接入，32张图库与90项回归／构建、实际桌面／手机检查完成，请用户看4185最新预览；没有替换8080或发布生产。确认后再按runbook独立授权更新／提交，不将当前源码预览当上线。
- **视频仍未完成**：10套分镜没有对应生成视频，本轮明确答复，不再把“视频条目”误称成片。需可用视频生成服务及预算，或可核实展示／再分发许可的素材；每段保存真实生成记录或明确反推模板再上播放器，不用海报动画、无关旧电影或编造原prompt填补。

- **本轮视频缺口（最高优先）**：用户要9–12段有吸引力的实际视频，现在只完成10套分镜，0段新成片。等待用户在已发问题中选择自己的视频服务（须确认接口和预算）、上传可展示片段或可核实再分发授权的现成素材；逐条拿到真实样片／真实对应prompt／来源与许可后，再上播放器与卡片，不用海报、二维运镜、无关影片或后写模板冒充。现成短剧若未公开原始提示词只能标“反推参考模板”，不能声称作者原prompt。
- **本轮验收／发布**：请用户查看4185最新作品与提示词布局及六张新图；确认后才补媒体并决定是否更新8080。当前新版未Docker构建或部署，无新提交／远端CI；保留原8080、数据库与包号并行工作。系统剪贴板、TXT落地、真实账号模型配置和收费生成仍不能沿用模拟测试；不增加自动采集、调用或定时任务。

- **最新本轮收口优先**：169项测试、类型／lint／构建及新工作区已完成。用户确认后才替换本机8080（重新备份、保留旧镜像和精确CORS层、仅更新应用、比对镜像与运行容器以及业务／设置摘要），不沿用前次更新授权作为本轮已部署证据。生产与提交推送须另行明确授权，不能把构建新测试镜像当发布完成。
- **真实工作区验收**：用户在同一站点真实登录→顶部配方／工坊→检查右侧展示、主站主题同步、已有Key／模型下拉、显式应用、取消与退出清理。没有Key时原密钥管理创建，不替用户创建测试Key或发收费生成；模型目录可查询，实际文字／生图／计费需要用户授权。外部独立URL、SynaRoute桌面账号SSO与跨站统一退出不在本轮。
- **视觉与素材后续**：请用户确认原创成年男女与新网格／完整预览；若必须按三条抖音逐镜分析，先修Chrome连接或用户上传可播放视频，当前未知完整视频内容。现有4段授权参考是开放动画电影，不称真人模型演示，待用户提供许可素材或授权模型／预算后可增原创真人电影短片。工坊生成API、Skills实际运行和定时任务仍未启用。

- **一体化最新待办（覆盖旧“必须独立托管”）**：本机8080构建／部署／数据保护／HTTP／桌面及手机页面验证已完成，用户可直接打开/recipes/和/studio/。下一步由用户在8080真实登录后核验控制台切换、配方配置弹窗及已有Key／模型授权，再决定收费调用；本轮keys基线0，不创建Key或发收费请求冒充验收。IAB未显示选择弹窗，需在用户正常浏览器确认，系统剪贴板／TXT实际落地仍待。正式生产需另行授权提交推送并核对新SHA的CI／GHCR，旧sha-25071d4不含本次集成，不拿旧CI替代；完整web测试的旧logo.png夹具失败单列处理，不在本轮修无关用例。统一清除已授权Key、视频API仍未完成，工坊定时任务保持关闭，不混入包号发布。

- **最新版本更新门槛**：25071d4 功能已推送，核实 GHCR 36736764218、CI 36736764231 和 Security Scan 36736764322 成功终态，再由用户备份部署配置／数据库并在 /sub2api-deploy 执行 ./update.sh sha-25071d4；标签仍须核对本轮成功任务日志，不在构建中拉取或直接使用 latest。发布后验工坊三图／鼠标与键盘置顶、用户及管理员皇冠、免登录官网下载入口、真实登录返回家族首页。生产、真实官网下载安装／协议导入和模型授权未验；配方／工坊仍须托管并设置公开地址，完整跨站 SSO 未新增，包号专项不在此功能提交中。

- **本轮家族视觉发布（优先于下方待授权历史）**：已获提交推送授权，按限定暂存版本验证后推送 origin/main；核对本轮 SHA 的 CI、安全扫描与 GHCR，再提供实际镜像标签供用户备份后更新。包号专项仍另行处理，独立配方／工坊地址和精确来源仍须配置；不把重发主站镜像当全部产品已托管或跨站 SSO 已实现。皇冠真实侧栏、生产登录回跳、官网实际下载及真实模型授权仍待用户验收。

- **SynaRoute 官网／下载入口待发布**：用户在4184/home#synaroute确认新增两个公开入口；链接与布局已在源码和浏览器验证，官网实时可达性／安装包下载／真实协议导入仍待用户网络下验证。本轮未改原导入失败提示或增加可靠安装检测，不能仅凭焦点断言未安装。确认后另行授权提交／发布，与未确认包号专项隔离；网站换域名时更新family/synarouteLinks.ts，不把第三方同名产品地址替入。

- **皇冠图标待发布验收**：本轮家族首页侧栏皇冠仅在源码，尚未进入8080或远端镜像；用户确认后再授权提交／发布。更新时验证用户及管理员个人区、展开／收起和双主题，仪表盘四宫格保持不变，不沿用旧镜像截图当新图标验收。

- **工坊三图待用户视觉确认与发布**：用户查看4184/home#studio或最新studio-balanced桌面／手机截图，确认深发色参考写真与三张等宽放大的布局；不能再沿用被否决的雨夜两图、金发人物或1.3:1:1卡片。默认无遮挡、鼠标悬停／键盘置顶及手机尺寸点按已取证，不等于用户视觉认可或实机触摸验收。本轮仅本地源码，321fe99不含这些改动；提交推送／Docker部署仍须新的用户授权，保留并行包号改动。独立产品实际发布地址仍未知；先托管，再按FAMILY_PORTAL“在哪里配置”创建MAGIC_RECIPES_URL／MAGIC_RECIPES_ORIGIN／MAGIC_STUDIO_URL仓库Variables并重建镜像，不把本机预览或回退默认站名当线上配置。

- **本轮发布检查**：功能321fe99及其镜像已发布，Go Unit通过，CI36712632123集成仍运行；先核验集成终态再由用户备份并执行./update.sh sha-321fe99，不标全绿或直接更新latest。此版仅产品相关，包号专项仍在工作区、须另行确认提交。三个MAGIC_*仓库Variables均未配置；独立配方／工坊静态页面分别托管并设置公开构建变量后需重建镜像，当前生产入口保持禁用，不宣称推原站镜像即可运行全部产品。

  - **产品展台视觉验收**：本轮替代旧产品卡片，最终94dba406镜像已部署8080，三项修复评分ship不等于用户视觉批准。用户在8080/home切换四个预览并确认入口表达，再用自己的账号验API控制台／顶部产品坞及配置授权；完整真实组件的模拟账户夹具只证明布局，不能当真实登录或SSO证据。生产仍不部署。

- **首页最新验收方向**：打开8080/home确认原品牌首屏及下方子产品介绍；用用户自己的账号验证默认进API控制台、顶部胶囊和显式打开配方／工坊，SynaRoute只进入密钥配置。不再要求登录返回产品目录。生产地址／域名仍待确认，未配置地址应禁用；后台home_content是否覆盖默认首页需发布时核实。真实配置授权、模型权限／计费、跨窗口闭环仍待用户测试；不代建Key或收费调用，工坊定时任务继续关闭。

- [ ] 多专属组修正与具体400诊断需同版本前后端发布／本轮CI；用户已确认两位成员都选全仍失败，不再要求重复勾选。线上公开管理页仍无新诊断；需核验实际后端版本与失败请求。此前严格单组规则已被本轮同批成员多专属组要求替代。

- [ ] 线上包号400：需本次保存account_id、group_id、user_ids及现场数据库配置来确认具体失败项；新诊断补丁尚未发布。源码／隔离库可排除Save使用缓存及完整标准独占配置必然失败，不能替代线上取证。

- **独立产品登录的剩余边界**：最新首页验收按本节首条；若需要独立母品牌域名或SynaRoute桌面账号SSO，先确认授权协议、产品账号能力与退出策略，再另行实施。不跨域共享JWT或宣称现有密钥导入就是SSO；工坊已是公开内容首版，正式发布地址仍待确认。

- **家族与配方发布配置**：确认同源/recipes/或独立子域名，以及服务器宝塔／1Panel／原生Nginx等管理方式；按FAMILY_PORTAL.md设置仓库MAGIC_RECIPES_URL／MAGIC_RECIPES_ORIGIN公开Variables（已接Docker构建ARG），独立来源还需精确API CORS。两端新版本验收后再由用户授权提交发布、查本次CI镜像、备份并部署；配方单文件仍单独托管，不能覆盖原站server块，不用运行容器变量冒充前端构建生效，也不把本机4178／候选域名写进生产入口。当前未操作仓库变量、域名或服务器，生产部署未授权。

- **魔法配方登录选择验收与发布**：从8080家族首页／控制台打开配方后，4178模型设置使用本机8080；用用户自己的账号选择已有有效Key／分组／模型并明确授权，再自行运行，验证真实目录、跨窗口、权限与计费。精确4178 CORS已HTTP验证，但没有真实账号闭环证据。线上需确认来源、配置VITE_MAGIC_RECIPES_ORIGIN并授权两端发布；生产仍未更新。手动连接与复制继续保留，视频接入待另行适配，不代建Key或收费请求。

- **第四产品本地验收**：魔法工坊16条资源／4个玩法已实施，4179仍独立Node预览，已接入本机家族介绍及控制台打开入口。待确认名称、界面、人群并在目标工具实测；约40条内容、站内生成、正式托管与Docker仍待决定。定时任务暂不启用，不自动创建调度；说明见product-samples/magic-studio/README.md，不把入口集成当模型／Skill效果通过。

- **魔法配方模型接入验收**：用户在4178顶部“模型设置”填写自己的文字或生图连接，应用后整理提示词并运行，验证实际Key／分组模型权限、CORS与计费；当前只证明模拟兼容接口可用，不代表真实服务接通。视频需要提供服务商任务提交／查询格式后另行适配；正式托管未决定。历史“是否接模型”已由本轮要求更新为自带接口，复制流程继续保留。真实剪贴板／TXT下载落地与配方模型效果仍待，不用接口模拟测试替代。

- **接入教程剩余验收**：前端检查与公开／预览验收已完成，8080已随本机源码镜像包含教程；生产未更新。后续用实际分组密钥分别验证Anthropic Messages、Responses与至少一个Chat Completions客户端及/help/guide正文；客户端版本变化重新核对官方文档，不把HTTP入口存在当真实调用成功。

- **专项修复下一步（覆盖下方“先修复”待办）**：四项代码及定向回归已完成；待用户授权提交推送或本机Docker测试。发布时查本轮SHA的CI／GHCR并备份，前后端同版更新，旧缓存页缺expected_updated_at会安全拒绝，需刷新。管理员实际验收两个成员移除一人、剩余者调用、旧表单409及明确恢复；不要把隔离测试镜像当8080部署。全量Go／线上请求根因未验证。

- **包号发布前阻断项**：不要单独发布当前成员移除补丁。先修复deploy/DEDICATED_AUDIT.md中的Key恢复／停用状态不一致、全员状态耦合及旧表单覆盖撤销，再补期望行为回归；管理页补配置诊断。当前仅完成审查，待实施；严格独占及历史隔离保留，生产历史403/503仍无现场证据。不自动提交推送部署。

- **当前优先**：移除包号成员400的修复待用户授权提交／推送或本机部署；不要将sha-43b715a当作已含此修复的版本。发布后验两个用户直接移除一个、剩余者请求可用及被移除者Key停用；旧版不要仅停用Key就误称可以保存（旧校验仍计入停用Key）。

- **发布最后门槛**：代码43b715a15及镜像sha-43b715a已推送发布；仍需确认CI36657101544的集成测试终态。本轮不能标全绿；成功后用户可按UPDATE_GUIDE备份并运行 `./update.sh sha-43b715a`。生产403/503仍无现场日志，不宣称本次解决历史故障。

- **本轮提交推送已获授权**：推送后核实对应代码SHA的CI／安全扫描／GHCR，再提供固定镜像标签；迁移242发布前备份数据库，保持独占规则，不直接回滚旧单用户镜像。

- **2026-09-30后续授权部署完成**：8080现已更新，本轮此前“需授权本机部署”待办关闭。用户可登录测试游客入口隐藏、FAQ右侧、包号成员／改绑／删除与费用显示；提交推送及生产仍未执行，不能拿生产站点验本机新版。

- **2026-09-30 本轮待验收**：管理员真机确认多人选择、名称、重选账号／分组和删除流程，深浅主题与手机；用户确认FAQ留在右侧及智力效率费用展示。当前源码未部署到8080或生产，下一步需用户授权提交／推送或本机部署；发布时查本轮SHA的CI，不沿用fa45b20。242上线先备份，不直接回退旧单用户镜像。两个生产错误缺日志，保持原因未定，不绕过安全校验。
- [x] **游客功能发布验证**：代码fa45b2023的CI（含Go单测／集成）、安全扫描、GHCR全绿；使用镜像sha-fa45b20。旧记录「Go单测仍待」已由本条取代。
- [ ] **游客功能生产更新**：由用户备份后在/sub2api-deploy执行./update.sh sha-fa45b20；生产首页配置、完整登录选购与支付恢复仍需用户验收，不自动下单或部署。
- [ ] **游客适配复核**：宽屏偏左问题已修并部署8080（ea6e3a5ecc3c），请用户以截图对应窗口刷新确认；不把旧版无溢出检查等同完整视觉验收。

- 魔法配方创作扩展：待用户体验生图／视频配方并确认视觉；需另行授权模型实测（8份共32个设计样例，建议至少64次，尚未执行）。仍需验证系统剪贴板与TXT下载落地；本轮不部署。
- [x] **游客控制台预览入口**：首页主CTA与公共导航已添加，独立只读预览完成并通过179项定向回归与浏览器验收。
- [x] **游客控制台预览本机发布**：用户授权后已部署8080，镜像ecb3961845e8；新版入口、个人接口匿名认证、数据保留及浏览器交互通过。待用户操作与视觉确认。
- [x] **配方教程**：已将使用流程、外部AI运行步骤和常见问题加入页面，27项回归及真实浏览器往返／手机检查通过；不代表新增模型执行能力。
- [x] **游客官网本机Docker部署**：8080已运行`e7a582285cbd`，Go embed构建成功，三服务健康；公开真实套餐／FAQ／登录提示验收通过，备份和旧镜像保留。
- [ ] **游客官网剩余验收**：Go unit测试仍待本轮未来提交CI；登录后选中套餐、注册回跳和支付恢复完整联调尚未实测，不做未经确认的下单扣款。用户确认公开套餐文案／权益及最终视觉；生产仍由用户授权后手动升级。
- [x] **配方确认框**：已用主题内弹框替代覆盖／清空的原生confirm，24项回归及真实浏览器取消／键盘／手机检查通过；视觉由用户确认，刷新旧页前先保存输入。
- [x] **游客官网源码**：公开首页／套餐／FAQ及可取消登录提示已完成，163项定向测试通过，保持当前品牌，不开放个人数据。
- [ ] **游客官网后端与业务验收**：本机8080新后端已部署，真实目录和匿名认证边界通过，替代此前仅4175预览阶段。剩余关闭开关、登录／注册后的完整回跳、确认购买和支付恢复联调见上方待办；未来提交必须检查对应CI，不沿用旧结果。
- [ ] **游客官网发布确认**：用户确认首页／FAQ文案与布局；目标站点检查home_content是否覆盖官网。无自动生产部署，未确认运营政策不发布承诺。
- [x] **配方交互原型**：独立页面已生成并本地启动，入口product-samples/magic-recipes/site/index.html，维护说明见site/README.md；19项核心／DOM回归通过，不等同模型效果或生产验收。
- [ ] **配方原型验收**：用户确认界面及6份流程，在实际目标AI工具检查粘贴与TXT文件下载；浏览器自动化尚不能证实系统剪贴板／文件落地。最终托管与是否接中转站另行决定，未经授权不部署生产或开公共模型执行。
- [x] **大学配方样品**：学生定位限定大学；已替换为论文精读、编程实验复盘、期末复习，工作系列不变。入口product-samples/magic-recipes/README.md，六份内容草稿，非已上线产品。
- [ ] **大学配方评审**：只在大学范围确认优先专业／课程，不再扩展中小学。先核对内容与难度，再按README执行授权模型测试；多轮配方需验证理解反馈、修正和真实实验记录边界。六份共24个样例、至少48次首轮抽样，付费调用须另获授权，不用文档检查代替学习效果。
- [ ] **日夜海报定稿**：雾钛青深色与雾白日间两版已新增官网／微信并排双二维码（with-wechat目录），微信码标签按用户要求更新为「客服咨询」，待用户确认；正式发布前手机实扫两个二维码，确认目标与有效性。原图与此前单二维码成品均保留。
- [x] **魔法配方首批内容样品**：3份已完成，入口product-samples/magic-recipes/README.md，当前为待评审草稿，不是自动化程序。
- [ ] **魔法配方评审与实测**：用户已认可内容方向；现为工作3份＋大学3份，按README的24个样例至少各运行两次，须先授权模型及预算，留存输出与失败记录；验证通过后才讨论制作自动化及发布。未经授权不开放免费执行或托管用户密钥。
- [x] **第三产品方向初选**：用户选择魔法配方并授权先做3份样品；此前魔法实验室仅为历史建议，不继续实施。获客人群、实际转化与平台开发仍未确认，不承诺知名度增长。
- [ ] 修正版镜像sha-a1b9137已发布，等待CI36375537042后端单测／集成测试完成；仅此项仍未确认。生产升级仍由用户手动备份执行，不沿用首轮失败CI或旧镜像标签。
- [ ] 本次包号发布：核对功能提交对应CI和GHCR成功后提供固定sha镜像标签；生产由用户备份并手动升级，真实账号联调仍待。
- [ ] 本机新版已部署：用户验证包号管理／我的专属账号及推广清理；此前“授权部署后验收”现进入验收阶段。真实账号隔离、额度采样、续期撤销仍待联调；Go单测仍待执行。
- [ ] 推广清理：授权部署后验收用户菜单、首页／用量页脚及三处代理选择区域；当前仅源码变化，运行站点未更新。
- [x] **包号首期源码**：用户已选Claude与ChatGPT/Codex；管理员绑定／续期／撤销、用户状态与额度、调度隔离和回归代码已加入，文档见DEDICATED_ACCOUNTS.md。仅标准模式，复用手工预配置的独占分组，不改变计费。
- [ ] **包号上线前验证**：先在Go1.27环境gofmt本轮变更并核对Wire，再执行Go单测／集成／lint／embed构建和真实数据库迁移；取得授权后本机新镜像验收两名用户隔离、旧Key／备用路由／影子账号拒绝、HTTP／WS到期撤销、Claude／Codex真实额度、深浅色与手机。当前未提交推送／未部署，不沿用旧CI；启用绑定后旧镜像没有隔离保护，回滚先暂停相关账号与Key。
- [ ] **宣传海报用户确认**：2026-09-28雾钛青竖版初稿已交付，待确认文案/版式；对外发布前手机实扫并确认目标站点。未自动上传、发布或发送给第三方。
- [x] **SynaRoute咨询范围与初查**：官网和客户端均已检查，结果见第四节2026-09-27只读评估；是否实施及改动范围待用户决定，真实后端/安装包验收未执行。
- [ ] **异常登录现场取证**：取得告警所属系统（本站/SSH/其他）、时间时区、3–5条脱敏记录；保留并关联生产审计与反代日志的action/path/status_code/actor/auth_method/IP/UA/request_id，区分失败尝试、2FA中间状态及非本人成功操作。核对生产镜像、代理头清洗与源站暴露、管理员API Key是否启用及使用记录；不要发送密码、完整Token、Cookie、密钥或.env。证据不足不得宣称安全或已被入侵；未核对代理拓扑前不直接切换转发IP开关。
- [x] **说明订单菜单显示条件**：已对照上游菜单、管理员设置store、设置表单及路由守卫；无需为解释开关改业务或重建镜像。
- [ ] **两段说明精简上线**：源代码与回归断言已修改，待本机镜像重建及视觉验收；不要将当前8080页面当成此轮新版。
- [ ] **核验用户实际菜单展开／套餐直达页**：先区分「订阅管理」（用户订阅分配）与折叠的「订单管理」→「订阅套餐」（商品）。直达`http://127.0.0.1:8080/admin/orders/plans`。若展开后仍缺子项或直达页无「创建套餐」，须取实际页面／控制台证据再定位；不能用源码存在或HTTP200代替交互验证。
- [x] **智力页面反馈修正（代码）**：移除外链与频率提示、明确矩阵排序并增加选项、改为可读三指标对比；修复支付居中容器宽度适配。76项回归通过，真实视觉仍待。
- [x] **反馈修正版本机更新**：8080运行`cfa0a12c8bf6`；镜像一致、三服务健康、HTTP新资源、订阅新分组与全部设置保持检查通过。无收费套餐创建，无生产操作。
- [x] **新增智力效率源代码**：接入用户指定JSON，提供矩阵／效率图、图标、30分钟浏览器同步与失败保留；真实数据解析及70项定向回归通过，详见CUSTOMIZATIONS。
- [x] **智力效率本机Docker发布**：用户已授权且更新完成，8080运行`1a7a31d0e424`；构建、镜像一致性、三服务健康、HTTP新资源与数据／设置保持检查通过，旧镜像及数据库备份保留。
- [ ] **智力效率视觉验收**：浏览器控制工具认证错误，仍需用户在真实浏览器检查登录、深浅色／手机／筛选／对比分页／网络错误及精简说明。用户已授权本轮提交推送，生产由用户更新，不以旧本机镜像验收最新代码。
- [ ] **配置订阅套餐**：本机原生订阅功能已开启但套餐0个；用户需给出分组、售价、币种、周期、日／周／月额度及续费规则后再配置。不要照搬参考截图，不创建测试售卖套餐或实际订单。
- [x] **本轮视觉收口**：用户明确「暂时先这样」，保留统一认证双栏、雾钛青、核心M与扁平Logo、魔法家族／多模型 API 服务；不继续自动改风格。这是阶段性接受，不等于全部业务验收完成。
- [x] **代码及交接整理**：本轮用户授权提交并推送到自己的origin/main；不推upstream、不发布生产。已有文档纳入版本管理，本机.env／override／数据库备份／设置快照／日志不入仓。推送结果以Git远端HEAD核验为准。
- [x] **新Logo复制**：C:/Users/Administrator/Pictures/mofamily/mofa-mark-flat.png，16313字节；源文件与副本SHA256一致，旧mofa-logo-local.png保留。
- [x] **本机部署与已测范围**：镜像b4fa2d8e97a1仍运行8080；真实登录、空态业务页、认证多尺寸及日夜主题已测。品牌名称／副标题／Logo来自本机设置，非代码硬编码。
- [x] **维护边界与故障记录**：展示接缝已登记CUSTOMIZATIONS和custom/UPGRADE；Docker代理防复发见deploy/LOCAL_DOCKER_RUNBOOK.md。上游仍需合并核对，不能承诺零冲突。
- [x] **本轮提交推送及镜像发布**：功能提交d592f9481已推送，GHCR镜像sha-d592f94构建推送成功；只用此标签升级，不使用随后纯文档提交的SHA猜镜像标签。
- [x] **本轮CI最后结果**：d592f9481对应CI运行36264121775最终success（含Go单测／集成测试），安全扫描36264121782及镜像36264121792均success；通过GitHub API针对同一功能SHA核验。本机未跑Go单测，不混淆本地与CI结果。
- [ ] **带数据业务验收**：在获授权环境检验真实密钥／用量表格、分页、图表、错误／警告路径和上游请求；空库截图与临时DOM探针不能代替。
- [ ] **首次引导与完整交互回归**：完整走查引导、键盘／焦点、长文案、加载与错误状态；目前仅验证关闭引导后关键操作。
- [ ] **生产升级（用户操作）**：本轮用户确认自行Docker更新；待本次GHCR构建成功后，按deploy/DEPLOY_CUSTOM.md及deploy/UPDATE_GUIDE.md备份数据库、记录旧镜像、指定新sha标签升级、健康检查并准备回退。检查智力效率新入口／排序／对比／两段说明，以及支付开启后的订阅套餐配置。站点名／Logo需在生产独立配置；推送不等于部署。
- [ ] **后续独立任务**：SynaRoute保留紫色，留用户新会话；可选Logo矢量重绘及进一步品牌体验打磨，不自行开启。
- [x] **升级方式保持Docker + update.sh**：不采用App内在线更新，上游预编译二进制会覆盖定制；普通重启不会还原可写层，镜像重建才会，历史澄清保持有效。

## 六、文档地图
- [deploy/DEDICATED_BILLING_PLAN.md](deploy/DEDICATED_BILLING_PLAN.md) — 包号免扣源码证据、共享限制／计量、验证与局限（镜像sha-acd73d7已发布，真实业务待用户更新验收）。
- [BRAND_IMPLEMENTATION.md](BRAND_IMPLEMENTATION.md) — **最新已批准设计的执行计划**：素材路径、色表、阶段顺序、代码接线、防坑、验收与新对话提示。
- [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md) — **定制唯一权威清单**（新增文件 / 接线改动 / 行为修改 / 自检清单）。
- [SYNC.md](SYNC.md) — 同步上游 + 构建镜像 + 部署 runbook + 4 条红线。
- [deploy/DEPLOY_CUSTOM.md](deploy/DEPLOY_CUSTOM.md) — 定制镜像部署清单（登录/起服/升级/回滚/迁移）。
- [deploy/UPDATE_GUIDE.md](deploy/UPDATE_GUIDE.md) — 日常更新 / 回滚操作速查卡（模式 A latest / B 钉 sha、常见坑）。
- [frontend/src/custom/README.md](frontend/src/custom/README.md) — 前端叠加层三种用法 + 影子替换代价。
