# LifeOS Mini App — UX/UI Plan

**Version:** 0.2  
**Status:** Draft  
**Scope:** Telegram Mini App (`web/miniapp/`)  
**Related:** [ROADMAP.md](../roadmap/ROADMAP.md) · [ARCHITECTURE.md](../architecture/ARCHITECTURE.md) · [BACKLOG.md](../roadmap/BACKLOG.md)  
**Backend этап (отдельный промпт):** [BACKEND_PROMPT.md](BACKEND_PROMPT.md) — auth, overview API, HTTPS / static IP / ngrok. Фронт и бэкенд вести раздельно.  
**Frontend lead (self-prompt):** [FRONTEND_LEAD_PROMPT.md](FRONTEND_LEAD_PROMPT.md) — зона клиента, анти-бэкенд boundary.

---

## 1. Зачем Mini App

Бот — быстрый capture (NL + reply keyboard).  
Mini App — **обзор, структура и визуальные действия**, где чат неудобен:

| Задача | Лучше в боте | Лучше в Mini App |
|--------|--------------|------------------|
| «Потратил 400 на еду» | ✅ NL | ок как secondary |
| Закрыть 5 задач сегодня | ок | ✅ чеклист |
| Дерево сферы → проект → задачи | тяжело | ✅ |
| Кольцо бюджета / категории | плохо | ✅ |
| Настройки quiet hours | ок текстом | ✅ формы |
| Reviews / аналитика | текст | ✅ карточки |

**Правило разделения:** capture → бот; browse / complete / visual feedback → Mini App.  
Полный feature parity с ботом — **цель Phase 4**, не MVP Mini App.

---

## 2. Принципы UX

1. **Telegram-native** — `BackButton`, цвета `--tg-theme-*`, safe-area, haptic на commit-действиях. Не изобретать «веб-приложение внутри».
2. **Один главный job на экран** — Home = сегодня; finance sheet = одна транзакция; project = задачи проекта.
3. **Большие тач-цели** (≥44px), минимум шагов до «done».
4. **Offline-tolerant UX** — понятный error/retry; токен/сессия без сюрпризов.
5. **Бот не дублируется экранами ради галочки** — нет смысла тащить все NL-интенты в формы, пока нет явной боли.
6. **Визуальный якорь** — finance ring + сегодняшние задачи. Остальное — спокойный список/дерево без карточного шума.

---

## 3. Текущее состояние (baseline)

| Есть | Нет / дыры |
|------|------------|
| 2 таба: Главная, Сферы | BackButton / MainButton |
| Upcoming tasks + complete | Settings / Habits / Calendar / Health / Career / Notes |
| Finance card + sheet (income/expense) | `POST /auth/telegram-webapp`, `GET /finance/overview` в backend |
| Sphere → project → tasks | CRUD сфер; edit task (due/priority); triage |
| TG theme CSS vars + haptic | Sheet motion, page transitions |
| Dev auth через API key | Logout / token refresh UX |

Stack: React 19 · Vite · Tailwind 4 · TanStack Query · `@twa-dev/sdk` · HashRouter (`base: /app/`).

---

## 4. Information Architecture

### 4.1 Навигация v1 (после полировки)

```
┌─────────────────────────────────────┐
│  [TG header / BackButton when nest] │
│                                     │
│           <screen content>          │
│                                     │
├─────────┬─────────┬─────────────────┤
│ Главная │ Сферы   │ Ещё             │
└─────────┴─────────┴─────────────────┘
```

| Таб | Job | Содержимое |
|-----|-----|------------|
| **Главная** | «Что важно сегодня» | Greeting, upcoming tasks, priorities snippet, finance widget |
| **Сферы** | Структура жизни | Tree → sphere detail → project detail |
| **Ещё** | Всё остальное без перегруза nav | Habits, Calendar, Analytics, Settings, Health/Career/Notes (поэтапно) |

Finance **не** отдельный таб в v1 — виджет на Home + deep-link из «Ещё» позже при росте.

### 4.2 Карта экранов

```
AuthGate
├─ Home
│  ├─ Task complete (inline)
│  ├─ → Spheres (все задачи)
│  └─ FinanceSheet (income | expense)
├─ Spheres
│  ├─ SphereDetail
│  │  ├─ CreateProject sheet
│  │  └─ ArchiveProject confirm
│  └─ ProjectDetail
│     ├─ CreateTask sheet
│     └─ CompleteTask inline
└─ More
   ├─ HabitsToday
   ├─ CalendarToday
   ├─ Analytics
   └─ Settings
      ├─ Review times
      ├─ Quiet hours
      └─ Spheres CRUD
```

