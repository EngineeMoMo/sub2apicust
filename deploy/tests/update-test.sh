#!/usr/bin/env bash
set -euo pipefail

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd "$TEST_DIR/.." && pwd)"
TEST_TEMP_BASE="$(cd "${TMPDIR:-/tmp}" && pwd -P)"
TEST_ROOT="$(mktemp -d "$TEST_TEMP_BASE/sub2api-update-test.XXXXXX")"
FIXTURE="$TEST_DIR/fixtures/update-docker.sh"
REPOSITORY=ghcr.io/engineemomo/sub2apicust
SOURCE=https://github.com/EngineeMoMo/sub2apicust
CASE_DIR=""
CASE_COUNT=0
cleanup() {
  if [[ "$TEST_ROOT" == "$TEST_TEMP_BASE"/sub2api-update-test.* && -d "$TEST_ROOT" && ! -L "$TEST_ROOT" ]]; then
    rm -rf -- "$TEST_ROOT"
  fi
}
trap cleanup EXIT

fail() { printf 'FAIL (%s): %s\n' "$CASE_DIR" "$*" >&2; cat "$CASE_DIR/output" >&2; exit 1; }

add_image() {
  printf 'config-%s sha256:%s %s %s %s %s\n' "$1" "$1" "${4:-$REPOSITORY}" "$2" "$3" "${5:-$SOURCE}" >> "$CASE_DIR/images"
}

new_case() {
  CASE_DIR="$TEST_ROOT/$1"
  mkdir "$CASE_DIR"
  cp "$DEPLOY_DIR/update.sh" "$CASE_DIR/update.sh"
  printf 'services:\n  sub2api:\n    image: %s:sha-bbbbbbb\n' "$REPOSITORY" > "$CASE_DIR/docker-compose.local.yml"
  cp "$CASE_DIR/docker-compose.local.yml" "$CASE_DIR/docker-compose.override.yml"
  : > "$CASE_DIR/images"; : > "$CASE_DIR/commands"; : > "$CASE_DIR/used"; : > "$CASE_DIR/output"
  add_image a sha-aaaaaaa 20
  add_image b sha-bbbbbbb 30
  add_image c sha-ccccccc 10
  add_image d sha-ddddddd 5
  add_image postgres 18-alpine 1 postgres other
  add_image chat latest 1 yidadaa/chatgpt-next-web other
  add_image dangling '<none>' 1
  add_image foreign '<none>' 1 other/image other
  printf 'config-a %s:sha-aaaaaaa\n' "$REPOSITORY" > "$CASE_DIR/current"
  printf 'config-b %s:sha-bbbbbbb\n' "$REPOSITORY" > "$CASE_DIR/next"
  CASE_COUNT=$((CASE_COUNT + 1))
}

run_update() {
  local result=0
  BASH_ENV="$FIXTURE" UPDATE_TEST_ROOT="$CASE_DIR" bash "$CASE_DIR/update.sh" "$@" > "$CASE_DIR/output" 2>&1 || result=$?
  return "$result"
}

assert_image() { grep -q "^config-$1 " "$CASE_DIR/images" || fail "image $1 was removed"; }
assert_removed() { if grep -q "^config-$1 " "$CASE_DIR/images"; then fail "image $1 was retained"; fi; }
assert_rollback() { grep -Fq "config-$1 sha256:$1 $REPOSITORY rollback-previous " "$CASE_DIR/images" || fail "rollback is not $1"; }
assert_no_removals() { if grep -q 'docker <image> <rm>' "$CASE_DIR/commands"; then fail 'images were removed after a failed update'; fi; }
assert_safe() {
  assert_image postgres; assert_image chat; assert_image foreign
  if grep -Eq '<prune>|<--force>|<-f> <sha256|<volume>|<system>' "$CASE_DIR/commands"; then fail 'global or forced cleanup attempted'; fi
  [[ ! -d "$CASE_DIR/.sub2api-update.lock" ]] || fail 'update lock was not released'
}

new_case fixed_upgrade
run_update sha-bbbbbbb || fail 'upgrade failed'
assert_image a; assert_image b; assert_rollback a
assert_removed c; assert_removed d; assert_removed dangling; assert_safe

new_case repeated_update
add_image a rollback-previous 20
printf 'config-b %s:sha-bbbbbbb\n' "$REPOSITORY" > "$CASE_DIR/current"
run_update || fail 'repeated update failed'
assert_image a; assert_rollback a; assert_removed c; assert_safe

