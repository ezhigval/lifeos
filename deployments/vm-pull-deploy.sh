#!/usr/bin/env bash
# Pull-deploy for a ~2GB VM. The image build is capped so it cannot OOM-kill
# sshd or nginx. Safe to run from lifeos-deploy.timer every few minutes.
set -euo pipefail

BRANCH="${LIFEOS_BRANCH:-main}"
BASE=/opt/lifeos
REPO="$BASE/repo"
SHA_FILE="$BASE/.last_deploy_sha"
SELF="$BASE/bin/vm-pull-deploy.sh"
LOCK=/var/lock/lifeos-deploy.lock
IMAGE=lifeos-app:deploy

exec 9>"$LOCK"
if ! flock -n 9; then
  echo "deploy already running"
  exit 0
fi

log() { echo "[lifeos-deploy] $*"; }

apply_oom_scores() {
  local unit score dir name pid
  for unit in ssh.service sshd.service nginx.service docker.service containerd.service; do
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
  for name in sshd nginx dockerd containerd; do
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

install_nginx_site() {
  local src=""
  if [[ -f "$REPO/deployments/nginx/lifeos.conf" ]]; then
    src="$REPO/deployments/nginx/lifeos.conf"
  elif [[ -f "$BASE/bin/nginx-lifeos.conf" ]]; then
    src="$BASE/bin/nginx-lifeos.conf"
  else
    install -d -m 755 "$BASE/bin"
    src="$BASE/bin/nginx-lifeos.conf"
    cat > "$src" <<'NGINX'
# Fallback copy of deployments/nginx/lifeos.conf, used when that file
# is not in the checked-out revision yet.
server {
    listen 80 default_server;
    server_name _;
    gzip off;
    client_max_body_size 8m;
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Accept-Encoding $http_accept_encoding;
        proxy_buffering on;
        proxy_read_timeout 120s;
    }
}
NGINX
  fi
  command -v nginx >/dev/null 2>&1 || return 0
  install -d -m 755 /etc/nginx/sites-available /etc/nginx/sites-enabled
  if [[ -f /etc/nginx/sites-available/lifeos.conf ]]; then
    cp -a /etc/nginx/sites-available/lifeos.conf /etc/nginx/sites-available/lifeos.conf.bak
  fi
  install -m 644 "$src" /etc/nginx/sites-available/lifeos.conf
  ln -sfn /etc/nginx/sites-available/lifeos.conf /etc/nginx/sites-enabled/lifeos.conf
  rm -f /etc/nginx/sites-enabled/default /etc/nginx/sites-enabled/lifeos
  if nginx -t; then
    if systemctl reload nginx; then
      log "nginx site reloaded"
      return 0
    fi
    if systemctl start nginx; then
      log "nginx was down; started"
      return 0
    fi
    log "nginx reload and start failed"
    return 1
  fi
  log "nginx -t failed; restoring previous site"
  if [[ -f /etc/nginx/sites-available/lifeos.conf.bak ]]; then
    mv /etc/nginx/sites-available/lifeos.conf.bak /etc/nginx/sites-available/lifeos.conf
  fi
  return 1
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

# return 0 built, 2 skipped (retry later), 1 failed
build_app() {
  local avail_kb disk_kb
  avail_kb=$(awk '/MemAvailable:/ {print $2}' /proc/meminfo)
  disk_kb=$(df -Pk / | awk 'NR==2 {print $4}')
  log "MemAvailable=${avail_kb}kB disk_avail=${disk_kb}kB"
  if [[ "${avail_kb:-0}" -lt 400000 ]]; then
    log "skip build: under 400MB available; timer will retry"
    return 2
  fi
  docker image prune -f >/dev/null 2>&1 || true
  docker builder prune -f >/dev/null 2>&1 || true
  disk_kb=$(df -Pk / | awk 'NR==2 {print $4}')
  log "disk_avail after prune=${disk_kb}kB"
  if [[ "${disk_kb:-0}" -lt 1500000 ]]; then
    log "skip build: under 1.5GB free disk; timer will retry"
    return 2
  fi
  log "docker build (classic builder, memory cap)"
  local help flags=()
  help=$(docker build --help 2>&1 || true)
  if grep -q -- '--memory ' <<<"$help" || grep -q -- '--memory=' <<<"$help"; then
    flags+=(--memory=1200m)
  fi
  if grep -q -- '--memory-swap' <<<"$help"; then
    flags+=(--memory-swap=2200m)
  fi
  if grep -q -- '--oom-score-adj' <<<"$help"; then
    flags+=(--oom-score-adj=800)
  fi
  if ! (
    cd "$REPO"
    # BuildKit ignores --memory and the compile can OOM the host.
    DOCKER_BUILDKIT=0 docker build \
      "${flags[@]}" \
      -f deployments/Dockerfile \
      -t "$IMAGE" \
      .
  ); then
    return 1
  fi
  cat > "$BASE/docker-compose.buildimage.yml" <<EOF
# Points compose at the memory-capped image. Generated by vm-pull-deploy.sh.
services:
  app:
    image: ${IMAGE}
EOF
  return 0
}

apply_oom_scores
install_nginx_site || true

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
# an older image (build skipped or OOM'd after the file was written).
image_stale=0
if [[ -f "$REPO/web/miniapp/index.html" ]] && grep -q '__LIFEOS_MARK' "$REPO/web/miniapp/index.html"; then
  if docker exec lifeos-app-1 grep -q '__LIFEOS_MARK' /app/web/index.html 2>/dev/null; then
    image_stale=0
  else
    image_stale=1
    log "running image has no __LIFEOS_MARK; will build even if the stamp matches"
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
    log "build of $REMOTE failed recently; skipping for an hour"
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

install_nginx_site || true
ensure_postgres_localhost

build_rc=0
build_app || build_rc=$?
if [[ "$build_rc" -eq 2 ]]; then
  exit 0
fi
if [[ "$build_rc" -ne 0 ]]; then
  printf '%s\t%s\n' "$REMOTE" "$(date +%s)" > "$fail_stamp"
  log "build failed"
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
