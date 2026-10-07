# Changelog

## 1.0.0-alpha.1 — 2026-10-05

Первый публичный альфа-срез. Идентификатор десктоп-пакета: `LifeOS_alpha_1.0.0`.

- Один процесс: Telegram-бот, REST `/api/v1`, Mini App.
- Задачи, проекты, сферы, планирование, финансы, привычки, календарь, заметки, здоровье, карьера, напоминания.
- Локальный запуск через Docker Compose и десктоп-пакет (`make package`).
- Постоянный HTTPS для Telegram: свой домен и named Cloudflare Tunnel. Инструкция: [docs/deploy/vm.md](docs/deploy/vm.md).
- Автодеплой на ВМ по таймеру, секреты только в `/opt/lifeos/.env` и `/opt/lifeos/secrets`.
