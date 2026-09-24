# 定制版部署清单（GHCR 镜像）

本清单用于部署 **你自己的定制镜像**（不是上游 `weishaw/sub2api`）。
镜像由私有仓的 `.github/workflows/custom-image.yml` 在 push 到 `main` 时构建并推送到：

```
ghcr.io/enginemomo/sub2apicust        # 镜像名全小写
```

标签：`latest`（main 最新）、`sha-<7位提交>`（可精确回滚）、`vX.Y.Z`（打 tag 时）。

> 前提：仓库 Actions 里 `custom-image.yml` 已跑绿（Package 已生成）。首次可在
> 仓库 → Packages 确认 `sub2apicust` 包存在。

---

## 一、服务器一次性准备

**1. 装 Docker + Compose v2**（Docker 20.10+ / Compose v2+）。

**2. 登录 GHCR**（私有镜像 pull 前必须登录一次）：
先在 GitHub 建一个 **classic PAT**，勾选 `read:packages`（仅读包即可）。然后：

```bash
echo <你的PAT> | docker login ghcr.io -u EngineeMoMo --password-stdin
```

**3. 建部署目录并取三个文件**（用本仓 `deploy/` 里的）：

```bash
mkdir -p ~/sub2api-deploy && cd ~/sub2api-deploy
# 从你的私有仓 raw 取（把 main 换成你要部署的分支/tag）
BASE=https://raw.githubusercontent.com/EngineeMoMo/sub2apicust/main/deploy
curl -sSL $BASE/docker-compose.local.yml -o docker-compose.local.yml
curl -sSL $BASE/.env.example -o .env
curl -sSL $BASE/docker-compose.override.yml.example -o docker-compose.override.yml
mkdir -p data postgres_data redis_data
chmod 600 .env
```

> `docker-compose.override.yml` 已被 .gitignore 排除，只存在于服务器本地，含镜像指向与开关。

---

## 二、改两个文件

**1. `docker-compose.override.yml`** —— 把镜像指到你的 GHCR，并确认已关在线更新：

```yaml
services:
  sub2api:
    image: ghcr.io/enginemomo/sub2apicust:latest   # 或钉某个 sha-xxxx / vX.Y.Z
    environment:
      - DISABLE_ONLINE_UPDATE=true
```

**2. `.env`** —— 至少改这几项（其余保持默认即可）：

```bash
POSTGRES_PASSWORD=<强密码>          # 必填
JWT_SECRET=<openssl rand -hex 32>   # 建议：固定后重启不掉登录
TOTP_ENCRYPTION_KEY=<openssl rand -hex 32>  # 建议：固定后 2FA 不失效
ADMIN_EMAIL=you@example.com         # 可选：首次自动建管理员
ADMIN_PASSWORD=<管理员密码>          # 可选：留空则看日志取随机密码
TZ=Asia/Shanghai
SERVER_PORT=8080
```

生成密钥：`openssl rand -hex 32`（跑三次分别填 JWT / TOTP / 也可当 POSTGRES_PASSWORD）。

> ⚠️ `.env` 含密钥，别提交进任何仓库（本地保管或放你的私有 ops 仓）。

---

## 三、启动（注意：用了 local.yml 必须同时 `-f` override）

`docker compose` 只有对**默认文件名** `docker-compose.yml` 才会自动叠加 `docker-compose.override.yml`；
用 `docker-compose.local.yml` 时**必须两个 `-f` 都写**，否则镜像指向和开关不生效：

```bash
cd ~/sub2api-deploy
docker compose -f docker-compose.local.yml -f docker-compose.override.yml pull
docker compose -f docker-compose.local.yml -f docker-compose.override.yml up -d
docker compose -f docker-compose.local.yml -f docker-compose.override.yml logs -f sub2api
```

> 命令较长，建议存个别名：
> ```bash
> alias s2='docker compose -f docker-compose.local.yml -f docker-compose.override.yml'
> ```
> 之后就是 `s2 pull` / `s2 up -d` / `s2 logs -f sub2api` / `s2 ps`。

**访问**：`http://<服务器IP>:8080`。若没设 `ADMIN_PASSWORD`，取随机密码：
```bash
docker compose -f docker-compose.local.yml -f docker-compose.override.yml logs sub2api | grep -i "admin password"
```

**自检**：登录后到「管理员 → 版本徽标」应**只显示当前版本、没有「立即更新」按钮**（说明在线更新已关）。

---

## 四、升级（走定制流程，不走 App 内更新）

日常升级 = 在私有仓合上游 / 改代码 → CI 出新 `latest` 镜像 → 服务器重新 pull。
把 `update.sh` 一起放到部署目录后，一条命令即可（等价于 App 内「立即更新」，但走镜像方式）：

```bash
./update.sh            # 拉 :latest 并重建 + 健康自检 + 清理旧镜像
```

不用脚本时的等价手动命令（`s2` 为前面的别名）：
```bash
s2 pull && s2 up -d
```

镜像里的前端已 embed，无需单独处理。

> 说明：App 内那个「立即更新」按钮是给**二进制/systemd 安装**设计的（原地换二进制），
> Docker 部署里用不了（容器内非 root 改不了 /app，且会被下次 pull 覆盖），因此已用
> `DISABLE_ONLINE_UPDATE=true` 关闭。`update.sh` 是 Docker 下的等价替代。

## 五、回滚

镜像每个提交都有 `sha-xxxxxxx` 标签，回滚只需带标签跑一次脚本（会自动改 override 的 `image:` 并备份原文件）：

```bash
./update.sh sha-1c0a69c     # 或 ./update.sh v0.2.8
```

手动方式：把 override 的 `image:` 改到旧标签，再 `s2 pull && s2 up -d`。

## 六、数据备份/迁移

用的是本地目录版（`data/` `postgres_data/` `redis_data/`），整包搬走即可：
```bash
s2 down
cd .. && tar czf sub2api-deploy.tar.gz sub2api-deploy/
# 拷到新机解包，重新 docker login ghcr.io 后 s2 up -d
```