Nested routes: Telegram **BackButton** = `navigate(-1)` / up to parent. Bottom nav скрывает Back на корневых табах.

---

## 5. Design system (Mini App)

### 5.1 Theme

Источник правды — Telegram WebApp theme:

| Token | Назначение |
|-------|------------|
| `--tg-theme-bg-color` | фон экрана |
| `--tg-theme-secondary-bg-color` | поверхности списков / sticky bars |
| `--tg-theme-text-color` / `hint-color` | текст / secondary |
| `--tg-theme-button-color` / `button-text-color` | primary CTA, active nav |
| `--tg-theme-link-color` | ссылки |

Бренд-акценты **поверх** темы (не ломая light/dark клиента):

| Token | Значение | Зачем |
|-------|----------|-------|
| `--lifeos-income` | green (сейчас `#22c55e`) | доход, success, complete |
| `--lifeos-expense` | red/coral | расход |
| `--lifeos-priority-*` | urgent/high/medium/low | точки приоритета |
| category palette | `lib/categories.ts` | legend / chips |

Fallback slate (`#0f172a` / `#1e293b`) — только вне Telegram / пока тема не пришла.

### 5.2 Typography

Оставить **system UI** в Mini App (читается нативно в TG). Не тащить marketing-шрифты.  
Иерархия:

- Screen title: 20–22 / semibold  
- Section: 13 / medium · hint color · uppercase optional sparingly  
- Body: 15–16 / regular  
- Meta: 12–13 / hint  
- Money: `tabular-nums`, semibold

### 5.3 Components

| Примитив | Статус | Правило |
|----------|--------|---------|
| `Button` | есть | primary / secondary / ghost / danger / income / expense |
| `Sheet` | есть | добавить enter/exit + swipe-to-dismiss + focus amount field |
| `Skeleton` | есть | на все async-блоки |
| `ListRow` | нет → добавить | title + meta + trailing action; без лишних «карточек» |
| `EmptyState` | частично | один паттерн: короткий текст + одно действие |
| `Confirm` | нет | `showPopup` / `showConfirm` TG API или лёгкий sheet |
| `Header` | есть | wired Back через TG; settings → More/Settings |

Карточки — **только** где есть отдельный интерактивный модуль (Finance ring). Списки задач/проектов — rows, не card grid.

### 5.4 Motion (минимум, осознанно)

1. Sheet: slide-up + fade overlay (~200ms)  
2. Task complete: checkbox + haptic success + collapse/strike  
3. Nav active: color transition  

Без page-transition theatre и без лишней анимации дерева.

### 5.5 Haptics

| Событие | Feedback |
|---------|----------|
| Toggle expand / period | light |
| Complete task / save tx / create | success |
| Destructive (archive) | warning/error после confirm |
| Validation fail | error |

---

## 6. Ключевые UX-flows

### 6.1 Home — Today

1. Greeting (`tgUser.first_name`) + дата  
2. Блок «Задачи» (incomplete today + priorities, cap ~7)  
3. Finance widget (период, ring, +/−)  
4. Пустые состояния с одним CTA («Добавить в боте» / «Создать в проекте»)

**Не пихать** на Home: расписание, health stats, career, notes.

### 6.2 Complete task

Tap checkbox → optimistic UI → API → haptic success.  
Ошибка → rollback + toast/banner.

### 6.3 Finance capture

1. + / − → Sheet  
2. Keypad amount → category/description → Save  
3. Invalidate overview/cash-flow → update ring  

Нужен стабильный `GET /finance/overview` (или явный контракт на cash-flow + categories), иначе legend пустой.

### 6.4 Sphere → Project → Task

Tree на корне; detail screens для CRUD-действий.  
Create project/task — sheet, не отдельный route.  
Archive — confirm.

### 6.5 Auth

Production: `initData` → `POST /api/v1/auth/telegram-webapp` → JWT.  
Dev: API key + telegram_id (как сейчас).  
Показать явный error + «открой из бота», без бесконечного спиннера.

---

## 7. Backend gaps (блокеры UX)

| Gap | Зачем UI |
|-----|----------|
| `POST /api/v1/auth/telegram-webapp` | prod-вход без Vite secrets |
| `GET /api/v1/finance/overview` (или расширить cash-flow) | ring + categories по периодам |
| OpenAPI sync для spheres | контракт клиента |
| JWT expiry handling | тихий re-auth через initData |
| Публичный HTTPS (static IP / port-forward / ngrok) | Telegram WebApp URL |

