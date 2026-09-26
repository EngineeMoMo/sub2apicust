# 日常更新操作手册（定制版 sub2api）

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
