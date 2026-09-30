# 项目规则 — sub2api 定制 fork（本仓专属；通用规则见 ~/.claude/CLAUDE.md）

## 交接协议（每个会话必走）
- **开工先读**：[HANDOFF.md](HANDOFF.md)（交接中枢：用户要求 / 准则 / 当前状态 / 待办）、
  [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md)（定制唯一权威清单）、[SYNC.md](SYNC.md)（同步 / 构建 / 部署 runbook）。
- **收工前更新** HANDOFF.md 的「四、当前状态 / 五、待办」；任何新定制同时登记 CUSTOMIZATIONS.md。
- 交接走文件，别把状态只留在会话里。

## 本仓最关键的几条（详见 HANDOFF.md「三、用户准则」）
- 2026-09-30发布进展：功能43b715a15已推送、镜像sha-43b715a与安全扫描成功；CI36657101544集成测试仍待终态，其余步骤已通过。不要把下方授权／未推送历史当现状，也不要把镜像成功等同全部CI通过。生产未部署。
- 2026-09-30最新收口：用户确认包号继续严格独占并授权提交推送origin/main，生产仍由用户自行发布；本轮CI／GHCR需按新提交查证，product-samples不混入提交。
- 2026-09-30用户后续授权已部署本机8080，镜像be2f26f98e8c，三服务健康，迁移242已生效；此前“仅隔离测试”属历史。证据output/console-deploy-20260930，生产未改、未提交推送，详见HANDOFF最新状态。
- 2026-09-30未发布改动：多人包号迁移242保留旧Key隔离标记，不可删标记绕过403或直接回退旧单用户版本；规则与诊断见 `deploy/DEDICATED_TROUBLESHOOTING.md`，进展以HANDOFF四／五为准。本轮Docker仅用于隔离编译与测试，不等于8080或生产已更新。
- 🔴 只说有证据的话，没证据就说「我不知道」；回答分「已核实 / 推断 / 不知道」并附取证方式。
- 全程中文。
- 定制优先新增文件；改上游文件必打 `[CUSTOM]` 注释并登记 CUSTOMIZATIONS.md；配置走 gitignore。
- **本机无 Go**：后端改动未编译，靠 CI 验证（`go build -tags embed` + `go test -tags=unit`）；提交时如实标注。
- 同步上游先 `git merge-tree` 演算再合并（rerere 已开），合并后过自检清单 + 冒烟。
- 换肤只改 `frontend/src/custom/theme.css`；上游 `teal-*` 语义分类色刻意不改。
