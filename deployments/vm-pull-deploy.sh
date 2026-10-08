#!/usr/bin/env bash
# Pull-deploy for a ~2GB VM. The image is built on GitHub, never on this host:
# the Alpine package CDN is blocked here, and a local compile can OOM sshd.
# lifeos-deploy.service runs this script. Schedule is lifeos-deploy.timer
# (daily 04:15 UTC, Persistent=true). Unchanged SHA skips the pull, including
# after boot.
set -euo pipefail

BRANCH="${LIFEOS_BRANCH:-main}"
BASE=/opt/lifeos
REPO="$BASE/repo"
SHA_FILE="$BASE/.last_deploy_sha"
SELF="$BASE/bin/vm-pull-deploy.sh"
LOCK=/var/lock/lifeos-deploy.lock
IMAGE=lifeos-app:deploy
GHCR_IMAGE="${LIFEOS_GHCR_IMAGE:-ghcr.io/ezhigval/lifeos}"
RELEASE_TAG="${LIFEOS_IMAGE_RELEASE:-vm-image}"

log() { echo "[lifeos-deploy] $*"; }

apply_oom_scores() {
  local unit score dir name pid
  for unit in ssh.service sshd.service docker.service containerd.service; do
    if systemctl cat "$unit" >/dev/null 2>&1; then
      case "$unit" in
        docker.service|containerd.service) score=-500 ;;
        *) score=-900 ;;
      esac
      dir="/etc/systemd/system/${unit}.d"
      install -d -m 755 "$dir"
      printf '[Service]\nOOMScoreAdjust=%s\n' "$score" > "$dir/lifeos-oom.conf"
    fi
  done
  systemctl daemon-reload || true
  # Drop-ins apply on next start. Set the running processes now so a build
  # does not require an ssh restart.
  for name in sshd dockerd containerd; do
    for pid in $(pgrep -x "$name" 2>/dev/null || true); do
      case "$name" in
        dockerd|containerd) echo -500 > "/proc/$pid/oom_score_adj" 2>/dev/null || true ;;
        *) echo -900 > "/proc/$pid/oom_score_adj" 2>/dev/null || true ;;
      esac
    done
  done
  if command -v docker >/dev/null 2>&1; then
    docker update --oom-score-adj -200 lifeos-app-1 lifeos-postgres-1 >/dev/null 2>&1 || true
  fi
}

# A previous install left a local HTTP server in front of the app.
# The only public door is the named tunnel to 127.0.0.1:8080.
retire_nginx() {
  rm -f \
    "$BASE/bin/nginx-lifeos.conf" \
    /etc/nginx/sites-enabled/lifeos \
    /etc/nginx/sites-enabled/lifeos.conf \
    /etc/nginx/sites-available/lifeos.conf \
    /etc/nginx/sites-available/lifeos.conf.bak
  if systemctl cat nginx.service >/dev/null 2>&1; then
    systemctl disable --now nginx >/dev/null 2>&1 || true
    log "stopped local http server; public path is the tunnel"
  fi
}

bind_loopback_ports() {
  python3 - "$REPO/deployments/docker-compose.yml" "$BASE/docker-compose.override.yml" <<'PY'
import pathlib, sys
replacements = (
    ('"5433:5432"', '"127.0.0.1:5433:5432"'),
    ('"0.0.0.0:5433:5432"', '"127.0.0.1:5433:5432"'),
    ('"8080:8080"', '"127.0.0.1:8080:8080"'),
    ('"0.0.0.0:8080:8080"', '"127.0.0.1:8080:8080"'),
)
for raw in sys.argv[1:]:
    path = pathlib.Path(raw)
    if not path.is_file():
        continue
    text = path.read_text()
    new = text
    for old, repl in replacements:
        new = new.replace(old, repl)
    if new != text:
        path.write_text(new)
        print(f"[lifeos-deploy] loopback publish updated")
PY
}

bind_postgres_localhost() {
  local compose="$REPO/deployments/docker-compose.yml"
  [[ -f "$compose" ]] || return 0
  python3 - "$compose" <<'PY'
import pathlib, sys
path = pathlib.Path(sys.argv[1])
text = path.read_text()
old = '      - "5433:5432"\n'
new = '      - "127.0.0.1:5433:5432"\n'
if old in text:
    path.write_text(text.replace(old, new, 1))
    print("[lifeos-deploy] postgres publish set to 127.0.0.1:5433")
PY
}

