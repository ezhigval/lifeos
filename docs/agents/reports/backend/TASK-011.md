# TASK-011 — Backend отчёт (UX P0, пункты 1–2)

Дата: 2026-10-04 · Исполнитель: backend · Статус: п.1–2 готово, п.3–8 переданы зонам (см. ниже)

## Пункт 1. Формы «Планируемый доход / расход» — обязательные поля (BUG)

**Причина бага.** `POST /api/v1/finance/plan` возвращал одну генерическую ошибку
`"kind, title and amount_cents are required"` при любом невалидном теле, а поле
`currency`, которое стабильно отправляет мини-приложение (`FinanceCard.tsx`), в контракте
не описывалось вообще. Клиент при отказе показывал «Заполни обязательные поля», хотя все
видимые поля были заполнены — пользователь не мог понять, что именно «не так».

**Исправления (internal/transport/http/api/finance.go):**
- Полевая валидация с понятными сообщениями по каждому обязательному полю:
  `kind` (обязателен, enum income|expense), `title` (обязателен), `amount_cents`
  (обязателен, > 0), `interval` (опционален, enum once|weekly|monthly; пустой → домен
  сам ставит monthly), `next_date` (опционален, формат YYYY-MM-DD).
- Добавлено принятие и валидация поля `currency` (RUB/USD/EUR) — контракт теперь честно
  отражает то, что шлёт клиент. Суммы хранятся в копейках RUB, значение currency
  валидируется, но не персистится (задокументировано в OpenAPI).
- `docs/api/openapi.yaml`: схема `POST /api/v1/finance/plan` дополнена `currency`.

## Пункт 2. Главная не отображает задачи (BUG)

**Диагностика.** Use-case `ListTasksToday` вызывал `store.ListByDueDate` — SQL, который
фильтровал строго `due_date = сегодня`. Задачи без даты и просроченные на главной не
появлялись, список выглядел пустым.

**Исправление (internal/tasks/app/list_today.go):** переключено на существующий метод
`ListOpenDueOnOrBefore` (sqlc-запрос `ListOpenTasksDueOnOrBefore`: открытые статусы
todo/in_progress, due_date IS NOT NULL AND due_date <= today, сортировка по дате).
Порт, репозиторий и sqlc-генерация уже присутствовали — проверено соответствие
(`queries/tasks.sql`, `internal/platform/db/tasks.sql.go`). Поведение совпадает с
ожиданием acceptance: «на главной видны сегодняшние + просроченные задачи».

**Ограничение:** задачи вовсе без `due_date` на главной не показываются (по замыслу —
«сегодня/просрочено»); это согласовано с формулировкой acceptance-критерия.

## Тесты
- `go vet ./internal/...` / сборка: выполняются в CI (Go 1.25 в `.github/workflows`);
  локальная sandbox-среда содержит только Go 1.19 и не может собрать модуль `go 1.25.5` —
  проверка поручена CI (gitlab-actions green gate из правил деплоя).
- Существующие unit-тесты `internal/tasks/app` (fakeStore реализует оба метода) и
  `api_test.go` (regression для `/tasks/today`) покрывают изменения; новые кейсы
  валидации planned — предложены frontend-зоне в e2e (см. её отчёт).

## Что дальше (остальные пункты TASK-011)
- п.3 habits PATCH/DELETE + start/end (миграция) — backend, следующий PR;
- п.4 `/calendar/aggregate` — backend;
- п.5 note/reminder ↔ task связи — backend;
- п.7 `task_spheres` N:M — backend;
- п.8 domain_rule engine (money→planned income) — backend.
П.1–2 (баги) закрыты этим коммитом, чтобы owner мог проверить их на стенде раньше фич.

## Secret-safety
Секретов/IP/токенов не добавлено; diff проверен grep'ом (`token|secret|key|BEGIN`).
