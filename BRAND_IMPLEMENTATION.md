# 雾钛青品牌落地计划 — 新对话执行入口

> 2026-09-26：用户已确认执行；阶段一单文件主题及登录页验证已完成，完整业务验收待补，详见第 9 节。
> 批准方向：魔法家族 / 中转站采用「雾钛青」；SynaRoute 保留紫色产品识别，统一基础设计规范，不强制全改青色。
> 不要再从 A/B/C/D 开始选风格；历史候选不是当前执行目标。批准设计方向不等于批准生产部署、全站重构、改名或改默认主题行为。

## 1. 新对话先做什么

1. 读 AGENTS.md、HANDOFF.md、CLAUDE.md、CUSTOMIZATIONS.md、SYNC.md 和本文。
2. 执行 git status -sb / git diff，保留本轮未提交文档，不覆盖他人变更。
3. 阅读最终概念稿与本节下方代码锚点，核对项目主题接线与样式优先级。
4. 从「阶段一：真实业务页换肤」开始；先拿登录、密钥页面验证，不再画一轮无关概念稿。
5. 本轮只修改交接文档，业务代码尚未开始；不得宣称新主题已上线或已构建通过。

### 本地基线（交接时实测）
- 项目：D:\ccfile\sub2apicust。
- 分支 main，HEAD 0d1e3e7b8；本地 origin/main 引用相同。本轮未 fetch，不能推出当前远端状态。
- 产品仍是电光蓝主题：theme.css 的 primary-500 = 30 139 255；不是最终雾钛青。
- main.ts 当前未保存 theme 时默认深色，已保存选择则沿用；本次不改该策略。
- 本轮留有文档改动；不自动提交、推送、开分支或部署，相关操作需要明确授权。

## 2. 已定设计原则

- 核心 M 设计保留；用户允许 Logo 换颜色。原图狮子、皇冠轮廓先一并保留，不自行重新画成另一个 M。
- 用户说 Logo 由朋友设计，无可编辑矢量源文件。当前预览是原图去背景遮罩重着色，不是已完成的矢量源稿。
- 母品牌与中转站：雾白 / 炭青底，低饱和钛青强调；科技感克制，不用全屏浓蓝、粒子、闪烁、扫描线或重金属效果。
- 品牌首屏承担吸引力；密钥、账单、用量等操作区承担可读性与效率，不放大 Logo、不铺装饰背景。
- 增大正文与说明的可读性，清楚区分主操作、次要操作、焦点与禁用状态；不靠极小极淡的字制造高级感。
- 深浅外观是一套品牌的两种模式，保留用户选择。预览默认浅色不代表要求产品默认浅色。
- SynaRoute 后续保留紫色，允许细调饱和度；统一背景层次、字级、间距、圆角、品牌背书，不在同一屏让紫与青抢主操作。
- 不改变模型/厂商/渠道等分类色，不把成功、警告、错误全部染成品牌青色。
- MOFA、魔法 API 是概念稿中的暂定呈现，未单独确认正式改名；沿用后台已配置的实际品牌名称。

## 3. 最终参考与素材

### 设计参考（本机已存在，不在仓库 Git 中）
- 最终互动稿：C:\Users\Administrator\.codex\visualizations\2026\09\26\01a0dd78-fd99-7e82-bbcc-5d095ed76b1c\mofa-aurora-system.html
- 本地浏览器检查用包装页：同目录 mofa-aurora-preview.html。
- 去背景中间素材：同目录 mofa-logo-reference.png 与 mofa-logo-preview.webp。
- 原图：C:\Users\Administrator\Pictures\微信图片_20260924212414_92_137.jpg。
- 海报参考：C:\Users\Administrator\Pictures\微信图片_20260926093050_93_137.jpg。
- HTML 是会话概念稿，含示例账户/示例密钥与演示交互；不能当成实际 Vue 页面、真实功能完成证据或生产代码直接复制。
- 使用前 Test-Path 确认素材可读；如换机器导致文件缺失，本文色表仍可用，Logo 原图应向用户补取，不编造已查看图稿。
- 若继续修改会话预览，先完整读取 visualize 技能；不要把 window.openai 或预览专用依赖带进产品。

### 参考色表（概念稿实际值）

| 角色 | 日间 | 夜间 |
| --- | --- | --- |
| 页面背景 | #F5F7F7 | #162124 |
| 内容表面 | #FFFFFF | #1D2C30 |
| 主文字 | #1D3033 | #E9F1F0 |
| 次文字 | #596E72 | #B0C3C3 |
| 边框 | #DCE5E5 | #36494D |
| 主操作背景 | #096B68 | #A1D9CE |
| 主操作文字 | #FFFFFF | #173B38 |
| 选中/柔和背景 | #E4F1EE | #283F40 |
| 基础标志色 | #225652 | #B4D7CF |

这些是视觉目标，不是一份可机械粘贴到 primary-500 的 Tailwind 色阶。要补齐当前 primary/dark 色阶并逐项验证透明度、hover、focus、disabled、渐变与前景色组合。

