# ADR-007: Built-in Scheduler

## Status
Accepted (revised v0.2)

## Context
Напоминания, обзоры, recurring tasks. Cron на хосте — внешняя зависимость, не переносится с app.

## Decision
In-process scheduler: goroutine + 1-minute ticker + **unified `scheduled_jobs` table**.

All timed work (reminders, morning/evening reviews) — rows in `scheduled_jobs` with `job_type`.

## Consequences
**+** Атомарность с app lifecycle  
**+** Единый execution path — проще тесты и отладка  
**+** Тестируемо с fake clock  
**−** При multi-instance нужен distributed lock (Redis, compose profile `cache`, Phase 2)

## v1
Single instance. No Redis required. `FOR UPDATE SKIP LOCKED` готов к multi-instance.

## Revision v0.2
- Убрана отдельная execution-таблица `reminders`
- Domain `Reminder` → `ScheduledJob` с `job_type=reminder`
- Scheduler queries только `scheduled_jobs`

## Job Types (MVP)
- `reminder`
- `morning_review`
- `evening_review`
