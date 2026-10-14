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
IMAGE_REPOSITORY=ghcr.io/engineemomo/sub2apicust
IMAGE_SOURCE=https://github.com/EngineeMoMo/sub2apicust
BEFORE_IMAGE="${IMAGE_REPOSITORY}:update-before"
ROLLBACK_IMAGE="${IMAGE_REPOSITORY}:rollback-previous"
PREVIOUS_IMAGE_ID=""
CURRENT_IMAGE_ID=""
ROLLBACK_IMAGE_ID=""
CLEANUP_READY=false
override_next=""

if ! mkdir .sub2api-update.lock 2>/dev/null; then
  echo "✗ 更新锁已存在，请确认没有其他更新进程；中断遗留的锁需确认后移除。" >&2
  exit 1
fi
trap 'rmdir .sub2api-update.lock 2>/dev/null || true; [[ -z "$override_next" ]] || rm -f "$override_next"' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

project_image() {
  case "$1" in
    "${IMAGE_REPOSITORY}:"*|"${IMAGE_REPOSITORY}@"*) return 0 ;;
    *) return 1 ;;
  esac
}

prepare_image_cleanup() {
  local container_id image_ref image_id
  container_id=$("${COMPOSE[@]}" ps -a -q "$SERVICE") || return 1
  if [[ -n "$container_id" ]]; then
    [[ "$container_id" != *$'\n'* ]] || return 1
    image_ref=$(docker container inspect --format '{{.Config.Image}}' "$container_id") || return 1
    project_image "$image_ref" || return 1
    image_id=$(docker container inspect --format '{{.Image}}' "$container_id") || return 1
    PREVIOUS_IMAGE_ID=$(docker image inspect --format '{{.Id}}' "$image_id") || return 1
    [[ -n "$PREVIOUS_IMAGE_ID" ]] || return 1
    docker image tag "$PREVIOUS_IMAGE_ID" "$BEFORE_IMAGE" || return 1
  fi
  CLEANUP_READY=true
}

remove_unused_image() {
  local image_ref="$1" image_id containers
  if ! image_id=$(docker image inspect --format '{{.Id}}' "$image_ref"); then
    echo "⚠ 无法检查镜像，保留：$image_ref" >&2
    return 0
  fi
  [[ -n "$image_id" ]] || return 0
  if [[ "$image_id" == "$CURRENT_IMAGE_ID" || "$image_id" == "$ROLLBACK_IMAGE_ID" ]]; then
    return 0
  fi
  if ! containers=$(docker ps --all --quiet --filter "ancestor=$image_ref"); then
    echo "⚠ 无法检查容器引用，保留：$image_ref" >&2
    return 0
  fi
  if [[ -n "$containers" ]]; then
    echo "→ 保留其他容器引用的镜像：$image_ref"
    return 0
  fi
  echo "→ 删除未使用的旧镜像：$image_ref"
  if ! docker image rm "$image_ref"; then
    echo "⚠ 镜像未删除，请检查上述错误：$image_ref" >&2
  fi
}

