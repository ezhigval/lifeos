# TASK-011 п.7 — N:M задачи↔сферы/проекты + домен-правила (прогресс)

Дата: 2026-10-04. Статус: миграция+domain готовы, далее app/repo/http/wire/OpenAPI/FE.

## Сделано
1. **Миграция** `migrations/00033_task_spheres_and_rules.sql`:
   - `task_spheres(task_id, sphere_id, PK)` — многие-ко-многим;
   - `sphere_domain_links(sphere_id, link_type, ref_id)` — привязка сферы к сущностям
     (`finance_category`, `habit`, `workspace`) — база для п.8 и мостов карьера→воркспейс;
   - индекс по task_spheres.sphere_id.
2. **Domain** `internal/tasks/domain/task.go`: поле `SphereIDs []string` в Task.

## План (продолжение этой же ветки)
- sqlc-запросы attach/detach/list + ручной db-ген; repo: SaveSpheres/ListByUser+spheres.
- app: EditTaskInput.SphereIDs, TaskDTO.SphereIDs, ListTasks* пробрасывают связи.
- HTTP: payload принимает sphere_ids; OpenAPI схема Task.
- FE: многовыбор сфер в TaskDetail; бейджи сфер в карточках; каскадный picker (ws→сфера→проект→задача).
- П.8 правила: registry domain-rules (money salary-task → planned income; health project→habit via sphere_domain_links; career project→workspace bridge), запуск при create/update task, идемпотентность через origin-ссылку.

## Деплой/push
Push в GitHub блокируется: репозиторий приватный, токена нет (см. .git/GITHUB_PUSH_SETUP.md).
SSH на ВМ недоступен из среды (порт 22 filtered). Коммиты локальные, секреты не попадают (скан чистый).