## 4. 已核实的代码接线与风险

- frontend/src/main.ts：style.css 之后导入 custom/theme.css，已存在主题接缝，不再重复导入。
- frontend/tailwind.config.js：primary / dark 色阶及部分阴影、渐变引用 CSS 变量。
- frontend/src/custom/theme.css：现有品牌色阶与 .dark body 双径向光晕，是换肤唯一编辑入口。
- frontend/src/style.css：公共 .btn-primary 当前是 primary 渐变 + text-white；.card 当前含暗色半透明、圆角和阴影。
- frontend/src/components/layout/AppLayout.vue：自身有 bg-gray-50 / dark:bg-dark-950 和独立 bg-mesh-gradient 装饰层。只去掉 body 光晕不等于去掉页面装饰。
- frontend/src/views/user/KeysView.vue：已有 AppLayout、TablePageLayout、筛选、操作区、DataTable，不复制重建这些业务结构。
- frontend/src/custom/routes.ts：新增定制页面的路由集合；只有独立品牌页阶段才按需使用。
- frontend/vite.config.ts 提供页面 alias 影子替换示例；本任务不以替换核心业务页为默认方案。

### 特别防坑
- 不得把夜间 primary 色阶直接整体调成浅薄荷色却保留所有 text-white，否则可能造成浅底白字。
- 应优先在 theme.css 中使用明确的语义变量及少量有依据的公共组件覆盖；分别检查按钮背景/前景，而非全局一刀切。
- 不全局替换所有 gray/teal 类，不使用大批 !important、位置选择器、深层 DOM 选择器去掩盖结构问题。
- 主色变量也会影响图表和现有高亮，用实际页面检查，不能把合并无文本冲突等同于没有视觉回归。
- 如单文件主题层不足以达到某项效果，明确报告效果差距及最小接缝改动；不默默扩展到上游组件重构。

## 5. 分阶段执行与交付门槛

### 阶段一：真实业务页换肤（下一对话优先）
- [x] 记录真实登录页深浅基线；受保护业务页不可访问，未伪造基线。
- [x] 仅在 frontend/src/custom/theme.css 调整色阶、适量公共样式与装饰强度。
- [x] 保留业务页面布局、字段、按钮行为、API 与路由；不改后端/数据库/依赖。
- [x] 登录主按钮浅深前景组合已验证；分类色与成功/错误 CSS 未改，真实业务状态待验。
- [x] 交付真实登录页截图与主题 diff，说明与概念稿差距；不包含未登录业务页。
- [x] 更新 CUSTOMIZATIONS.md 中实际主题描述、风险和验证方法；不再把旧电光蓝登记成当前新主题。

完成门槛：真实页面能使用、无明显可读性退化，且改动面可解释。若本机没有可用后端，先做登录/公开页面；业务页面验证明确标为未完成，不使用概念假表格冒充实测。

### 阶段二：Logo 资产与品牌入口
- [ ] 从原图准备透明底日间深钛青、夜间浅银青、单色备用资产，保留核心 M 与原有构图。
- [ ] 检查 16/24/32px 图标、导航与登录页尺寸；没有做好矢量重绘前只交付真实的位图，不把 PNG 包进 SVG 后称作矢量稿。
- [ ] Logo / 站点名优先走已有后台设置；不要跨组件写死名称或未确认的英文名。
- [ ] 如现有设置仅接受一个 Logo，先交付对深浅背景都可用的单资产；不要擅自给后台加日夜双 Logo 字段。

### 阶段三：独立品牌展示页
- [ ] 先检查现有首页、后台自定义首页设置和路由，选最少侵入的接入方式。
- [ ] 品牌页新增在 frontend/src/custom/ 范围，样式局部隔离；共享配色不意味着覆盖业务 DOM。
- [ ] 不把概念稿顶部的预览切换器、示例密钥、模拟账户复制到线上。
- [ ] 产品介绍使用真实能力；没有确认的数据不写延迟、稳定率、用户数等指标。
- [ ] 新页面可以先用独立预览路由验收；不擅自替换线上根路由或规划新域名。
- [ ] 不引入装饰用大依赖，不把营销页样式泄漏到表格/后台。

阶段二/三在阶段一实际效果稳定后推进，不同时大改多个系统。

### 阶段四：SynaRoute 配套（独立后续任务）
- [ ] 保留紫色产品识别，沿用相近的底色、字级、间距、圆角与母品牌背书。
- [ ] 当前只持有其官网观察记录，没有在本仓核查 SynaRoute 工程位置或源码；不要猜路径或顺带改另一个项目。
- [ ] 等当前中转站验收后，再在对应项目独立推进。不强加登录、统一账号、云同步或自家 API 绑定。

## 6. 验证与升级保护