new_case latest_upgrade
add_image a latest 20
printf 'config-a %s:latest\n' "$REPOSITORY" > "$CASE_DIR/current"
printf 'config-b %s:latest\n' "$REPOSITORY" > "$CASE_DIR/next"
run_update || fail 'latest upgrade failed'
assert_image a; assert_image b; assert_rollback a; assert_removed c; assert_safe

new_case latest_without_sha_alias
awk '$1 != "config-a"' "$CASE_DIR/images" > "$CASE_DIR/images.next"
mv "$CASE_DIR/images.next" "$CASE_DIR/images"
add_image a latest 20
printf 'config-a %s:latest\n' "$REPOSITORY" > "$CASE_DIR/current"
printf 'config-b %s:latest\n' "$REPOSITORY" > "$CASE_DIR/next"
run_update || fail 'latest-only upgrade failed'
assert_image a; assert_image b; assert_rollback a; assert_removed c; assert_safe
run_update || fail 'latest-only repeated update failed'
assert_image a; assert_image b; assert_rollback a; assert_safe

new_case rollback
add_image a rollback-previous 20
printf 'config-b %s:sha-bbbbbbb\n' "$REPOSITORY" > "$CASE_DIR/current"
printf 'config-a %s:sha-aaaaaaa\n' "$REPOSITORY" > "$CASE_DIR/next"
run_update sha-aaaaaaa || fail 'rollback failed'
assert_image a; assert_image b; assert_rollback b; assert_removed c; assert_safe

new_case first_same_version
printf 'config-b %s:sha-bbbbbbb\n' "$REPOSITORY" > "$CASE_DIR/current"
run_update || fail 'initial repeated update failed'
assert_image a; assert_rollback a; assert_removed c; assert_safe

new_case first_install
: > "$CASE_DIR/current"
run_update || fail 'first install failed'
assert_image b; assert_image a; assert_rollback a; assert_removed c; assert_safe

new_case containers_using_old_images
printf 'sha256:c running-other\nsha256:d stopped-other\n' > "$CASE_DIR/used"
run_update || fail 'upgrade failed'
assert_image c; assert_image d; assert_removed dangling; assert_safe

new_case additional_tags
add_image a sha-1111111 20
add_image b latest 30
add_image c v0.2.8 10
run_update || fail 'upgrade failed'
grep -Fq "$REPOSITORY sha-1111111 " "$CASE_DIR/images" || fail 'rollback alias was removed'
grep -Fq "$REPOSITORY latest " "$CASE_DIR/images" || fail 'current alias was removed'
grep -Fq "$REPOSITORY v0.2.8 " "$CASE_DIR/images" || fail 'manual version tag was removed'
assert_safe

for failure in health pull up; do
  new_case "failed_$failure"
  touch "$CASE_DIR/fail-$failure"
  if run_update; then fail 'failed update reported success'; fi
  assert_no_removals; assert_image a; assert_image c; assert_safe
done

new_case snapshot_failure
touch "$CASE_DIR/fail-tag"
run_update || fail 'snapshot failure blocked application update'
assert_no_removals; assert_safe

new_case missing_new_container
touch "$CASE_DIR/disappear"
run_update || fail 'cleanup failure blocked application update'
assert_no_removals; assert_safe

new_case image_list_failure
touch "$CASE_DIR/fail-ls"
run_update || fail 'list failure blocked application update'
assert_image c; assert_image d; assert_safe

new_case image_inspection_failure
printf '%s:sha-ccccccc\n' "$REPOSITORY" > "$CASE_DIR/fail-inspect"
run_update || fail 'inspection failure blocked application update'
assert_image c; assert_removed d; assert_safe

new_case container_check_failure
touch "$CASE_DIR/fail-ps"
run_update || fail 'reference check failure blocked application update'
assert_image c; assert_image d; assert_image dangling; assert_safe

new_case removal_failure
printf '%s:sha-ccccccc\n' "$REPOSITORY" > "$CASE_DIR/fail-rm"
run_update || fail 'removal failure blocked application update'
assert_image c; assert_removed d; assert_safe
grep -q '镜像未删除' "$CASE_DIR/output" || fail 'removal failure was hidden'

