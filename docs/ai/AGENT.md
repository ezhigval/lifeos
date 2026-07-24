# Диалоговый агент LifeOS

Агент принимает свободный текст **и расшифрованный голос/кружочки** в Telegram, может **уточнять**, **болтать**, **вызывать tools** (use cases домена) и отвечать по факту выполнения.

## Поток (Telegram)

```
текст | voice | video_note | audio | photo | image-document
  → resolveIncomingText (STT / Vision)
    → (FSM drafts: задача/проект/сфера)
    → rulebased known? → dispatchIntent
    → иначе ConversationalAgent.Handle
         ├ system prompt + personal memories
         ├ JSON action: ask | reply | tool
         ├ tool → domain use case
         └ до 4 tool-раундов
    → ответ в Telegram
```

Intent-resolver (rulebased→LLM classify) — fallback, если агент выключен или упал.

## Голос / медиа

| Вход | Поведение |
|------|-----------|
| `voice`, `audio`, `video_note` | download → Whisper STT → тот же agent/intent path; UX «Слушаю…» |
| `photo` / image-`document` без caption | download → Vision → короткая RU-команда → agent path; UX «Смотрю фото…» |
| `photo` / `document` с caption | берётся caption (vision не вызывается) |
| не-image `document` без caption | просьба добавить подпись |
| стикеры | игнор |

Env:
- STT: `LIFEOS_STT_ENABLED=true`, ключ `LIFEOS_STT_API_KEY` (или `LIFEOS_LLM_API_KEY`), модель `whisper-large-v3-turbo` (Groq).
- Vision: `LIFEOS_VISION_ENABLED=true`, ключ `LIFEOS_VISION_API_KEY` (или LLM), модель `meta-llama/llama-4-scout-17b-16e-instruct`.

См. [docs/ops/LLM.md](../ops/LLM.md).

## Tools

| Tool | Действие |
|------|----------|
| `task.*` | create / list_today / complete / cancel / reschedule / reschedule_all |
| `finance.*` | income / expense / debts / cash_flow / list_plan / create_planned |
| `reminder.*` / `habit.*` / `note.*` | напоминания, привычки, заметки (+ delete) |
| `calendar.*` / `project.*` / `sphere.*` | календарь, проекты (+ archive/tasks/progress), сферы |
| `plan.*` | set_availability / triage |
| `health.*` | weight / steps / sleep |
| `career.*` | contacts / skills |
| `settings.*` | morning_review / evening_review / quiet_hours |
| `query.priorities` / `analytics.summary` | обзоры |
| `memory.*` | личная память |

Tools дергают **use cases**, не сырой HTTP.

## Общение

Болтовня / «что умеешь?» → `type=reply` без tools. Действия с побочными эффектами — только через tools. STT-текст может быть неточным — агент интерпретирует мягко и уточняет (`ask`).

## Дальше

- UI-чат Mini App (отложен)
- nightly learning → few-shot
- шифрование `user_memories.value` at rest