compose() {
  local -a args=(docker compose --env-file "$BASE/.env" -p lifeos -f "$REPO/deployments/docker-compose.yml")
  if [[ -f "$BASE/docker-compose.override.yml" ]]; then
    args+=(-f "$BASE/docker-compose.override.yml")
  fi
  if [[ -f "$BASE/docker-compose.buildimage.yml" ]]; then
    args+=(-f "$BASE/docker-compose.buildimage.yml")
  fi
  "${args[@]}" "$@"
}

backup_postgres() {
  local cid out bytes
  cid=$(docker ps -q -f name=lifeos-postgres -f status=running | head -n 1 || true)
  if [[ -z "$cid" ]]; then
    log "no running postgres container; skip backup"
    return 0
  fi
  install -d -m 700 "$BASE/backups"
  out="$BASE/backups/lifeos-$(date -u +%Y%m%dT%H%M%SZ).sql.gz"
  log "pg_dump before schema or port change"
  if ! docker exec "$cid" pg_dump -U lifeos -d lifeos --no-owner | gzip -c > "$out"; then
    rm -f "$out"
    log "pg_dump failed; not changing postgres or schema"
    exit 1
  fi
  gzip -t "$out"
  bytes=$(wc -c < "$out")
  if [[ "$bytes" -lt 100 ]]; then
    rm -f "$out"
    log "backup too small ($bytes bytes); not changing postgres"
    exit 1
  fi
  log "backup bytes=$bytes path=$out"
  ls -1t "$BASE/backups"/lifeos-*.sql.gz 2>/dev/null | tail -n +4 | xargs -r rm -f
}

# Recreate postgres only when Docker still publishes 5433 on every interface.
# The app talks to postgres over the compose network, not this host port.
# If an override file puts 0.0.0.0 back, do not recreate on every timer tick.
ensure_postgres_localhost() {
  bind_loopback_ports
  bind_postgres_localhost
  local ports stamp when now
  ports=$(docker ps --format '{{.Names}} {{.Ports}}' | grep lifeos-postgres || true)
  if [[ "$ports" != *0.0.0.0:5433* && "$ports" != *"[::]:5433"* ]]; then
    return 0
  fi
  stamp="$BASE/.postgres_rebind_fail"
  if [[ -f "$stamp" ]]; then
    when=$(cat "$stamp")
    now=$(date +%s)
    if [[ $((now - when)) -lt 3600 ]]; then
      log "postgres still on 0.0.0.0; not recreating again this hour ($ports)"
      return 0
    fi
  fi
  log "rebind postgres off 0.0.0.0 ($ports)"
  backup_postgres
  compose up -d postgres
  ports=$(docker ps --format '{{.Names}} {{.Ports}}' | grep lifeos-postgres || true)
  if [[ "$ports" == *0.0.0.0:5433* || "$ports" == *"[::]:5433"* ]]; then
    date +%s > "$stamp"
    log "postgres still published on every interface after rebind ($ports)"
  else
    rm -f "$stamp"
    log "postgres host port is no longer 0.0.0.0"
  fi
}

mem_available_kb() {
  awk '/MemAvailable:/ {print $2}' /proc/meminfo
}

disk_free_kb() {
  df -Pk / | awk 'NR==2 {print $4}'
}

ghcr_reachable() {
  timeout 8 bash -c 'echo >/dev/tcp/ghcr.io/443' >/dev/null 2>&1
}

pin_image() {
  cat > "$BASE/docker-compose.buildimage.yml" <<EOF
# Points compose at the image published by GitHub Actions.
services:
  app:
    image: ${IMAGE}
EOF
}

# return 0 fetched, 2 not ready (retry later), 1 failed
fetch_image() {
  local sha="$1"
  local avail_kb disk_kb ref tmp http
  avail_kb=$(mem_available_kb)
  disk_kb=$(disk_free_kb)
  log "MemAvailable=${avail_kb}kB disk_avail=${disk_kb}kB"
  if [[ "${avail_kb:-0}" -lt 100000 ]]; then
    log "skip pull: under 100MB available; timer will retry"
    return 2
  fi
  # Drop layers left by the old on-VM build. That path cannot succeed here.
  docker image prune -f >/dev/null 2>&1 || true
  docker builder prune -af >/dev/null 2>&1 || true
  disk_kb=$(disk_free_kb)
  log "disk_avail after prune=${disk_kb}kB"
  if [[ "${disk_kb:-0}" -lt 800000 ]]; then
    log "skip pull: under 800MB free disk; timer will retry"
    return 2
  fi

  ref="${GHCR_IMAGE}:sha-${sha}"
  if ghcr_reachable; then
    log "docker pull $ref"
    if timeout 180 docker pull "$ref"; then
      docker tag "$ref" "$IMAGE"
      pin_image
      return 0
    fi
    log "ghcr pull failed for $ref"
  else
    log "ghcr.io:443 closed from this host"
  fi

  tmp=$(mktemp)
  log "download release asset lifeos-${sha}.tar.gz"
  http=$(curl -sS -L --retry 3 --retry-delay 2 --connect-timeout 15 --max-time 300 \
    -o "$tmp" -w '%{http_code}' \
    "https://github.com/ezhigval/lifeos/releases/download/${RELEASE_TAG}/lifeos-${sha}.tar.gz" || true)
  if [[ "$http" == "200" ]] && gzip -t "$tmp"; then
    if docker load -i "$tmp"; then
      rm -f "$tmp"
      docker image inspect "$IMAGE" >/dev/null
      pin_image
      return 0
    fi
    rm -f "$tmp"
    log "docker load failed"
    return 1
  fi
  rm -f "$tmp"
  if [[ "$http" == "404" ]]; then
    log "image for $sha is not published yet"
    return 2
  fi
  log "release download failed http=$http"
  return 1
}

