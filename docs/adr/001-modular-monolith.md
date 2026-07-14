# ADR-001: Modular Monolith over Microservices

## Status
Accepted

## Context
LifeOS — personal single-user system на домашнем Mac. Нужна долгосрочная расширяемость без операционной сложности microservices.

## Decision
Используем **modular monolith**: один deployable binary, bounded contexts как Go modules внутри `internal/`.

## Consequences
**+** Простой деплой, одна БД, низкая latency  
**+** Чёткие границы через package structure  
**−** Нужна дисциплина import rules  
**−** Scale horizontal ограничен (достаточно для 1 user)

## Alternatives Considered
- Microservices — отклонено: overkill
- Plain monolith без границ — отклонено: tech debt
