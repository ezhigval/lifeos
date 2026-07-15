LifeOS desktop package
======================

Содержимое
----------
  Start.command / Start.bat   — запуск (migrate + serve + логи)
  Stop.command  / Stop.bat    — остановка
  Logs.command  / Logs.bat    — хвост логов
  Settings.command / .bat     — редактор settings.env
  settings.env.example        — шаблон настроек
  bin/lifeos[.exe]            — сервер
  web/miniapp/dist            — Mini App
  migrations/                 — схема БД
  logs/                       — lifeos.log

Быстрый старт (macOS)
---------------------
  1. Нужен PostgreSQL (локально или Docker).
  2. Двойной клик Settings.command — токен бота, JWT, DATABASE_URL.
  3. Двойной клик Start.command — окно Terminal с логами.
  4. Mini App: укажи LIFEOS_MINIAPP_URL (HTTPS, …/app/) или работай в polling без Mini App.

Быстрый старт (Windows)
-----------------------
  То же через Start.bat / Settings.bat / Logs.bat.

Postgres одной командой
-----------------------
  docker run --name lifeos-pg -e POSTGRES_USER=lifeos -e POSTGRES_PASSWORD=lifeos \
    -e POSTGRES_DB=lifeos -p 5432:5432 -d postgres:16

Остановка / логи
----------------
  Stop.* убивает процесс из run/lifeos.pid
  Logs.* = tail -f logs/lifeos.log