- 先跑与实际改动相关的检查，再跑 frontend 的 pnpm run build（包含 i18n 检查、vue-tsc、Vite）；是否执行遵守当前会话审批设置。
- 按需使用 pnpm run lint:check（不要用带 --fix 的全仓 lint 制造无关改动），本轮无后端修改就不虚构 Go 编译结果。
- 必看：登录、密钥、用量、一个管理员页面；分别检查深浅主题、空状态、错误/警告、弹窗、焦点、禁用态。
- 至少检查 360px、768px、桌面宽度；页面不横向溢出，表格需要时可在自身容器横滚。
- 正文与操作不能为了造型变小、变淡；有真实用户试用反馈后再判断舒适度，不能靠设计稿断言更护眼或转化率提升。
- 核对现有主题选择持久化、SynaRoute 导入设置与深链接、在线更新禁用等既有定制不受影响。
- 上游更新后：按既有 merge-tree / 定制清单 / 构建 / 冒烟流程检查，额外检查公共 CSS 类与主题变量是否仍匹配。
- 实现结束前更新 HANDOFF.md、CUSTOMIZATIONS.md。没有新 CI/服务器证据时，仅报告本地检查结果。
- 本任务不包含自动部署。用户明确授权提交/推送后才进入 CI 出镜像阶段；服务器仍由用户按 deploy/UPDATE_GUIDE.md 手动更新。

## 7. 当前完成程度

- [x] 用户确认最终设计方向与 SynaRoute 保留紫色的品牌分工。
- [x] 最终概念稿含首页/控制台、日夜模式；浏览器检查过切换、搜索、空状态、1024px/360px 布局。
- [x] 本轮只整理交接计划，保留前述验证范围。
- [x] 真实 Vue 单文件主题实施与前端构建完成。
- [ ] Logo 生产资产、独立品牌页和真实业务页验收未完成；SynaRoute 留后续新会话。
- [ ] 没有本任务的新提交、新镜像或部署；不能引用旧 58b154f53 CI 作为这次主题完成证据。

## 8. 可直接用于新对话的提示

请先读取 AGENTS.md、HANDOFF.md、CLAUDE.md、CUSTOMIZATIONS.md、SYNC.md 和 BRAND_IMPLEMENTATION.md，再核对 git status 与现有 diff。用户已确认执行「雾钛青」设计，保留核心 M，SynaRoute 保留紫色并在后续独立统一基础规范。不要重新选风格。现在从 BRAND_IMPLEMENTATION.md 阶段一开始：在真实页面落地主题，换肤只改 frontend/src/custom/theme.css，保留上游业务结构、语义分类色和现有默认深色策略。不要复制概念稿重建业务页，不改后端，不自动提交、推送或部署。完成相关验证后展示真实效果，并如实标注无法验证的部分，更新交接文件。

## 9. 阶段一实施验证记录（2026-09-26）

### 已核实
- 开工 git status -sb：main；已有 AGENTS.md / CUSTOMIZATIONS.md / HANDOFF.md 修改，以及未跟踪 BRAND_IMPLEMENTATION.md，均保留续写。产品代码仅 theme.css 改动；未动上游组件、Logo、main.ts、API、路由、依赖或后端。
- 最终 pnpm run build 成功：i18n 3/3、vue-tsc、Vite 1062 模块。存在 Browserslist 数据较旧、动态/静态混合导入与 >500kB 分块警告，未为此扩大修改范围。
- pnpm exec vitest run src/utils/__tests__/synaRouteImport.spec.ts：11/11 通过；只做回归，不修改 SynaRoute。
- 本地 Vite http://127.0.0.1:3000/login 的真实 Vue 登录页：双主题，各 360/768/1440 × 900，documentElement.scrollWidth 不超过 innerWidth；已人工查看桌面/手机截图。
- 清除 theme 后刷新仍默认深色；写入 light/dark 后刷新分别保留。未修改策略；这里验证的是存储恢复，不是点击业务页主题切换器。
- 主按钮 computedStyle：日间 #096B68 / #FFFFFF，夜间 #A1D9CE / #173B38，无渐变；按 sRGB 相对亮度计算文字对比 6.34:1 / 7.76:1。hover 分别 #084F4D / #C4E4DB。
- 输入焦点 2px 品牌色 outline + 3px offset；空表单原生有效性检查仍阻止提交。禁用外观通过临时设置 DOM disabled 验证 opacity=0.5、cursor=not-allowed，随后还原；不代表验证了真实登录 loading 流程。
- GET /api/v1/settings/public 返回 500；认证布局等待设置，故截图没有加载品牌 Logo/标题，仅显示默认 Sub2API 页脚。未用假设置或示例账户掩盖问题，核心 M 资产未改。
- /keys、/usage、/admin/dashboard 实际重定向到 /login?redirect=对应路径，无业务页/图表/弹窗/错误警告的完整验证证据。
- 截图与 JSON 证据目录：C:/Users/Administrator/.codex/visualizations/2026/09/26/01a0ddbc-a03f-7550-93c1-ccbf9a0d3fb9。before-dark.png / before-light.png 为修改前；login-{dark,light}-{1440,768,360}.png 为最终效果；verification.json、state-verification.json 为观测值。首份 JSON 焦点为 transition 中间值，最终稳定值以 state-verification.json 为准。

