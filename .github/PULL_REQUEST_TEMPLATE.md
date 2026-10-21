## What does this PR do?

Fixes # (issue)

## Type of change

- [ ] Bug fix
- [ ] Feature
- [ ] Refactor / architecture boundary change
- [ ] Docs only

## Checklist

- [ ] `make ci` passes locally (build, test, lint)
- [ ] Migrations added if schema changed (goose Up/Down, next number after 00037)
- [ ] New bounded context respects import rules (domain ← app ← transport; no cross-context infra imports)
- [ ] No secrets committed (`bash scripts/check-no-secrets.sh`)
- [ ] Docs updated: ARCHITECTURE.md / DOMAIN_MODEL.md / CHANGELOG.md if behavior changed
- [ ] OpenAPI spec updated if HTTP contract changed (`docs/api/openapi.yaml`)
