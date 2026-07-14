# TASK-007 — Stage 3.2 Assistant presentation (Telegram)

**Agent:** telegram  
**Date:** 2026-07-14  
**Branch:** `cursor/docs-sync-orchestration-fe85`  
**Base SHA:** `a2fafe9c80e902b9878f1626ff84f7d6f714cf2e`  
**Scope:** `internal/transport/telegram/**`  
**Status:** DONE

---

## Verdict

Soft harden for free-form assistant/review/triage HTML under `parse_mode=HTML`. Structured `Format*` paths already escaped user fields. Presentation-layer sanitize complements Backend `reviewsafe` (in flight) without depending on it.

---

## Findings

| Path | Before | Action |
|------|--------|--------|
| Most `Format*` (titles, notes, tags, …) | `html.EscapeString` | OK — left as-is |
| `FormatDashboard` | Expects pre-safe HTML body | Documented; no sanitize (would double-escape) |
| `FormatTriage` | Passthrough; unused by `ActionTriage` | Sanitize + wire into actions |
| Weekly/monthly review dispatch | Raw query HTML string | Wrap with `FormatAssistantHTML` |
| Morning/evening notify (`cmd`) | Source-side escape | Backend/`reviewsafe` (not telegram Format*) |

---

## Soft harden

- **`FormatAssistantHTML`**: full escape → restore allowlist (`b|strong|i|em`) → truncate to 3500 runes (dashboard wrapper headroom).
- **`FormatTriage`** → `FormatAssistantHTML`; **`ActionTriage`** uses it.
- **`IntentReviewWeekly` / `IntentReviewMonthly`** present via `FormatAssistantHTML`.

---

## Tests

```text
go test ./internal/transport/telegram/...   # OK
go vet ./internal/transport/telegram/...    # OK
```

---

## Commits

- `c796e397649d113eb924545694577a67ed6c736e` — `fix(telegram): sanitize assistant/review HTML for parse_mode`
