# Changelog

All notable changes to this project are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Desktop window clients for macOS and Windows (beta, tags `desktop-v*`).
- Conversational agent with LLM tool calling (`LIFEOS_LLM_ENABLED`), STT for voice messages, Vision for photos.
- Agent memory (`internal/memory`) and opt-in anonymized learning events (`internal/learning`) — migration 00030.
- Telegram login codes for Mini App web auth — migration 00037.
- Task↔sphere links and recurring rules — migrations 00035–00036; home widget settings — 00034.
- Architecture roadmap: [docs/architecture/ROADMAP.md](docs/architecture/ROADMAP.md).

## [1.0.0-alpha.1] — 2026-10-05

Первый публичный альфа-срез. Идентификатор десктоп-пакета: `LifeOS_alpha_1.0.0`.

- Один процесс: Telegram-бот, REST `/api/v1`, Mini App.
- Задачи, проекты, сферы, планирование, финансы, привычки, календарь, заметки, здоровье, карьера, напоминания.
- Локальный запуск через Docker Compose и десктоп-пакет (`make package`).
- Постоянный HTTPS для Telegram: свой домен и named Cloudflare Tunnel. Инструкция: [docs/deploy/vm.md](docs/deploy/vm.md).
- Автодеплой на ВМ по таймеру, секреты только в `/opt/lifeos/.env` и `/opt/lifeos/secrets`.

[Unreleased]: https://github.com/ezhigval/lifeos/compare/LifeOS_alpha_1.0.0...HEAD
[1.0.0-alpha.1]: https://github.com/ezhigval/lifeos/releases/tag/LifeOS_alpha_1.0.0
