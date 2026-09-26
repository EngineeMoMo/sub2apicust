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
见 HANDOFF「五、待办」——最新任务是「品牌配色微调（减压抑）」，第一版已提交，待微调 + 用户最终确认。