new_case foreign_application
printf 'config-chat yidadaa/chatgpt-next-web:latest\n' > "$CASE_DIR/current"
printf 'config-chat yidadaa/chatgpt-next-web:latest\n' > "$CASE_DIR/next"
run_update || fail 'foreign application update failed'
assert_no_removals; assert_safe

new_case changed_repository
printf 'config-chat yidadaa/chatgpt-next-web:latest\n' > "$CASE_DIR/next"
run_update || fail 'repository switch failed'
assert_no_removals; assert_image a; assert_safe

new_case legacy_compose
touch "$CASE_DIR/legacy"
run_update || fail 'legacy compose update failed'
assert_rollback a; assert_removed c; assert_safe
grep -q '^docker-compose' "$CASE_DIR/commands" || fail 'legacy compose was not used'

# [CUSTOM] 指定标签只改 sub2api 服务下本项目镜像那一行，其他服务、引号、注释保留。
new_case tag_multi_service
cat > "$CASE_DIR/docker-compose.override.yml" <<EOF
services:
  sub2api:
    # image: $REPOSITORY:sha-0000000
    image: "$REPOSITORY:sha-aaaaaaa"   # 当前版本
    environment:
      DISABLE_ONLINE_UPDATE: "true"
  postgres:
    image: postgres:18-alpine
  redis:
    image: 'redis:8-alpine'
EOF
run_update sha-bbbbbbb || fail 'multi-service tag switch failed'
grep -Fq "    image: \"$REPOSITORY:sha-bbbbbbb\"   # 当前版本" "$CASE_DIR/docker-compose.override.yml" || fail 'sub2api image tag was not switched in place'
grep -Fq "    image: postgres:18-alpine" "$CASE_DIR/docker-compose.override.yml" || fail 'postgres image was changed'
grep -Fq "    image: 'redis:8-alpine'" "$CASE_DIR/docker-compose.override.yml" || fail 'redis image was changed'
grep -Fq "    # image: $REPOSITORY:sha-0000000" "$CASE_DIR/docker-compose.override.yml" || fail 'comment line was changed'
ls "$CASE_DIR"/docker-compose.override.yml.bak.* > /dev/null 2>&1 || fail 'override backup was not kept'
if ls "$CASE_DIR"/docker-compose.override.yml.next.* > /dev/null 2>&1; then fail 'temporary override was left behind'; fi
assert_safe

new_case tag_foreign_image
printf 'services:\n  sub2api:\n    image: sub2api-custom:local\n' > "$CASE_DIR/docker-compose.override.yml"
cp "$CASE_DIR/docker-compose.override.yml" "$CASE_DIR/override.before"
if run_update sha-bbbbbbb; then fail 'tag switch on a foreign image was allowed'; fi
cmp -s "$CASE_DIR/override.before" "$CASE_DIR/docker-compose.override.yml" || fail 'foreign override was modified'
grep -q '不是' "$CASE_DIR/output" || fail 'foreign image refusal was not explained'
assert_no_removals; assert_safe

new_case tag_invalid
cp "$CASE_DIR/docker-compose.override.yml" "$CASE_DIR/override.before"
if run_update 'bad;tag|x'; then fail 'invalid tag was accepted'; fi
cmp -s "$CASE_DIR/override.before" "$CASE_DIR/docker-compose.override.yml" || fail 'override was modified by an invalid tag'
if ls "$CASE_DIR"/docker-compose.override.yml.bak.* > /dev/null 2>&1; then fail 'invalid tag created a backup'; fi
assert_no_removals; assert_safe

new_case tag_missing_image
printf 'services:\n  postgres:\n    image: postgres:18-alpine\n' > "$CASE_DIR/docker-compose.override.yml"
if run_update sha-bbbbbbb; then fail 'tag switch without a sub2api image was allowed'; fi
grep -Fq "image: postgres:18-alpine" "$CASE_DIR/docker-compose.override.yml" || fail 'unrelated image was changed'
assert_no_removals; assert_safe

new_case concurrent_update
mkdir "$CASE_DIR/.sub2api-update.lock"
if run_update; then fail 'concurrent update was allowed'; fi
assert_no_removals
[[ -d "$CASE_DIR/.sub2api-update.lock" ]] || fail 'another updater lock was removed'

printf 'PASS: %s update and image retention scenarios\n' "$CASE_COUNT"
