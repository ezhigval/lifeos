# Domain Model

**Version:** 0.5  
**Status:** Current (synced to code, 2026-10-22)  
**See also:** [ARCHITECTURE.md](ARCHITECTURE.md)

---

## Shared Kernel

Пакет: `internal/platform/ids`

```
UserID, TaskID, ProjectID, SphereID, …   // typed uuid aliases
Priority      enum: low | medium | high | urgent
```

`Money`: `{ Amount int64 /* копейки */, Currency string }` — default `RUB`.

---

## Identity

**Aggregate:** `User`

| Field | Type | Notes |
|-------|------|-------|
| ID | UserID | |
| TelegramID | int64 | unique |
| DisplayName | string | |
| Timezone | string | IANA, default Europe/Moscow |
| CreatedAt | time.Time | |

---

## Settings

**Aggregate:** `UserSettings` (1:1 с User)

| Field | Type |
|-------|------|
| UserID | UserID |
| MorningReviewAt | time (HH:MM) |
| EveningReviewAt | time |
| QuietHoursStart | optional time |
| QuietHoursEnd | optional time |
| Language | string (ru) |

---

## Tasks

**Aggregate:** `Task`

| Field | Type | Notes |
|-------|------|-------|
| ID | TaskID | |
| UserID | UserID | |
| Title | string | название |
| Description | optional string | |
| Status | todo \| in_progress \| done \| cancelled | |
| Priority | Priority | |
| DueDate | optional date | **дата реализации** |
| DurationMinutes | optional int | **длительность** (оценка, минуты > 0) |
| Tags | []string | **хештеги** без `#` (для фильтрации) |
| ProjectIDs | []ProjectID | принадлежность к проектам (M:N `task_projects`) |
| CompletedAt | optional time | |
| DeletedAt | optional time | soft-delete |
| CreatedAt | time.Time | **дата создания** (авто) |
| UpdatedAt | time.Time | |

**Invariants:** title не пустой; done → completed_at; cancelled нельзя complete; duration > 0 если задана; done/cancelled нельзя edit/reschedule.

**Жизненный цикл:**

| Действие | Поведение |
|----------|-----------|
| Add | создать задачу (title, due_date, duration, tags, projects) |
| Complete | status → done |
| Cancel | status → cancelled |
| Edit | изменение полей открытой задачи |
| Reschedule | смена due_date |
| Auto-reschedule | вечерний обзор: открытые с due_date ≤ сегодня → завтра + уведомление в Telegram |

---

## Spheres

**Aggregate:** `LifeSphere` — области жизни (Карьера, Здоровье, …)

| Field | Type |
|-------|------|
| ID | SphereID |
| UserID | UserID |
| Name | string |
| SortOrder | int |

---

## Projects

**Aggregate:** `Project` — контейнер задач и KPI (заменяет бывшие Goals)

| Field | Type |
|-------|------|
| ID | ProjectID |
| UserID | UserID |
| Name | string |
| Description | optional |
| Outcome | optional |
| Status | active \| archived \| completed |
| TargetValue | optional decimal |
| CurrentValue | decimal |
| Unit | optional string |
| TargetDate | optional date |
| SphereIDs | []SphereID (M:N через `project_spheres`) |

**Value Object:** `Progress` — percent / remaining от target.

---

## Planning

**Entity:** `DayAvailability` — доступность пользователя на день.

---

## Notifications

**Aggregate:** `ScheduledJob`

| job_type | Description |
|----------|-------------|
| `reminder` | One-shot reminder |
| `morning_review` | Daily morning review |
| `evening_review` | Daily evening review (+ auto-reschedule incomplete tasks → tomorrow + notify) |

Reminder = `ScheduledJob` с `job_type=reminder`.

---

## Finance

**Aggregates:** `FinanceTransaction`, `Debt`, `Category`

Деньги хранятся в `amount_cents int64`.

---

## Habits