### 差距与未知
- 保留原 AuthLayout 网格、柔和光晕、原尺寸/排版，未复制概念稿；没有重着色 Logo，也没有上线品牌首页。非公共 btn-primary 的内联 primary 按钮仍使用深钛青底，未强制全站变浅青。
- 未全局覆盖 gray，所以明确写死的中性色仍为上游值；去除的是 body 额外光晕，不是所有页面装饰。业务图表及透明叠加效果需有后端后再检验。
- 无真实登录、密钥调用、管理员操作、生产部署或新 CI 证据；不能据本地构建声称全站验收完成。当前仅完成阶段一可验证部分。
- 后续：提供可用测试后端及授权账号后补业务验收；用户确认视觉后再决定阶段二。SynaRoute 不在本会话继续。
- 未提交、推送、部署或修改配置；预览仅绑定本机。交接及定制清单已更新。

## 10. 本地 Docker 后端已部署（2026-09-26）

- Docker Desktop 4.92.0 / Engine 29.8.0 / Compose v5.5.1，context=desktop-linux；从本工作区 Dockerfile 构建镜像 sub2apicust:local-theme，不依赖旧 GHCR 镜像。当前雾钛青前端随 Go embed 一起打包，无需另开 Vite。
- 三个容器均 healthy：sub2apicust-local-sub2api-1、sub2apicust-local-postgres-1、sub2apicust-local-redis-1。健康检查 http://127.0.0.1:8080/health 返回 200 / {"status":"ok"}，公开设置接口 200。
- 应用只发布 127.0.0.1:8080；PostgreSQL / Redis 不发布主机端口。使用新建 sub2apicust-local_* 三个命名卷，未导入生产数据，也没有配置真实上游账户。
- 已核实容器内前端构建（i18n 3/3、vue-tsc、Vite）及 Go build -tags embed 成功。本轮未跑 Go 单元测试，不能表述为后端单测通过；宿主机仍未安装 Go。
- 后端 --version：0.2.8，commit=docker，构建时间 2026-09-26T13:24:35Z。镜像 ID 可用 docker image inspect sub2apicust:local-theme 核查。
- 配置：deploy/.env（随机密码及密钥）与 deploy/docker-compose.override.yml（独立 Compose 文件）均经 git check-ignore 确認忽略，不提交、不回显密钥。管理员 admin@sub2api.local，密码读取 .env 的 ADMIN_PASSWORD。该 Compose 文件必须用下方 -f 独立启动，不与上游默认 compose 自动合并。
- 容器 printenv 已验证 DISABLE_ONLINE_UPDATE=true，BATCH_IMAGE_ENABLED / BATCH_IMAGE_QUEUE_ENABLED / BATCH_IMAGE_VERTEX_ENABLED 均 false。
- 构建首次因 auth.docker.io token 请求超时失败；docker pull docker/dockerfile:1.7 成功后，为重试命令进程设置 HTTP_PROXY / HTTPS_PROXY=http://127.0.0.1:7897，随后完整构建成功。未修改系统代理；后续重建若再遇网络超时，先确认该代理仍运行，不将 127.0.0.1 代理地址直接配置进容器。
- 真实浏览器登录成功，进入 /dashboard；随后出现部署与运营合规确认和新手引导。未代用户填写确认短语、未接受或绕过合规流程。需用户自行阅读并决定；业务页完整验收仍未完成，阻塞原因已从缺后端更新为首次登录确认待用户处理。
- 默认品牌为新库 Sub2API；未从生产复制站点名或 M Logo，也未修改既有核心 M 资产。登录截图 docker-login-dark.png、首次登录截图 docker-first-login.png 在第 9 节证据目录。
- 本地部署不等于生产发布：未提交、推送、触发 CI 或部署生产；SynaRoute 未改。

### 本地日常命令
在 deploy 目录执行（新开终端应能找到 docker；旧终端可用绝对路径 C:\Users\Administrator\AppData\Local\Programs\DockerDesktop\resources\bin\docker.exe）：

```powershell
docker compose -p sub2apicust-local --env-file .env -f docker-compose.override.yml ps
docker compose -p sub2apicust-local --env-file .env -f docker-compose.override.yml stop
docker compose -p sub2apicust-local --env-file .env -f docker-compose.override.yml up -d
```

主题源码修改后需加 --build 重建容器，或另启 Vite 开发预览；不要使用 down -v，它会删除本地测试数据卷。当前未配置自动重启策略，重开 Docker 后若服务没启动，执行 up -d。

## 11. 截图问题修复与真实页面复验（2026-09-26）

