#!/usr/bin/env bash
# [CUSTOM] 定制版一键更新脚本，见项目根 CUSTOMIZATIONS.md / deploy/DEPLOY_CUSTOM.md。
#
# 作用：拉取你 CI 从自己仓库构建的最新 GHCR 镜像并重建 sub2api 容器 —— 等价于
# App 内那个「立即更新」，但走 Docker 镜像方式（持久、契合容器部署）。
#
# 用法（在你的部署目录里，与 docker-compose*.yml 同级）：
#   ./update.sh              # 拉取 :latest 并重建
#   ./update.sh sha-1c0a69c  # 更新/回滚到指定镜像标签
#
# 前提：已 docker login ghcr.io；override 里 image 指向你的 GHCR 镜像。
set -euo pipefail

cd "$(dirname "$0")"

# --- 选择 compose 命令（v2 优先，回退 legacy）-------------------------------
if docker compose version >/dev/null 2>&1; then
  DC=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  DC=(docker-compose)
else
  echo "✗ 未找到 docker compose / docker-compose" >&2
  exit 1
fi

# --- 组装 -f 文件列表（base + override，与部署时保持一致）--------------------
FILES=()
if   [[ -f docker-compose.local.yml ]]; then FILES+=(-f docker-compose.local.yml)
elif [[ -f docker-compose.yml       ]]; then FILES+=(-f docker-compose.yml)
else echo "✗ 当前目录没有 docker-compose.local.yml 或 docker-compose.yml" >&2; exit 1
fi
# override 若存在必须显式带上：用 local.yml 时 compose 不会自动叠加 override
[[ -f docker-compose.override.yml ]] && FILES+=(-f docker-compose.override.yml)

COMPOSE=("${DC[@]}" "${FILES[@]}")
SERVICE=sub2api

# --- 可选：把镜像标签钉到参数指定的版本（回滚/指定版）------------------------
TAG="${1:-}"
if [[ -n "$TAG" ]]; then
  if [[ ! -f docker-compose.override.yml ]]; then
    echo "✗ 指定标签需要 docker-compose.override.yml 里有 image: 行" >&2; exit 1
  fi
  # 只改 image 的标签部分（最后一个冒号后的内容），并留一份备份便于回退
  cp docker-compose.override.yml "docker-compose.override.yml.bak.$(date +%s)"
  sed -i -E "s|(image:[[:space:]]*[^[:space:]]+):[^[:space:]]+|\1:${TAG}|" docker-compose.override.yml
  echo "→ 镜像标签已切到 :${TAG}"
fi

echo "→ 当前版本：$(${COMPOSE[@]} exec -T "$SERVICE" /app/sub2api --version 2>/dev/null || echo 未运行/未知)"
echo "→ 拉取最新镜像…"
"${COMPOSE[@]}" pull "$SERVICE"

echo "→ 重建容器…"
"${COMPOSE[@]}" up -d "$SERVICE"

# --- 健康自检：等 /health 就绪（最多 ~60s）----------------------------------
echo -n "→ 等待健康检查"
for _ in $(seq 1 30); do
  if "${COMPOSE[@]}" exec -T "$SERVICE" wget -q -T 3 -O /dev/null http://localhost:8080/health 2>/dev/null; then
    echo " ✓ 已就绪"
    break
  fi
  echo -n "."
  sleep 2
done

echo "→ 清理悬空旧镜像（保留在用的）…"
docker image prune -f >/dev/null 2>&1 || true

echo "✓ 更新完成。查看日志： ${DC[*]} ${FILES[*]} logs -f ${SERVICE}"
