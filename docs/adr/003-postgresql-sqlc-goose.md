# ADR-003: PostgreSQL + SQLC + Goose

## Status
Accepted

## Context
Нужен надёжный relational store, type-safe queries, versioned migrations.

## Decision
- **PostgreSQL 16** — primary store
- **pgx/v5** — driver + connection pool
- **SQLC** — generate Go from SQL
- **Goose** — migrations

## Consequences
**+** Compile-time query safety  
**+** SQL остаётся явным и reviewable  
**−** SQLC требует rebuild при изменении queries  
**−** Сложные dynamic queries — raw SQL вручную

## Alternatives
- GORM — отклонено: magic, слабее контроль SQL
- Ent — отклонено: ORM overhead для Clean Arch