### 已完成
- 产品代码只改 frontend/src/custom/theme.css。根因：原 .dark .text-gradient 与 AuthLayout 的 scoped 类具有相同类级权重，懒加载样式更晚覆盖 color；修为 html.dark .text-gradient，指定浅青 color 并清除 background-image。没有绑定构建 hash、深层 DOM 或 !important。
- 输入普通边框/焦点排除 .input-error，补明确错误红边与红色 focus-visible，保留语义而非全部染青。主按钮 .animate-spin 继承前景；真实登录期间实测 disabled=true，文字/加载图标同为 rgb(23,59,56)。该测试仅将登录网络请求延迟900ms后原样发送，未伪造后端响应。
- 当前本地数据库品牌设置：site_name=GPT中转站，site_subtitle=魔法家族；使用原批准素材 mofa-logo-preview.webp，保持核心 M、狮子及皇冠，叠加炭青圆角底板改善浅色可辨识性。单 PNG 360×360 / 42290 字节，输出 output/brand/mofa-logo-local.png（Git 忽略）；通过原 site_logo 设置存为 data URL，不改 Vue、不伪称矢量、不安装新依赖。生产设置未改。
- 两轮最终修正均经 Docker 完整构建：i18n 3/3、vue-tsc、Vite、Go build -tags embed 通过；本轮 Go 单元测试未运行。保留数据库卷，仅重建应用容器。
- 最终容器登录页 computedStyle：夜间 h1=#A1D9CE，background-image=none；日间保留深钛青渐变。错误输入无焦点及有焦点均为 #EF4444；焦点 outline=2px 红色。错误样式测试通过临时 DOM 添加已有 input-error 类并还原，不等于业务错误流程全覆盖。
- 真实已认证 /keys、/usage、/admin/dashboard：双主题，各 360/768/1440×900 无页面横向溢出；观察期间 API 无 HTTP >=400。主题通过持久化后整页加载核验，也点击实际侧栏“深色模式”按钮确认 class 和 localStorage 同步。新手引导仅按 Escape 退出，不改变合规记录。
- 真实创建密钥弹窗正常打开/取消，未创建密钥或真实调用上游。当前库无真实账号/用量，不将空态截图描述为带数据图表或业务全流程验收。
- 截图证据：第 9 节目录 fixed-login-{dark,light}-{1440,768,360}.png、fixed-{keys,usage,dashboard}-{dark,light}-{1440,768,360}.png；repair-verification.json。最终桌面管理员截图等待主题过渡结束，避免把动画中间帧当成文字对比问题。

### 设置接口连带变化（已恢复，不修改上游代码）
- 首次仅 PUT 三个品牌字段后，对比完整 GET 设置发现 oidc_connect_use_pkce / oidc_connect_validate_id_token 意外从 true 变为 false；追踪现有 OIDCSecurityWriteDefaults 与禁用 OIDC 分支后确认局部 PUT 无法将其恢复。
- 已在专用 sub2apicust-local PostgreSQL 中事务更新这两项回 true，并重建应用刷新缓存。之后 GET 对比确认只有 site_name / site_subtitle / site_logo 三项业务设置发生变化；后续更换底板 Logo 再次确认只有 site_logo 改变。未启用 OIDC、未改真实认证设置，也未扩展修上游该逻辑。下次保存后台设置需留意并复核该行为。

### 仍未完成
- 不是概念稿像素级复刻：上游布局、字段、语义分类色与默认深色策略均保留；Logo 仍为位图，独立品牌首页、统一品牌设计规范及 SynaRoute 不在本轮。
- 用户视觉确认、带数据表格/图表、真实上游调用和生产部署尚未完成。没有提交/推送/CI 发布或生产更新。当前修复已在本机 8080 服务生效，可刷新验证。

## 12. 概念布局落地与独立前端验收（2026-09-26）

### 本轮完成与边界
- 新增 `/brand`，不替换根 `/`、不覆盖后台首页配置。认证页保留原业务插槽，加桌面品牌双栏、手机堆叠与主题按钮。
- 原批准 M／狮子／皇冠位图通过 alpha mask 着色，非矢量重绘。品牌名与图标读取原公开设置，SynaRoute 不动。
- AppLayout/AppHeader 仅加样式类，不复制业务页。视觉全在 custom/theme.css，接缝登记见 CUSTOMIZATIONS，维护步骤见 frontend/src/custom/UPGRADE.md。
- 修复长英文无空格站名的移动端面板标题／页脚溢出。

### 已核实与取证
- 定向 Vitest：品牌组件4项、升级契约5项、既有SynaRoute11项，共20项通过；定向ESLint通过。
- pnpm run build：i18n 3/3、vue-tsc、Vite 1070模块通过；既有Browserslist及分块警告未扩大范围处理。
- 实际Vue页面 /brand、/login、/register、/forgot-password：dark/light × 360/768/1440，24组无页面横向溢出，观察无pageerror；后三者输入控件仍分别为2、3、1个。
- 实际点击主题按钮后刷新，html class与localStorage保持light；匿名CTA跳转 /login?redirect=/dashboard，沿用原守卫。英文长站名通过额外注入配置夹具检查，修复后360宽无溢出，非真实后台配置。
- 浏览器API请求隔离返回503，无登录或写数据操作。这是前端降级布局验收，不是后端验收；截图默认Sub2API，因为未连接后台设置。
- 证据目录：C:/Users/Administrator/.codex/visualizations/2026/09/26/01a0ddbc-a03f-7550-93c1-ccbf9a0d3fb9/；frontend-only-verification.json、frontend-only-{brand,login}-{dark,light}-{360,1440}.png。

