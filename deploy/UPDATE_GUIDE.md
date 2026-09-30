# 日常更新操作手册（定制版 sub2api）

## 2026-09-30 家族视觉与下载入口更新（最新）

功能提交 `25071d4d547c69d6c933bbcb0c276585b23b4b68` 已推送 origin/main：工坊三张等宽图片与鼠标／点按／键盘置顶、最新深发人像、家族首页独立皇冠图标、SynaRoute 官网与下载链接，以及发布地址配置说明。只包含 16 个家族相关文件，不包含工作区包号专项，未新增跨站 SSO 或一体化产品托管。实际暂存快照 72 项／8 文件、类型／相关 lint／Vite 构建通过；真实生产登录与模型授权仍需用户验收。

**记录时发布尚未完成**：[镜像构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36736764218)在 Build and push、[CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36736764231)的 Go Unit 仍运行；[安全扫描](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36736764322)以及 CI 的前端／Go lint／shell／release-helpers 已成功，不标全部通过，不把旧提交的成功结果沿用到本轮。待三项成功并核对镜像任务日志中实际标签／摘要，再备份数据库和部署配置、记录旧标签后执行：

```bash
cd /sub2api-deploy
./update.sh sha-25071d4
```

上面是本轮功能提交的预期标签，不代表构建中已能拉取；后续交接文档 `[skip ci]` 提交不生成替代镜像。三项 `MAGIC_RECIPES_URL`／`MAGIC_RECIPES_ORIGIN`／`MAGIC_STUDIO_URL` 仓库变量名称只读查询仍为空，配方／工坊入口保持待配置，需按 `frontend/src/custom/FAMILY_PORTAL.md` 设置实际已托管地址并重建。上线后验证 `/health`、三图及交互、皇冠与仪表盘图标区分、SynaRoute 公开链接、真实登录后的家族首页／API 控制台往返；代理不代操作生产。

## 2026-09-30 家族产品与接入教程版本

功能提交 `321fe9943b700384b10c2fc1368ed24c6b90eb97` 已推送origin/main：保留品牌首屏，新增互动产品展台／控制台产品坞、配方模型设置与显式登录配置、工坊独立源码及接入教程。未混入工作区待确认的包号专项修复；本机8080此前整工作区镜像不等于本次发布范围。

**发布门槛**：[本轮镜像构建](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36712632118)已成功，实际标签 `ghcr.io/engineemomo/sub2apicust:sha-321fe99`、摘要 `sha256:c8dd5b34edacf29eec9e30f53f35f5de8c5898706b86715b81d70b8a73db4147`，从成功任务日志核验。[本轮CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36712632123)的Go Unit、frontend、Go lint、shell、release-helpers及安全扫描成功，集成仍运行；先确认集成成功，不把镜像发布等同全部CI通过。实际暂存源码已通过210项原站、54项配方及21项工坊测试与类型／lint／Vite构建。

集成成功后，备份数据库与部署配置，记录当前标签，在服务器执行（后续仅文档的skip-ci提交不生成新镜像，仍使用此功能标签）：

```bash
cd /sub2api-deploy
./update.sh sha-321fe99
```

独立配方／工坊不随原站镜像自动托管。仓库 `MAGIC_RECIPES_URL`、`MAGIC_RECIPES_ORIGIN`、`MAGIC_STUDIO_URL` 尚未配置（只读API核对变量名称），本次生产入口将显示待配置并禁用；需先单独托管、确认HTTPS或同源路径、设置公开构建变量后重建。独立配方还需精确API CORS，不放松认证或来源限制，不把本机4178／4179地址用于生产。具体见 `frontend/src/custom/FAMILY_PORTAL.md`。

更新后核查 `/health`、`/home`、`/family`、`/api`、`/guide`及真实登录后的产品切换。真实Key授权／计费及用户视觉批准仍待，不自动调用收费模型；后台自定义 `home_content` 仍可能覆盖默认首页。代理未操作生产。

## 2026-09-30 多人包号与控制台版本

代码 `43b715a15f599f5ae2125bc923849d57e4b3522c` 已推送；GHCR及安全扫描成功，镜像 `ghcr.io/engineemomo/sub2apicust:sha-43b715a` 已生成（摘要 `sha256:dc36b117e524262faddf0c13f0dad5c9ff510e385e69ddc2860145a7ea6566af`）。

