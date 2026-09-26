# AGENTS.md — 给接手本仓的 AI 编码代理（Codex 等）

本仓是开源 sub2api 的**定制 fork**。**开工前必读、收工前必更**以下交接文件（与 `CLAUDE.md` 同源，Codex 也照此遵守）：

- **[HANDOFF.md](HANDOFF.md) — 交接中枢：用户要求 / 准则 / 当前状态 / 待办。先看这里。**
- [CLAUDE.md](CLAUDE.md) — 项目规则与交接协议。
- [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md) — 定制唯一权威清单（改上游文件必登记）。
- [SYNC.md](SYNC.md) / [deploy/UPDATE_GUIDE.md](deploy/UPDATE_GUIDE.md) — 同步上游 / 部署 / 日常升级。

## 核心准则（详见 HANDOFF「三」）
- 🔴 只说有证据的话，没证据就说「我不知道」；回答分「已核实 / 推断 / 不知道」并附取证方式。
- 全程中文。
- 定制优先新增文件；改上游文件必打 `[CUSTOM]` 注释并登记 `CUSTOMIZATIONS.md`；配置走 gitignore。
- 本机无 Go：后端改动未编译，靠 CI 验证（`go build -tags embed` + `go test -tags=unit`）。
- **换肤只改 `frontend/src/custom/theme.css`**；上游 `teal-*` 语义分类色刻意不改。

## 当前进行中
本机Docker重建前必读 [deploy/LOCAL_DOCKER_RUNBOOK.md](deploy/LOCAL_DOCKER_RUNBOOK.md)：先确认现有代理可用，再给实际构建进程设置代理；失败先排查，不盲目重复构建，不把旧镜像页面当新版验收。
用户已确认雾钛青，保留核心M与默认深色，SynaRoute不动。先读BRAND_IMPLEMENTATION第16节与HANDOFF最新有效状态：用户再次否决登录布局后，2026-09-27已重排为统一双栏主体；本机8080运行b4fa2d8e97a1，三服务healthy，61项测试与30组真实认证布局、实际登录通过。用户视觉确认及带数据业务验收仍待。未提交／推送／生产部署，旧镜像和数据库备份保留。样式只在theme.css，AuthLayout仅新增共用主体容器，接缝见CUSTOMIZATIONS与custom/UPGRADE.md。

最新品牌补充（本机2026-09-27 00:58）：先读BRAND_IMPLEMENTATION第17节。本机站点名现为「魔法家族」，副标题「多模型 API 服务」，Logo为新增mofa-mark-flat.png的data URL；保留原M／狮冠。仅三键配置变化，其余设置已比对不变；镜像仍b4fa2d8e97a1，生产未改。名称与视觉待用户确认，不再把GPT中转站当当前站名。

本轮收口决定优先于以上历史状态：用户已表示「暂时先这样」，授权提交并推送到origin/main、更新待办和复制Logo。保持当前视觉与多模型品牌，不自动再做设计；生产不部署。待办以HANDOFF第五节为准，CI结果要检查本轮提交，不能沿用旧结果。Logo副本在C:/Users/Administrator/Pictures/mofamily/mofa-mark-flat.png。