### 待确认／未验证
- 本节替代第11节“独立品牌首页未做”的历史状态；第11节8080截图是旧镜像，不含本轮布局。
- 用户要求暂不运行Docker；本轮没有Docker命令、容器构建／启动、提交、推送或部署。之后统一部署须重新授权。
- 新控制台样式的已认证带数据复验、真实API调用、完整业务错误态、用户视觉确认仍待完成。契约测试不代表未来上游零冲突。

## 13. 本机Docker部署尝试（2026-09-26 23:28）

- 用户本轮重新授权本机Docker部署；第12节“不运行Docker”是此前限制，不再作为本机操作禁令。生产发布、提交、推送仍未授权。
- 已核实Docker Desktop desktop-linux上下文为本机命名管道。沿用sub2apicust-local、127.0.0.1:8080及原三个数据卷。
- 旧镜像保留标签 sub2apicust:before-brand-20260926（原镜像ID前缀6353bb5bfecb）；PostgreSQL备份 output/brand/before-brand-20260926.dump，595798字节，Git忽略。
- 两次compose build均失败：请求 https://auth.docker.io/token，连接157.240.20.8:443超时，node:24-alpine元数据获取失败。没有进入本轮镜像编译，更没有启动新版应用。日志 output/brand/docker-brand-build.log、docker-brand-build-retry.log。
- 系统代理检查显示ProxyEnable=0；发现clash-verge/verge-mihomo进程，但未验证代理端口可用，也未擅自更改代理、DNS或Docker设置。网络问题的最终根因尚未确认。
- 为备份临时启动过postgres与redis，构建失败后均已正常停止。最终三个原容器均停止，原应用镜像和数据卷保留；不能将此状态称为部署成功。
- 下一步：恢复Docker Hub访问后重试本地构建，成功后再启动应用、检查health、品牌配置与真实页面。带数据业务验收仍未完成。

### 用户再次要求重试
- 2026-09-26第三次构建依然失败，错误与前两次一致：auth.docker.io/token连接157.240.20.8:443超时，node:24-alpine元数据获取失败。证据 output/brand/docker-brand-build-retry2.log。
- 此次仅检查引擎并重试build，没有启动容器、改代理或DNS；现有备份、旧镜像和数据卷均未改动。新版尚未部署；需先恢复仓库网络访问。

## 14. 代理修复与本机新版部署成功（2026-09-26 23:50）

> **部署后用户视觉验收未通过**：当前效果与原稿差距大，控制台仍接近上游换肤。以下构建／健康／空态检查仅证明技术运行，不证明设计还原。后续必须对照原稿与实际大屏视口继续结构性还原，详见HANDOFF最新反馈。

### 网络与构建
- 用户授权修改Docker代理。备份原Docker设置到 output/brand/docker-settings-before-proxy.json；通过Docker Desktop界面设置HTTP/HTTPS代理为 http://127.0.0.1:7897，并排除localhost、127.0.0.1、::1、postgres、redis、host.docker.internal。不改Windows系统代理、DNS或Clash配置。
- 主机经现有Clash授权接口测试HTTP200；仅保存Docker设置后构建仍超时。用户关闭Docker后重新启动，httpproxy.log确认registry连接走7897。
- 再给本次PowerShell构建进程设置HTTP_PROXY/HTTPS_PROXY为同一代理及NO_PROXY后，基础镜像元数据正常取得，完整构建成功。因同时发生重启和客户端代理调整，不声称已单独证明所有失败的唯一根因。临时环境变量未写入系统全局。
- 成功构建日志 output/brand/docker-brand-build-client-proxy.log：i18n 3/3、vue-tsc、Vite及Go build -tags embed通过；本轮未运行Go单元测试。

### 部署与数据
- 使用原compose项目sub2apicust-local执行up -d --no-build sub2api，保留原三个数据卷。仅应用容器重建，PostgreSQL与Redis沿用原容器。
- 新镜像／容器Image一致：sha256:9c8c1dfb87bccab594cdc49a063b85f1bba941ab8604a0a2ce2a0e0de0a027da。三个服务均healthy，应用只绑定127.0.0.1:8080。
- 原回退镜像sub2apicust:before-brand-20260926保留；新增部署前即时数据库备份 output/brand/before-brand-final-20260926.dump（596221字节），保留早前595798字节备份。备份和配置均Git忽略。
- 没有提交／推送／生产发布，没有修改后台设置API或创建业务密钥。

