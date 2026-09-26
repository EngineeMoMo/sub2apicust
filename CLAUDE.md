# 项目规则 — sub2api 定制 fork（本仓专属；通用规则见 ~/.claude/CLAUDE.md）

## 交接协议（每个会话必走）
- **开工先读**：[HANDOFF.md](HANDOFF.md)（交接中枢：用户要求 / 准则 / 当前状态 / 待办）、
  [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md)（定制唯一权威清单）、[SYNC.md](SYNC.md)（同步 / 构建 / 部署 runbook）。
- **收工前更新** HANDOFF.md 的「四、当前状态 / 五、待办」；任何新定制同时登记 CUSTOMIZATIONS.md。
- 交接走文件，别把状态只留在会话里。

## 本仓最关键的几条（详见 HANDOFF.md「三、用户准则」）
- 🔴 只说有证据的话，没证据就说「我不知道」；回答分「已核实 / 推断 / 不知道」并附取证方式。
- 全程中文。
- 定制优先新增文件；改上游文件必打 `[CUSTOM]` 注释并登记 CUSTOMIZATIONS.md；配置走 gitignore。
- **本机无 Go**：后端改动未编译，靠 CI 验证（`go build -tags embed` + `go test -tags=unit`）；提交时如实标注。
- 同步上游先 `git merge-tree` 演算再合并（rerere 已开），合并后过自检清单 + 冒烟。
- 换肤只改 `frontend/src/custom/theme.css`；上游 `teal-*` 语义分类色刻意不改。
