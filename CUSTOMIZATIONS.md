# CUSTOMIZATIONS — 本 fork 相对上游的所有改动登记

> 这是本 fork 的**唯一权威清单**。每一处偏离上游的改动都必须登记在这里。
> 同步上游（`git merge upstream/main`）解冲突时，逐条对照本表核对：
> **每一处定制是否仍然存在、是否被上游改动覆盖**。合并里定制被静默吞掉，是这类
> 维护最常见、也最难发现的失效——本表是唯一防线。
>
> 维护约定：
> - 优先**新增文件**（新 handler / 新 Vue 组件 / 新 service），新文件永不冲突。
> - 不得不改上游文件时，改动**尽量小、尽量集中**，并在下方「接线改动」逐处登记。
> - 配置类定制（`.env` / `config.yaml` / `docker-compose.override.yml`）走 gitignore，
>   不进本 fork，不登记在这里；它们记在你的私有 ops 仓。

---

## 一、新增文件（低风险，merge 不冲突）

> 格式：`路径` — 用途 — 引入日期

### SynaRoute 深链接一键导入（2026-09-22）
按 SynaRoute 文档「深链接一键导入」（`synaroute://`）在 API Keys 页面加「导入到 SynaRoute」按钮，
**形态完全对齐上游已有的「导入到 CCS」（`ccswitch://`）**：并排放在密钥行操作里，并配一个管理员开关。
规格文档：`C:\Users\Administrator\Desktop\temp\demo\SynaRoute\docs\20-深链接一键导入.md`。
- `frontend/src/utils/synaRouteImport.ts` — 纯函数：URL 构造 + 由 platform 推导分类/协议/端点（镜像 `ccswitchImport.ts`）
- `frontend/src/utils/__tests__/synaRouteImport.spec.ts` — 单测（11 用例，已过）

### 部署文档（2026-09-22）
- `deploy/DEPLOY_CUSTOM.md` — 定制版 GHCR 镜像部署清单（登录/起服/升级/回滚/迁移）
- `deploy/update.sh` — Docker 下的一键更新脚本（`pull && up -d` + 健康自检 + 旧镜像清理；带标签参数可回滚）。用户明确选择「Docker 镜像更新」而非启用 App 内按钮（那按钮是 systemd 安装用的原地换二进制，Docker 下权限失败且会被 pull 覆盖）。

### 前端定制叠加层（换肤 + 新增/替换页面）（2026-09-24）
用户方向：「换肤 + 少数页面」。设计为 `frontend/src/custom/` 叠加层，把冲突面收敛到 4 个接缝文件（见下节）。
- `frontend/src/custom/theme.css` — 品牌变量层：`--color-primary-*`（电光蓝，500=#1e8bff）+ `--color-dark-*`（深海军蓝底，减压抑）+ `.dark body` 电光蓝径向光晕。换肤只改这里；logo/站点名走管理员后台设置（无需改码）。
- `frontend/src/custom/routes.ts` — 新增页面路由集中处（`customRoutes`）。
- `frontend/src/custom/views/CustomDemoView.vue` — 脚手架演示页（验证链路用，可删）。
- `frontend/src/custom/README.md` — 叠加层三种用法 + 影子替换代价说明。
- 新增页面/影子替换页放 `custom/views/`、`custom/components/`（本身零冲突）。

## 二、接线改动（会冲突，重点核对）

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
**为什么**：品牌是深蓝底 + 银色 + 电光蓝（logo/海报），首屏应呈现深色才对味。
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
