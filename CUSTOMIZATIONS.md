# CUSTOMIZATIONS — 本 fork 相对上游的所有改动登记

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
