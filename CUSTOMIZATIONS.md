# CUSTOMIZATIONS — 本 fork 相对上游的所有改动登记

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

> 格式：`路径` — 用途 — 引入日期

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
