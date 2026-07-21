# Dogfood — 14 дней (Valentin / Mac)

**Цель:** закрыть gate **G2 → G3** реальной ежедневной нагрузкой: Telegram-бот + Mini App.  
**Стек:** локальный LifeOS на Mac + tunnel + Telegram.  
**Баги:** сразу в `docs/agents/inbox/<agent>/` по шаблону ниже (приоритет **P0** → BLOCKED/OPEN после triage Architect).

Связано: [ROADMAP.md](ROADMAP.md) · [LOCAL_DEV.md](../miniapp/LOCAL_DEV.md)

---

## Перед стартом (ops)

- [ ] `git pull` на рабочей ветке / `main` после merge
- [ ] `.env` заполнен (DB, bot token, `LIFEOS_JWT_SECRET`, `LIFEOS_MINIAPP_URL` после tunnel)
- [ ] `make stack-up` — Postgres + API + bot + cloudflared; `/health` ок
- [ ] В Telegram: **`/start`** сразу после поднятия tunnel (свежий URL на reply keyboard)
- [ ] `make verify-webapp-auth` — initData → JWT зелёный
- [ ] Mini App открывается с кнопки бота (не из старого сообщения в истории)

> После рестарта `cloudflared`: снова `make stack-up` → `/start` → `verify-webapp-auth`. Старый hostname = Error 1033.

---

## Как заводить баг

Файл: `docs/agents/inbox/<backend|frontend|telegram>/BUG-DOGFOOD-NNN-<slug>.md`  
либо копить в существующий `TASK-008-*` после triage.

```markdown
# BUG-DOGFOOD-NNN: краткий заголовок

- **Agent:** backend | frontend | telegram
- **Status:** DRAFT
- **Priority:** P0 | P1 | P2
- **Stage:** dogfood
- **Day:** N (1–14)
- **Date:** YYYY-MM-DD

## Repro
1. …
2. …

## Expected
…

## Actual
…

## Zone
Bot NL | reply KB | Mini App screen | auth/tunnel | review/reminder

## Notes / screenshot
(опционально путь к скрину или текст ошибки)
```

Приоритет: **P0** — блокер дня / потеря данных / auth; **P1** — ломает сценарий, есть обход; **P2** — UX/косметика.

---

## День 1 — Bootstrap + вход

**Ops**

- [ ] `make stack-up`
- [ ] `/start` в боте
- [ ] `make verify-webapp-auth`

**Бот**

- [ ] `/start` → приветствие / dashboard
- [ ] Reply keyboard на месте
- [ ] Кнопка Mini App (web_app) открывает актуальный URL

**Mini App**

- [ ] Home грузится, auth без вечного спиннера
- [ ] Bottom nav: Home / Сферы / … кликабельны

---

## День 2 — Задачи (NL + Home)

**Бот**

- [ ] NL: создать задачу («Купить молоко завтра»)
- [ ] Задача видна в списке / dashboard
- [ ] Отметить выполненной (кнопка или NL)

**Mini App**

- [ ] Home: задача из бота видна
- [ ] Создать / открыть / закрыть задачу в UI
- [ ] Task detail открывается без пустого экрана

---

## День 3 — Привычки

**Бот**

- [ ] NL: создать привычку
- [ ] Затрекать привычку за сегодня

**Mini App**

- [ ] Экран Habits: список + streak
- [ ] Создать / отметить / не сломать empty state

---

## День 4 — Финансы

**Бот**

- [ ] NL расход («Потратил 400 на еду»)
- [ ] Ответ бота понятный, сумма верная

**Mini App**

- [ ] Finance ring на Home / карточке
- [ ] Period picker (месяц)
- [ ] Добавить транзакцию из sheet
- [ ] Legend / категории не «мертвые»

---

## День 5 — Сферы и проекты

**Mini App**

- [ ] Spheres: список сфер
- [ ] Зайти в сферу → проекты → задачи
- [ ] Создать задачу внутри проекта

**Бот** (по возможности)

- [ ] NL / кнопки, связанные со сферой (если есть в keyboard) — не падает

---

## День 6 — Напоминания

**Бот**

- [ ] NL: «напомни вечером …» / «утром …»
- [ ] Подтверждение с **локальным** временем
- [ ] Напоминание реально приходит
- [ ] Отмена напоминания

