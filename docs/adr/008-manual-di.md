# ADR-008: Manual Dependency Injection

## Status
Accepted

## Context
Нужен DI без magic. Wire/Fx добавляют complexity.

## Decision
Ручной wiring в `cmd/lifeos/serve.go`. Конструкторы `NewX(deps)` для каждого use case.

## Consequences
**+** Явный граф зависимостей  
**+** Нет code generation для DI  
**−** serve.go растёт → разбить на `wire_*.go` при необходимости

## Revisit When
> 30 use cases → рассмотреть Wire
