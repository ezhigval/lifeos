#!/usr/bin/env bash
# Exercises fetch_image without Docker, GitHub, or root.
set -euo pipefail

root=$(cd "$(dirname "$0")/.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

bin="$tmp/bin"
mkdir -p "$bin"
cat > "$bin/docker" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${DOCKER_LOG:?}"
case "$1" in
  pull)
    if [[ "${DOCKER_PULL_OK:-0}" == 1 ]]; then
      exit 0
    fi
    exit 1
    ;;
  load)
    if [[ "${DOCKER_LOAD_OK:-1}" == 1 ]]; then
      exit 0
    fi
    exit 1
    ;;
  image)
    if [[ "$2" == "inspect" && "$3" != "lifeos-app:deploy" ]]; then
      exit 1
    fi
    ;;
esac
exit 0
EOF
chmod 755 "$bin/docker"

cat > "$bin/curl" <<'EOF'
#!/usr/bin/env bash
out=""
prev=""
for arg in "$@"; do
  if [[ "$prev" == "-o" ]]; then
    out=$arg
  fi
  prev=$arg
done
code="${CURL_CODE:-404}"
if [[ "$code" == "200" ]]; then
  echo hello | gzip > "$out"
else
  printf 'nope' > "$out"
fi
printf '%s' "$code"
EOF
chmod 755 "$bin/curl"

export PATH="$bin:$PATH"
export DOCKER_LOG="$tmp/docker.log"
export LIFEOS_DEPLOY_LIB=1
# shellcheck disable=SC1091
source "$root/deployments/vm-pull-deploy.sh"

BASE="$tmp/base"
mkdir -p "$BASE"
SHA=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
ghcr_reachable() { return 1; }
mem_available_kb() { echo 500000; }
disk_free_kb() { echo 2000000; }

assert_rc() {
  local want=$1
  shift
  set +e
  "$@"
  local got=$?
  set -e
  if [[ "$got" -ne "$want" ]]; then
    echo "want rc $want got $got: $*" >&2
    exit 1
  fi
}

: > "$DOCKER_LOG"
export DOCKER_PULL_OK=1 DOCKER_LOAD_OK=1 CURL_CODE=404
ghcr_reachable() { return 0; }
assert_rc 0 fetch_image "$SHA"
grep -q "pull ${GHCR_IMAGE}:sha-${SHA}" "$DOCKER_LOG"
grep -q "tag ${GHCR_IMAGE}:sha-${SHA} lifeos-app:deploy" "$DOCKER_LOG"
grep -q 'image: lifeos-app:deploy' "$BASE/docker-compose.buildimage.yml"
rm -f "$BASE/docker-compose.buildimage.yml"

: > "$DOCKER_LOG"
export DOCKER_PULL_OK=0
ghcr_reachable() { return 0; }
export CURL_CODE=404
assert_rc 2 fetch_image "$SHA"
if grep -q '^load ' "$DOCKER_LOG"; then
  echo "404 must not docker load" >&2
  exit 1
fi

: > "$DOCKER_LOG"
ghcr_reachable() { return 1; }
export CURL_CODE=200
export DOCKER_LOAD_OK=1
assert_rc 0 fetch_image "$SHA"
grep -q '^load ' "$DOCKER_LOG"
grep -q 'image: lifeos-app:deploy' "$BASE/docker-compose.buildimage.yml"

: > "$DOCKER_LOG"
export CURL_CODE=200
export DOCKER_LOAD_OK=0
assert_rc 1 fetch_image "$SHA"

: > "$DOCKER_LOG"
export CURL_CODE=500
assert_rc 1 fetch_image "$SHA"

mem_available_kb() { echo 1000; }
assert_rc 2 fetch_image "$SHA"
mem_available_kb() { echo 500000; }

disk_free_kb() { echo 1000; }
assert_rc 2 fetch_image "$SHA"

# Oldest assets beyond the two newest are the ones the workflow deletes.
printf '%s\n' \
  $'1\tlifeos-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.tar.gz' \
  $'2\tnotes.txt' \
  $'3\tlifeos-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.tar.gz' \
  $'4\tlifeos-cccccccccccccccccccccccccccccccccccccccc.tar.gz' |
  awk -F '\t' 'BEGIN { keep = 2 } $2 ~ /^lifeos-[0-9a-f]{40}\.tar\.gz$/ { ids[++n] = $1 } END { for (i = 1; i <= n - keep; i++) print ids[i] }' > "$tmp/drop.txt"
got=$(cat "$tmp/drop.txt")
if [[ "$got" != "1" ]]; then
  echo "asset cleanup want id 1 got [$got]" >&2
  exit 1
fi

echo "vm-pull-deploy tests ok"