### 真实页面验收
- /health 返回200及status=ok；公开设置返回200，site_name=GPT中转站，M图标保留。
- /brand、/login分别dark/light×360/1440，共8组无横向溢出；新BrandHome和认证双栏存在，不是旧容器或前端fixture。
- 用已有本地管理员正常登录成功，首次引导仅按Escape关闭；未接受新合规条款或绕过认证。
- /keys、/usage、/admin/dashboard分别dark/light×360/1440，共12组无页面横向溢出，业务API观察无HTTP>=400，无pageerror；新工作台样式钩子存在。当前为空数据，未做付费调用、带数据图表或完整业务错误路径验收。
- 截图与记录在 C:/Users/Administrator/.codex/visualizations/2026/09/26/01a0ddbc-a03f-7550-93c1-ccbf9a0d3fb9/：docker-new-*.png、docker-new-verification.json。全部来自真实8080后端，未拦截API。
- 现可打开 http://127.0.0.1:8080/brand 和 http://127.0.0.1:8080/login 由用户确认；原根首页仍未替换。

### 后续本机重建
- 原因分级、网络预检、可复制的进程级代理命令、失败停止与镜像一致性验收已整理到 deploy/LOCAL_DOCKER_RUNBOOK.md；后续重建以此清单为入口。
- 保持Clash的7897端口可用；Docker代理回退可在Resources > Proxies选择原System proxy，并按备份核对。不要整文件覆盖其他后来产生的设置。
- 构建命令所处PowerShell会话设置HTTP_PROXY与HTTPS_PROXY=http://127.0.0.1:7897；NO_PROXY=localhost,127.0.0.1,::1,postgres,redis,host.docker.internal，再执行原本地compose build及up命令。不要把这组本机代理地址带到生产。

## 15. 真实业务结构还原与本机部署（2026-09-27）

### 本轮实际变化
- 对照批准原稿，认证页改为顶部品牌栏＋大屏双栏；控制台改为路径顶栏、独立正文标题与右侧操作、统一筛选／表格／分页面板。不是再次只替换颜色。
- 保留完整管理员导航、全部业务列／操作／权限／事件／API及默认深色；SynaRoute不改。原稿是简化示意，并无同样登录模板，因此不声称像素级复制或删除业务功能来凑图。
- 标题计算从上游页头抽到 custom/brand/useWorkspaceHeading，保留计费模式与自定义菜单权限规则；新增 WorkspaceHeading 和4项测试。接缝逐项登记 CUSTOMIZATIONS 与 custom/UPGRADE.md。
- 实测修正 scoped 样式造成的深色表头蓝灰残留、筛选按钮圆角／底色；仅在 theme.css 使用明确选择器，不改 sticky 定位、语义色或错误边框。

### 已核实的验证
- 7测试文件／61项通过，定向ESLint通过；i18n、vue-tsc与Vite构建通过。日志 output/brand/structure-build-final.log。浏览器数据提醒为既有Browserslist过旧警告，未改依赖。
- Docker构建退出0（含Go embed编译），日志 output/brand/docker-structure-build.log；未运行Go单元测试。
- 新镜像和运行容器均为 sha256:30804c47efc0c5b9b7303ef27203c290fffa97cdacc403eec46677ec7d32eeb3。三个容器healthy，/health返回status=ok，仍仅127.0.0.1:8080。
- 本次备份 output/brand/before-structure-20260927.dump，596871字节；旧9c8c镜像保留为 sub2apicust:before-structure-20260927。之前更早镜像及备份也保留。未改设置API、数据库结构、数据卷或创建业务密钥。
- 新8080正常登录；五业务页 keys/usage/admin dashboard/user dashboard/admin users × 双主题 × 360/768/1440/2560，40组稳定布局无页面横向溢出，每页正文一个h1，观察无pageerror/API>=400。另认证三页同样矩阵24组无横向溢出。
- 120ms过快采样曾出现两次仪表盘过渡瞬时溢出；等650ms完成响应式过渡后40组重跑通过，最终截图等800ms再拍，不拿半动画截图交付。
- 初次引导遮罩曾拦截自动点击；通过正常Escape关闭（不移除DOM绕过），随后创建弹窗打开并取消、列设置11项可见、手机菜单left=0/width=224且跳转usage成功。未验证首次引导完整流程。
- 多行长名称仅在开发端临时DOM探针：50行、scrollHeight5900、clientHeight677、scrollTop300、表头sticky且页面无横溢；刷新即恢复，没有数据持久化，不当作真实带数据验收。
- 实图和JSON位于 C:/Users/Administrator/.codex/visualizations/2026/09/26/01a0ddbc-a03f-7550-93c1-ccbf9a0d3fb9/，文件前缀 deployed-structure-*；来自本机Docker8080，无API拦截。

### 仍需明确的边界
- 用户视觉确认仍待；真实密钥／用量为空，带数据图表、分页和完整错误链路尚未验收，不宣称全部业务完成。
- 上游仍可同步，但增加展示接缝意味着需要合并核对和回归，不保证未来零冲突。
- 未提交、未推送、未生产发布；已有未提交文档全部保留。