# Tests source this file. A normal run falls through.
if [[ "${LIFEOS_DEPLOY_LIB:-}" == 1 ]]; then
  return 0 2>/dev/null || exit 0
fi

exec 9>"$LOCK"
if ! flock -n 9; then
  echo "deploy already running"
  exit 0
fi

apply_oom_scores
retire_nginx || true

cd "$REPO"
# github.com:22 from this VM often times out. Fail fast and keep going
# with the checkout that is already on disk.
if ! GIT_SSH_COMMAND="ssh -o ConnectTimeout=8 -o StrictHostKeyChecking=accept-new" \
  git fetch --quiet origin "$BRANCH"; then
  log "git fetch failed; using the checkout already on disk"
fi
REMOTE=$(git rev-parse "origin/$BRANCH")
LOCAL=$(cat "$SHA_FILE" 2>/dev/null || echo none)
# The stamp can say the SHA is deployed while the container still serves
# an older image (pull skipped after the file was written).
image_stale=0
if [[ -f "$REPO/web/miniapp/index.html" ]] && grep -q '__LIFEOS_MARK' "$REPO/web/miniapp/index.html"; then
  if docker exec lifeos-app-1 grep -q '__LIFEOS_MARK' /app/web/index.html 2>/dev/null; then
    image_stale=0
  else
    image_stale=1
    log "running image has no __LIFEOS_MARK; will pull even if the stamp matches"
  fi
fi
if [[ "$REMOTE" == "$LOCAL" && "$image_stale" -eq 0 ]]; then
  ensure_postgres_localhost
  log "no changes ($REMOTE), skipping"
  exit 0
fi

fail_stamp="$BASE/.last_build_fail"
if [[ -f "$fail_stamp" ]]; then
  prev=$(cut -f1 "$fail_stamp")
  when=$(cut -f2 "$fail_stamp")
  now=$(date +%s)
  if [[ "$prev" == "$REMOTE" && $((now - when)) -lt 3600 ]]; then
    ensure_postgres_localhost
    log "image fetch of $REMOTE failed recently; skipping for an hour"
    exit 0
  fi
fi

log "deploying $LOCAL -> $REMOTE"
git checkout -f "$BRANCH"
git reset --hard "origin/$BRANCH"

repo_script="$REPO/deployments/vm-pull-deploy.sh"
if [[ -f "$repo_script" && -f "$SELF" ]] && ! cmp -s "$repo_script" "$SELF"; then
  install -m 755 "$repo_script" "$SELF"
  log "deploy script updated from repo; re-exec"
  exec "$SELF"
fi

retire_nginx || true
ensure_postgres_localhost

fetch_rc=0
fetch_image "$REMOTE" || fetch_rc=$?
if [[ "$fetch_rc" -eq 2 ]]; then
  exit 0
fi
if [[ "$fetch_rc" -ne 0 ]]; then
  printf '%s\t%s\n' "$REMOTE" "$(date +%s)" > "$fail_stamp"
  log "image fetch failed"
  exit 1
fi
rm -f "$fail_stamp"

backup_postgres
compose up -d postgres
compose run --rm --no-deps --no-build app migrate up
compose up -d --no-build --force-recreate app
docker image prune -f >/dev/null 2>&1 || true
ok=0
for _ in 1 2 3 4 5 6; do
  if curl -fsS http://127.0.0.1:8080/health >/dev/null; then
    ok=1
    break
  fi
  sleep 3
done
if [[ "$ok" -ne 1 ]]; then
  printf '%s\t%s\n' "$REMOTE" "$(date +%s)" > "$fail_stamp"
  log "health check failed after deploy"
  exit 1
fi
printf '%s\n' "$REMOTE" > "$SHA_FILE"
log "health OK ($REMOTE)"
