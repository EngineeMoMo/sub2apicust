# 雾钛青：上游同步核对清单

本文件只描述维护操作，不授权提交、推送或部署。定制降低冲突面，但不保证未来零冲突。

## 展示层边界

- `/brand` 独立公开入口，不抢占 `/`；后台模式仍受上游路由守卫限制。
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

## 局部回退

先备份当前 diff。仅移除本次品牌组件引用、mofa 展示钩子与 /brand 路由，回退对应新增视觉规则；保留原插槽、上游组件与既有错误态修复。不要 reset 整仓、覆盖整个布局文件、删除数据库卷或回滚其他人的未提交文档。

## 当前部署限制（2026-09-26）

当前本机8080运行b4fa2d8e97a1；品牌三项配置为魔法家族／多模型 API 服务／扁平M，详见BRAND_IMPLEMENTATION第16—18节。用户已阶段性接受并授权提交、推送自己的fork；生产仍不部署。真实带数据业务验收和本轮CI结果待核验，旧镜像与数据库备份保留；网络预检遵循deploy/LOCAL_DOCKER_RUNBOOK.md。
