# Domain Model

**Version:** 0.4  
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

## AI (ports, не domain)

```go
type IntentResolver interface {
    Resolve(ctx context.Context, input ResolveInput) (ResolvedIntent, error)
}

type Assistant interface {
    Summarize(ctx context.Context, req SummaryRequest) (string, error)
}
```

Реализации: `rulebased` (default), `ollama` (optional), `composite`.

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
```

**Integration:** shared `UserID`, слабая связь через IDs, cross-context reads в `internal/query/`.

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