**Aggregates:** `Habit`, `HabitLog`

---

## Calendar

**Aggregate:** `CalendarEvent`

---

## Knowledge

**Aggregate:** `Note` (с тегами)

---

## Health

**Entities:** `WeightEntry`, `StepsEntry`, `SleepEntry`

---

## Career

**Entities:** `Contact`, `Skill`

---

## Memory

**Пакет:** `internal/memory` · **Таблица:** `user_memories` (миграция 00030)

**Aggregate:** `Memory`

| Field | Type | Notes |
|-------|------|-------|
| ID | MemoryID | |
| UserID | UserID | |
| Kind | enum | `preference` \| `fact` \| `alias` \| `pattern` |
| Key | string | unique вместе с (UserID, Kind) |
| Value | string | |
| Confidence | float64 | default 0.7 |
| Source | string | default `agent` |
| CreatedAt / UpdatedAt | time.Time | |

Инварианты: непустые `Key`/`Value`, валидный `Kind`; пара `(user_id, kind, key)` уникальна — повторная запись обновляет существующую память. Флаг `UserSettings.memory_enabled` (default `true`) включает подмешивание памяти в контекст агента.

---

## Learning (анонимное обучение, opt-in)

**Пакет:** `internal/learning` · **Таблица:** `anon_learning_events` (миграция 00030)

**Entity:** `Event`

| Field | Type | Notes |
|-------|------|-------|
| ID | LearningEventID | |
| AnonSubject | string | HMAC-SHA256 от UserID + salt; raw user id никогда не сохраняется |
| Type | string | тип сигнала (например, tool call) |
| ToolOrIntent | string? | |
| Success | *bool | nil — неизвестно |
| AskRounds | int | число уточняющих вопросов |
| Model | string | имя LLM-модели |
| Meta | JSONB | только грубые флаги, без PII |
| CreatedAt | time.Time | |

Включение: `UserSettings.learning_opt_in` (default `false`). События пишутся append-only; сырые сообщения, имена и суммы в `Meta` запрещены доменным правилом.

---

## AI (ports, не domain)

```go
type IntentResolver interface {
    Resolve(ctx context.Context, input ResolveInput) (ResolvedIntent, error)
}

type Assistant interface {
    Summarize(ctx context.Context, req SummaryRequest) (string, error)
}

type ConversationalAgent interface {   // диалог с tools (LLM function calling)
    RunTurn(ctx context.Context, req AgentRequest) (AgentReply, error)
}

type STT interface {                   // расшифровка голосовых сообщений
    Transcribe(ctx context.Context, audio AudioAttachment) (string, error)
}

type Vision interface {                // описание фотографий/скриншотов
    Describe(ctx context.Context, img ImageAttachment, hint string) (string, error)
}
```

Реализации: `rulebased` (default), `ollama` (optional), `openai` (Groq по умолчанию), `composite`.
STT/Vision включаются флагами конфигурации; без LLM система остаётся полностью работоспособной.

---

## Context Map

```
Identity → Settings
    ↓
Tasks ↔ Projects ↔ Spheres
    ↓
Planning, Finance, Habits, Calendar, Knowledge, Health, Career
    ↓
Notifications (scheduled_jobs)
    ↓
query/ (read models: priorities, reviews, analytics)

AI (ports) ──→ Memory (контекст агента: предпочтения/факты)
          └──→ Learning (append-only анонимные сигналы, opt-in)
```

**Integration:** shared `UserID`, слабая связь через IDs, cross-context reads в `internal/query/`.
`Memory` и `Learning` — leaf-контексты: никто не зависит от них, кроме `ai`/transport.

---

## Ubiquitous Language

| RU | Domain term |
|----|-------------|
| Задача | Task |
| Проект | Project |
| Сфера | LifeSphere |
| Напоминание | ScheduledJob (reminder) |
| Обзор | Review job |
| Долг | Debt |