Полный промпт и чеклист деплоя: **[BACKEND_PROMPT.md](BACKEND_PROMPT.md)**.  
Без auth endpoint Mini App в Telegram нестабилен — **P0 на бэкенд-этапе**; фронт может полироваться параллельно на dev-auth.

---

## 8. Roadmap внедрения

### Phase A — Foundation (сделать из scaffold продукт)

**Цель:** надёжный вход, нативная навигация, стабильный Home.

| ID | Work | Priority |
|----|------|----------|
| MA-A1 | Backend: telegram-webapp auth | P0 |
| MA-A2 | Telegram `BackButton` на nested routes | P0 |
| MA-A3 | Таб «Ещё» (заглушки + Settings stub) | P1 |
| MA-A4 | Finance overview API + wire legend/periods | P0 |
| MA-A5 | Sheet motion + swipe dismiss | P1 |
| MA-A6 | Error/retry banners for queries | P1 |
| MA-A7 | Убрать Vite leftovers (react/vite svg noise) | P2 |
| MA-A8 | Fix nested button a11y в ProjectCard | P1 |

**Exit:** открытие из бота → Home с задачами и finance без пустых нулей на текущем месяце.

### Phase B — Core loops

**Цель:** закрыть главные daily jobs без бота.

| ID | Work | Priority |
|----|------|----------|
| MA-B1 | CreateTask sheet: title + optional priority/due | P1 |
| MA-B2 | Task complete из tree preview (не только bullets) | P1 |
| MA-B3 | Priorities с реальным taskId + complete | P1 |
| MA-B4 | Habits today + track | P1 |
| MA-B5 | Calendar today + create event sheet | P2 |
| MA-B6 | Settings: review times, quiet hours | P2 |
| MA-B7 | Spheres CRUD в Settings | P2 |

**Exit:** утренний цикл (задачи + привычки + finance) целиком в Mini App.

### Phase C — Depth & parity

**Цель:** приближение к боту по обзору доменов.

| ID | Work | Priority |
|----|------|----------|
| MA-C1 | Analytics / reviews screens | P2 |
| MA-C2 | Notes list/search | P2 |
| MA-C3 | Health (weight/steps/sleep) | P3 |
| MA-C4 | Career contacts/skills | P3 |
| MA-C5 | Reminders list | P3 |
| MA-C6 | Triage / day availability UI | P3 |
| MA-C7 | Debts UI (finance) | P2 |

**Exit:** Roadmap «Mini App: full feature parity with bot» можно закрывать по чеклисту доменов.

### Phase D — Polish

- Light/dark sync edge cases  
- MainButton для primary sheet CTA (опционально)  
- Performance: virtualize long task lists  
- E2E smoke: auth → complete task → add expense  
- Accessibility: focus order in sheets, reduced motion  

---

## 9. Acceptance criteria (Definition of Done по фазам)

### Phase A
- [ ] Auth через Telegram initData без `.env` хаков  
- [ ] Back из project/sphere = TG BackButton  
- [ ] Finance текущего периода показывает суммы и категории  
- [ ] Нет dead Settings gear без экрана  

### Phase B
- [ ] Создать задачу в проекте с приоритетом  
- [ ] Закрыть задачу на Home и в Project  
- [ ] Track habit за ≤2 тапа  
- [ ] Изменить quiet hours и увидеть sync с ботом  

### Phase C
- [ ] Каждый домен бота либо имеет Mini App screen, либо явно отклонён ADR/note  

---

## 10. Anti-goals

- Не делать dashboard «всего сразу» на Home  
- Не копировать reply-menu бота 1:1 в bottom nav (7 табов — смерть)  
- Не строить design system ради DS (никакого full shadcn dump без нужды)  
- Не блокировать Phase A ожиданием полного parity  

---

## 11. Связь с roadmap

| Doc | Update |
|-----|--------|
| ROADMAP Phase 4 | выполнять по Phases A→C этого плана |
| MILESTONES Mini App | scaffold → A done → B done |
| BACKLOG Icebox «Telegram Mini App UI» | вынести в активный backlog как MA-* |

Рекомендуемый следующий шаг после принятия плана: **MA-A1 + MA-A4** (auth + finance overview), параллельно MA-A2 (BackButton).
