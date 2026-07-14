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

## Future
`LLMResolver` implements same interface, выбирается в composition root.
