# 同步上游 & 构建部署 Runbook

本 fork 的目标：**保留定制的同时能同步上游 sub2api 更新。**
定制策略为「配置 + 新增功能」，尽量不改上游核心逻辑。

- 上游（只读）：`upstream` → https://github.com/Wei-Shaw/sub2api.git
- 你的私有仓：`origin` → （建仓后填）
- 定制清单：见 [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md)

---

## 0. 一次性初始化（已完成的部分打勾）

- [x] `git clone` 上游为基线
- [x] `git remote rename origin upstream`（上游只读）
- [x] `git remote set-url --push upstream DISABLED_NO_PUSH_TO_UPSTREAM`（防误推上游）
- [x] `git config rerere.enabled true`（记住冲突解法）
- [ ] 建私有仓并设为 `origin`：
  ```bash
  # 方式一：已有私有空仓
  git remote add origin <你的私有仓地址>
  git push -u origin main

  # 方式二：用 gh 一条命令建私有仓并推送（需已登录 gh）
  gh repo create <名字> --private --source=. --remote=origin --push
  ```
- [ ] 在 GitHub 私有仓 Settings → Actions 确认 Actions 已启用（GHCR 构建用内置 `GITHUB_TOKEN`，无需额外密钥）

---

## 1. 日常：同步上游更新

```bash
git fetch upstream
git checkout -b sync-test main          # 先在测试分支试合，别直接动 main
git merge upstream/main
```

- 有冲突：解冲突（rerere 会自动套用你以前的解法），然后 `git commit`。
- 解完后**打开 [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md)，逐条核对定制是否还在**。
- 跑自检（见 CUSTOMIZATIONS.md 末尾清单）：build + unit test 通过。

验证通过后合回 main 并推送：

```bash
git checkout main
git merge sync-test
git branch -d sync-test
git push origin main                     # push 会触发 CI 构建镜像（见下）
```

## 2. 构建镜像（CI 自动 / 手动）

`.github/workflows/custom-image.yml` 会在 **push 到 main** 或 **手动触发** 时，
用根 `Dockerfile` 构建并推到 `ghcr.io/<你的用户名或组织>/<仓库名>`。

标签规则：
- `latest` —— main 最新
- `sha-xxxxxxx` —— 每个 commit（可精确回滚）
- 打 tag `v*` 时 —— 该版本号

手动触发：私有仓 → Actions → “Build custom image (GHCR)” → Run workflow。

> GHCR 私有镜像默认需要登录才能 pull。首次在服务器上：
> ```bash
> echo <GHCR_PAT> | docker login ghcr.io -u <你的用户名> --password-stdin
> ```
> PAT 需要 `read:packages` 权限。

## 3. 部署 / 升级（服务器上）

服务器的部署目录里放 `deploy/docker-compose.yml`（或 `.local.yml`）+ 你的
`.env` + `docker-compose.override.yml`（override 把 image 指向你的 GHCR 镜像，
见 `deploy/docker-compose.override.yml.example`）。

```bash
cd <部署目录>
docker compose pull                      # 拉你 CI 刚推的定制镜像
docker compose up -d
docker compose logs -f sub2api           # 冒烟
```

## 4. 🔴 红线（务必遵守）

1. **App 内「在线更新」已被关闭**（`DISABLE_ONLINE_UPDATE=true`，在 `docker-compose.override.yml` 里）。
   它原本拉的是**上游**预编译产物，会覆盖你的定制镜像。关闭后：前端不再出现「立即更新」按钮，
   后端也拒绝应用/回滚。定制版一律走「本 runbook 的 merge + CI 构建 + compose pull」升级。
   如需把更新检查指向自己的仓库，设 `UPDATE_GITHUB_REPO=your-org/your-repo`（且需你自己发 Release）。
2. **密钥不进本 fork。** `.env` / `config.yaml` 里的 `JWT_SECRET`、`POSTGRES_PASSWORD`、
   `TOTP_ENCRYPTION_KEY` 等已被 gitignore；单独放私有 ops 仓或加密备份。
3. **Go 版本硬钉 1.27.0**（`backend/go.mod` + 3 个 Dockerfile + CI 断言）。上游升级时随 merge
   带过来即可；**你自己若改 Dockerfile，记得同步这几处**，否则 build 时才炸。
4. **改 ent/schema 要 `go generate ./ent` 并提交生成码；加前端依赖要提交 `pnpm-lock.yaml`。**
