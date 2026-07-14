# ER Diagram

**Version:** 0.3  
**See also:** [ARCHITECTURE.md](../architecture/ARCHITECTURE.md)

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

    projects ||--o{ task_projects : links
    tasks ||--o{ task_projects : links
    projects ||--o{ project_spheres : in
    life_spheres ||--o{ project_spheres : contains

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
        timestamptz completed_at
        timestamptz deleted_at
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

    processed_updates {
        bigint update_id PK
        timestamptz processed_at
    }
```

---

## Finance

`finance_transactions`, `finance_categories`, `debts` — см. миграции 00008–00009.

---

## Habits

`habits`, `habit_logs` — миграция 00010.

---

## Extended Domains

| Domain | Tables |
|--------|--------|
| Calendar | `calendar_events` |
| Knowledge | `notes`, `note_tags` |
| Health | `health_weight`, `health_steps`, `health_sleep` |
| Career | `career_contacts`, `career_skills` |

---

## Indexes (ключевые)

| Table | Index | Purpose |
|-------|-------|---------|
| tasks | (user_id, due_date) WHERE deleted_at IS NULL | Today view |
| scheduled_jobs | (run_at, status) WHERE status='pending' | Scheduler poll |
| processed_updates | (update_id) | Telegram idempotency |
| life_spheres | (user_id, sort_order) | Sphere list |

## Notes

- Деньги: `amount_cents int64`, currency default `RUB`
- `processed_updates` — idempotency для Telegram
- `scheduled_jobs` — единая очередь reminders + reviews