**发布门槛**：记录时[本轮CI](https://github.com/EngineeMoMo/sub2apicust/actions/runs/36657101544)的集成测试仍在运行，Go单测、Go lint、前端及其他任务通过；请先确认该CI全部成功，不把镜像成功当作全量测试通过。

通过后先备份数据库和部署配置，再在服务器执行：

```bash
cd /sub2api-deploy
./update.sh sha-43b715a
```

本次新增迁移242（多成员与删除隔离），保持一账号一专属分组的独占规则。上线后检查登录隐藏游客入口、FAQ右侧正文、包号名称／多人／改绑／撤销后删除及任务费用显示。不要直接回滚旧单用户镜像；详细规则与回退边界见[包号排查说明](DEDICATED_TROUBLESHOOTING.md)。生产尚未由代理操作。

## 2026-09-28 游客功能已验证版本

代码提交 `fa45b2023c01e5c84084dcbcb32d5a7c5bbe213a` 的 CI（含Go单测／集成）、安全扫描及GHCR构建全部通过，发布镜像 `ghcr.io/engineemomo/sub2apicust:sha-fa45b20`。本次包括公开官网／真实套餐／FAQ、游客控制台及宽屏手机适配；不开放个人数据接口。

服务器由用户操作：先备份数据库与部署配置，并记录当前镜像标签，再执行：

```bash
cd /sub2api-deploy
./update.sh sha-fa45b20
```

更新后检查 `/health`、首页游客入口、`/preview`、`/plans`、`/faq` 和登录回跳；如配置了自定义 `home_content`，其优先级仍保留。未执行生产更新或真实支付验收。后续仅文档提交不生成新镜像，不能替换此处已验证标签。

> [CUSTOM] **Windows本机源码构建**另见 [LOCAL_DOCKER_RUNBOOK.md](LOCAL_DOCKER_RUNBOOK.md)：已记录Docker Hub授权超时与代理预检。本文下方仍是服务器镜像更新流程；不要把本机127.0.0.1:7897代理配置带到生产。

> 速查卡：以后升级 / 回滚照这份敲即可。完整部署、首次平移、数据备份/恢复见 [DEPLOY_CUSTOM.md](DEPLOY_CUSTOM.md)。
>
> 你的环境（已确认）：部署目录 `/sub2api-deploy`，compose 用 `docker-compose.local.yml`，
> 定制镜像 `ghcr.io/engineemomo/sub2apicust`（owner 是 EngineeMoMo，中间**两个 e**），DB 用户/库均 `sub2api`。

---

## 更新分两步

### 第一步：出新镜像（代码侧，一般由维护者 / AI 协助）
上游出新版、或要改定制（换色 / 加功能）时 → 在 fork 仓库合并或改代码 → `git push` → CI 自动构建新镜像推到 GHCR。
（需要 Go 编译校验，本机若无 Go 就靠 CI。）构建完成后你会拿到新的镜像标签 `sha-xxxxxxx`。

### 第二步：服务器部署（你操作）
进部署目录：
```bash
cd /sub2api-deploy
```
按 override 里 image 钉的是 `latest` 还是 `sha-xxxx`，选一种模式：

**模式 A ｜ 用 latest（省心，日常一条命令）**
override 的 image 是 `...:latest` 时：
```bash
./update.sh
```
每次自动拉最新定制镜像并重建。

**模式 B ｜ 钉 sha（可控，当前就是这个）**
override 钉了固定 `sha-xxxx` 时，不带参数的 `./update.sh` 只会重拉同一版、**不升级**。
升级到新版本要带新 sha（维护者会给你）：
```bash
./update.sh sha-<新提交>
```
脚本会自动切 override 标签 + 备份原文件，再拉取重建。

**从模式 B 切到模式 A**（以后只敲 `./update.sh`）：把 override 重写为 latest：
```bash
cat > /sub2api-deploy/docker-compose.override.yml <<'EOF'
services:
  sub2api:
    image: ghcr.io/engineemomo/sub2apicust:latest
    environment:
      - DISABLE_ONLINE_UPDATE=true
EOF
```

`update.sh` 每次都会：拉镜像 → 重建容器 → **启动时自动跑数据库迁移** → 等 `/health` 就绪 → 清理悬空旧镜像。

---

## 回滚
切回上一个正常的版本：
```bash
cd /sub2api-deploy && ./update.sh sha-<上一个正常的提交>
```
迁移是前向增量的，回滚镜像通常安全（旧代码忽略新列）；只有怀疑迁移本身出问题才用数据库备份恢复，见 DEPLOY_CUSTOM.md 第七节。

---

## 每次升级后验证
```bash
docker compose -f docker-compose.local.yml -f docker-compose.override.yml exec -T sub2api /app/sub2api --version
```
- 版本号是你要的 `0.2.8-custom.xxx`；
- 浏览器：主题正常、老数据都在、后台**无「立即更新」按钮**；
- 设过 logo / 站点名的，检查显示正常。

---

## 常见坑
- **`pull` 报 not found**：多半是镜像 owner 拼错——正确是 `engineemomo`（EngineeMoMo，中间**两个 e**），别写成 `enginemomo`。
- **`./update.sh` 跑了但版本没变**：override 钉了固定 sha（模式 B），要 `./update.sh sha-<新>` 或切 latest（模式 A）。
- **`pull` 报 unauthorized / denied**：GHCR 登录过期 → `docker login ghcr.io -u EngineeMoMo`（用有 `read:packages` 的 PAT；**别用 `echo <token> |` 的写法**，token 会进 shell history）。
- **升级前想备份数据库**：
```bash
docker compose -f docker-compose.local.yml exec -T postgres pg_dump -U sub2api sub2api | gzip > backup-$(date +%F).sql.gz
```
