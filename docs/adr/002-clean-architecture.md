# ADR-002: Clean Architecture Layering

## Status
Accepted (revised v0.2)

## Context
Система должна пережить смену transport (Telegram → Web/Mobile) и пережить годы разработки.

## Decision
Каждый bounded context: `domain` → `app` → `infra`. Transport в `internal/transport/` вызывает app use cases.

## Consequences
**+** Бизнес-логика тестируется без БД и Telegram  
**+** Замена адаптеров без изменения domain  
**−** Больше файлов и boilerplate на старте

## Rules
1. Domain imports: `time`, `errors`, `fmt`, `internal/platform/ids` only
2. Repository interfaces — в `domain` или `infra` (консистентно внутри модуля)
3. Reader ports для cross-context — у потребителя в `app` (e.g. `ProjectReader` in tasks/app)
4. Use case = один публичный `Execute` method
5. Cross-context reads — `internal/query/`, не import чужого domain

## Revision v0.2
- `application/` переименован в `app/` (Go convention)
- `transport/` перемещён в `internal/transport/`
- Убрано нереалистичное ограничение «domain imports только context, time, errors, fmt» без typed IDs
