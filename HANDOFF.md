# 交接文档（HANDOFF）— sub2api 定制 fork

> **本文件是本项目的交接中枢。** 每个接手的智能体（含 Claude）：
> - **开工前必读**：本文件 + [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md) + [SYNC.md](SYNC.md)。
> - **收工前必更**：更新下方「四、当前状态」「五、待办」；任何新定制同时登记 CUSTOMIZATIONS.md。
> - 交接一律走本文件——**别把状态只留在会话里**（会话会丢，文件不会）。
>
> 最后更新：2026-09-26。

---

## 一、用户是谁 / 怎么协作
- 语言：**全程中文**（回复、文档、注释）。
- 自部署运维者，目标：把开源 sub2api 定制成自有品牌的商用「GPT中转站」，并长期跟随上游。
- 用户**手动部署**；智能体负责改代码、验证、提交推送，**不代替用户上生产**。

## 二、用户要求（项目意图）
1. **核心**：定制 sub2api 的同时能持续同步上游更新、且定制不丢。
   路线 = **方案 A：配置 + 新增功能**，尽量不动上游核心 → 私有 fork + override 构建。
2. **镜像**：GHCR + GitHub Actions CI（push main 自动构建）。
3. **已交付定制**（细节见 CUSTOMIZATIONS.md）：
   - SynaRoute 深链接一键导入（`synaroute://`，镜像上游 CCS 导入，带管理员开关 `hide_synaroute_import_button`）。
   - 关闭 App 内「在线更新」（`DISABLE_ONLINE_UPDATE=true`；可选把更新源指向自己仓库 `UPDATE_GITHUB_REPO`）。
   - 前端定制叠加层 `frontend/src/custom/`（换肤 + 新增/影子替换页面，零冲突）。
   - **品牌换肤**：电光蓝主题（`--color-primary-500=#1877f0`）+ 默认深色；logo/站点名走**管理员后台设置**（无需改码）。
4. **品牌**：魔法家族 / GPT中转站；域名 https://ai.mofamilys.com ；银色狮子+皇冠「M」logo；深蓝底 + 电光蓝发光风。
5. **升级方式**：**手动 Docker 更新**（`deploy/update.sh`），不用 App 内按钮。

## 三、用户准则（红线，务必遵守）
1. 🔴 **只说有证据的话；没证据就说「我不知道」**（2026-09-07 用户原话：「以后都拿证据事实说话，不要理论和以为，不确定的就不回答，找到实质性证据再回答」）。
   - 回答分栏：**已核实**（附取证方式：grep 命中数 / 读了哪个文件哪几行 / 命令完整输出）/ **推断**（附「要什么证据才能确认」）/ **不知道**。
   - 「代码里存在能产生该症状的路径」≠「用户遇到的就是它」；用户报的现场要在**那台机器**上取证。
   - 写「已核对」必须同时写核对方式，否则别写那三个字。
2. **定制维护**：优先新增文件；不得不改上游文件时改动尽量小并打 `[CUSTOM]` 注释，且**逐处登记 CUSTOMIZATIONS.md**。配置类定制（.env / override）走 gitignore，不进仓。
3. **本机无 Go 工具链** → 后端改动**未编译**，靠 CI 验证（`gofmt -w ./...` + `go build -tags embed ./cmd/server` + `go test -tags=unit ./...`）；提交时如实标注「未编译」。
4. **同步上游**：先 `git merge-tree --write-tree main upstream/main` 演算冲突，再真合并；rerere 已开；合并后过 CUSTOMIZATIONS.md 末尾自检清单 + 登录发一个请求冒烟。
5. **换肤**：品牌主色只改 `frontend/src/custom/theme.css`；上游组件里的 `teal-*` 若是**语义分类色**（deepseek、模型徽标、渠道状态）**刻意不改**（改了会撞色/丢语义）。
6. 会话上下文超 **200k** 先压缩再继续，压缩后先核对工作树再接着做。

## 四、当前状态（每次收工更新）
- 分支 `main`，最新提交 `83457c64a`（品牌换肤：电光蓝 + 默认深色），已推 `origin/main`；工作树干净。
- fork 接线：`origin`=https://github.com/EngineeMoMo/sub2apicust.git ；`upstream`=Wei-Shaw/sub2api（**push 已禁用**，只读）。
- 上游同步：落后本地上游快照 **207 提交**（约 0.2.8 era）。**2026-09-26 实测** `git merge-tree` 合并演算 = **零冲突 + 11 处关键定制全存活**。
  - ⚠️ 该演算针对**上次拉到的上游快照**；当日 github.com 从本机**拉不动**（连接重置/超时），**最新上游未复验**。
- 前端：`pnpm run build` 通过（typecheck + i18n + vite）。本机预览：`pnpm -C frontend run dev` → :3000（无后端时仅外观预览，数据页会跳登录）。
- 后端 SynaRoute + 关在线更新改动：**仍未编译**，待 CI 的 Go 构建/单测。

## 五、待办 / 下一步
- [ ] 网络恢复后 `git fetch upstream` 重拉 → 重跑 merge-tree 复验最新上游 → 正式 `git merge upstream/main`（0.2.8）→ 推送触发 CI。
- [ ] 后端改动经 CI 的 Go 编译 + 单测验证。
- [ ] 用户按 `deploy/DEPLOY_CUSTOM.md` 部署新镜像；后台设置站点名 + 上传 logo（建议透明底 PNG/SVG，≤300KB，约 80×80）。
- [ ] SYNC.md 里「更新源指向自己仓库」（`UPDATE_GITHUB_REPO` + 自发 Release）尚未启用，按需再做。

## 六、文档地图
- [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md) — **定制唯一权威清单**（新增文件 / 接线改动 / 行为修改 / 自检清单）。
- [SYNC.md](SYNC.md) — 同步上游 + 构建镜像 + 部署 runbook + 4 条红线。
- [deploy/DEPLOY_CUSTOM.md](deploy/DEPLOY_CUSTOM.md) — 定制镜像部署清单（登录/起服/升级/回滚/迁移）。
- [frontend/src/custom/README.md](frontend/src/custom/README.md) — 前端叠加层三种用法 + 影子替换代价。
