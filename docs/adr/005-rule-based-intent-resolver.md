# ADR-005: Rule-Based Intent Resolver (AI Stub)

## Status
Accepted

## Context
NL-команды нужны с первого дня. LLM добавляет зависимость, latency, cost.

## Decision
Интерфейс `IntentResolver` с реализацией `RuleBasedResolver` (regex + keywords + fuzzy time parsing).  
Output: typed `ResolvedIntent` structs per intent (не `map[string]any`).

## Consequences
**+** Офлайн, детерминировано, тестируемо  
**+** Use cases не знают про LLM  
**−** Ограниченное понимание свободного текста  
**−** Нужно итеративно расширять правила

## Future / Stage 3.2
Optional Ollama implements the same ports. Composition root (`cmd/lifeos/cmd/resolver.go`) keeps **rule-based primary**; LLM runs only for `unknown` when `LIFEOS_LLM_ENABLED=true`. Failures/timeouts degrade silently. Reviews: LLM primary with `reviewsafe` HTML sanitization + templateassistant fallback.