## 16. 用户否决后重排认证页（2026-09-27 00:43，本机时间）
- 用户截图指出上一版排版不好，要求重新设计布局；上一版仍属视觉未通过，不将技术通过写成用户认可。
- 仅重排认证展示：1120px统一双栏面板，品牌文案与180px核心M整合在左半区，右半区去掉嵌套卡片并左对齐标题；标题上限42px，手机上下堆叠且隐藏大幅装饰。雾钛青、核心M、默认深色保留；业务页、SynaRoute和认证处理未改。
- 上游只在AuthLayout新增main.mofa-auth-content容器，视觉规则仍全在theme.css；升级契约测试补容器检查，原插槽保留。
- 61项定向测试、ESLint、i18n／vue-tsc／Vite构建通过；日志output/brand/auth-redesign-build.log。Docker构建成功，日志output/brand/docker-auth-layout-build.log，未运行Go单测。
- 当前镜像／容器一致sha256:b4fa2d8e97a1b4690b708cbaa5c9fd06c853ea6956cba1243be2568843e2abf0，三个服务healthy，/health=ok，仅绑定本机8080。旧30804镜像保留为sub2apicust:before-auth-layout-20260927；即时数据库备份before-auth-layout-20260927.dump为597228字节。
- 新8080登录／注册／找回密码×双主题×360/768/1440/2024/2560共30组无横向溢出、无pageerror；使用已有本地管理员实际登录成功跳转/dashboard。注册和找回仅测布局，没有发送邮件／创建账号。
- 最终真实截图docker-auth-redesign-{light,dark}-{360,2024}.png及auth-redesign-verification.json位于既有visualizations输出目录，无API拦截。用户视觉确认仍待；带数据业务验收继续保持未完成状态。未提交、推送或生产部署。

## 17. 多模型品牌名称与扁平Logo（本机2026-09-27 00:58）
- 用户指出顶部银色立体Logo与扁平页面不统一，且业务不只GPT。当前采用主名称「魔法家族」、副标题「多模型 API 服务」。用户尚未最终确认名称和视觉，不将其写成已认可。
- 从原透明M位图导出扁平头像frontend/src/custom/assets/mofa-mark-flat.png；保留狮子／皇冠／M，未生成新标志；512px、16313字节，青色标志／浅雾青背景，小尺寸减少金属细节。
- 只新增素材及更新本机配置，无产品代码改动，无需重建镜像；运行仍为b4fa2d8e97a1。本次未重新跑编译或单测，上一轮61项结果为历史证据。
- 已备份output/brand/before-brand-identity-20260927.dump（607430字节），另完整设置快照settings-before-identity.json和受保护事务脚本apply-brand-identity-local.sql均在output/brand且Git忽略。快照可能含敏感设置，禁止提交或展示正文。
- 通过原后台编辑三项但保存前拦截审查：发现payment_enabled_types、default_platform_quotas发生非品牌变化，其余部分仅对象键顺序差异。未让该完整请求落库，失败提示属于被主动阻止的保存，不是已生效。
- 最终经授权执行本机SQL事务，先校验三项旧值MD5，再仅UPDATE site_name/site_subtitle/site_logo；事务对照其他277行key/value完全一致，任意校验失败会回滚。应用单容器重启清缓存，不重启数据库或修改数据卷。此为一次本地维护操作，不新增应用接口，不建议以后盲目复制旧值校验脚本。
- 核验：完整管理员设置响应304项比较，仅site_logo/site_name/site_subtitle变化；实际登录成功，侧栏为魔法家族，页签「登录 - 魔法家族」／「API 密钥 - 魔法家族」；/health=ok。登录双主题×360/1868四组无横向溢出且Logo加载成功。
- 真实截图identity-login-*、identity-lockup-*及identity-verification.json在既有visualizations目录。最终截图未拦截API；早前仅中止审查用设置PUT请求。
- 未修改SynaRoute、支付、额度、认证逻辑、默认深色或生产；未提交／推送。生产若需采用新名称与Logo，另行授权配置三项，不把本机已生效说成生产已更新。

## 18. 阶段性收口与提交授权
- 用户明确「暂时先这样」，本轮保留当前布局与多模型品牌，不继续重选风格或新增开发；授权提交、推送自己的fork、更新待办及复制新Logo，不授权生产部署。
- 新Logo已复制到C:/Users/Administrator/Pictures/mofamily/mofa-mark-flat.png（16313字节），SHA256与源文件一致，旧Logo未覆盖。
- 当前待办统一到HANDOFF第五节：本轮CI／镜像核验、带数据业务验收、首次引导与完整交互、用户决定生产升级及独立配置品牌、后续SynaRoute独立任务。旧的视觉否决记录保留为过程证据，不再当作最新指令。
- 提交仅含品牌源代码、素材、测试与交接；本机设置、SQL、快照、备份和日志保持Git忽略。当前本机镜像无需因文档提交而再次重建；推送可触发仓库现有CI与GHCR构建，不代表生产升级。
- 收口提交前复跑：7个测试文件61项通过，定向ESLint通过，i18n／vue-tsc／Vite构建通过；日志output/brand/precommit-build.log（Git忽略）。保留既有Browserslist过期和大chunk警告，不为提交收口扩改依赖或拆包。Go单测与本轮远端CI仍待单独核验。
