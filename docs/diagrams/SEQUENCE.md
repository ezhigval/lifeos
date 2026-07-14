# Sequence Diagrams

**Version:** 0.3  
**See also:** [ARCHITECTURE.md](../architecture/ARCHITECTURE.md)

---

## 1. Incoming Telegram Message

```mermaid
sequenceDiagram
    actor User
    participant TG as Telegram API
    participant Poller as Long Polling
    participant H as MessageHandler
    participant IR as IntentResolver
    participant UC as UseCase
    participant DB as PostgreSQL
    participant FMT as ResponseFormatter

    User->>TG: "Добавь задачу купить фильтр"
    TG->>Poller: Update
    Poller->>Poller: Check allowed user_id
    Poller->>Poller: Idempotency (processed_updates)
    Poller->>H: HandleIncomingMessage(ctx, text)
    H->>IR: Resolve(ctx, text)
    IR-->>H: ResolvedIntent{task.create, title:"купить фильтр"}
    H->>UC: CreateTask.Execute(ctx, cmd)
    UC->>DB: BEGIN TX
    UC->>DB: INSERT tasks
    UC->>DB: INSERT domain_events (source=telegram)
    UC->>DB: COMMIT
    UC-->>H: TaskDTO
    H->>FMT: FormatTaskCreated(task)
    FMT-->>H: "✅ Задача создана: купить фильтр"
    H->>TG: sendMessage (HTML)
    TG->>User: Response
```

## 2. Reminder Scheduling (unified scheduled_jobs)

```mermaid
sequenceDiagram
    actor User
    participant TG as Telegram API
    participant H as MessageHandler
    participant IR as IntentResolver
    participant SR as ScheduleReminder UC
    participant DB as PostgreSQL
    participant SCH as Scheduler
    participant NS as TelegramNotifier

    User->>TG: "Напомни вечером позвонить"
    TG->>H: Update
    H->>IR: Resolve
    IR-->>H: ResolvedIntent{reminder.create, msg, fire_at: today 19:00}
    H->>SR: Execute
    SR->>DB: INSERT scheduled_jobs (job_type=reminder, pending)
    SR-->>H: OK
    H->>TG: "Напомню в 19:00"

    Note over SCH: Every 1 minute
    SCH->>DB: SELECT scheduled_jobs WHERE run_at <= now AND pending FOR UPDATE SKIP LOCKED
    DB-->>SCH: [job]
    SCH->>NS: Send(user, payload.message)
    NS->>TG: sendMessage
    SCH->>DB: UPDATE status=done
```

## 3. Morning Review (Scheduled)

```mermaid
sequenceDiagram
    participant SCH as Scheduler
    participant MR as MorningReview (query/)
    participant TR as TaskRepository
    participant GR as GoalRepository
    participant AS as Assistant (template)
    participant NS as TelegramNotifier
    participant TG as Telegram API
    actor User

    SCH->>DB: SELECT scheduled_jobs (job_type=morning_review)
    SCH->>MR: Execute(ctx, userID)
    MR->>TR: ListToday(userID)
    MR->>GR: ListActive(userID)
    MR->>AS: Summarize(tasks, goals)
    AS-->>MR: "Топ-3: ..."
    MR-->>SCH: ReviewMessage
    SCH->>NS: Send(userID, message)
    NS->>TG: sendMessage
    TG->>User: Morning review
    SCH->>DB: Reschedule next morning_review
```

## 4. Query Priorities

```mermaid
sequenceDiagram
    actor User
    participant TG as Telegram API
    participant H as Handler
    participant IR as IntentResolver
    participant QP as GetTopPriorities (query/)
    participant DB as PostgreSQL

    User->>TG: "Что сейчас самое важное?"
    TG->>H: Update
    H->>IR: Resolve
    IR-->>H: ResolvedIntent{query.priorities}
    H->>QP: Execute
    QP->>DB: SELECT tasks (urgent/high, due today or overdue)
    QP->>DB: SELECT goals (active, low progress)
    QP->>QP: Rank by algorithm
    QP-->>H: PriorityList
    H->>TG: Formatted response (HTML)
    TG->>User: Ranked list
```

## 5. Overloaded Day Triage

```mermaid
sequenceDiagram
    actor User
    participant TG as Telegram API
    participant H as Handler
    participant TD as TriageDay UC
    participant DB as PostgreSQL

    User->>TG: "Сегодня полный завал"
    TG->>H: Update
    H->>TD: Execute
    TD->>DB: SELECT all tasks for today
    TD->>TD: Classify: must / should / can_defer
    TD-->>H: TriageProposal
    H->>TG: Message + InlineKeyboard [Перенести low] [Ок]
    User->>TG: Click "Перенести low"
    TG->>H: CallbackQuery
    H->>TD: ApplyDefer(lowPriorityIDs)
    TD->>DB: UPDATE due_date = tomorrow
    H->>TG: "Перенесено N задач"
```

## 5b. Evening auto-reschedule

```mermaid
sequenceDiagram
    participant SCH as Scheduler
    participant RT as Runtime
    participant AR as AutoRescheduleIncomplete
    participant DB as PostgreSQL
    participant TG as TelegramNotifier

    SCH->>RT: evening_review job
    RT->>RT: Evening review text
    RT->>AR: Execute(user)
    AR->>DB: SELECT open tasks due_date <= today
    AR->>DB: UPDATE due_date = tomorrow
    RT->>TG: review text + list of moved tasks
```

## 6. Application Bootstrap

```mermaid
sequenceDiagram
    participant Main as cmd/lifeos
    participant CFG as Config
    participant PG as PostgreSQL
    participant APP as App
    participant SCH as Scheduler
    participant TG as Telegram Poller

    Main->>CFG: Load env
    Main->>PG: Connect + ping
    Main->>APP: Wire dependencies (manual DI)
    Main->>SCH: Start background
    Main->>TG: Start long polling
    Main->>Main: Wait SIGTERM
    Main->>TG: Stop (drain)
    Main->>SCH: Stop
    Main->>PG: Close
```

## 7. Future REST API (same use cases)

```mermaid
sequenceDiagram
    actor Client
    participant API as Chi HTTP
    participant AUTH as JWT Middleware
    participant H as TaskHandler
    participant UC as CreateTask UC
    participant DB as PostgreSQL

    Client->>API: POST /api/v1/tasks + JWT
    API->>AUTH: Validate token
    AUTH->>H: CreateTaskRequest
    H->>UC: Execute (same as Telegram path)
    UC->>DB: INSERT
    UC-->>H: TaskDTO
    H-->>Client: 201 JSON
```

> Phase 3 only. MVP has no REST endpoints.
