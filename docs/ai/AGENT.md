# Диалоговый агент LifeOS

Агент принимает свободный текст в Telegram, может **уточнять**, **болтать по делу**, **вызывать tools** (use cases домена) и отвечать по факту выполнения.

## Поток

```
пользователь
  → (FSM drafts: задача/проект/сфера — без изменений)
  → ConversationalAgent.Handle
       ├ system prompt + personal memories
       ├ JSON action: ask | reply | tool
       ├ tool → domain use case (task/finance/habit/reminder/memory)
       └ до 4 tool-раундов, затем финальный reply/ask
  → ответ в Telegram
  → (opt-in) anon_learning_events без PII
```

Intent-resolver (rulebased→LLM classify) остаётся fallback, если агент выключен или упал.

## Tools

| Tool | Действие |
|------|----------|
| `task.create` / `list_today` / `complete` / `cancel` / `reschedule` / `reschedule_all` | задачи |
| `finance.expense` / `income` / `list_debts` / `create_debt` / `pay_debt` / `cash_flow` / `list_plan` / `create_planned` | финансы |
| `reminder.create` / `cancel` | напоминания |
| `habit.create` / `track` / `list` | привычки |
| `note.create` / `list` / `search` | заметки |
| `calendar.create` / `list_today` | календарь |
| `project.create` / `list` | проекты |
| `sphere.list` / `create` | сферы |
| `plan.set_availability` / `triage` | планирование дня |
| `health.record_*` / `latest_*` (weight/steps/sleep) | здоровье |
| `career.contact_*` / `skill_*` | карьера |
| `query.priorities` / `analytics.summary` | обзоры |
| `memory.save` / `recall` | личная память |

Tools дергают **use cases**, не сырой HTTP — тот же путь, что Mini App/API.

## Персональная память

Таблица `user_memories` (CASCADE с user). Виды: `preference`, `fact`, `alias`, `pattern`.

- Хранится **только** в тенанте пользователя
- Флаг `user_settings.memory_enabled` (default true)
- При `/delete` пользователя память удаляется вместе с аккаунтом

Не уходит в анонимный learning.

## Анонимное обучение

Таблица `anon_learning_events`:

- `anon_subject` = HMAC-SHA256(user_id, `LIFEOS_LEARNING_SALT`) — без обратимого id
- Пишет только при `learning_opt_in=true`
- Meta: число tools, имена tools, waiting — **без** текста сообщений, сумм, названий задач

Используется для анализа: где агент часто ask'ает, какие tools фейлятся — улучшение промпта/few-shot.

## Env

```env
LIFEOS_LLM_ENABLED=true
LIFEOS_LLM_AGENT_ENABLED=true
LIFEOS_LLM_PROVIDER=ollama   # или openai (Groq)
LIFEOS_OLLAMA_URL=http://127.0.0.1:11434
LIFEOS_OLLAMA_MODEL=llama3.2
LIFEOS_LEARNING_SALT=change-me-in-prod
```

Opt-in обучения и memory flags — колонки в `user_settings` (API для переключения — следующий шаг).

## Безопасность

1. Модель **не** пишет в БД напрямую — только через tools с валидацией
2. Персональные факты изолированы по `user_id`
3. Learning без сырого текста
4. Соль learning обязана быть уникальной на инстанс
5. «отмена» сбрасывает `await_agent_turn`

## Дальше

- `POST /api/v1/assistant/chat` для Mini App
- note.delete / project.archive / settings.* tools
- nightly job: агрегаты learning → кандидаты в few-shot
- шифрование `user_memories.value` at rest (envelope key per user)