cleanup_old_images() {
  local container_id image_ref image_id candidates dangling tag
  [[ "$CLEANUP_READY" == true ]] || return 0
  container_id=$("${COMPOSE[@]}" ps -a -q "$SERVICE") || return 1
  [[ -n "$container_id" && "$container_id" != *$'\n'* ]] || return 1
  image_ref=$(docker container inspect --format '{{.Config.Image}}' "$container_id") || return 1
  project_image "$image_ref" || return 1
  image_id=$(docker container inspect --format '{{.Image}}' "$container_id") || return 1
  CURRENT_IMAGE_ID=$(docker image inspect --format '{{.Id}}' "$image_id") || return 1
  [[ -n "$CURRENT_IMAGE_ID" ]] || return 1

  if [[ -n "$PREVIOUS_IMAGE_ID" && "$PREVIOUS_IMAGE_ID" != "$CURRENT_IMAGE_ID" ]]; then
    docker image tag "$BEFORE_IMAGE" "$ROLLBACK_IMAGE" || return 1
  fi
  ROLLBACK_IMAGE_ID=$(docker image inspect --format '{{.Id}}' "$ROLLBACK_IMAGE" 2>/dev/null || true)
  if [[ -z "$ROLLBACK_IMAGE_ID" || "$ROLLBACK_IMAGE_ID" == "$CURRENT_IMAGE_ID" ]]; then
    ROLLBACK_IMAGE_ID=""
    candidates=$(docker image ls "$IMAGE_REPOSITORY" --filter "before=$CURRENT_IMAGE_ID" --format '{{.Repository}}:{{.Tag}}') || return 1
    while IFS= read -r image_ref; do
      project_image "$image_ref" || continue
      tag="${image_ref##*:}"
      [[ "$tag" =~ ^sha-[0-9a-f]{7,40}$ ]] || continue
      image_id=$(docker image inspect --format '{{.Id}}' "$image_ref") || return 1
      [[ -n "$image_id" && "$image_id" != "$CURRENT_IMAGE_ID" ]] || continue
      docker image tag "$image_ref" "$ROLLBACK_IMAGE" || return 1
      ROLLBACK_IMAGE_ID="$image_id"
      break
    done <<< "$candidates"
  fi

  echo "→ 保留当前镜像：$CURRENT_IMAGE_ID"
  if [[ -n "$ROLLBACK_IMAGE_ID" ]]; then
    # [CUSTOM] Bash 3 在 UTF-8 locale 下可能将相邻中文符号读入变量名，显式界定边界。
    echo "→ 保留回退镜像：${ROLLBACK_IMAGE}（${ROLLBACK_IMAGE_ID}）"
  fi
  if [[ -n "$PREVIOUS_IMAGE_ID" ]]; then
    docker image rm "$BEFORE_IMAGE" >/dev/null || return 1
  fi
  candidates=$(docker image ls "$IMAGE_REPOSITORY" --format '{{.Repository}}:{{.Tag}}') || return 1
  dangling=$(docker image ls --quiet --no-trunc --filter dangling=true --filter "label=org.opencontainers.image.source=$IMAGE_SOURCE") || return 1
  while IFS= read -r image_ref; do
    project_image "$image_ref" || continue
    tag="${image_ref##*:}"
    [[ "$tag" == latest || "$tag" =~ ^sha-[0-9a-f]{7,40}$ ]] || continue
    remove_unused_image "$image_ref"
  done <<< "$candidates"
  while IFS= read -r image_id; do
    [[ -n "$image_id" ]] || continue
    remove_unused_image "$image_id"
  done <<< "$dangling"
}

if ! prepare_image_cleanup; then
  echo "⚠ 无法安全记录本项目旧镜像，本次更新跳过自动镜像清理。" >&2
fi

