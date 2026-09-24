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
- `frontend/src/custom/theme.css` — 品牌色 CSS 变量层（`--color-primary-*`，通道值）。换肤只改这里。默认值=上游 Teal，视觉零变化。
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

### 前端定制叠加层的 4 个接缝（2026-09-24）
> 全部打 `[CUSTOM]` 注释。**已跑通完整 `pnpm run build`**（含 check:i18n + vue-tsc + vite build），
> 产物 CSS 确含 `--color-primary-500: 20 184 166` 定义与 `rgb(var(--color-primary-500)/…)` 引用，
> 演示路由 `custom-demo` 已进产物 → 换肤与路由链路验证通过。
- `frontend/tailwind.config.js` — `primary` 色阶改为 `rgb(var(--color-primary-*) / <alpha-value>)`（引用 theme.css 变量）。⚠️ 上游若改 primary 色阶会冲突；解冲突时保留变量引用形式。
- `frontend/vite.config.ts` — `resolve.alias` 由对象改为**数组形式**，并留「影子替换」注释示例（覆盖项须在 `'@'` 之前）。
- `frontend/src/main.ts` — `import './style.css'` 后加 `import './custom/theme.css'`。
- `frontend/src/router/index.ts` — import `customRoutes` + 在 404 兜底前 `...customRoutes` 展开。

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

- 除上面这一处外：无。

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
