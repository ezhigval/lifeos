# ER Diagram

**Version:** 0.4  
**Status:** Current (synced to migrations 00001–00037, 2026-10-22)  
**See also:** [ARCHITECTURE.md](../architecture/ARCHITECTURE.md), [DOMAIN_MODEL.md](../architecture/DOMAIN_MODEL.md)

> Goals удалены (00024): KPI и иерархия — в `projects`.  
> Tasks связаны с projects через `task_projects` (M:N).

---

## Core Schema

```mermaid
erDiagram
    users ||--|| user_settings : has
    users ||--o{ tasks : owns
    users ||--o{ projects : owns
    users ||--o{ life_spheres : owns
    users ||--o{ day_availability : has
    users ||--o{ scheduled_jobs : owns
    users ||--o{ domain_events : generates
    users ||--o{ user_memories : remembers
    users ||..o{ anon_learning_events : anonymized_to

    projects ||--o{ task_projects : links
    tasks ||--o{ task_projects : links
    projects ||--o{ project_spheres : in
    life_spheres ||--o{ project_spheres : contains
    tasks ||--o{ task_spheres : in
    life_spheres ||--o{ task_spheres : contains

    users {
        uuid id PK
        bigint telegram_id UK
        varchar display_name
        varchar timezone
        timestamptz created_at
    }

    tasks {
        uuid id PK
        uuid user_id FK
        varchar title
        varchar status
        varchar priority
        date due_date
        int duration_minutes
        text[] tags
        timestamptz completed_at
        timestamptz deleted_at
        timestamptz created_at
    }

    projects {
        uuid id PK
        uuid user_id FK
        varchar name
        varchar status
        numeric target_value
        numeric current_value
        varchar unit
        date target_date
    }

    life_spheres {
        uuid id PK
        uuid user_id FK
        varchar name
        int sort_order
    }

    task_projects {
        uuid task_id PK_FK
        uuid project_id PK_FK
    }

    project_spheres {
        uuid project_id PK_FK
        uuid sphere_id PK_FK
    }

    scheduled_jobs {
        uuid id PK
        uuid user_id FK
        varchar job_type
        jsonb payload
        timestamptz run_at
        varchar status
    }

    user_memories {
        uuid id PK
        uuid user_id FK
        varchar kind  -- preference|fact|alias|pattern
        varchar key
        text value
        float confidence
        varchar source
        timestamptz created_at
        timestamptz updated_at
        varchar user_id_kind_key PK
    }

    anon_learning_events {
        uuid id PK
        varchar anon_subject  -- HMAC-SHA256(userID, salt)
        varchar event_type
        varchar tool_or_intent
        bool success
        int ask_rounds
        varchar model
        jsonb meta  -- no PII
        timestamptz created_at
    }

    task_spheres {
        uuid task_id PK_FK
        uuid sphere_id PK_FK
    }

    processed_updates {
        bigint update_id PK
        timestamptz processed_at
    }
```

---

## Extended Domains

| Domain | Tables | Migrations |
|--------|--------|------------|
| Calendar | `calendar_events` | 00012 |
| Knowledge | `notes`, `note_tags` | 00014–00015, 00033 |
| Health | `health_weight_logs`, `health_step_logs`, `health_sleep_logs` | 00016–00018 |
| Career | `career_contacts`, `career_skills` | 00019–00020 |
| Auth (sessions) | `telegram_sessions` | 00007, 00030 |
| Finance | `finance_transactions`, `finance_categories`, `debts`, `planned_cashflows` | 00008–00009, 00029 |
| Habits | `habits`, `habit_logs` | 00010, 00032 |
| Memory (агент) | `user_memories` | 00030 |
| Learning (анонимное, opt-in) | `anon_learning_events` | 00030 |
| Task extensions | `task_projects`, `task_spheres`, `sphere_domain_links` | 00023, 00035–00036 |
| Home widgets | `user_settings.home_widgets` (JSONB) | 00034 |
| Telegram login codes | `telegram_login_codes` | 00037 |

---

## Indexes (ключевые)

| Table | Index | Purpose |
|-------|-------|---------|
| tasks | (user_id, due_date) WHERE deleted_at IS NULL | Today view |
| scheduled_jobs | (run_at, status) WHERE status='pending' | Scheduler poll |
| processed_updates | (update_id) | Telegram idempotency |
| life_spheres | (user_id, sort_order) | Sphere list |
| user_memories | UNIQUE (user_id, kind, key) | Idempotent memory writes |
| anon_learning_events | (anon_subject), (created_at) | Analytics per subject |

## Notes

- Деньги: `amount_cents int64`, currency default `RUB`
- `processed_updates` — idempotency для Telegram
- `scheduled_jobs` — единая очередь reminders + reviews
- `anon_learning_events` связан с `users` мягко (HMAC-subject, без FK) — анонимность сильнее схемы