**Mini App**

- [ ] Reminders: список
- [ ] Создать напоминание из UI
- [ ] Отменить / удалить

---

## День 7 — Reviews (утро / вечер) + mid-week

**Бот / scheduler**

- [ ] Утренний обзор пришёл (или триггер / команда, если настроено)
- [ ] Вечерний обзор пришёл
- [ ] HTML в сообщении читаемый (без сырого `<b>` / обрезки мусора)
- [ ] Невыполненные задачи корректно фигурируют в вечернем обзоре

**Ops / регресс**

- [ ] После сна Mac: stack жив или `make stack-up` + `/start` + `verify-webapp-auth`

**Сводка mid-week**

- [ ] Список P0/P1 за дни 1–7 → inbox (DRAFT)

---

## День 8 — Календарь

**Mini App**

- [ ] Calendar: события на сегодня / неделе
- [ ] Создать событие
- [ ] Правка / удаление без crash

**Бот** (если есть NL/кнопки календаря)

- [ ] Создать событие текстом — корректный ack

---

## День 9 — Настройки

**Mini App · Settings**

- [ ] Обзоры (morning/evening/weekly/monthly) читаются/меняются
- [ ] Quiet hours
- [ ] Язык / timezone (если доступно)
- [ ] CRUD сфер из Settings

**Проверка**

- [ ] Изменение quiet hours не ломает следующие напоминания/reviews

---

## День 10 — More: Notes / Health / Debts / Analytics

- [ ] Notes: базовый flow
- [ ] Health: вес / шаги / сон
- [ ] Debts: платежи
- [ ] Analytics: сводка
- ~~Career~~ — UI снят (backend оставлен в icebox)
**Mini App**

- [ ] Notes: создать / открыть / удалить
- [ ] Health: базовый flow
- [ ] Career: базовый flow
- [ ] Debts: создать / отметить
- [ ] Analytics: экран грузится, пустые данные не «error»

---

## День 11 — Сквозные сценарии

- [ ] Задача создана в боте → видна и закрывается в Mini App
- [ ] Расход в боте → отражается в Finance ring
- [ ] Привычка в Mini App → трек в боте (или наоборот)
- [ ] Reminder: создать в Mini App → отменить в боте (или наоборот)
- [ ] Нет двойных записей / «призраков»

---

## День 12 — Tunnel / auth resilience

- [ ] Остановить origin / перезапустить tunnel (`make stack-up`)
- [ ] Старая кнопка Mini App из истории → ожидаемо 1033 / мер
- [ ] **`/start`** → новая кнопка работает
- [ ] `make verify-webapp-auth` снова зелёный
- [ ] После re-auth Home и Finance не требуют ручного clear storage (или задокументировать workaround)

---

## День 13 — Weekly / monthly + свободный день

**Бот**

- [ ] Weekly review (команда / авто) — текст ок
- [ ] Monthly review (если доступен) — текст ок

**Свободный dogfood**

- [ ] Обычный день «как живёшь»: 3+ захвата NL + 1 сессия Mini App ≥10 мин
- [ ] Записать все P0/P1 в inbox

---

## День 14 — Gate review

- [ ] Пройтись по неотмеченным чекбоксам дней 1–13
- [ ] Все **P0** разложены по `inbox/<agent>/` (или закрыты)
- [ ] Краткий итог в `docs/agents/reports/` или комментарий Architect: go / no-go для G2→G3
- [ ] При go: обновить [ROADMAP.md](ROADMAP.md) (галочка dogfood gate) + [MILESTONES.md](MILESTONES.md)

---

## Быстрый ops-cheat sheet

```bash
make stack-up              # полный стек + свежий tunnel
# Telegram → /start
make verify-webapp-auth    # auth Mini App
curl -sS http://127.0.0.1:8080/health
```

| Симптом | Действие |
|---------|----------|
| 1033 / Mini App не открывается | `stack-up` → `/start` → не жать старую кнопку |
| Auth крутится / 401 | `verify-webapp-auth`; проверить bot token / JWT secret |
| Бот молчит | health + логи bot; polling не в двух процессорах |
| Reply KB без Mini App | `/start`; `LIFEOS_MINIAPP_URL` в `.env` |