# --- 可选：把镜像标签钉到参数指定的版本（回滚/指定版）------------------------
TAG="${1:-}"
if [[ -n "$TAG" ]]; then
  if [[ ! -f docker-compose.override.yml ]]; then
    echo "✗ 指定标签需要 docker-compose.override.yml 里有 image: 行" >&2; exit 1
  fi
  # [CUSTOM] 标签只允许 Docker 合法字符，避免拼进替换内容时改坏 override。
  if [[ ! "$TAG" =~ ^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$ ]]; then
    echo "✗ 镜像标签格式无效：${TAG}（示例：sha-1c0a69c、latest）" >&2; exit 1
  fi
  # [CUSTOM] 只改 sub2api 服务下、且指向本项目镜像仓库的那一行 image:；
  # postgres/redis 等其他服务及注释行不动。找不到或找到多处都不改文件。
  override_next=$(mktemp docker-compose.override.yml.next.XXXXXX)
  replace_status=0
  awk -v repo="$IMAGE_REPOSITORY" -v tag="$TAG" -v q="'" '
    function indent(s) { match(s, /^[ \t]*/); return RLENGTH }
    {
      line = $0; trimmed = line; sub(/^[ \t]+/, "", trimmed)
      if (trimmed == "" || substr(trimmed, 1, 1) == "#") { print line; next }
      level = indent(line)
      if (in_service && level <= service_level) in_service = 0
      if (trimmed ~ /^sub2api:[ \t]*(#.*)?$/) { in_service = 1; service_level = level; print line; next }
      if (in_service && trimmed ~ /^image:/) {
        value = trimmed; sub(/^image:[ \t]*/, "", value)
        quote = ""
        if (substr(value, 1, 1) == "\"" || substr(value, 1, 1) == q) { quote = substr(value, 1, 1); value = substr(value, 2) }
        if (index(value, repo ":") != 1) { failed = 3; exit 3 }
        rest = substr(value, length(repo) + 2)
        end = length(rest) + 1
        for (i = 1; i <= length(rest); i++) {
          c = substr(rest, i, 1)
          if (c == " " || c == "\t" || c == "\"" || c == q || c == "#") { end = i; break }
        }
        prefix = substr(line, 1, length(line) - length(trimmed))
        print prefix "image: " quote repo ":" tag substr(rest, end)
        found++
        next
      }
      print line
    }
    END {
      if (failed) exit failed
      if (found == 0) exit 2
      if (found > 1) exit 4
    }' docker-compose.override.yml > "$override_next" || replace_status=$?
  case "$replace_status" in
    0) ;;
    2) rm -f "$override_next"; echo "✗ override 的 sub2api 服务下没有 image: 行，未修改文件" >&2; exit 1 ;;
    3) rm -f "$override_next"; echo "✗ override 的 sub2api 镜像不是 ${IMAGE_REPOSITORY}，未修改文件；本地构建等情况请手动改标签" >&2; exit 1 ;;
    *) rm -f "$override_next"; echo "✗ override 的 sub2api 服务下有多处 image: 或文件无法解析，未修改文件" >&2; exit 1 ;;
  esac
  # 备份原文件便于回退；cp 写回保留原文件权限。BSD/GNU sed -i 不兼容，统一不用。
  cp docker-compose.override.yml "docker-compose.override.yml.bak.$(date +%s)"
  cp "$override_next" docker-compose.override.yml
  rm -f "$override_next"
  echo "→ 镜像标签已切到 :${TAG}"
fi

echo "→ 当前版本：$(${COMPOSE[@]} exec -T "$SERVICE" /app/sub2api --version 2>/dev/null || echo 未运行/未知)"
echo "→ 拉取最新镜像…"
"${COMPOSE[@]}" pull "$SERVICE"

echo "→ 重建容器…"
"${COMPOSE[@]}" up -d "$SERVICE"

# --- 健康自检：等 /health 就绪（最多 ~60s）----------------------------------
echo -n "→ 等待健康检查"
HEALTHY=false
for _ in $(seq 1 30); do
  if "${COMPOSE[@]}" exec -T "$SERVICE" wget -q -T 3 -O /dev/null http://localhost:8080/health 2>/dev/null; then
    echo " ✓ 已就绪"
    HEALTHY=true
    break
  fi
  echo -n "."
  sleep 2
done

if [[ "$HEALTHY" != true ]]; then
  echo "✗ 健康检查超时，未清理任何镜像。请检查容器日志并使用原版本回退。" >&2
  exit 1
fi

echo "→ 清理本项目旧镜像（保留当前、上一版及其他容器引用的镜像）…"
if ! cleanup_old_images; then
  echo "⚠ 本次更新已就绪，但镜像清理未完成；旧镜像保留，请检查上述错误。" >&2
fi

echo "✓ 更新完成。查看日志： ${DC[*]} ${FILES[*]} logs -f ${SERVICE}"
