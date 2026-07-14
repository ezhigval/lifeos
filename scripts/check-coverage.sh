#!/usr/bin/env bash
# Fail if critical Go packages drop below statement-coverage minimums.
#
# Usage: scripts/check-coverage.sh [coverprofile]
# Default: coverage.out from `go test ./... -coverprofile=coverage.out`
#
# Parses `go tool cover -func` package totals by matching each configured
# package path prefix against function rows and comparing the package-level
# percentage reported when that package was tested (embedded via recomputing
# from coverprofile block counts).
set -euo pipefail

PROFILE="${1:-coverage.out}"
if [[ ! -f "$PROFILE" ]]; then
  echo "check-coverage: missing cover profile: $PROFILE" >&2
  echo "hint: run  go test ./... -coverprofile=coverage.out  first" >&2
  exit 1
fi

# package path (repo-relative) -> minimum percent (inclusive)
# Keep aligned with docs/agents/reports/backend/TASK-009.md
MIN_TASKS_DOMAIN=75
MIN_IDENTITY_DOMAIN=80
MIN_AI_RULEBASED=65
MIN_FINANCE_DOMAIN=80
MIN_FINANCE_APP=70
MIN_PROJECTS_DOMAIN=80
MIN_SPHERES_DOMAIN=80
MIN_TASKS_APP=65

package_pct() {
  local pkg="$1"
  # Cover profile lines: path:start.col,end.col statements count
  # mode: set | count | atomic
  awk -v pkg="/${pkg}/" '
    NR == 1 { next }
    index($1, pkg) {
      # $1 = file:range, $2 = statements, $3 = count
      n = split($1, a, ":")
      # last field before range is awkward; fields are: file:line.col,line.col stmts count
      stmts = $(NF-1) + 0
      count = $NF + 0
      total += stmts
      if (count > 0) covered += stmts
    }
    END {
      if (total == 0) {
        print ""
        exit 1
      }
      printf "%.1f", 100.0 * covered / total
    }
  ' "$PROFILE"
}

check_one() {
  local pkg="$1"
  local min="$2"
  local pct
  if ! pct="$(package_pct "$pkg")"; then
    echo "check-coverage: no coverage data for $pkg" >&2
    return 1
  fi
  printf '%-36s %7s%%  (min %s%%)\n' "$pkg" "$pct" "$min"
  awk -v a="$pct" -v m="$min" 'BEGIN { exit !(a+0 >= m+0) }' || {
    echo "  FAIL: $pkg ${pct}% < ${min}%" >&2
    return 1
  }
}

fail=0
echo "Coverage gate ($PROFILE):"
check_one "internal/tasks/domain" "$MIN_TASKS_DOMAIN" || fail=1
check_one "internal/identity/domain" "$MIN_IDENTITY_DOMAIN" || fail=1
check_one "internal/ai/rulebased" "$MIN_AI_RULEBASED" || fail=1
check_one "internal/finance/domain" "$MIN_FINANCE_DOMAIN" || fail=1
check_one "internal/finance/app" "$MIN_FINANCE_APP" || fail=1
check_one "internal/projects/domain" "$MIN_PROJECTS_DOMAIN" || fail=1
check_one "internal/spheres/domain" "$MIN_SPHERES_DOMAIN" || fail=1
check_one "internal/tasks/app" "$MIN_TASKS_APP" || fail=1

if [[ "$fail" -ne 0 ]]; then
  echo "check-coverage: FAILED" >&2
  exit 1
fi
echo "check-coverage: OK"
