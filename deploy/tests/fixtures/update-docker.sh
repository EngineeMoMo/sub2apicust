update_mock_resolve() {
  awk -v ref="$1" '$1 == ref || $2 == ref || $3 ":" $4 == ref { print; exit }' "$UPDATE_TEST_ROOT/images"
}

update_mock_tag() {
  local source target row repository tag
  source="$1"
  target="$2"
  row=$(update_mock_resolve "$source")
  [[ -n "$row" ]] || return 1
  repository="${target%:*}"
  tag="${target##*:}"
  awk -v ref="$target" '$3 ":" $4 != ref' "$UPDATE_TEST_ROOT/images" > "$UPDATE_TEST_ROOT/images.next"
  printf '%s\n' "$row" | awk -v repository="$repository" -v tag="$tag" '{ $3 = repository; $4 = tag; print }' >> "$UPDATE_TEST_ROOT/images.next"
  mv "$UPDATE_TEST_ROOT/images.next" "$UPDATE_TEST_ROOT/images"
}

update_mock_compose() {
  while [[ "${1:-}" == -f ]]; do shift 2; done
  case "${1:-}" in
    version) [[ ! -f "$UPDATE_TEST_ROOT/legacy" ]] ;;
    ps)
      [[ ! -f "$UPDATE_TEST_ROOT/fail-compose-ps" ]] || return 1
      [[ ! -s "$UPDATE_TEST_ROOT/current" ]] || printf '%s\n' app
      ;;
    pull)
      [[ ! -f "$UPDATE_TEST_ROOT/fail-pull" ]] || return 1
      if [[ -s "$UPDATE_TEST_ROOT/next" ]]; then
        update_mock_tag "$(cut -d ' ' -f 1 "$UPDATE_TEST_ROOT/next")" "$(cut -d ' ' -f 2 "$UPDATE_TEST_ROOT/next")"
      fi
      ;;
    up)
      [[ ! -f "$UPDATE_TEST_ROOT/fail-up" ]] || return 1
      cp "$UPDATE_TEST_ROOT/next" "$UPDATE_TEST_ROOT/current"
      [[ ! -f "$UPDATE_TEST_ROOT/disappear" ]] || : > "$UPDATE_TEST_ROOT/current"
      ;;
    exec)
      if [[ "$*" == *wget* ]]; then
        [[ ! -f "$UPDATE_TEST_ROOT/fail-health" ]]
      else
        printf '%s\n' '0.2.8-custom.mock'
      fi
      ;;
    *) printf 'Unexpected compose command: %s\n' "$*" >&2; return 99 ;;
  esac
}

docker() {
  local command row reference image_id tag ids filter before repository dangling source format
  printf 'docker' >> "$UPDATE_TEST_ROOT/commands"
  printf ' <%s>' "$@" >> "$UPDATE_TEST_ROOT/commands"
  printf '\n' >> "$UPDATE_TEST_ROOT/commands"
  command="${1:-} ${2:-}"
  case "$command" in
    'compose '*) shift; update_mock_compose "$@" ;;
    'container inspect')
      [[ -s "$UPDATE_TEST_ROOT/current" ]] || return 1
      if [[ "$*" == *'.Config.Image'* ]]; then
        cut -d ' ' -f 2 "$UPDATE_TEST_ROOT/current"
      else
        cut -d ' ' -f 1 "$UPDATE_TEST_ROOT/current"
      fi
      ;;
    'image inspect')
      reference="${!#}"
      if [[ -f "$UPDATE_TEST_ROOT/fail-inspect" ]] && grep -Fxq "$reference" "$UPDATE_TEST_ROOT/fail-inspect"; then return 1; fi
      row=$(update_mock_resolve "$reference")
      [[ -n "$row" ]] || return 1
      printf '%s\n' "$row" | cut -d ' ' -f 2
      ;;
    'image tag')
      [[ ! -f "$UPDATE_TEST_ROOT/fail-tag" ]] || return 1
      update_mock_tag "$3" "$4"
      ;;
    'image ls')
      [[ ! -f "$UPDATE_TEST_ROOT/fail-ls" ]] || return 1
      shift 2
      repository=""; before=""; dangling=false; source=""
      while [[ $# -gt 0 ]]; do
        case "$1" in
          --filter)
            filter="$2"
            case "$filter" in
              before=*) row=$(update_mock_resolve "${filter#before=}"); before=$(printf '%s\n' "$row" | cut -d ' ' -f 5) ;;
              dangling=true) dangling=true ;;
              label=org.opencontainers.image.source=*) source="${filter#label=org.opencontainers.image.source=}" ;;
              *) return 99 ;;
            esac
            shift 2
            ;;
          --format) shift 2 ;;
          --quiet|--no-trunc) shift ;;
          *) repository="$1"; shift ;;
        esac
      done
      awk -v repository="$repository" -v before="$before" -v dangling="$dangling" -v source="$source" '
        (repository == "" || $3 == repository) && (before == "" || $5 < before) &&
        (dangling != "true" || $4 == "<none>") && (source == "" || $6 == source) { print }
      ' "$UPDATE_TEST_ROOT/images" | sort -k5,5nr | awk -v dangling="$dangling" '{
        if (dangling == "true") { if (!seen[$2]++) print $2 }
        else print $3 ":" $4
      }'
      ;;
    'image rm')
      [[ $# == 3 && "$3" != -* ]] || return 99
      reference="$3"
      if [[ -f "$UPDATE_TEST_ROOT/fail-rm" ]] && grep -Fxq "$reference" "$UPDATE_TEST_ROOT/fail-rm"; then return 1; fi
      row=$(update_mock_resolve "$reference")
      [[ -n "$row" ]] || return 1
      image_id=$(printf '%s\n' "$row" | cut -d ' ' -f 2)
      if [[ "$reference" == sha256:* ]]; then
        awk -v image_id="$image_id" '$2 != image_id' "$UPDATE_TEST_ROOT/images" > "$UPDATE_TEST_ROOT/images.next"
      else
        awk -v reference="$reference" '$3 ":" $4 != reference' "$UPDATE_TEST_ROOT/images" > "$UPDATE_TEST_ROOT/images.next"
      fi
      mv "$UPDATE_TEST_ROOT/images.next" "$UPDATE_TEST_ROOT/images"
      printf 'Deleted: %s\n' "$reference"
      ;;
    'ps '*)
      [[ ! -f "$UPDATE_TEST_ROOT/fail-ps" ]] || return 1
      filter="${!#}"
      [[ "$filter" == ancestor=* ]] || return 99
      row=$(update_mock_resolve "${filter#ancestor=}")
      [[ -n "$row" ]] || return 1
      image_id=$(printf '%s\n' "$row" | cut -d ' ' -f 2)
      if [[ -s "$UPDATE_TEST_ROOT/current" ]]; then
        row=$(update_mock_resolve "$(cut -d ' ' -f 1 "$UPDATE_TEST_ROOT/current")")
        [[ "$(printf '%s\n' "$row" | cut -d ' ' -f 2)" != "$image_id" ]] || printf '%s\n' app
      fi
      awk -v image_id="$image_id" '$1 == image_id { print $2 }' "$UPDATE_TEST_ROOT/used"
      ;;
    *) printf 'Unexpected Docker command: %s\n' "$*" >&2; return 99 ;;
  esac
}

docker-compose() {
  printf 'docker-compose <%s>\n' "$*" >> "$UPDATE_TEST_ROOT/commands"
  update_mock_compose "$@"
}

sleep() { :; }
