# TASK-011 п.7-8 — N:M задачи↔сферы + домен-правила (готово)

Дата: 2026-10-05. Статус: п.7 и п.8 реализованы, протестированы, запушены в GitHub.

## п.7 — N:M задачи↔сферы (завершено)
- Миграции `00035_task_spheres_and_rules.sql` (task_spheres, sphere_domain_links) +
  `00036` (репарация FK на life_spheres).
- Domain Task.SphereIDs; sqlc task_spheres (Insert/Delete/List) + SpheresExist.
- Repo tasks: SetSpheres + заполнение SphereIDs на всех чтениях; spheres repo AllExist.
- App: CreateTask/EditTask принимают SphereIDs (валидация через sphere checker).
- HTTP payload принимает sphere_ids; OpenAPI схема Task обновлена.
- FE: многовыбор сфер в TaskDetail, бейджи сфер в TaskCard, UpcomingTasks подтягивает ['spheres'].
- Агентский tool task.create: резолвинг sphere/project по имени.

## п.8 — домен-правила (завершено)
- `internal/tasks/app/domain_rules.go`: DomainRules.Apply вызывается best-effort
  после create/edit задачи:
  - «Деньги» + заголовок ~зарплата → ежемесячный план дохода «Зарплата» (idempotent по ListPlanned);
  - «Здоровье» → daily habit-трекер «Спорт / здоровье (авто)» (idempotent по FindByName);
  - сумма плана — placeholder 1 копейка (домен finance требует >0), пользователь правит в плане.
- Wire в runtime.go: NewDomainRules(sphereRepo, financeRepo, habitRepo) + WithDomainRules.
- Тесты domain_rules_test.go (4 кейса, включая идемпотентность) — зелёные.

## Не сделано (следующие задачи)
- Карьера→воркспейс мост через sphere_domain_links (registry link_type 'workspace').
- Каскадный picker ws→сфера→проект→задача в FE.
- Фильтры главной/календаря по сфере.
