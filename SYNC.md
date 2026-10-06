# 同步上游 & 构建部署 Runbook

2026-10-06本地同步v0.2.13：稳定tag3040209f2＋随后0.2.13 VERSION元数据；合并保留WAP扫码、游客目录与上游verify限流。新增余额在途预占必须跳过已认证包号凭证，否则零余额包号并发被误拒；接缝与回归见CUSTOMIZATIONS。新老241迁移按完整文件名并存，独立PG新库／升级／幂等通过，禁止重命名既有已应用迁移。本轮尚未推送或部署，完整integration已按用户授权补跑通过（含安全审计专用PG／Redis补测），6项既有／外部依赖跳过及新SHA CI仍待，不沿用历史CI结果。

2026-10-06支付宝安全补强同步要求：保留CUSTOMIZATIONS登记的官方通知严格字段／配置公钥／总额、查询订单绑定、精确分金额及新扫码HTTPS／可信站点规则；不得恢复金额回退或用Referer扩大可信Host。路由Guard在限流前禁止缓存，302正文为空。运行TestCustomAlipay及支付履约回归，真实仓储并发测试也要保留；发布与真实扫码边界见deploy/ALIPAY_SECURITY_REVIEW.md。生产由用户部署，不沿用历史CI结论。

2026-10-06我的订阅提示源码完成未发布：同步保留SubscriptionsView的custom卡片与独立列表／重置时间读取、原progress接口真实嵌套结构和服务端resets_at，不恢复本地固定小时推算；15秒共享显示时钟、标签返回更新与卸载清理、过期实际日期及最后周期结束提示保持。样式仅theme.css，原平台语义色保持；Makefile增加subscription-timing.spec.ts，本轮27文件399项、类型／lint／Vite及深浅／手机浏览器模拟夹具通过。未改后端或实际部署，真实账号及本次代码发布门禁待用户要求发布后核验，不能借旧CI／镜像说明线上已改；详见CUSTOMIZATIONS、custom/UPGRADE和HANDOFF。

2026-10-06电脑端支付宝 WAP 扫码已完成源码，未发布。同步时保留默认关闭的桌面专用开关、官方实例 WAP 分支、订单分钟期限与 qrcode snapshot、32字节令牌免登录短入口及官方目标校验；手机两项旧配置、通知幂等、个人订单认证不得覆盖。按CUSTOMIZATIONS最新节逐处检查，运行新 `TestCustomAlipayDesktopWap` 与Makefile前端关键回归。前后端同版构建，无新迁移；关闭设置只影响新订单，旧二维码仍依订单期限校验。Windows全量两项已知失败不算CI成功，发布时核验本次实际代码SHA的unit／integration、安全和镜像；真实App／小额实付待用户按 [验收说明](deploy/ALIPAY_DESKTOP_WAP.md) 检查。已有更新脚本专项保留，生产由用户部署。

2026-10-02配方图片打开修复已发布：功能4bf1436fb同SHA CI／安全／GHCR全绿，固定镜像sha-4bf1436，摘要见HANDOFF，含上一版包号免扣。同步保留自有site的dialog原图按钮、Base64下载入口与材料／连接失效清理；不要恢复data地址的新窗口链接。运行配方四组Node测试和build-products.test.mjs，并检查独立及同源内嵌页的打开／Esc／窄屏。此轮无新上游接缝或CSP变更，生产由用户备份更新。

2026-10-01包号免扣版已发布：功能acd73d716的CI（unit／integration、真实PG、lint、前端372项）、安全与GHCR全绿，使用固定镜像sha-acd73d7，详见[更新手册](deploy/UPDATE_GUIDE.md)。同步上游须保留迁移243、私有准入凭证、独立参考计量、共享限制与WS帧保护，并实际执行带DEDICATED_TEST_POSTGRES_DSN的PG回归；普通计费与已有多人多组隔离一并验。生产由用户备份后更新，更新脚本专项仍未提交，不用纯文档提交SHA猜镜像。

本 fork 的目标：**保留定制的同时能同步上游 sub2api 更新。**
定制策略为「配置 + 新增功能」，尽量不改上游核心逻辑。

- 上游（只读）：`upstream` → https://github.com/Wei-Shaw/sub2api.git
- 你的私有仓：`origin` → https://github.com/EngineeMoMo/sub2apicust.git
- 定制清单：见 [CUSTOMIZATIONS.md](CUSTOMIZATIONS.md)

---

## 0. 一次性初始化（已完成的部分打勾）

- [x] `git clone` 上游为基线
- [x] `git remote rename origin upstream`（上游只读）
- [x] `git remote set-url --push upstream DISABLED_NO_PUSH_TO_UPSTREAM`（防误推上游）
- [x] `git config rerere.enabled true`（记住冲突解法）
- [x] 建私有仓并设为 `origin`（已完成：EngineeMoMo/sub2apicust，已 `push -u origin main`）：
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
