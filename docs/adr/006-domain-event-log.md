# ADR-006: Domain Event Log

## Status
Accepted

## Context
Нужен аудит, будущая аналитика, возможность event sourcing projections.

## Decision
Append-only таблица `domain_events`. Use cases публикуют события в той же TX что и мутация.

## Consequences
**+** Полная история изменений  
**+** Analytics read models позже  
**−** Дополнительная запись на каждую мутацию  
**−** Не full event sourcing (state всё ещё в tables)

## Event Schema
`aggregate_type`, `aggregate_id`, `event_type`, `payload` (JSONB), `occurred_at`, `source` (telegram | scheduler | cli)
