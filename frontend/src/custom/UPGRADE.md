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
