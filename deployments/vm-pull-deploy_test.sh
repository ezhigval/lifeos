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
if [[ "${CURL_CODE+x}" == x ]]; then
  code=$CURL_CODE
else
  code=404
fi
if [[ "$code" == "200" && "${CURL_PLAIN:-0}" != 1 ]]; then
  echo hello | gzip > "$out"
elif [[ -n "$out" ]]; then
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
if ! grep -q '^load ' "$DOCKER_LOG"; then
  echo "corrupt docker load must be attempted" >&2
  exit 1
fi

: > "$DOCKER_LOG"
export CURL_CODE=200
export CURL_PLAIN=1
export DOCKER_LOAD_OK=1
assert_rc 1 fetch_image "$SHA"
if grep -q '^load ' "$DOCKER_LOG"; then
  echo "non-gzip 200 must not docker load" >&2
  exit 1
fi
unset CURL_PLAIN

: > "$DOCKER_LOG"
export CURL_CODE=000
export DOCKER_LOAD_OK=1
assert_rc 2 fetch_image "$SHA"
if grep -q '^load ' "$DOCKER_LOG"; then
  echo "http 000 must not docker load" >&2
  exit 1
fi

: > "$DOCKER_LOG"
export CURL_CODE=
assert_rc 2 fetch_image "$SHA"
if grep -q '^load ' "$DOCKER_LOG"; then
  echo "empty http status must not docker load" >&2
  exit 1
fi

: > "$DOCKER_LOG"
export CURL_CODE=500
assert_rc 2 fetch_image "$SHA"
if grep -q '^load ' "$DOCKER_LOG"; then
  echo "http 500 must not docker load" >&2
  exit 1
fi

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

REPO="$root"
BASE="$tmp/pullbase"
mkdir -p "$BASE"
IMAGE=lifeos-app:deploy
write_pulled_compose
grep -q 'image: lifeos-app:deploy' "$BASE/docker-compose.pulled.yml"
if grep -q 'dockerfile:' "$BASE/docker-compose.pulled.yml"; then
  echo "pulled compose still has a build stanza" >&2
  exit 1
fi
grep -q 'postgres:' "$BASE/docker-compose.pulled.yml"
grep -q '127.0.0.1:8080:8080' "$BASE/docker-compose.pulled.yml"
grep -q "$BASE/.env" "$BASE/docker-compose.pulled.yml"
if grep -F -q '../.env' "$BASE/docker-compose.pulled.yml"; then
  echo "pulled compose still points env_file at ../.env" >&2
  exit 1
fi

printf '%s\n' \
  'TELEGRAM_BOT_TOKEN=supersecret' \
  'LIFEOS_HTTP_PROXY=http://127.0.0.1:8081' \
  'LIFEOS_MINIAPP_URL=https://local-ai-assist.ru/app/' \
  > "$BASE/.env"
pin_image
grep -q 'host.docker.internal:host-gateway' "$BASE/docker-compose.buildimage.yml"
grep -q 'LIFEOS_HTTP_PROXY: "http://host.docker.internal:8081"' "$BASE/docker-compose.buildimage.yml"
if grep -q 'supersecret' "$BASE/docker-compose.buildimage.yml"; then
  echo "buildimage compose leaked a token" >&2
  exit 1
fi
origin=$(public_origin)
if [[ "$origin" != "https://local-ai-assist.ru" ]]; then
  echo "public origin [$origin]" >&2
  exit 1
fi

printf '%s\n' 'LIFEOS_HTTP_PROXY=http://proxy.example:8081' > "$BASE/.env"
pin_image
if grep -q 'LIFEOS_HTTP_PROXY:' "$BASE/docker-compose.buildimage.yml"; then
  echo "non-loopback proxy must stay in the env file" >&2
  exit 1
fi
grep -q 'host.docker.internal:host-gateway' "$BASE/docker-compose.buildimage.yml"

printf '%s\n' '<script src="/app/assets/index-DaXJohFr.js"></script>' | miniapp_script_src > "$tmp/src.txt"
if [[ "$(cat "$tmp/src.txt")" != "/app/assets/index-DaXJohFr.js" ]]; then
  echo "script src parse failed" >&2
  exit 1
fi
printf '%s\n' '<html>no script</html>' | miniapp_script_src > "$tmp/src.txt"
if [[ -n "$(cat "$tmp/src.txt")" ]]; then
  echo "missing script src must be empty" >&2
  exit 1
fi

echo "vm-pull-deploy tests ok"
